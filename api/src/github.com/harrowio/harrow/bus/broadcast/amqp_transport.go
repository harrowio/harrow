package broadcast

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/harrowio/harrow/logger"
	"github.com/streadway/amqp"
)

var (
	ErrMalformedMessagePayload = errors.New("can't unmarshal message payload")
)

type AMQPTransport struct {
	sync.Mutex
	consumerName string
	autoDelete   bool
	url          string
	c            *amqp.Connection
	busses       []BroadcastMessageType
	terminator   chan chan error

	log logger.Logger

	exclusiveWorkQueue bool
	consumeRoutingKey  string
}

// NewAMQPTransport returns an AMQPTransport which implements
// broadcast.Source{} and broadcast.Sink{}. This AMQPTransport must be closed
// cleanly by the caller (unless the process is about to exit), this Close() call
// is not defined to be part of either the broadcast.Source{} or
// broadcast.Sink{} interfaces as it is not common to all sources or sinks.
func NewAMQPTransport(url, consumerName string) *AMQPTransport {
	mt := &AMQPTransport{
		consumerName:       consumerName,
		url:                url,
		busses:             []BroadcastMessageType{Create, Change},
		exclusiveWorkQueue: true,
		consumeRoutingKey:  "#",
	}
	mt.dial()
	return mt
}

func (self *AMQPTransport) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *AMQPTransport) SetLogger(l logger.Logger) {
	self.log = l
}

func (self *AMQPTransport) OnlyTable(tableName string) *AMQPTransport {
	self.consumeRoutingKey = tableName
	return self
}

func (self *AMQPTransport) ShareWork() *AMQPTransport {
	self.exclusiveWorkQueue = false
	return self
}

func NewAutoDeletingAMQPTransport(url, consumerName string) *AMQPTransport {
	mt := &AMQPTransport{
		consumerName:      consumerName,
		autoDelete:        true,
		url:               url,
		busses:            []BroadcastMessageType{Create, Change},
		consumeRoutingKey: "#",
	}
	mt.dial()
	return mt
}

func (mt *AMQPTransport) dial() {
	for {
		conn, err := amqp.Dial(mt.url)
		if err != nil {
			mt.Log().Error().Msgf("broadcast - Unable to amqp.Dial: %s", err)
			time.Sleep(1 * time.Second)
		} else {
			mt.c = conn
			return
		}
	}
}

func (mt *AMQPTransport) queueForConsumer(exchangeName, consumerName string) (string, error) {

	var err error
	var c *amqp.Channel

	var queueName string = fmt.Sprintf("%s-%s", exchangeName, consumerName)

	c, err = mt.newChannel()
	defer c.Close()
	if err != nil {
		return "", err
	}

	if _, err = c.QueueDeclare(
		queueName,     // queueName
		true,          // durable
		mt.autoDelete, // autoDelete
		false,         // exclusive
		false,         // noWait
		mt.args(),     // args
	); err != nil {
		return "", err
	}

	routingKey := "#"
	mt.Log().Debug().Msgf("consumeroutingkey=%q", mt.consumeRoutingKey)
	if mt.consumeRoutingKey != "#" {
		switch exchangeName {
		case string(Create):
			routingKey = fmt.Sprintf("%s.created", mt.consumeRoutingKey)
		case string(Change):
			routingKey = fmt.Sprintf("%s.changed", mt.consumeRoutingKey)
		default:
			routingKey = "#"
		}
	}
	mt.Log().Debug().Msgf("routingkey=%q", routingKey)
	if err = c.QueueBind(
		queueName,    // queueName
		routingKey,   // routingKey
		exchangeName, // exchangeName
		false,        // noWait
		nil,          // args
	); err != nil {
		return "", err
	}

	return queueName, nil
}

func (mt *AMQPTransport) workChan(tepy BroadcastMessageType) (<-chan amqp.Delivery, error) {
	c, err := mt.newChannel()
	if err != nil {
		return nil, fmt.Errorf("unable to open channel: %s", err)
	}
	queueName, err := mt.queueForConsumer(string(tepy), mt.consumerName)
	if err != nil {
		return nil, fmt.Errorf("unable to declare queue: %s", err)
	}
	work, err := c.Consume(
		queueName, // queueName
		"",        // consumerName
		false,     // autoAck
		mt.exclusiveWorkQueue, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return nil, fmt.Errorf("unable to consume: %s", err)
	}
	return work, nil

}

