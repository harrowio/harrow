package runner

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"net/http"
	_ "net/http/pprof"

	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/config"
	"github.com/rs/zerolog"
)

var (
	ProgramName string         = "runner"
	log         zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()
)

func Main() {

	// Define flags with sane defaults as far as possible
	connStr := flag.String("connect", "ssh://root@host:port", "lxd host to connect to")
	flag.Parse()

	// Set up handler for signals from the operating system (e.g CTRL+C)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	listener, err := net.Listen("tcp", ":0")
	go func(l net.Listener) {
		if err != nil {
			panic(err)
		}
		log.Info().Msgf("debug http server on port %d", l.Addr().(*net.TCPAddr).Port)
		http.Serve(l, nil)
	}(listener)

	// Get configuration from ENV (see config package)
	config := config.GetConfig()

	// we use this to emit some activitiyes when things happen
	activityBus := activity.NewAMQPTransport(config.AmqpConnectionString(), fmt.Sprintf("runner-%s", connStr))
	defer activityBus.Close()

	// Configure the runner with the things we have (log, interval, etc)
	// and start it in a goroutine
	runner := &Runner{
		config:       config,
		errs:         make(chan error),
		interval:     60,
		log:          log.With().Str("host", *connStr).Logger(),
		activitySink: activityBus,
	}

	if err := runner.SetLXDConnStr(*connStr); err != nil {
		log.Fatal().Msgf("unable to set runner conn str: %s", err)
	}

	log.Info().Msgf("starting runner on host %s", *connStr)
	go runner.Start()

	// Wait for the runner to return an error or for an OS signal and then
	// continue.
Wait:
	for {
		select {
		case e := <-runner.errs:
			if err == nil {
				log.Error().Msgf("runner sent a nil error (signals successful completion)")
			} else {
				log.Error().Msgf("runner sent an error: %s", e)
			}

			// Close resources, we don't exit main, so deferred functions are not run
			activityBus.Close()
			listener.Close()

			runner.Stop() // premature stop, running syscall.Exec will mean we never continue

			executable, _ := os.Executable()
			execErr := syscall.Exec(executable, os.Args, os.Environ())
			if execErr != nil {
				panic(execErr)
			}

		case s := <-sig:
			log.Error().Msgf("received signal: %s", s)
			break Wait
		}
	}

	// Stop the runner before we exit, as far as possible we'll wait for it to
	// clean up after itself.
	runner.Stop()
}
