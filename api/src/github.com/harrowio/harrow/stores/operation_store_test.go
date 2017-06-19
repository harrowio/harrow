package stores_test

import (
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"

	"testing"
	"time"

	helpers "github.com/harrowio/harrow/test_helpers"

	"github.com/jmoiron/sqlx"
)

type TestParams struct {
	tx           *sqlx.Tx
	organization *domain.Organization
	project      *domain.Project
	repository   *domain.Repository
	job          *domain.Job
	environment  *domain.Environment
	task         *domain.Task
}

func setupOperationStoreTest(t *testing.T) *TestParams {

	tx := helpers.GetDbTx(t)

	organization := helpers.MustCreateOrganization(t, tx, &domain.Organization{Name: "Example Org"})
	project := helpers.MustCreateProject(t, tx, &domain.Project{Name: "Example Project", OrganizationUuid: organization.Uuid})
	repository := helpers.MustCreateRepository(t, tx, &domain.Repository{ProjectUuid: project.Uuid, Name: "Git Core Tools", Url: "git://git.kernel.org/pub/scm/git/git.git"})
	task := helpers.MustCreateTask(t, tx, &domain.Task{ProjectUuid: project.Uuid, Name: "Example Task", Type: domain.TaskTypeTest})
	env := helpers.MustCreateEnvironment(t, tx, &domain.Environment{
		ProjectUuid: project.Uuid,
		Variables: domain.EnvironmentVariables{
			M: map[string]string{
				"FOO": "BAR",
			},
		},
	})

	job := helpers.MustCreateJob(t, tx, &domain.Job{
		Name:            "Example Job",
		EnvironmentUuid: env.Uuid,
		TaskUuid:        task.Uuid,
	})

	return &TestParams{
		tx:           tx,
		organization: organization,
		task:         task,
		project:      project,
		repository:   repository,
		job:          job,
		environment:  env,
	}
}

func (self *TestParams) newOperationWithTimestamps(t *testing.T, created, finished time.Time) *domain.Operation {
	operationStore := stores.NewDbOperationStore(self.tx)
	operation := &domain.Operation{
		WorkspaceBaseImageUuid: "31b0127a-6d63-4d22-b32b-e1cfc04f4007",
		JobUuid:                &self.job.Uuid,
		Type:                   domain.OperationTypeJobScheduled,
		Uuid:                   uuidhelper.MustNewV4(),
	}
	if _, err := operationStore.Create(operation); err != nil {
		t.Fatal(err)
	}

	if _, err := self.tx.Exec(`UPDATE operations SET finished_at = $2 WHERE uuid = $1`, operation.Uuid, finished); err != nil {
		t.Fatal(err)
	}

	if _, err := self.tx.Exec(`UPDATE operations SET created_at = $2 WHERE uuid = $1`, operation.Uuid, created); err != nil {
		t.Fatal(err)
	}

	return operation
}

func Test_OperationStoreStore_SucessfullyCreating(t *testing.T) {

	test := setupOperationStoreTest(t)
	defer test.tx.Rollback()

	op := &domain.Operation{
		Uuid:           "44444444-3333-4333-a333-333333333333",
		RepositoryUuid: &test.repository.Uuid,
		Type:           domain.OperationTypeGitAccessCheck,
		WorkspaceBaseImageUuid: "31b0127a-6d63-4d22-b32b-e1cfc04f4007",
		TimeLimit:              900,
	}

	savedOp := helpers.MustCreateOperation(t, test.tx, op)

	if savedOp.Uuid != op.Uuid {
		t.Fatalf("Expected savedOp.Uuid != op.Uuid, got %v != %v", savedOp.Uuid, op.Uuid)
	}

}

func Test_OperationStoreStore_SucessfullySavingJobUuid(t *testing.T) {

	test := setupOperationStoreTest(t)
	defer test.tx.Rollback()
	jobUuid := test.job.Uuid

	op := &domain.Operation{
		// Uuid:                   "54444444-3333-4333-a333-333333333333",
		Type:                   domain.OperationTypeJobScheduled,
		JobUuid:                &jobUuid,
		WorkspaceBaseImageUuid: "31b0127a-6d63-4d22-b32b-e1cfc04f4007",
	}
	// t.Logf("JobUuid: %s\n*op.JobUuid: %s\n", jobUuid, test.job.Uuid)
	savedOp := helpers.MustCreateOperation(t, test.tx, op)

	if *savedOp.JobUuid != test.job.Uuid {
		t.Fatalf("Expected savedOp.JobUuid != test.job.Uuid, got %v != %v", *savedOp.JobUuid, test.job.Uuid)
	}
}

func Test_OperationStoreFindByUuid_ReadsFatalError(t *testing.T) {
	test := setupOperationStoreTest(t)
	defer test.tx.Rollback()

	e := "FAT ALF"
	op := &domain.Operation{
		Uuid:           "44444444-3333-4333-a333-333333333333",
		RepositoryUuid: &test.repository.Uuid,
		Type:           domain.OperationTypeGitAccessCheck,
		WorkspaceBaseImageUuid: "31b0127a-6d63-4d22-b32b-e1cfc04f4007",
		TimeLimit:              900,
	}

	store := stores.NewDbOperationStore(test.tx)
	if id, err := store.Create(op); err != nil {
		t.Fatal(err)
	} else {
		test.tx.MustExec(`UPDATE operations SET fatal_error = $2 WHERE uuid = $1;`, id, e)
		found, err := store.FindByUuid(id)
		if err != nil {
			t.Fatal(err)
		}

		if found.FatalError == nil || *found.FatalError != e {
			t.Fatalf("FatalError != %q: %q\n", e, found.FatalError)
		}
	}
}

