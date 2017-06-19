package buildStatusWorker

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"github.com/harrowio/harrow/bus/broadcast"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/stores"
)

const ProgramName = "build-status-worker"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {

	c := config.GetConfig()
	db, err := c.DB()
	if err != nil {
		log.Fatal().Err(err)
	}

	broadcastBus := broadcast.NewAMQPTransport(c.AmqpConnectionString(), "build-status-worker")
	activityCreations, err := broadcastBus.Consume(broadcast.Create)
	defer broadcastBus.Close()

	for message := range activityCreations {
		if message.Table() != "activities" {
			message.RejectForever()
			continue
		}

		activityID, err := strconv.Atoi(message.UUID())
		if err != nil {
			log.Info().Msgf("invalid activity id: %q", message.UUID())
			message.RejectForever()
			continue
		}

		activity, err := LoadActivity(db, activityID)
		if err != nil {
			log.Info().Msgf("error loading activity: %s", err)
			message.RejectForever()
			continue
		}

		if !strings.HasPrefix(activity.Name, "operation.") {
			message.Acknowledge()
			continue
		}

		if activity.Name == "operation.started" {
			message.Acknowledge()
			continue
		}

		log.Info().Msgf("handling %s@%d\n", activity.Name, activityID)
		operation, ok := activity.Payload.(*domain.Operation)
		if !ok {
			log.Info().Msgf("payload for %s is %t (want %t), skipping", activity.Name, activity.Payload, operation)
			message.RejectForever()
			continue
		}

		command := exec.Command("/srv/harrow/bin/harrow-any", "report-build-status-to-github", "--operation-uuid", operation.Uuid)
		go run(message, log, command)
	}
}

func LoadActivity(db *sqlx.DB, id int) (*domain.Activity, error) {
	tx, err := db.Beginx()
	defer tx.Rollback()
	if err != nil {
		return nil, err
	}

	return stores.NewDbActivityStore(tx).FindActivityById(id)
}

func run(msg broadcast.Message, log logger.Logger, command *exec.Cmd) {
	log.Info().Msgf("run %s %s\n", command.Path, strings.Join(command.Args, " "))
	output, err := command.CombinedOutput()
	if err != nil {
		log.Info().Msgf("error %s: %s", err, output)
		msg.RejectForever()
	} else {
		log.Info().Msgf("success: %s", output)
		msg.Acknowledge()
	}
}
