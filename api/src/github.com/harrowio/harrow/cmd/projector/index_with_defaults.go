package projector

import (
	"reflect"
	"strings"
)

type IndexWithDefaults struct {
	wrappedIndex Index
	defaults     map[reflect.Type]reflect.Value
}

func NewIndexWithDefaults(wrapped Index) *IndexWithDefaults {
	return &IndexWithDefaults{
		wrappedIndex: wrapped,
		defaults:     map[reflect.Type]reflect.Value{},
	}
}

func (index *IndexWithDefaults) Update(handler func(tx IndexTransaction) error) error {
	return index.wrappedIndex.Update(func(tx IndexTransaction) error {
		tx = index.newTransaction(tx)
		return handler(tx)
	})
}

func (index *IndexWithDefaults) SetDefault(value interface{}) *IndexWithDefaults {
	kind := reflect.TypeOf(value)
	index.defaults[kind] = reflect.ValueOf(value)
	return index
}

func (index *IndexWithDefaults) newTransaction(tx IndexTransaction) IndexTransaction {
	return &IndexWithDefaultsTransaction{
		parent:    index,
		wrappedTx: tx,
	}
}

func (index *IndexWithDefaults) loadDefault(value interface{}) {
	kind := reflect.TypeOf(value)
	destination := reflect.ValueOf(value)
	defaultValue, found := index.defaults[kind]

	if !found {
		return
	}

	// Shallow copy of struct values:
	//
	//     *destination = *defaultValue
	reflect.Indirect(destination).Set(reflect.Indirect(defaultValue))
}

type IndexWithDefaultsTransaction struct {
	parent    *IndexWithDefaults
	wrappedTx IndexTransaction
}

func (tx *IndexWithDefaultsTransaction) Get(uuid string, dst interface{}) error {
	err := tx.wrappedTx.Get(uuid, dst)
	if err == nil {
		return nil
	}

	if !strings.Contains(err.Error(), "not_found") {
		return err
	}

	tx.parent.loadDefault(dst)

	return nil
}

func (tx *IndexWithDefaultsTransaction) Put(uuid string, src interface{}) error {
	return tx.wrappedTx.Put(uuid, src)
}