func Test_OperationStore_FindAllByRepositoryUuid_SkipsArchived(t *testing.T) {
	test := setupOperationStoreTest(t)
	defer test.tx.Rollback()

	repoStore := stores.NewDbRepositoryStore(test.tx)
	repoStore.ArchiveByUuid(test.repository.Uuid)
	store := stores.NewDbOperationStore(test.tx)

	// should not find any Operation because the Repository was archived
	ops, err := store.FindAllByRepositoryUuid(test.repository.Uuid)
	if err != nil {
		t.Fatal(err)
	}
	if len(ops) != 0 {
		t.Errorf("len(ops)=%d, want %d", len(ops), 0)
	}
}

func Test_OperationStore_FindAllByProjectUuid_SkipsArchivedJobs(t *testing.T) {
	test := setupOperationStoreTest(t)
	defer test.tx.Rollback()

	jobStore := stores.NewDbJobStore(test.tx)
	jobStore.ArchiveByUuid(test.job.Uuid)

	store := stores.NewDbOperationStore(test.tx)
	// should not find any Operation because the Job of this Projetc was archived
	ops, err := store.FindAllByProjectUuid(test.project.Uuid)
	if err != nil {
		t.Fatal(err)
	}
	if len(ops) != 0 {
		t.Errorf("len(ops)=%d, want %d", len(ops), 0)
	}
}

func Test_OperationStore_FindAllByJobUuid_SkipsArchived(t *testing.T) {
	test := setupOperationStoreTest(t)
	defer test.tx.Rollback()

	jobStore := stores.NewDbJobStore(test.tx)
	jobStore.ArchiveByUuid(test.job.Uuid)

	store := stores.NewDbOperationStore(test.tx)
	// should not find any Operation because the Job of this Projetc was archived
	ops, err := store.FindAllByJobUuid(test.job.Uuid)
	if err != nil {
		t.Fatal(err)
	}
	if len(ops) != 0 {
		t.Errorf("len(ops)=%d, want %d", len(ops), 0)
	}
}

func TestDbOperationStore_FindPreviousOperation_returnsMostRecentFinishedOperationForSameJob(t *testing.T) {
	test := setupOperationStoreTest(t)
	defer test.tx.Rollback()
	operationStore := stores.NewDbOperationStore(test.tx)
	now := time.Now()
	previous := test.newOperationWithTimestamps(
		t,
		now.Add(-2*time.Hour),
		now.Add(-2*time.Hour).Add(1*time.Minute),
	)
	current := test.newOperationWithTimestamps(
		t,
		now.Add(-2*time.Minute),
		now,
	)

	found, err := operationStore.FindPreviousOperation(current.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got := found; got == nil {
		t.Fatalf("found is nil")
	}

	if got, want := found.Uuid, previous.Uuid; got != want {
		t.Errorf(`found.Uuid = %v; want %v`, got, want)
	}
}

func TestDbOperationStore_FindPreviousOperation_returnsNilAndNoErrorIfThereIsNoPreviousOperation(t *testing.T) {
	test := setupOperationStoreTest(t)
	defer test.tx.Rollback()
	operationStore := stores.NewDbOperationStore(test.tx)
	now := time.Now()
	current := test.newOperationWithTimestamps(
		t,
		now.Add(-2*time.Minute),
		now,
	)

	found, err := operationStore.FindPreviousOperation(current.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := found, (*domain.Operation)(nil); got != want {
		t.Errorf(`found = %v; want %v`, got, want)
	}

}

func TestDbOperationStore_FindPreviousOperation_ignoresUnfinishedOperationsInTheSameTimeRange(t *testing.T) {
	test := setupOperationStoreTest(t)
	defer test.tx.Rollback()
	operationStore := stores.NewDbOperationStore(test.tx)
	now := time.Now()
	previous := test.newOperationWithTimestamps(
		t,
		now.Add(-2*time.Hour),
		now.Add(-2*time.Hour).Add(1*time.Minute),
	)
	unfinished := test.newOperationWithTimestamps(
		t,
		now.Add(-1*time.Hour),
		now.Add(-1*time.Hour),
	)
	current := test.newOperationWithTimestamps(
		t,
		now.Add(-2*time.Minute),
		now,
	)

	if _, err := test.tx.Exec(`UPDATE operations SET finished_at = NULL WHERE uuid = $1`, unfinished.Uuid); err != nil {
		t.Fatal(err)
	}

	found, err := operationStore.FindPreviousOperation(current.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got := found; got == nil {
		t.Fatalf("found is nil")
	}

	if got, want := found.Uuid, previous.Uuid; got != want {
		t.Errorf(`found.Uuid = %v; want %v`, got, want)
	}
}

func TestDbOperationStore_FindPreviousOperation_ignoresArchivedOperationsInTheSameTimeRange(t *testing.T) {
	test := setupOperationStoreTest(t)
	defer test.tx.Rollback()
	operationStore := stores.NewDbOperationStore(test.tx)
	now := time.Now()
	previous := test.newOperationWithTimestamps(
		t,
		now.Add(-2*time.Hour),
		now.Add(-2*time.Hour).Add(1*time.Minute),
	)
	archived := test.newOperationWithTimestamps(
		t,
		now.Add(-1*time.Hour),
		now.Add(-1*time.Hour),
	)
	current := test.newOperationWithTimestamps(
		t,
		now.Add(-2*time.Minute),
		now,
	)

	if _, err := test.tx.Exec(`UPDATE operations SET archived_at = NOW() WHERE uuid = $1`, archived.Uuid); err != nil {
		t.Fatal(err)
	}

	found, err := operationStore.FindPreviousOperation(current.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got := found; got == nil {
		t.Fatalf("found is nil")
	}

	if got, want := found.Uuid, previous.Uuid; got != want {
		t.Errorf(`found.Uuid = %v; want %v`, got, want)
	}
}
