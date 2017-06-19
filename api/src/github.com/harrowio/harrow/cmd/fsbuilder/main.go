package fsbuilder

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	redis "gopkg.in/redis.v2"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/bus/activity"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/fsbuilder"
	"github.com/harrowio/harrow/fsbuilder/rootfs"
	"github.com/harrowio/harrow/stores"
)

const ProgramName = "fsbuilder"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {
	c := config.GetConfig()
	interval := 2 * time.Second
	operationUuid := flag.String("operation-uuid", "", "The operation to build a rootfs for")

	flag.Parse()
	if *operationUuid == "" {
		flag.Usage()
		os.Exit(2)
	}
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		log.Error().Msg("cowardly refusing to write compressed archive to a terminal.")
	}
	var err error
	db, err := c.DB()
	if err != nil {
		log.Error().Msgf("Unable to get db: %s", err)
	}

	deadline := time.After(config.InstanceDeadline)
	go func() {
		<-deadline
		deadlineReached(c, db, operationUuid)
	}()

	tx, err := db.Beginx()

	if err != nil {
		log.Error().Msgf("unable to start tx: %s", err)
	}
	defer tx.Rollback()

	for {
		if err := buildRootFs(c, os.Stdout, tx, *operationUuid); err != nil {
			log.Error().Msgf("unable to build rootfs: %s", err)
			time.Sleep(interval)
		} else {
			return
		}
	}
}

func buildRootFs(c *config.Config, out io.WriteCloser, tx *sqlx.Tx, uuid string) error {
	redisClient := redis.NewTCPClient(c.RedisConnOpts(1))
	defer redisClient.Close()
	ss := stores.NewRedisSecretKeyValueStore(redisClient)
	config := fsbuilder.NewConfig(ss, tx)
	builder := rootfs.NewBuilder(config)
	builder.SetLogger(log)
	reader, err := builder.Build(uuid)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, reader)
	if err != nil {
		return err
	}
	return out.Close()
}

func deadlineReached(c *config.Config, db *sqlx.DB, operationUuid *string) {
	log.Error().Msgf("Operation(%s) timed out, aborting", *operationUuid)
	tx, err := db.Beginx()
	if err != nil {
		panic(fmt.Sprintf("Unable to create tx, can't mark operation as timed out: %s", err))
	}
	defer tx.Rollback()
	store := stores.NewDbOperationStore(tx)
	store.MarkAsTimedOut(*operationUuid)
	err = tx.Commit()
	if err != nil {
		panic(fmt.Sprintf("Unable to commit tx, can't mark operation as timed out: %s", err))
	}

	activityBus := activity.NewAMQPTransport(c.AmqpConnectionString(), "harrow/fsbuilder")
	defer activityBus.Close()
	activityBus.Publish(activities.OperationTimedOut(*operationUuid))

	// sysexits.h: #define EX_TEMPFAIL	75	/* temp failure; user is invited to retry */
	os.Exit(75)
}
