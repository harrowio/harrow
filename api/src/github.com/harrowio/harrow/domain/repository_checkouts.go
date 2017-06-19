package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type RepositoryCheckout struct {
	Ref  string
	Hash string
}

type RepositoryCheckouts struct {
	Refs map[string][]*RepositoryCheckout `json:"refs"`
}

func NewRepositoryCheckouts() *RepositoryCheckouts {
	return &RepositoryCheckouts{
		Refs: map[string][]*RepositoryCheckout{},
	}
}

func (self *RepositoryCheckouts) HandleEvent(payload EventPayload) {
	if payload.Get("event") != "checkout" {
		return
	}
	ref, hash := payload.Get("ref"), payload.Get("hash")
	repoUuid := payload.Get("repository")

	if ref == "" || hash == "" || repoUuid == "" {
		return
	}

	checkout := &RepositoryCheckout{
		Ref:  ref,
		Hash: hash,
	}

	self.Refs[repoUuid] = append(self.Refs[repoUuid], checkout)
}

func (self *RepositoryCheckouts) Ref(repositoryUuid string) string {
	checkouts := self.Refs[repositoryUuid]
	if len(checkouts) > 0 {
		return checkouts[len(checkouts)-1].Ref
	}
	return ""
}

func (self *RepositoryCheckouts) Hash(repositoryUuid string) string {
	checkouts := self.Refs[repositoryUuid]
	if len(checkouts) > 0 {
		return checkouts[len(checkouts)-1].Hash
	}
	return ""
}

func (self *RepositoryCheckouts) Value() (driver.Value, error) {
	data, err := json.Marshal(self)
	return data, err
}

func (self *RepositoryCheckouts) Scan(data interface{}) error {
	src := []byte{}
	switch raw := data.(type) {
	case []byte:
		src = raw
	default:
		return fmt.Errorf("RepositoryCheckouts: cannot scan from %T", data)
	}
	if err := json.Unmarshal(src, self); err != nil {
		return err
	}
	if self.Refs == nil {
		self.Refs = make(map[string][]*RepositoryCheckout)
	}
	return nil
}
