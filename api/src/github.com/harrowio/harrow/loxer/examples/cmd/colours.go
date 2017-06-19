package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/harrowio/harrow/bus/logevent"
	"github.com/harrowio/harrow/loxer"
)

type Event struct {
	loxer.Event
	Fd uintptr
	T  int64
}

func (e *Event) String() string {
	data, err := json.Marshal(e.Event)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("fd#%d %s", e.Fd, string(data))
}

func main() {

	stop := make(chan chan bool)
	events := make(chan logevent.Message)
	go func(events chan logevent.Message, stop chan chan bool) {
		for {
			select {
			case stopped := <-stop:
				stopped <- true
				return
			case e := <-events:
				data, err := json.Marshal(e)
				if err != nil {
					panic(err)
				}
				fmt.Println(string(data))
			}
		}
	}(events, stop)

	cmd := exec.Command("colours.sh")

	stdoutHandler := func(fd uintptr, events chan logevent.Message) func(loxer.Event) {
		return func(e loxer.Event) { events <- logevent.Message{O: "1234", E: loxer.SerializedEvent{e}} }
	}(os.Stdout.Fd(), events)
	stderrHandler := func(fd uintptr, events chan logevent.Message) func(loxer.Event) {
		return func(e loxer.Event) { events <- logevent.Message{O: "1234", E: loxer.SerializedEvent{e}} }
	}(os.Stderr.Fd(), events)

	cmd.Stdout = loxer.NewLexer(stdoutHandler)
	cmd.Stderr = loxer.NewLexer(stderrHandler)

	if err := cmd.Run(); err != nil {
		panic(err)
	}

	// Stop and wait for the consumer to finish
	stopped := make(chan bool)
	stop <- stopped
	<-stopped

}
