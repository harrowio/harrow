package activity

import (
	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/streadway/amqp"
)

type AMQPMessage struct {
	delivery amqp.Delivery
	activity *domain.Activity
}

func NewAMQPMessage(delivery amqp.Delivery) *AMQPMessage {
	return &AMQPMessage{
		delivery: delivery,
	}
}

func (self *AMQPMessage) Parse() error {
	if self.activity != nil {
		return nil
	}

	activity, err := activities.UnmarshalJSON(self.delivery.Body)
	if err != nil {
		return err
	}

	self.activity = activity
	return nil
}

func (self *AMQPMessage) Activity() *domain.Activity {
	return self.activity
}

func (self *AMQPMessage) Acknowledge() error {
	return self.delivery.Ack(false)
}
