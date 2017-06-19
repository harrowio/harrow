package http

import (
	"testing"

	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/domain"
)

func Test_BusActivitySink_emitsActivitiesOverActivityBus(t *testing.T) {
	bus := activity.NewMemoryTransport()
	sink := NewBusActivitySink(bus)
	activity := domain.NewActivity(1, "test.run")
	sink.EmitActivity(activity)

	if got, want := len(bus.Published()), 1; got != want {
		t.Errorf("len(bus.Published()) = %d; want %d", got, want)
	}
}
