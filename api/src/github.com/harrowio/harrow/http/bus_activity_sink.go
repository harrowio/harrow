package http

import (
	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/domain"
)

type BusActivitySink struct {
	sink activity.Sink
}

func NewBusActivitySink(bus activity.Sink) *BusActivitySink {
	return &BusActivitySink{
		sink: bus,
	}
}

func (self *BusActivitySink) EmitActivity(activity *domain.Activity) {
	if err := self.sink.Publish(activity); err != nil {
		// ctxt.Log().Info().Msgf("busactivitysink.emitactivity: %s", err)
	}
}
