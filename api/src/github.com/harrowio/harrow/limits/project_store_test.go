package limits

import "github.com/harrowio/harrow/domain"

type DummyProjectStore struct {
	byId map[string]*domain.Project
}

func NewDummyProjectStore() *DummyProjectStore {
	return &DummyProjectStore{
		byId: map[string]*domain.Project{},
	}
}

func (self *DummyProjectStore) Add(project *domain.Project) *DummyProjectStore {
	self.byId[project.Uuid] = project
	return self
}

func (self *DummyProjectStore) FindByUuid(uuid string) (*domain.Project, error) {
	project, found := self.byId[uuid]
	if !found {
		return nil, new(domain.NotFoundError)
	}

	return project, nil
}

func (self *DummyProjectStore) FindByMemberUuid(uuid string) (*domain.Project, error) {
	panic("not implemented")
}

func (self *DummyProjectStore) FindByOrganizationUuid(uuid string) (*domain.Project, error) {
	panic("not implemented")
}

func (self *DummyProjectStore) FindByNotifierUuid(uuid string, notifierType string) (*domain.Project, error) {
	panic("not implemented")
}

func (self *DummyProjectStore) FindByJobUuid(uuid string) (*domain.Project, error) {
	panic("not implemented")
}

func (self *DummyProjectStore) FindByTaskUuid(uuid string) (*domain.Project, error) {
	panic("not implemented")
}

func (self *DummyProjectStore) FindByRepositoryUuid(uuid string) (*domain.Project, error) {
	panic("not implemented")
}

func (self *DummyProjectStore) FindByEnvironmentUuid(uuid string) (*domain.Project, error) {
	panic("not implemented")
}

func (self *DummyProjectStore) FindByWebhookUuid(uuid string) (*domain.Project, error) {
	panic("not implemented")
}

func (self *DummyProjectStore) FindByNotificationRule(notifierType string, notifierUuid string) (*domain.Project, error) {
	panic("not implemented")
}
