package logevent

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/loxer"
	redis "gopkg.in/redis.v2"
)

type RedisTransport struct {
	sync.Mutex
	log        logger.Logger
	redsi      *redis.Client
	terminator chan chan error
}

func NewRedisTransport(redsi *redis.Client, log logger.Logger) *RedisTransport {
	return &RedisTransport{
		redsi: redsi,
		log:   log,
	}
}

type redisConsumer struct {
	log           logger.Logger
	transport     *RedisTransport
	operationUUID string
	messages      chan *Message
	newLength     chan int64
	lastMsgSent   int64
	redsi         *redis.Client
	ps            *redis.PubSub
}

func newRedisConsumer(transport *RedisTransport, operationUUID string) (*redisConsumer, error) {
	transport.redsi.ConfigSet("notify-keyspace-events", "Kl")
	consumer := &redisConsumer{
		operationUUID: operationUUID,
		messages:      make(chan *Message, 1),
		newLength:     make(chan int64, 1),
		transport:     transport,
		log:           transport.log,
	}
	ps := transport.redsi.PubSub()
	err := ps.Subscribe(consumer.subscriptionChannel())
	if err != nil {
		return nil, err
	}
	consumer.ps = ps
	transport.terminator = make(chan chan error)
	return consumer, nil
}

func (self *redisConsumer) Consume() error {

	log.Printf("PRedis Consumer Consume()")
	go func() {
		defer func() {
			self.transport.Lock()
			self.transport.terminator = nil
			self.transport.Unlock()
		}()
		for {
			select {
			case errors := <-self.transport.terminator:
				err := self.terminate()
				errors <- err
				return
			case n, ok := <-self.newLength:
				if !ok {
					return
				}
				payloads, err := self.transport.redsi.LRange(self.operationUUID, self.lastMsgSent, n).Result()
				if err != nil {
					self.log.Warn().Msgf("unable to lrange: %s\n", err)
					self.terminate()
					return
				}
				for _, e := range payloads {
					msg := new(Message)

					err := json.Unmarshal([]byte(e), msg)
					if err != nil {
						self.log.Warn().Msgf("unable to unmarshal %s: %s\n", e, err)
						self.terminate()
						return
					}
					if msg.Event().EventType() == "eof" {
						self.terminate()
						return
					}
					self.messages <- msg
					self.lastMsgSent++
				}
			default:
				r, err := self.ps.ReceiveTimeout(100 * time.Millisecond)
				if err != nil {
					// return on non-neterrs and neterrs that are no timeouts
					if neterr, isNeterr := err.(net.Error); (isNeterr && !neterr.Timeout()) || !isNeterr {
						self.log.Error().Msgf("ps.receivetimeout: %s", err)
						self.terminate()
						return
					}
				}
				// if we hit a timeout, this will return false and we continue
				if _, ok := r.(*redis.Message); ok {
					lastMsg, err := self.transport.redsi.LLen(self.operationUUID).Result()
					if err != nil {
						self.log.Warn().Msgf("Unable to get new list lenght", err)
						self.terminate()
						return
					}
					self.newLength <- lastMsg
				}
			}
		}
	}()

	lastMsg, err := self.transport.redsi.LLen(self.operationUUID).Result()
	if err != nil {
		return err
	}
	self.newLength <- lastMsg

	return nil

}

func (self *redisConsumer) terminate() error {
	err := self.ps.Unsubscribe("__keyspace@0__:" + self.operationUUID)
	self.ps.Close()
	close(self.newLength)
	close(self.messages)
	return err
}

func (self *redisConsumer) subscriptionChannel() string {
	return fmt.Sprintf("__keyspace@0__:%s", self.operationUUID)
}

func (self *RedisTransport) Consume(operationUUID string) (<-chan *Message, error) {
	consumer, err := newRedisConsumer(self, operationUUID)
	if err != nil {
		return nil, err
	}
	err = consumer.Consume()
	if err != nil {
		return nil, err
	}
	return consumer.messages, nil
}

func (self *RedisTransport) Publish(operationUUID string, fd int, time int64, event loxer.Event) error {
	e := loxer.SerializedEvent{event}
	msg := &Message{
		O:  operationUUID,
		T:  time,
		FD: fd,
		E:  e,
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	self.redsi.RPush(operationUUID, string(payload))
	return nil
}

// synthesize an EofEvent, signalling running consumers to close the connecion
func (self *RedisTransport) EOF(operationUUID string) error {
	return self.Publish(operationUUID, -1, time.Now().UnixNano(), loxer.EOF)
}

func (self *RedisTransport) Expire(operationUUID string) error {
	self.log.Debug().Msgf("expiring logs for %s", operationUUID)
	err := self.redsi.Expire(operationUUID, 1*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("self.redsi.expire(%s): %s", operationUUID, err)
	}
	return nil
}

func (self *RedisTransport) Close() error {
	var err error
	self.log.Debug().Msgf("closing redistransport")
	self.Lock()
	if self.terminator != nil {
		errors := make(chan error)
		self.terminator <- errors
		err = <-errors
	}
	self.Unlock()
	self.log.Debug().Msgf("redistransport closed, err=%#v", err)
	return err
}
