package domain

import "testing"

type mockProjectStore struct {
	byId          map[string]*Project
	byWebhookUuid map[string]*Project
	callsByEnv    []string
}

func (self *mockProjectStore) FindByUuid(uuid string) (*Project, error) {
	project, found := self.byId[uuid]
	if !found {
		return nil, new(NotFoundError)
	}
	return project, nil
}

func (self *mockProjectStore) FindByEnvironmentUuid(environmentUuid string) (*Project, error) {
	if self.callsByEnv == nil {
		self.callsByEnv = make([]string, 0)
	}
	self.callsByEnv = append(self.callsByEnv, environmentUuid)
	return &Project{}, nil
}

func (self *mockProjectStore) FindByJobUuid(uuid string) (*Project, error) {
	return &Project{}, nil
}

func (self *mockProjectStore) FindByMemberUuid(uuid string) (*Project, error) {
	return &Project{}, nil
}

func (self *mockProjectStore) FindByRepositoryUuid(uuid string) (*Project, error) {
	return &Project{}, nil
}

func (self *mockProjectStore) FindByOrganizationUuid(uuid string) (*Project, error) {
	return &Project{}, nil
}

func (self *mockProjectStore) FindByNotifierUuid(uuid, notifierType string) (*Project, error) {
	return &Project{}, nil
}

func (self *mockProjectStore) FindByTaskUuid(uuid string) (*Project, error) {
	return &Project{}, nil
}

func (self *mockProjectStore) FindByWebhookUuid(uuid string) (*Project, error) {
	project, found := self.byWebhookUuid[uuid]
	if found {
		return project, nil
	} else {
		return nil, new(NotFoundError)
	}
}

func (self *mockProjectStore) FindByNotificationRule(notifierType string, notifierUuid string) (*Project, error) {
	return &Project{}, nil
}

func Test_RandomTotpSecret_returns16CharacterLongString(t *testing.T) {
	// 16 charactes is what is expected by authentication devices
	secret := RandomTotpSecret()
	if got, want := len(secret), 16; got != want {
		t.Fatalf("len(%q) = %d; want = %d", secret, got, want)
	}
}
