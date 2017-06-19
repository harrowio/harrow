package operationRunner

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	redis "gopkg.in/redis.v2"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/bus/broadcast"
	"github.com/harrowio/harrow/cast"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/limits"
	"github.com/harrowio/harrow/stores"
)

const (
	SshTimeout = 10 * time.Second
	Interval   = 5 * time.Second
)

type acquisitionFun func(c *config.Config) (string, string, error)

const ProgramName = "operation-runner"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {
	c := config.GetConfig()
	connectTo := flag.String("connect", "ssh://root@virthost.harrow.io/1", "Connection string")
	flag.Parse()

	connectionInfo, err := url.Parse(*connectTo)
	if err != nil {
		panic(err)
	}
	lxd := LXDAcquisition{ConnectTo: connectionInfo, BaseImage: "harrow-baseimage"}
	lxd.log = log
	db, err := c.DB()
	if err != nil {
		log.Fatal().Msgf("Unable to open db: %s", err)
	}
	defer db.Close()

	acquisitionFuns := make(map[string]acquisitionFun)
	acquisitionFuns["lxd"] = lxd.MustTakeInstance

	machinedConfig := c.MachinedConfig()
	implName := machinedConfig.AcquisitionImpl
	mustTakeInstance, ok := acquisitionFuns[implName]
	if !ok {
		log.Fatal().Msgf("Unknown acquisition impl: %s", implName)
	}

	keyValueStore := stores.NewRedisKeyValueStore(redis.NewTCPClient(c.RedisConnOpts(0)))
	activityBus := activity.NewAMQPTransport(c.AmqpConnectionString(), "machined")
	defer activityBus.Close()

	containerId, host, err := mustTakeInstance(c)
	log.Info().Msgf("impl=%s instance=%v host=%v err=%s", implName, containerId, host, err)
	if err != nil {
		log.Fatal().Msgf("%s: error taking instance: %s", implName, err)
	}
	intTerm := make(chan os.Signal, 1)
	signal.Notify(intTerm, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	go func(c chan os.Signal) {
		signal := <-c
		log.Info().Msgf("Got signal %s, exiting", signal)
		terminateInstance(implName, containerId, host)
		os.Exit(0)
	}(intTerm)

	source := broadcast.NewAMQPTransport(c.AmqpConnectionString(), "machined")
	source.ShareWork().OnlyTable("operations")
	defer source.Close()

	log.Info().Msg("listening for a message")
	msg, err := source.ConsumeOne(broadcast.Create)
	if err != nil {
		panic(err)
	}

	emitStatus(db, msg.UUID(), "vm.acquired", "VM acquired, uploading user script...")

	// handleMessage will Ack/Reject
	handleMessage(c, activityBus, db, source, containerId, host, *connectTo, msg, keyValueStore)
	log.Info().Msgf("done: %s", msg.UUID())
}

func watchForCancellations(c *config.Config, db *sqlx.DB, operationUuid string, stop <-chan bool) {
	source := broadcast.NewAutoDeletingAMQPTransport(c.AmqpConnectionString(), fmt.Sprintf("machined-cancellations-%s", operationUuid))
	source.OnlyTable("activities")
	work, err := source.Consume(broadcast.Create)
	if err != nil {
		log.Error().Msgf("Error watching for cancellations: %s", err)
		return
	}

	for {
		select {
		case msg := <-work:
			if msg.Table() == "activities" {
				handleActivity(db, msg, operationUuid)
			} else {
				if err := msg.Acknowledge(); err != nil {
					log.Warn().Msgf("msg.Acknowledge(): %s", err)
				}
			}
		case <-stop:
			log.Info().Msgf("cancellation-watch: stopped for %q", operationUuid)
			if err := source.Close(); err != nil {
				log.Error().Msgf("source.Close(): %s", err)
			}
			return
		}
	}
}

func handleActivity(db *sqlx.DB, message broadcast.Message, operationUuid string) {
	tx, err := db.Beginx()
	if err != nil {
		log.Error().Msgf("db.Beginx(): %s", err)
		message.RejectForever()
		return
	}
	defer tx.Rollback()

	id, _ := strconv.Atoi(message.UUID())

	activityStore := stores.NewDbActivityStore(tx)
	activity, err := activityStore.FindActivityById(id)
	if err != nil {
		log.Warn().Msgf("activity not found")
		if err := message.RejectForever(); err != nil {
			log.Warn().Msgf("message.RejectForever(): %s", err)
		}
		tx.Rollback()
		return
	}

	if err := message.Acknowledge(); err != nil {
		log.Warn().Msgf("message.Acknowledge(): %s", err)
	}

	if activity.Name != "operation.canceled-by-user" {
		tx.Rollback()
		return
	}

	log.Debug().Msgf("processing %s@%d", activity.Name, activity.Id)

	payload, ok := activity.Payload.(*activities.OperationCanceledByUserPayload)
	if !ok {
		log.Fatal().Msgf("Invalid payload for activity: expected %T, got %T", payload, activity.Payload)
	}

	if payload.Uuid != operationUuid {
		tx.Rollback()
		return
	}

	operationStore := stores.NewDbOperationStore(tx)
	if err := operationStore.MarkAsCanceled(operationUuid); err != nil {
		log.Error().Msgf("Failed to mark operation as canceled: %s", err)
	}

	statusLogEntry := cast.NewStatusLogEntry("user.canceled", "Operation canceled by user")
	operation, err := operationStore.FindByUuid(operationUuid)
	if err != nil {
		log.Error().Msgf("handleActivity: Operation %q not found: %s", operationUuid, err)
		return
	}
	operation.HandleEvent(statusLogEntry.Payload)
	if err := operationStore.MarkStatusLogs(operationUuid, operation.StatusLogs); err != nil {
		log.Error().Msgf("MarkStatusLogs: %q: %s", operationUuid, err)
		return
	}

	tx.Commit()
}

func handleMessage(c *config.Config, activityBus activity.Sink, db *sqlx.DB, source broadcast.Source, containerId string, host string, connect string, msg broadcast.Message, keyValueStore stores.KeyValueStore) {

	tx, err := db.Beginx()
	if err != nil {
		log.Error().Msgf("db.Beginx(): %s", err)
		msg.RejectForever()
		return
	}
	defer tx.Rollback()
	operationStore := stores.NewDbOperationStore(tx)
	operation, err := operationStore.FindByUuid(msg.UUID())
	if err != nil {
		log.Error().Msgf("operationstore.findbyuuid(uuid): %s", err)
		msg.RejectForever()
		return
	}
	theLimits, err := newLimits(tx, keyValueStore)
	if err != nil {
		log.Error().Msgf("newlimits(tx): %s", err)
		msg.RejectOnce()
		return
	}
	if exceeded, err := theLimits.Exceeded(operation); exceeded {
		log.Info().Msgf("Limits exceeded for operation %q", operation.Uuid)
		activityBus.Publish(activities.OperationCanceledDueToBilling(operation.Uuid))
		msg.RejectForever()
		return
	} else if err != nil {
		log.Error().Msgf("error calculating limits: %s", err)
		return
	}
	stopCancellationWatcher := make(chan bool)
	go watchForCancellations(c, db, operation.Uuid, stopCancellationWatcher)
	defer func() { stopCancellationWatcher <- true }()

	if operation.IsUserJob() {
		err := handleUserJob(c, db, containerId, host, connect, msg.UUID())
		if err != nil {
			log.Error().Msgf("handleuserjob:%s", err)
			msg.RejectForever()
		}
	} else if operation.Category() == "notifier" {
		notifierType := strings.Split(*operation.NotifierType, "_")[0]
		err := spawnNotifierJob(msg.UUID(), notifierType)
		if err != nil {
			log.Error().Msgf("spawnnotifierjob:%s", err)
			msg.RejectForever()
		}
	} else {
		log.Error().Msgf("no handler found for %#v", operation)
		msg.RejectForever()
	}
	if err := msg.Acknowledge(); err != nil {
		log.Error().Msgf("failed to acknowledge message: %s", err)
	}
}

func handleUserJob(c *config.Config, db *sqlx.DB, containerId string, host string, connect string, uuid string) error {

	log.Info().Msgf("spawn controller for operation(%s) on %s (%s)", uuid, containerId, connect)
	implName := c.MachinedConfig().AcquisitionImpl
	err := (error)(nil)

	switch implName {
	case "lxd":
		err = spawnUserJobLXD(c, uuid, containerId, connect)
	default:
		err = fmt.Errorf("unknown acquistion implementation: %q", implName)
	}

	if err != nil {
		return fmt.Errorf("unable to spawn controller for operation(%s): %s", uuid, err)
	}
	return nil
}

func emitStatus(db *sqlx.DB, uuid, entryType, subject string) {

	tx, err := db.Beginx()
	if err != nil {
		log.Error().Msgf("unable to begin tx: %s", err)
		return
	}
	defer tx.Rollback()
	operationStore := stores.NewDbOperationStore(tx)
	operation, err := operationStore.FindByUuid(uuid)
	if err != nil {
		log.Error().Msgf("unable to load operation: %s", err)
		return
	}
	entry := cast.NewStatusLogEntry(entryType, subject)
	operation.HandleEvent(entry.Payload)

	err = operationStore.MarkStatusLogs(uuid, operation.StatusLogs)
	if err != nil {
		log.Error().Msgf("unable to mark status logs: %s", err)
	}
	err = tx.Commit()
	if err != nil {
		log.Error().Msgf("unable to commit tx: %s", err)
	}
}

func newLimits(tx *sqlx.Tx, keyValueStore stores.KeyValueStore) (*limits.Service, error) {
	organizationStore := stores.NewDbOrganizationStore(tx)
	projectStore := stores.NewDbProjectStore(tx)
	billingPlanStore := stores.NewDbBillingPlanStore(tx, stores.NewBraintreeProxy())
	billingHistoryStore := stores.NewDbBillingHistoryStore(tx, keyValueStore)
	billingHistory, err := billingHistoryStore.Load()
	if err != nil {
		return nil, err
	}
	limitsStore := stores.NewDbLimitsStore(tx)

	return limits.NewService(organizationStore, projectStore, billingPlanStore, billingHistory, limitsStore), nil
}

func terminateInstance(implName string, containerId string, host string) {
	err := (error)(nil)
	switch implName {
	case "lxd":
		err = startProcess([]string{"vmex-lxd", "-container-id", containerId, "-host", host})
	}

	if err != nil {
		log.Error().Msgf("terminateInstance(%s, %s, %s): %s", implName, containerId, host, err)
	}
}
