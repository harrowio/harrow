package broadcast

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

type AMQPMessage struct {
	d       amqp.Delivery
	isD     bool
	Payload map[string]string
	tepy    BroadcastMessageType
}

func newFromDelivery(d amqp.Delivery) (*AMQPMessage, error) {
	msg := &AMQPMessage{
		d:       d,
		isD:     true,
		tepy:    BroadcastMessageType(d.Exchange),
		Payload: map[string]string{},
	}

	err := json.Unmarshal(d.Body, msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}

func newAMQPMessage(tepy, table, uuid string) *AMQPMessage {
	msg := &AMQPMessage{
		Payload: map[string]string{
			"table": table,
			"uuid":  uuid,
		},
		tepy: BroadcastMessageType(tepy),
	}
	return msg
}

func (m AMQPMessage) UUID() string {
	return m.Payload["uuid"]
}

func (m AMQPMessage) Type() BroadcastMessageType {
	return m.tepy
}

func (m AMQPMessage) Table() string {
	return m.Payload["table"]
}

func (m AMQPMessage) Acknowledge() error {
	return m.d.Ack(false)
}

func (m AMQPMessage) RejectOnce() error {
	return m.d.Reject(true) // do re-enqueue
}

func (m AMQPMessage) RejectForever() error {
	return m.d.Reject(false) // don't re-enqueue
}

func (m AMQPMessage) RequeueAfter(d time.Duration) error {
	go func() {
		<-time.After(d)
		err := m.RejectOnce()
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

func (m AMQPMessage) String() string {
	return fmt.Sprintf("broadcast.AMQPMessage(tepy=%s,table=%s,uuid=%s)", m.tepy, m.Table(), m.UUID())
}
