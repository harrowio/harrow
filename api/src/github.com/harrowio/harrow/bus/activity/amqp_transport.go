package activity

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/harrowio/harrow/clock"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/streadway/amqp"
)

const exchangeName = "activities"

type AMQPTransport struct {
	consumerName string
	url          string
	consumers    []chan Message
	connection   *amqp.Connection
	log          logger.Logger
}

func NewAMQPTransport(url, consumerName string) *AMQPTransport {

	mt := &AMQPTransport{
		consumerName: consumerName,
		url:          url,
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

func (self *AMQPTransport) dial() {
	for i := 0; i < 10; i++ {
		conn, err := amqp.DialConfig(self.url, amqp.Config{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, 2*time.Second)
			},
		})
		if err != nil {
			self.Log().Error().Msgf("activities - unable to amqp.dial: %s", err)
			time.Sleep(1 * time.Second)
		} else {
			self.connection = conn
			return
		}
	}
	self.Log().Fatal().Msg("failed to dial amqp after 10 attempts\n")
}

func (self *AMQPTransport) consumeOn() (<-chan amqp.Delivery, error) {
	receiveOn, err := self.channel()
	if err != nil {
		return nil, err
	}
	queueName, err := self.declareConsumerQueue(receiveOn)
	if err != nil {
		return nil, err
	}

	consumerNameAutoGenerate := ""
	autoAck := false
	exclusive := true
	noLocal := false
	noWait := false
	args := (amqp.Table)(nil)
	consumeOn, err := receiveOn.Consume(
		queueName,
		consumerNameAutoGenerate,
		autoAck,
		exclusive,
		noLocal,
		noWait,
		args,
	)

	if err != nil {
		return nil, err
	}
	return consumeOn, nil

}

func (self *AMQPTransport) Consume() (chan Message, error) {
	out := make(chan Message)

	self.consumers = append(self.consumers, out)

	go self.handleWork(out)

	return out, nil
}

func (self *AMQPTransport) handleWork(out chan Message) {
	work, err := self.consumeOn()
	if err != nil {
		panic(fmt.Errorf("Unable to consumeOn: %s", err))
	}
	for {
		delivery, ok := <-work
		if !ok {
			self.Log().Error().Msg("activities - lost amqp connection, re-dialing")
			self.dial()
			work, err = self.consumeOn()
			if err != nil {
				panic(fmt.Sprintf("activities - Unable to open work channel: %s", err))
			}
			self.Log().Warn().Msg("activities - successfully reconnected")
			continue
		}
		message := NewAMQPMessage(delivery)
		if err := message.Parse(); err != nil {
			self.Log().Error().Msgf("invalid json: %s\n---\n%s\n---\n", err, delivery.Body)
			rejectAMQPDeliveryForever(delivery)
		} else {
			out <- message
		}
	}
}

func rejectAMQPDeliveryForever(delivery amqp.Delivery) { delivery.Reject(false) }

func (self *AMQPTransport) Close() error {
	for _, consumer := range self.consumers {
		close(consumer)
	}
	return self.connection.Close()
}

func (self *AMQPTransport) declareConsumerQueue(channel *amqp.Channel) (string, error) {
	queueName := exchangeName
	durable := true
	autoDelete := false
	exclusive := false
	noWait := false
	args := (amqp.Table)(nil)
	if _, err := channel.QueueDeclare(
		queueName,
		durable,
		autoDelete,
		exclusive,
		noWait,
		args,
	); err != nil {
		return "", err
	}

	routingKeyMatchAll := "#"
	if err := channel.QueueBind(
		queueName,
		routingKeyMatchAll,
		exchangeName,
		noWait,
		args,
	); err != nil {
		return "", err
	}

	return queueName, nil
}

func (self *AMQPTransport) channel() (*amqp.Channel, error) {

	var channel *amqp.Channel
	var err error
	for {
		channel, err = self.connection.Channel()
		if e, ok := err.(*net.OpError); ok {
			self.Log().Error().Msgf("activities - self.connection.channel(): *net.operror encountered, redialing: %s", e)
			self.dial()
			continue
		}
		if err != nil {
			return nil, err
		}
		break
	}

	kind := "topic"
	durable := true
	autoDelete := false
	internal := false
	noWait := false
	args := (amqp.Table)(nil)
	if err := channel.ExchangeDeclare(exchangeName, kind, durable, autoDelete, internal, noWait, args); err != nil {
		return nil, err
	}

	return channel, nil
}

func (self *AMQPTransport) Publish(activity *domain.Activity) error {
	sendOn, err := self.channel()
	if err != nil {
		return err
	}
	defer sendOn.Close()

	name, err := self.declareConsumerQueue(sendOn)
	if err != nil {
		self.Log().Warn().Msgf("error declaring queue %q: %s. messages might get lost.", name, err)
	}

	body, err := json.Marshal(activity)
	if err != nil {
		return err
	}
	amqpMessage := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Timestamp:    clock.Default.Now(),
		ContentType:  "application/json",
		Body:         body,
	}

	mandatory := true
	immediate := false
	return sendOn.Publish(exchangeName, activity.Name, mandatory, immediate, amqpMessage)
}
