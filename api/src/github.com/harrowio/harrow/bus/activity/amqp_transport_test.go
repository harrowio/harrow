package activity

import (
	"fmt"
	"sync"
	"testing"

	"net/http"
	_ "net/http/pprof"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
)

var (
	conf = config.GetConfig()
)

func init() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

func Test_AMQPTransport_endToEnd(t *testing.T) {
	bus := NewAMQPTransport(conf.AmqpConnectionString(), "amqp-transport-test")

	payloads := []*domain.Activity{
		domain.NewActivity(1, "test.run"),
		domain.NewActivity(2, "test.run"),
	}

	consumer, err := bus.Consume()
	if err != nil {
		t.Fatal(err)
	}
	producers := &sync.WaitGroup{}
	producers.Add(2)
	go func() {
		for i, payload := range payloads {
			if err := bus.Publish(payload); err != nil {
				t.Fatalf("[%d/%d] bus.Publish(%q): %s", i+1, len(payloads), payload.Name, err)
			}
		}
		producers.Done()
	}()

	activities := []*domain.Activity{}

	go func() {
		defer producers.Done()
		seen := 0
		for message := range consumer {
			activities = append(activities, message.Activity())
			if err := message.Acknowledge(); err != nil {
				t.Errorf("error: %s", err)
			}
			seen++
			if seen == len(payloads) {
				return
			}
		}
	}()

	producers.Wait()
	bus.Close()

	if got, want := len(activities), len(payloads); got != want {
		t.Fatalf("len(activities) = %d; want %d", got, want)
	}
}
