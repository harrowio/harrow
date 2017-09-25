package projector

import (
	"testing"

	"github.com/harrowio/harrow/domain"
)

type IndexTest struct {
	makeIndex func() Index
}

func NewIndexTest(makeIndex func() Index) *IndexTest {
	return &IndexTest{
		makeIndex: makeIndex,
	}
}

func (self *IndexTest) Run(t *testing.T) {
	self.test_Put_adds_object_that_can_be_returned_by_find(t)
}

func (self *IndexTest) test_Put_adds_object_that_can_be_returned_by_find(t *testing.T) {
	err := self.makeIndex().Update(func(tx IndexTransaction) error {
		toSave := &domain.Job{
			Uuid: "b0fb288c-41a0-409b-85a0-1b9d661effe7",
			Name: "foo",
		}

		if err := tx.Put(toSave.Uuid, toSave); err != nil {
			t.Fatal(err)
		}

		toLoad := &domain.Job{}
		if err := tx.Get(toSave.Uuid, toLoad); err != nil {
			t.Fatal(err)
		}

		if got, want := toLoad.Uuid, toSave.Uuid; got != want {
			t.Errorf(`toLoad.Uuid = %v; want %v`, got, want)
		}

		if got, want := toLoad.Name, toSave.Name; got != want {
			t.Errorf(`toLoad.Name = %v; want %v`, got, want)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestInMemoryIndex(t *testing.T) {
	NewIndexTest(func() Index { return NewInMemoryIndex() }).Run(t)
}

func TestIndexWithDefaults(t *testing.T) {
	NewIndexTest(func() Index { return NewIndexWithDefaults(NewInMemoryIndex()) }).Run(t)
}