func (mt *AMQPTransport) ConsumeOne(kind BroadcastMessageType) (Message, error) {
	deliveries, err := mt.workChan(kind)
	if err != nil {
		return nil, err
	}

	delivery := <-deliveries
	message, err := newFromDelivery(delivery)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (mt *AMQPTransport) Consume(tepy BroadcastMessageType) (<-chan Message, error) {
	out := make(chan Message)
	mt.Lock()
	mt.terminator = make(chan chan error)
	mt.Unlock()

	work, err := mt.workChan(tepy)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case errors := <-mt.terminator:
				err := mt.c.Close()
				errors <- err
				return
			case d, ok := <-work:
				if !ok {
					// attempt to reconnect
					mt.Log().Error().Msg("broadcast - lost amqp connection, re-dialing")
					mt.dial()
					w, err := mt.workChan(tepy)
					if err != nil {
						panic(fmt.Sprintf("broadcast - Unable to open work channel: %s", err))
					}
					work = w
					mt.Log().Warn().Msg("broadcast - successfully reconnected")
					continue
				}

				// http://godoc.org/github.com/streadway/amqp#Delivery
				msg, err := newFromDelivery(d)
				if err != nil {
					mt.Log().Warn().Msgf("invalid json received from amqp: %s\n", err)
					msg.RejectForever()
				}
				out <- msg
			}
		}
	}()

	return out, nil

}

func (mt *AMQPTransport) Publish(tepy, table, uuid string) error {

	msg := newAMQPMessage(tepy, table, uuid)
	c, err := mt.newChannel()
	defer c.Close()

	if err != nil {
		return fmt.Errorf("broadcast - newChannel: %s %T", err, err)
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	amqpMsg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
		ContentType:  "application/json",
		Body:         payload,
	}

	routingKey := ""
	switch BroadcastMessageType(tepy) {
	case Create:
		routingKey = fmt.Sprintf("%s.created", table)
	case Change:
		routingKey = fmt.Sprintf("%s.changed", table)
	default:
		panic(fmt.Sprintf("unknown broadcast type: %#v", tepy))
	}

	return c.Publish(
		string(msg.Type()), // exchange
		routingKey,
		true,    // mandatory
		false,   // immediate
		amqpMsg, // amqp.Publishing (msg)
	)
}

// newChannel uses AMQPTransport's amqp.Connection to create and
// return a new channel. It will return an error if it is unable to assert
// the predefined topology and AMQPTransport
func (mt *AMQPTransport) newChannel() (*amqp.Channel, error) {

	var c *amqp.Channel
	var err error
	for {
		c, err = mt.c.Channel()
		if e, ok := err.(*net.OpError); ok {
			mt.Log().Error().Msgf("broadcast - mt.c.channel(): *net.operror encountered, redialing: %s", e)
			mt.dial()
			continue
		}
		if err != nil {
			return nil, err
		}
		break
	}

	if err = mt.declareTopology(c); err != nil {
		return nil, err
	}

	return c, nil
}

// Close() is specific to the AMQPTransport which implements both
// broadcast.Source{} and broadcast.Sink{}
func (mt *AMQPTransport) Close() error {
	var err error
	mt.Lock()
	if mt.terminator != nil {
		errors := make(chan error)
		mt.terminator <- errors
		err = <-errors
	}
	mt.Unlock()
	return err
}

// declareTopology enforces the topology required for this transport
// if any of the function calls here fail, they will render the channel
// inoperable, and it is not safe for future use, thus the immediate
// return.
func (mt *AMQPTransport) declareTopology(c *amqp.Channel) error {

	var err error

	for _, b := range mt.busses {
		if err = c.ExchangeDeclare(
			string(b), // exchangeName
			"topic",   // kind
			true,      // durable
			false,     // autoDelete
			false,     // internal
			false,     // noWait
			nil,       // args
		); err != nil {
			return err
		}
	}

	return nil
}

func (mt *AMQPTransport) args() amqp.Table {
	if mt.autoDelete {
		return amqp.Table{"x-expires": int32(30 * 60 * 1000)}
	}

	return amqp.Table{}
}
