package activity

import "github.com/harrowio/harrow/domain"

type Message interface {
	Activity() *domain.Activity
	Acknowledge() error
}

type Sink interface {
	Publish(activity *domain.Activity) error
	Close() error
}

type Source interface {
	Consume() (chan Message, error)
	Close() error
}
