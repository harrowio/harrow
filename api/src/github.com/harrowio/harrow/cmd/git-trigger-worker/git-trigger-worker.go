package gitTriggerWorker

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/harrowio/harrow/bus/broadcast"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

const ProgramName = "git-trigger-worker"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {
	c := config.GetConfig()
	db, err := c.DB()
	if err != nil {
		log.Fatal().Err(err)
	}
	defer db.Close()

	bus := broadcast.NewAMQPTransport(c.AmqpConnectionString(), "git-trigger-worker")
	defer bus.Close()

	index := NewDbTriggerIndex(db)
	scheduler := NewDbScheduler(db)
	worker := NewGitTriggerWorker(
		index,
		scheduler,
	)
	worker.log = log

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGKILL, syscall.SIGTERM)

	messages, err := bus.Consume(broadcast.Create)
	if err != nil {
		log.Fatal().Err(err)
	}

	for {
		select {
		case message := <-messages:
			activity := activityForMessage(db, message)
			if activity == nil {
				continue
			}
			worker.HandleActivity(activity)
		case <-signals:
			return
		}
	}
}

func activityForMessage(db *sqlx.DB, message broadcast.Message) *domain.Activity {
	if message.Table() != "activities" {
		message.RejectForever()
		return nil
	}

	activityId, err := strconv.Atoi(message.UUID())
	if err != nil {
		log.Error().Msgf("invalid activity id: %q: %s\n", message.UUID(), err)
		message.RejectForever()
		return nil
	}

	tx := db.MustBegin()
	defer tx.Rollback()
	activities := stores.NewDbActivityStore(tx)

	activity, err := activities.FindActivityById(activityId)
	if err != nil {
		log.Error().Msgf("activity not found: id=%v\n", activityId)
		message.RejectForever()
		return nil
	}
	message.Acknowledge()
	return activity
}
