package projector

import (
	"fmt"
	"reflect"
	"sync"
)

type InMemoryIndex struct {
	lock *sync.RWMutex
	data map[string]reflect.Value
}

func NewInMemoryIndex() *InMemoryIndex {
	return &InMemoryIndex{
		lock: new(sync.RWMutex),
		data: map[string]reflect.Value{},
	}
}

func (self *InMemoryIndex) Update(do func(tx IndexTransaction) error) error {
	return do(self)
}

func (self *InMemoryIndex) Get(uuid string, dest interface{}) error {
	self.lock.RLock()
	defer self.lock.RUnlock()
	value, found := self.data[uuid]
	if !found {
		return fmt.Errorf("%T code=not_found key=%s", dest, uuid)
	}

	destValue := reflect.ValueOf(dest)
	if destValue.Kind() == reflect.Ptr {
		if value.Kind() == reflect.Ptr {
			destValue.Elem().Set(value.Elem())
		} else {
			destValue.Elem().Set(value)
		}
	} else {
		return fmt.Errorf("%s:%T: cannot load into non-pointer destination", uuid, dest)
	}

	return nil
}

func (self *InMemoryIndex) Put(uuid string, src interface{}) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	self.data[uuid] = reflect.ValueOf(src)
	return nil
}
