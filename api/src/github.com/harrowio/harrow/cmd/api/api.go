package api

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/http"
	"github.com/harrowio/harrow/stores"
	"github.com/rs/zerolog"

	redis "gopkg.in/redis.v2"
)

const ProgramName = "api"

func Main() {

	var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

	c := config.GetConfig()

	db, err := c.DB()
	if err != nil {
		log.Fatal().Msgf("error opening database handle %s", err)
	}

	bus := activity.NewAMQPTransport(c.AmqpConnectionString(), "harrow/http")
	defer bus.Close()

	kv := stores.NewRedisKeyValueStore(redis.NewTCPClient(c.RedisConnOpts(0)))
	ss := stores.NewRedisSecretKeyValueStore(redis.NewTCPClient(c.RedisConnOpts(1)))

	go http.ListenAndServe(log, db, bus, kv, ss, c)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	s := <-signals

	log.Info().Msgf("got signal %s, exiting", s)
}
