package activityWorker

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/config"
	"github.com/rs/zerolog"
)

const ProgramName = "activity-worker"

var logger zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {

	conf := config.GetConfig()
	bus := activity.NewAMQPTransport(conf.AmqpConnectionString(), "activity-worker")
	defer bus.Close()

	db, err := conf.DB()
	if err != nil {
		logger.Fatal().Err(err)
	}

	store := NewDbActivityStore(db)
	// store.SetLogger(logger)

	worker := NewActivityWorker(bus, store).
		AddMessageHandler(ListProjectMembers(db)).
		AddMessageHandler(MarkProjectUuid(db)).
		AddMessageHandler(MarkJobUuid(db)).
		AddMessageHandler(logMessage)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	worker.Start()

	select {
	case sig := <-signals:
		logger.Info().Msgf("Received %s", sig)
		logger.Info().Msgf("Stopping...")
		worker.Stop()
		logger.Info().Msgf("Stopped")
		os.Exit(0)
	}
}

func logMessage(msg activity.Message) {
	activity := msg.Activity()
	logger.Info().Msgf("received %q", activity.Name)
	logger.Info().Msgf("audience: %v", activity.Audience())
}
