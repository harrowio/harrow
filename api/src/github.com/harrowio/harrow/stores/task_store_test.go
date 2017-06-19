package stores_test

import (
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	helpers "github.com/harrowio/harrow/test_helpers"
)

func Test_TaskStore_FindByJobUuid_ReturnsTask(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()
	world := helpers.MustNewWorld(tx, t)

	job := world.Job("default")
	store := stores.NewDbTaskStore(tx)
	task, err := store.FindByJobUuid(job.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if task.Uuid != job.TaskUuid {
		t.Fatalf("Wrong task returned: expected %q, got %q", job.TaskUuid, task.Uuid)
	}
}

func Test_TaskStore_FindByJobUuid_IgnoresArchivedJobs(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()
	world := helpers.MustNewWorld(tx, t)

	job := world.Job("default")
	store := stores.NewDbTaskStore(tx)

	jobStore := stores.NewDbJobStore(tx)
	if err := jobStore.ArchiveByUuid(job.Uuid); err != nil {
		t.Fatal(err)
	}

	_, err := store.FindByJobUuid(job.Uuid)
	if err == nil {
		t.Fatal("Expected an error")
	}

	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatalf("Expected *domain.NotFoundError, got %s", err)
	}
}

func Test_TaskStore_Update_returnsDomainNotFoundError_whenUpdatingNonExistingTask(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbTaskStore(tx)
	got := store.Update(&domain.Task{})
	want, ok := got.(*domain.NotFoundError)
	if !ok {
		t.Errorf("err.(type) = %T; want %T", got, want)
	}
}
