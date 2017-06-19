package userScriptRunner

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/harrowio/harrow/cast"
)

type Config struct {
	basePort  int
	heartrate time.Duration
	args      []string
	help      bool
}

func (self *Config) NewCommand() *cast.Command {
	cmd := cast.NewWithHeartrate(self.heartrate, self.args[0], self.args[1:]...)
	cmd.SetBasePort(self.basePort)
	return cmd
}

const usageString = `
Run CMD with ARG with input and output redirections.

The following mapping is in place:

    stdin    /dev/null
    stdout   localhost:PORT+1
    stderr   localhost:PORT+2
`

const ProgramName = "user-script-runner"

func Main() {
	config := &Config{}
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.IntVar(&config.basePort, "port", 2000, "use this port for port number calculations")
	flags.DurationVar(&config.heartrate, "heartrate", cast.DefaultHeartrate, "interval between heartbeats")
	flags.BoolVar(&config.help, "help", false, "print usage information")
	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] CMD ARG...\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flags.PrintDefaults()
		fmt.Fprintf(os.Stderr, usageString)
	}

	flags.Parse(os.Args[1:])

	if config.help {
		flags.Usage()
		os.Exit(0)
	}

	config.args = flags.Args()
	if len(config.args) == 0 {
		flags.Usage()
		os.Exit(1)
	}

	cmd := config.NewCommand()
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
	}
	os.Exit(cmd.ExitStatus())
}
