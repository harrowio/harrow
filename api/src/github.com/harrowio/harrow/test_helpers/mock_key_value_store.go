package test_helpers

import (
	"errors"

	"github.com/harrowio/harrow/stores"
)

type MockKeyValueStore struct {
	data  map[string][]byte
	lists map[string][]string
}

type MockSecretKeyValueStore struct {
	data map[string][]byte
}

var errNotImplemented = errors.New("not implemented")

func (self *MockKeyValueStore) Get(key string) ([]byte, error) {
	if len(self.data[key]) == 0 {
		return nil, stores.ErrKeyNotFound
	}
	return self.data[key], nil
}

func (self *MockKeyValueStore) Exists(key string) (bool, error) {
	return len(self.data[key]) != 0, nil
}

func (self *MockKeyValueStore) Set(key string, data []byte) error {
	self.data[key] = data
	return nil
}

func (self *MockKeyValueStore) Del(key string) error {
	delete(self.data, key)
	return nil
}

func (self *MockKeyValueStore) LRange(key string, start, stop int64) ([]string, error) {
	return nil, errNotImplemented
}

func (self *MockKeyValueStore) RPush(key string, data string) error {
	self.lists[key] = append(self.lists[key], data)
	return nil
}

func (self *MockKeyValueStore) LPush(key string, data string) error {
	self.lists[key] = append([]string{data}, self.lists[key]...)
	return nil
}

func (self *MockKeyValueStore) Close() error {
	return errNotImplemented
}

func (self *MockKeyValueStore) SMembers(key string) ([]string, error) {
	return nil, errNotImplemented
}

func (self *MockKeyValueStore) SIsMember(key, member string) (bool, error) {
	return false, errNotImplemented
}

func (self *MockKeyValueStore) SAdd(key, member string) (int64, error) {
	return 0, errNotImplemented
}

func (self *MockKeyValueStore) SRem(key, member string) (int64, error) {
	return 0, errNotImplemented
}

func NewMockKeyValueStore() *MockKeyValueStore {
	return &MockKeyValueStore{
		data:  make(map[string][]byte),
		lists: map[string][]string{},
	}
}

func (self *MockSecretKeyValueStore) Get(key string, passphrase []byte) ([]byte, error) {
	if len(self.data[key]) == 0 {
		return nil, stores.ErrKeyNotFound
	}
	return self.data[key], nil
}

func (self *MockSecretKeyValueStore) Set(key string, passphrase, data []byte) error {
	self.data[key] = data
	return nil
}

func (self *MockSecretKeyValueStore) Del(key string) error {
	delete(self.data, key)
	return nil
}

func NewMockSecretKeyValueStore() *MockSecretKeyValueStore {
	return &MockSecretKeyValueStore{data: make(map[string][]byte)}
}
