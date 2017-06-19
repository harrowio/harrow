package stores_test

import (
	"testing"
	"time"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
)

func Test_JobStore_FindAllByUserUuid_returnsJobFromAllOfTheUsersProjects(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)

	projectUuids := []string{
		world.Project("public").Uuid,
		world.Project("private").Uuid,
	}

	expected := []*domain.Job{
		world.Job("default"),
		world.Job("other"),
	}

	store := stores.NewDbJobStore(tx)
	jobs, err := store.FindAllByProjectUuids(projectUuids)

	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(jobs), len(expected); got != want {
		t.Errorf("len(jobs) = %d; want %d", got, want)
	}
}

func Test_JobStore_FindAllByUserUuid_returnsEmptyListForEmptyList(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	projectUuids := []string{}

	expected := []*domain.Job{}

	store := stores.NewDbJobStore(tx)
	jobs, err := store.FindAllByProjectUuids(projectUuids)

	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(jobs), len(expected); got != want {
		t.Errorf("len(jobs) = %d; want %d", got, want)
	}
}

func Test_JobStore_Update_returnsDomainNotFoundError_whenUpdatingNonExistingJob(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbJobStore(tx)
	got := store.Update(&domain.Job{})
	want, ok := got.(*domain.NotFoundError)
	if !ok {
		t.Errorf("err.(type) = %T; want %T", got, want)
	}
}

func Test_JobStore_Update_updates_job(t *testing.T) {

	t.Skip("jobs_projects job.Name comes from the task name unless it starts with urn:")

	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)

	store := stores.NewDbJobStore(tx)

	job := *world.Job("default")
	job.Name = "changed"
	err := store.Update(&job)
	if err != nil {
		t.Errorf("err = %s; want %v", err, nil)
	}

	time.Sleep(500)

	reloaded, err := store.FindByUuid(job.Uuid)
	if err != nil {
		t.Errorf("err = %s; want %v", err, nil)
	}
	if reloaded.Name != "changed" {
		t.Errorf("reloaded.Name = %s; want %s", reloaded.Name, "changed")
	}
}
