package activityWorker

import (
	"sync"

	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type ActivityWorker struct {
	handlers []func(activity.Message)
	source   activity.Source
	store    ActivityStore

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

func (self *ActivityWorker) AddMessageHandler(handler func(activity.Message)) *ActivityWorker {
	self.handlers = append(self.handlers, handler)

	return self
}

func (self *ActivityWorker) Start() error {
	src, err := self.source.Consume()
	if err != nil {
		return err
	}

	go self.processMessages(src)
	return nil
}

func (self *ActivityWorker) Stop() {
	self.stop <- nil
	log.Info().Msgf("stopped")
}

func (self *ActivityWorker) processMessages(src chan activity.Message) {
	for {
		select {
		case <-self.stop:
			log.Info().Msgf("stopping")
			return
		case err := <-self.errors:
			log.Error().Msgf("error: %s", err)
		case msg, more := <-src:
			if !more {
				return
			}
			self.processMessage(msg)
		}
	}
}

func (self *ActivityWorker) processMessage(msg activity.Message) {
	for _, handler := range self.handlers {
		handler(msg)
	}

	if err := self.store.Store(msg.Activity()); err != nil {
		self.errors <- err
		return
	}

	if err := msg.Acknowledge(); err != nil {
		self.errors <- err
	}
}
