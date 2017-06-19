package activity

import "github.com/harrowio/harrow/domain"

type MemoryMessage struct {
	activity     *domain.Activity
	Acknowledged bool
}

func NewMemoryMessage(activity *domain.Activity) *MemoryMessage {
	return &MemoryMessage{
		activity:     activity,
		Acknowledged: false,
	}
}

func (self *MemoryMessage) Acknowledge() error {
	self.Acknowledged = true
	return nil
}

func (self *MemoryMessage) Activity() *domain.Activity {
	return self.activity
}

type MemoryTransport struct {
	publish chan Message
}

func NewMemoryTransport() *MemoryTransport {
	return &MemoryTransport{
		publish: make(chan Message, 1),
	}
}

func (self *MemoryTransport) Consume() (chan Message, error) {
	return self.publish, nil
}

func (self *MemoryTransport) Published() chan Message {
	return self.publish
}

func (self *MemoryTransport) Publish(activity *domain.Activity) error {
	self.publish <- NewMemoryMessage(activity)
	return nil
}

func (self *MemoryTransport) Close() error {
	close(self.publish)
	return nil
}
