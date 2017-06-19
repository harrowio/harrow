package stores

import (
	"encoding/json"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/jmoiron/sqlx"
)

type DbBillingHistoryStore struct {
	tx    *sqlx.Tx
	cache KeyValueStore
	log   logger.Logger
}

func NewDbBillingHistoryStore(tx *sqlx.Tx, cache KeyValueStore) *DbBillingHistoryStore {
	return &DbBillingHistoryStore{
		tx:    tx,
		cache: cache,
	}
}

func (self *DbBillingHistoryStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *DbBillingHistoryStore) SetLogger(l logger.Logger) {
	self.log = l
}

// Load returns the aggregation of all billing events in the form of a
// billing history.
func (self *DbBillingHistoryStore) Load() (*domain.BillingHistory, error) {

	history := self.LoadFromCache()
	err := NewDbBillingEventStore(self.tx).ReplayAllAfter(history.HandleEvent, history.Version)
	self.CacheHistory(history)
	return history, err
}

// LoadFromCache loads the cached version of the billing history.  If
// there is an error retrieving the cached version, a new empty
// instance is returned.
func (self *DbBillingHistoryStore) LoadFromCache() *domain.BillingHistory {
	history := domain.NewBillingHistory()
	data, err := self.cache.Get("billing-history")
	if err != nil {
		return history
	}

	if err := json.Unmarshal(data, history); err != nil {
		return domain.NewBillingHistory()
	}

	return history
}

// CacheHistory caches history in a key value store.  Errors are logged but
// not treated as fatal, because information can be recalculated in
// that case.
func (self *DbBillingHistoryStore) CacheHistory(history *domain.BillingHistory) {
	marshaled, err := json.Marshal(history)
	if err != nil {
		return
	}

	if err := self.cache.Set("billing-history", marshaled); err != nil {
	}
}
