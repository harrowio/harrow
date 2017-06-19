package broadcast

import "time"

type BroadcastMessageType string

const (
	Create BroadcastMessageType = "broadcast.create"
	Change BroadcastMessageType = "broadcast.change"
)

type Message interface {
	UUID() string
	Table() string
	Type() BroadcastMessageType
	Acknowledge() error
	RejectOnce() error
	RejectForever() error
	RequeueAfter(d time.Duration) error
}

type Sink interface {
	Publish(tepy, table, uuid string) error
	Close() error
}

type Source interface {
	Consume(tepy BroadcastMessageType) (<-chan Message, error)
	Close() error
}
