package projector

import (
	"testing"
	"time"

	"github.com/harrowio/harrow/domain"
)

func TestIndexWithDefaults_Get_fills_in_default_values_for_destination(t *testing.T) {
	index := NewIndexWithDefaults(NewInMemoryIndex())
	index.SetDefault(&domain.Job{
		Uuid:      "774660bc-ba74-4d4f-b149-332a4b03161e",
		CreatedAt: time.Now(),
		Name:      "test",
	})
	job := &domain.Job{}
	index.Update(func(tx IndexTransaction) error {
		unknownJobID := "40f67fe6-8471-4c80-aa42-8bc6a8ba5363"
		if err := tx.Get(unknownJobID, job); err != nil {
			return err
		}
		return nil
	})

	if got, want := job.Name, "test"; got != want {
		t.Errorf(`job.Name = %v; want %v`, got, want)
	}
}

func TestIndexWithDefaults_Get_does_not_fill_in_default_if_entry_is_found(t *testing.T) {
	index := NewIndexWithDefaults(NewInMemoryIndex())
	index.SetDefault(&domain.Job{
		Uuid:      "774660bc-ba74-4d4f-b149-332a4b03161e",
		CreatedAt: time.Now(),
		Name:      "test",
	})
	jobID := "40f67fe6-8471-4c80-aa42-8bc6a8ba5363"
	existingJob := &domain.Job{Uuid: jobID, Name: "found"}
	job := &domain.Job{}
	index.Update(func(tx IndexTransaction) error {
		tx.Put(jobID, existingJob)
		if err := tx.Get(jobID, job); err != nil {
			return err
		}
		return nil
	})

	if got, want := job.Name, "found"; got != want {
		t.Errorf(`job.Name = %v; want %v`, got, want)
	}
}
