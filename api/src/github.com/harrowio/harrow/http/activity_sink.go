package http

import (
	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/domain"
)

// An ActivitySink is used for emitting activities over some form of
// messaging system.
type ActivitySink interface {
	EmitActivity(activity *domain.Activity)
}

// NullSink is the default activity sink which does nothing
type NullSink struct{}

func NewNullSink() *NullSink                                  { return &NullSink{} }
func (sink *NullSink) EmitActivity(activity *domain.Activity) {}

// BusSink sends activities over an activity bus
type BusSink struct {
	bus activity.Sink
}

func NewBusSink(bus activity.Sink) *BusSink {
	return &BusSink{
		bus: bus,
	}
}

func (self *BusSink) EmitActivity(activity *domain.Activity) {
	if err := self.bus.Publish(activity); err != nil {
		// TODO: Raise an error here
	}
}
