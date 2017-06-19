package stores_test

import (
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"

	"testing"

	helpers "github.com/harrowio/harrow/test_helpers"
	"github.com/jmoiron/sqlx"
)

func setupEnvironmentStoreTest(t *testing.T) (*sqlx.Tx, *domain.Organization, *domain.Project) {

	var err error

	tx := helpers.GetDbTx(t)

	o := &domain.Organization{Name: "Example Org"}
	if o.Uuid, err = stores.NewDbOrganizationStore(tx).Create(o); err != nil {
		t.Fatal(err)
	}

	p := &domain.Project{
		Name:             "Example Project",
		OrganizationUuid: o.Uuid,
	}
	if p.Uuid, err = stores.NewDbProjectStore(tx).Create(p); err != nil {
		t.Fatal(err)
	}
	return tx, o, p
}

func Test_EnvironmentStore_CreatingANewEnvironmentSuccessfully(t *testing.T) {

	tx, _, p := setupRepositoryStoreTest(t)

	defer tx.Rollback()

	store := stores.NewDbEnvironmentStore(tx)

	e := &domain.Environment{
		Name:        "Example Environment",
		ProjectUuid: p.Uuid,
		Variables: domain.EnvironmentVariables{
			M: map[string]string{"LC_CTYPE": "C"},
		},
	}

	uuid, err := store.Create(e)
	if err != nil {
		t.Fatal(err)
	}

	e, err = store.FindByUuid(uuid)
	if e == nil {
		t.Fatalf("Expected e == nil, got %v", e)
	}
	if len(e.Uuid) == 0 {
		t.Fatalf("Expected len(e.Uuid) == 0, got %v", len(e.Uuid))
	}
	if len(e.Variables.M) != 1 {
		t.Fatalf("Lost the Variables to hstore's chasm %d", len(e.Variables.M))
	}

}

func Test_EnvironmentStore_CreatingANewEnvironmentDropsEmptyVariables(t *testing.T) {

	tx, _, p := setupRepositoryStoreTest(t)

	defer tx.Rollback()

	store := stores.NewDbEnvironmentStore(tx)

	e := &domain.Environment{
		Name:        "Example Environment",
		ProjectUuid: p.Uuid,
		Variables: domain.EnvironmentVariables{
			M: map[string]string{"": ""},
		},
	}

	uuid, err := store.Create(e)
	if err != nil {
		t.Fatal(err)
	}

	e, err = store.FindByUuid(uuid)
	if e == nil {
		t.Fatalf("Expected e == nil, got %v", e)
	}
	if len(e.Uuid) == 0 {
		t.Fatalf("Expected len(e.Uuid) == 0, got %v", len(e.Uuid))
	}
	if len(e.Variables.M) != 0 {
		t.Fatalf("Expected to not recieve any variables back, got %d", len(e.Variables.M))
	}

}

func Test_EnvironmentStore_FindByJobUuid_returnsEnvironment(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()
	world := helpers.MustNewWorld(tx, t)
	job := world.Job("default")
	store := stores.NewDbEnvironmentStore(tx)
	env, err := store.FindByJobUuid(job.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if env.Uuid != job.EnvironmentUuid {
		t.Fatalf("Wrong environment returned: %q, expected %q", env.Uuid, job.EnvironmentUuid)
	}
}

func Test_EnvironmentStore_FindByJobUuid_returnsNilForArchivedJob(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()
	world := helpers.MustNewWorld(tx, t)

	job := world.Job("default")
	jobStore := stores.NewDbJobStore(tx)
	jobStore.ArchiveByUuid(job.Uuid)

	store := stores.NewDbEnvironmentStore(tx)

	env, err := store.FindByJobUuid(job.Uuid)
	if env != nil {
		t.Fatalf("Expected no environment to be returned, got: %#v", env)
	}

	if err == nil {
		t.Fatal("Expected an error")
	}

	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatalf("Expected a *domain.NotFoundError, got %#v", err)
	}
}

func Test_EnvironmentStore_Update_returnsDomainNotFoundError_ifEnvironmentDoesNotExist(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbEnvironmentStore(tx)
	got := store.Update(&domain.Environment{})
	want, ok := got.(*domain.NotFoundError)
	if !ok {
		t.Errorf("err.(type) = %T; want %T", got, want)
	}
}
