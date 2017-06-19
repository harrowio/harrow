package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/harrowio/harrow/cast"
)

type Arguments []string

func (self Arguments) Parse() (cast.Payload, error) {
	result := cast.Payload{}
	for _, arg := range self {
		fields := strings.Split(arg, "=")
		key := fields[0]
		value := ""
		filename := strings.Split(key, "@")
		if len(filename) == 2 {
			key = filename[0]
			data, err := []byte(nil), (error)(nil)
			if filename[1] == "-" {
				data, err = ioutil.ReadAll(os.Stdin)
				defer os.Stdin.Close()
			} else {
				data, err = ioutil.ReadFile(filename[1])
			}

			if err != nil {
				return nil, err
			}
			value = string(data)
		} else if len(fields) > 1 {
			value = strings.Join(fields[1:], "=")
		}

		result[key] = append(result[key], value)
	}

	return result, nil
}

func (self Arguments) ToMessage(eventName string) (*cast.ControlMessage, error) {
	payload, err := self.Parse()
	if err != nil {
		return nil, err
	}
	payload.Set("event", eventName)
	message := &cast.ControlMessage{
		Type:    cast.Event,
		Payload: payload,
	}

	return message, nil
}

func main() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	dial := flag.String("dial", "localhost:2003", "send events to this address")
	flags.Parse(os.Args)
	if len(flags.Args()) < 2 {
		Usage()
		os.Exit(1)
	}

	args := flags.Args()
	eventName := args[1]
	message, err := Arguments(args[2:]).ToMessage(eventName)
	if err != nil {
		die(err)
	}

	out, err := net.Dial("tcp", *dial)
	if err != nil {
		json.NewEncoder(os.NewFile(3, "/dev/events")).Encode(message)
		return
	}
	defer out.Close()

	err = json.NewEncoder(out).Encode(message)
	if err != nil {
		die(err)
	}
}

func Usage() {
	fmt.Printf(`%s <event_name> [<arg>=<val>...]

Notifies the operation runner about <event_name> with the provided
arguments.
`, os.Args[0])
}

func die(thing interface{}) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], thing)
	os.Exit(1)
}
