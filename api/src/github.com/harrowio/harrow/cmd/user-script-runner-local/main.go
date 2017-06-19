package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/harrowio/harrow/cast"
)

const usageString = `
Run CMD with ARG with input and output redirections.

All output on the commands' stderr and stdout is captured and written
as a structured message in JSON format to stdout.  Additionally any
control messages written by CMD to file descriptor 3 will be part of
the output stream.
 `

func main() {
	help := false
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flags.BoolVar(&help, "help", false, "print usage information")
	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] CMD ARG...\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flags.PrintDefaults()
		fmt.Fprintf(os.Stderr, usageString)
	}

	flags.Parse(os.Args[1:])

	if help {
		flags.Usage()
		os.Exit(0)
	}

	args := flags.Args()
	if len(args) == 0 {
		flags.Usage()
		os.Exit(1)
	}

	cmd := cast.NewLocalCommand(os.Stdout, args[0], args[1:]...)
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
	}
	fmt.Printf("\n")
	os.Exit(cmd.ExitStatus())
}
