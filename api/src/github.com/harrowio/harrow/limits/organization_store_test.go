package limits

import "github.com/harrowio/harrow/domain"

type DummyOrganizationStore struct {
	byProjectId map[string]*domain.Organization
}

func NewDummyOrganizationStore() *DummyOrganizationStore {
	return &DummyOrganizationStore{
		byProjectId: map[string]*domain.Organization{},
	}
}

func (self *DummyOrganizationStore) Add(projectUuid string, organization *domain.Organization) *DummyOrganizationStore {
	self.byProjectId[projectUuid] = organization
	return self
}

// FindByProjectUuid returns the organization to which the project
// identified by projectUuid belongs.
func (self *DummyOrganizationStore) FindByProjectUuid(projectUuid string) (*domain.Organization, error) {
	if organization, found := self.byProjectId[projectUuid]; found {
		return organization, nil
	} else {
		return nil, new(domain.NotFoundError)
	}
}
