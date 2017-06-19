package limits

import "github.com/harrowio/harrow/domain"

type DummyLimitsStore struct {
	byOrganizationId map[string]*Limits
}

func NewDummyLimitsStore() *DummyLimitsStore {
	return &DummyLimitsStore{
		byOrganizationId: map[string]*Limits{},
	}
}

func (self *DummyLimitsStore) Add(organizationUuid string, limits *Limits) *DummyLimitsStore {
	self.byOrganizationId[organizationUuid] = limits
	return self
}

func (self *DummyLimitsStore) FindByOrganizationUuid(organizationUuid string) (*Limits, error) {
	if limits, found := self.byOrganizationId[organizationUuid]; found {
		return limits, nil
	} else {
		return nil, new(domain.NotFoundError)
	}
}
