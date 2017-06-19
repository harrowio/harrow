package ws

import (
	"fmt"

	redis "gopkg.in/redis.v2"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/logger"

	"github.com/harrowio/harrow/bus/logevent"
)

type dualSource struct {
	fSource logevent.Source
	rSource logevent.Source
	log     logger.Logger
}

// A logevent.Source that first tries FileTransport, and uses RedisTransport
// on failure
func NewDualSource(redsi *redis.Client, c *config.Config) *dualSource {
	s := new(dualSource)
	s.rSource = logevent.NewRedisTransport(redsi, s.Log())
	s.fSource = logevent.NewFileTransport(c, s.Log())
	return s
}

func (store *dualSource) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *dualSource) SetLogger(l logger.Logger) {
	store.log = l
}

func (self *dualSource) Consume(operationUUID string) (<-chan *logevent.Message, error) {
	msgs, err := self.fSource.Consume(operationUUID)
	if err != nil {
		msgs, err = self.rSource.Consume(operationUUID)
	}
	return msgs, err
}

func (self *dualSource) Close() error {
	err := self.fSource.Close()
	if err != nil {
		self.log.Error().Msgf("self.fSource.Close(): %s", err)
	}
	err = self.rSource.Close()
	if err != nil {
		return fmt.Errorf("self.rSource.Close(): %s", err)
	}
	return nil
}
