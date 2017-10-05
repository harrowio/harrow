package activityWorker

import (
	"sync"

	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
)

type ActivityWorker struct {
	handlers []func(activity.Message)
	source   activity.Source
	store    ActivityStore

	log logger.Logger

	running *sync.WaitGroup

	stop   chan error
	errors chan error
}

type ActivityStore interface {
	Store(activity *domain.Activity) error
}

type MemoryActivityStore struct {
	All   []*domain.Activity
	error error
}

func NewMemoryActivityStore() *MemoryActivityStore {
	return &MemoryActivityStore{
		All: []*domain.Activity{},
	}
}

func (self *MemoryActivityStore) Store(activity *domain.Activity) error {
	self.All = append(self.All, activity)
	return self.error
}

func (self *MemoryActivityStore) FailWith(err error) *MemoryActivityStore {
	self.error = err
	return self
}

type DbActivityStore struct {
	db *sqlx.DB
}

func NewDbActivityStore(db *sqlx.DB) *DbActivityStore {
	return &DbActivityStore{
		db: db,
	}
}

func (self *DbActivityStore) Store(activity *domain.Activity) error {
	tx, err := self.db.Beginx()
	defer tx.Rollback()

	if err != nil {
		return err
	}

	store := stores.NewDbActivityStore(tx)
	if err := store.Store(activity); err != nil {
		tx.Rollback()
		return err
	} else {
		return tx.Commit()
	}
}

func NewActivityWorker(source activity.Source, store ActivityStore) *ActivityWorker {
	return &ActivityWorker{
		source:   source,
		store:    store,
		running:  &sync.WaitGroup{},
		handlers: []func(activity.Message){},
		stop:     make(chan error),
		errors:   make(chan error, 1),
	}
}

func (self *ActivityWorker) SetLogger(l logger.Logger) {
	self.log = l
}

func (self *ActivityWorker) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *ActivityWorker) AddMessageHandler(handler func(activity.Message)) *ActivityWorker {

	self.Log().Debug().Msgf("adding message handler %v", handler)

	self.handlers = append(self.handlers, handler)

	return self
}

func (self *ActivityWorker) Start() error {

	self.Log().Debug().Msg("starting worker")

	src, err := self.source.Consume()
	if err != nil {
		return err
	}

	go self.processMessages(src)
	return nil
}

func (self *ActivityWorker) Stop() {
	self.stop <- nil
	self.Log().Info().Msgf("stopped")
}

func (self *ActivityWorker) processMessages(src chan activity.Message) {
	for {
		select {
		case <-self.stop:
			self.Log().Info().Msg("stopping")
			return
		case err := <-self.errors:
			self.Log().Error().Msgf("error: %s", err)
		case msg, more := <-src:
			if !more {
				self.Log().Info().Msg("no more on channel, returning")
				return
			}
			self.Log().Info().Msgf("processing message %s", msg.Activity())
			self.processMessage(msg)
		}
	}
}

func (self *ActivityWorker) processMessage(msg activity.Message) {

	self.Log().Debug().Msgf("applying %d handlers", len(self.handlers))

	for n, handler := range self.handlers {
		self.Log().Debug().Msgf("applying handler %d/%d", n, len(self.handlers))
		handler(msg)
		self.Log().Debug().Msgf("applying handler %d/%d ... done", n, len(self.handlers))
	}

	self.Log().Debug().Msgf("about to store activity from store %v", msg.Activity())
	if err := self.store.Store(msg.Activity()); err != nil {
		if _, ok := err.(*domain.ValidationError); ok {
			self.Log().Debug().Msgf("validation error, not interested in this.. skipping %s", err)
		} else {
			self.Log().Debug().Msgf("not a validation error, d'oh %s", err)
			self.errors <- err
		}
		return
	}

	self.Log().Debug().Msg("about to acknowledge message")
	if err := msg.Acknowledge(); err != nil {
		self.Log().Debug().Msgf("error acknowledging message", err)
		self.errors <- err
	}
	self.Log().Debug().Msg("acknowledged OK")
}
