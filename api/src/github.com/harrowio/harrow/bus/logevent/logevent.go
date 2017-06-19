package logevent

import "github.com/harrowio/harrow/loxer"

type Message struct {
	O  string
	FD int
	T  int64
	E  loxer.SerializedEvent
}

func (self *Message) OperationUUID() string {
	return self.O
}

func (self *Message) Event() loxer.Event {
	return self.E.Inner
}

type Sink interface {
	Publish(operationUUID string, fd int, time int64, event loxer.Event) error
	Close() error
}

type Source interface {
	Consume(operationUUID string) (<-chan *Message, error)
	Close() error
}
