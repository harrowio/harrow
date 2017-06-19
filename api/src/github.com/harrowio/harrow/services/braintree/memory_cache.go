package braintree

import braintreeAPI "github.com/lionelbarrow/braintree-go"

type MemoryCache struct {
	entries []*braintreeAPI.Plan
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		entries: nil,
	}
}

func (self *MemoryCache) Clear() {
	self.entries = nil
}

func (self *MemoryCache) Entries() ([]*braintreeAPI.Plan, error) {
	if self.entries == nil {
		return nil, ErrCacheMiss
	}

	return self.entries, nil
}

func (self *MemoryCache) Set(entries []*braintreeAPI.Plan) error {
	self.entries = entries
	return nil
}
