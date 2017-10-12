package controllerLXD

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/bus/broadcast"
	"github.com/harrowio/harrow/cast"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/ssh"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/stores"
)

const ProgramName = "controller-lxd"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {

	var conf *config.Config = config.GetConfig()
	operationUuid := flag.String("operation-uuid", "", "The operation to run")
	containerId := flag.String("container-id", "", "The id of the container to run the operation in")
	connectTo := flag.String("connect", "", "The URL to connect with")
	entrypoint := flag.String("entrypoint", "", "The command to run")

	flag.Parse()

	go func() {
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			panic(err)
		}
		log.Info().Msgf("debug http server on port %d", listener.Addr().(*net.TCPAddr).Port)
		http.Serve(listener, nil)
	}()

	connectionInfo, err := url.Parse(*connectTo)
	if err != nil {
		panic(err)
	}

	host := connectionInfo.Host
	user := connectionInfo.User.Username()
	if user == "" {
		user = "root"
	}

	db, err := conf.DB()
	if err != nil {
		log.Fatal().Msgf("unable to open db: %s", err)
	}
	defer db.Close()

	consumerId := fmt.Sprintf("controller-pid-%d", os.Getpid())
	activityBus := activity.NewAMQPTransport(conf.AmqpConnectionString(), consumerId)
	activitySink := NewBusActivitySink(activityBus)
	activitySink.log = log

	defer activityBus.Close()
	if *operationUuid == "" || host == "" || *entrypoint == "" {
		fmt.Fprint(os.Stderr, "Usage:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	log.Info().Msgf("operationuuid=%q host=%q entrypoint=%q\n", *operationUuid, host, *entrypoint)
	deadline := time.After(config.InstanceDeadline)
	go func() {
		<-deadline
		deadlineReached(db, conf, *connectTo, *operationUuid)
	}()

	tx := mustBeginTx(db)
	defer tx.Rollback()

	store := stores.NewDbOperationStore(tx)
	log.Debug().Msg("looking up operation...")
	operation, err := store.FindByUuid(*operationUuid)
	if err != nil {
		log.Fatal().Msgf("operation not found: %s", err)
	}
	log.Debug().Msg("looking up operation... (done)")

	log.Debug().Msg("checking if operation cancelled...")
	if operation.Status() == "canceled" {
		log.Info().Msgf("yes, operation canceled at %s. exiting.", operation.CanceledAt)
		os.Exit(0)
	} else {
		log.Debug().Msg("not cancelled, continuing")
	}

	operation.StartedAt = func() *time.Time { now := time.Now(); return &now }()
	mustCommitTx(tx)

	log.Debug().Msg("broadcasting op started")
	activitySink.EmitActivity(activities.OperationStarted(operation))
	log.Debug().Msg("broadcasting op started (done)")

	log.Debug().Msg("getting ssh config")
	addr := fmt.Sprintf("%s", host)
	sshConf, err := conf.GetSshConfig()
	if err != nil {
		log.Fatal().Msgf("unable to get ssh config: %s", err)
	}
	sshConf.User = user
	log.Debug().Msgf("done, dialing ssh %s %v", addr, sshConf)
	client, err := ssh.Dial("tcp", addr, sshConf)
	if err != nil {
		log.Fatal().Msgf("unable to open ssh connection: %s", err)
	}
	defer client.Close()

	log.Debug().Msg("starting upload user script")
	usu := userScriptUploader{log: log}
	if err := usu.uploadUserScript(client, *containerId); err != nil {
		fatalError := fmt.Errorf("unable to upload user script: %s", err)
		mustMarkFatal(db, *operationUuid, fatalError.Error())
		log.Fatal().Msgf("%s", fatalError)
	}
	log.Debug().Msg("user script uploaded")

	go watchForCancellations(conf, db, activitySink, *operationUuid)

	log.Debug().Msg("running user script")
	err = runUserScript(log, client, activitySink, db, *operationUuid, *entrypoint, conf)
	log.Debug().Msg("done, checking response type")
	switch e := err.(type) {
	case FatalError:
		mustMarkFatal(db, *operationUuid, e.Error())
		log.Fatal().Msgf("Fatal error: %s", e)
	case *ssh.ExitError:
		if n := e.ExitStatus(); n != 0 {
			log.Debug().Msgf("exit status: %d", n)
		}
	}
	log.Debug().Msgf("exiting cleanly")
}

func markFatal(db *sqlx.DB, operationUuid string, fatal string) error {
	tx := mustBeginTx(db)
	defer tx.Rollback()
	store := stores.NewDbOperationStore(tx)
	err := store.MarkAsFailed(operationUuid)
	if err != nil {
		return fmt.Errorf("MarkAsFailed: %s", err)
	}
	if err := store.MarkFatalError(operationUuid, fatal); err != nil {
		return fmt.Errorf("MarkFatalError: %s", err)
	}
	mustCommitTx(tx)

	return nil
}

func mustMarkFatal(db *sqlx.DB, operationUuid string, fatal string) {
	if err := markFatal(db, operationUuid, fatal); err != nil {
		log.Fatal().Msgf("markFatal: %s", err)
	}
}

func mustBeginTx(db *sqlx.DB) *sqlx.Tx {
	tx, err := db.Beginx()
	if err != nil {
		log.Fatal().Msgf("Unable to begin tx: %s", err)
	}
	return tx
}

func mustCommitTx(tx *sqlx.Tx) {
	log.Debug().Msg("entering mustCommitTx")
	if err := tx.Commit(); err != nil {
		log.Fatal().Msgf("Unable to commit tx: %s", err)
	}
	log.Debug().Msg("leaving mustCommitTx")
}

func deadlineReached(db *sqlx.DB, c *config.Config, connect string, operationUuid string) {
	log.Error().Msgf("operation(%s) timed out, aborting", operationUuid)
	tx := mustBeginTx(db)
	store := stores.NewDbOperationStore(tx)
	store.MarkAsTimedOut(operationUuid)
	mustCommitTx(tx)

	activityBus := activity.NewAMQPTransport(c.AmqpConnectionString(), "harrow/fsbuilder")
	defer activityBus.Close()
	activityBus.Publish(activities.OperationTimedOut(operationUuid))

	// deleteContainer(connect, operationUuid)

	// sysexits.h: #define EX_TEMPFAIL	75	/* temp failure; user is invited to retry */
	os.Exit(75)
}

func watchForCancellations(c *config.Config, db *sqlx.DB, activitySink ActivitySink, operationUuid string) {
	consumerId := fmt.Sprintf("controller-%s", operationUuid)
	broadcastBus := broadcast.NewAutoDeletingAMQPTransport(c.AmqpConnectionString(), consumerId)
	broadcastBus.OnlyTable("activities")
	creations, err := broadcastBus.Consume(broadcast.Create)
	if err != nil {
		mustMarkFatal(db, operationUuid, "Listening to activity creations: "+err.Error())
		return
	}

	go func() {
		defer broadcastBus.Close()

		for message := range creations {
			if message.Table() != "activities" {
				message.Acknowledge()
				continue
			}
			id, _ := strconv.Atoi(message.UUID())
			log.Debug().Msgf("received message id=%d table=%s", id, message.Table())
			tx := mustBeginTx(db)
			activityStore := stores.NewDbActivityStore(tx)
			activity, err := activityStore.FindActivityById(id)
			if err != nil {
				log.Warn().Msg("activity not found")
				if err := message.RejectForever(); err != nil {
					log.Warn().Msgf("message.RejectForever(): %s", err)
				}
				tx.Rollback()
				continue
			}

			if err := message.Acknowledge(); err != nil {
				log.Warn().Msgf("message.Acknowledge(): %s", err)
			}

			if activity.Name != "operation.canceled-by-user" {
				tx.Rollback()
				continue
			}

			log.Debug().Msgf("processing %s@%d", activity.Name, activity.Id)

			payload, ok := activity.Payload.(*activities.OperationCanceledByUserPayload)
			if !ok {
				log.Fatal().Msgf("invalid payload for activity: expected %t, got %t", payload, activity.Payload)
			}

			if payload.Uuid != operationUuid {
				tx.Rollback()
				continue
			}

			operationStore := stores.NewDbOperationStore(tx)
			if err := operationStore.MarkAsCanceled(operationUuid); err != nil {
				log.Error().Msgf("failed to mark operation as canceled: %s", err)
			}
			tx.Commit()

			statusLogEntry := cast.NewStatusLogEntry("user.canceled", "Operation canceled by user")
			handleEvent(db, statusLogEntry, activitySink, operationUuid)

			broadcastBus.Close()
			os.Exit(0)
		}
	}()
}
