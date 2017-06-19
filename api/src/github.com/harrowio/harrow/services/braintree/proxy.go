package braintree

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	hhttp "github.com/harrowio/harrow/http"
	"github.com/harrowio/harrow/stores"
	braintreeAPI "github.com/lionelbarrow/braintree-go"
)

var (
	ErrCacheMiss     = errors.New("braintree: cache miss")
	ErrCacheInternal = errors.New("braintree: cache internal error")
)

type Cache interface {
	Clear()

	// Entries returns all cache entries and ErrCacheMiss if the
	// cache is currently empty.
	Entries() ([]*braintreeAPI.Plan, error)

	// Set sets the entries in the cache.
	Set(entries []*braintreeAPI.Plan) error
}

type Proxy struct {
	cache  Cache
	client stores.Braintree
}

func NewProxy(cache Cache, apiClient stores.Braintree) *Proxy {
	return &Proxy{
		cache:  cache,
		client: apiClient,
	}
}

func (self *Proxy) FindAllPlans() ([]*braintreeAPI.Plan, error) {
	plans, err := self.cache.Entries()
	if err == ErrCacheMiss {
		plans, err = self.client.FindAllPlans()
		if err != nil {
			return nil, err
		}

		if err := self.cache.Set(plans); err != nil {
			return nil, err
		}
	}

	return self.cache.Entries()
}

func (self *Proxy) ClearCache() {
	self.cache.Clear()
}

func (self *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	action := fmt.Sprintf("%s %s",
		req.Method,
		filepath.Clean(req.URL.Path),
	)

	switch action {
	case "DELETE /cache":
		self.ClearCache()
	case "GET /":
		plans, err := self.FindAllPlans()
		if err != nil {
			self.respondWithJSON(w, hhttp.NewError(http.StatusInternalServerError, "internal", err.Error()))
		} else {
			self.respondWithJSON(w, plans)
		}
	default:
		self.respondWithJSON(w, hhttp.ErrNotFound)
	}
}

func (self *Proxy) respondWithJSON(w http.ResponseWriter, thing interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(thing)
}
