package stores_test

import (
	"time"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"

	"testing"

	helpers "github.com/harrowio/harrow/test_helpers"
	"github.com/jmoiron/sqlx"
)

func setupRepositoryStoreTest(t *testing.T) (*sqlx.Tx, *domain.Organization, *domain.Project) {

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

func Test_RepositoryStoreStore_SucessfullyCreating(t *testing.T) {

	tx, _, p := setupRepositoryStoreTest(t)

	defer tx.Rollback()

	repo := &domain.Repository{
		Uuid:        "33333333-3333-4333-a333-333333333333",
		ProjectUuid: p.Uuid,
		Name:        "Git Core Tools",
		Url:         "git://git.kernel.org/pub/scm/git/git.git",
	}

	if uuid, err := stores.NewDbRepositoryStore(tx).Create(repo); err != nil {
		t.Fatal(err)
	} else {
		if uuid != repo.Uuid {
			t.Fatalf("Expected uuid == repo.Uuid, got %v == %v", uuid, repo.Uuid)
		}
	}

}

func Test_RepositoryStoreStore_FindingAllByProjectUuid(t *testing.T) {

	tx, _, p := setupRepositoryStoreTest(t)
	defer tx.Rollback()

	repo := helpers.MustCreateRepository(t, tx, &domain.Repository{
		Uuid:        "33333333-3333-4333-a333-333333333333",
		ProjectUuid: p.Uuid,
		Name:        "Git Core Tools",
		Url:         "git://git.kernel.org/pub/scm/git/git.git",
	})

	repos, err := stores.NewDbRepositoryStore(tx).FindAllByProjectUuid(p.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	for _, r := range repos {
		if r.Uuid == repo.Uuid {
			return
		}
	}

	t.Fatalf("Expected to find repo in the list, got: %v", repos)

}

func Test_RepositoryStore_Update_returnsDomainNotFoundError_whenUpdatingNonExistingRepository(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbRepositoryStore(tx)
	got := store.Update(&domain.Repository{})
	want, ok := got.(*domain.NotFoundError)
	if !ok {
		t.Errorf("err.(type) = %T; want %T", got, want)
	}
}

func Test_RepositoryStore_FindAllByProjectUuidAndRepositoryName_returnsRepositoriesWithTheGivenNameInTheURL(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()
	store := stores.NewDbRepositoryStore(tx)
	world := helpers.MustNewWorld(tx, t)
	project := world.Project("public")
	repositories := []*domain.Repository{
		project.NewRepository("A on GitHub", "git@github.com:example/repository.git"),
		project.NewRepository("A on BitBucket", "https://bitbucket.org/example/repository.git"),
		project.NewRepository("B on GitHub", "git@github.com:example/other-repository.git"),
	}

	for _, repository := range repositories {
		if _, err := store.Create(repository); err != nil {
			t.Fatal(err)
		}
	}

	found, err := store.FindAllByProjectUuidAndRepositoryName(project.Uuid, "example/repository")
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(found), 2; got != want {
		t.Errorf(`len(found) = %v; want %v`, got, want)
	}

	names := []string{}
	for _, repository := range found {
		names = append(names, repository.Name)
	}

	if want := repositories[2].Name; helpers.InStrings(want, names) {
		t.Errorf("expected %q to not be present", want)
	}
}

func Test_RepositoryStore_FindByOldestMetadata_onlyRespectsAccessibleRepositories(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()
	store := stores.NewDbRepositoryStore(tx)
	world := helpers.MustNewWorld(tx, t)
	project := world.Project("public")
	repositories := []*domain.Repository{
		project.NewRepository("A on GitHub", "git@github.com:example/repository.git"),
		project.NewRepository("A on BitBucket", "https://bitbucket.org/example/repository.git"),
		project.NewRepository("B on GitHub", "git@github.com:example/other-repository.git"),
	}

	for _, repository := range repositories {
		if _, err := store.Create(repository); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := tx.Exec(`UPDATE repositories SET accessible = false`); err != nil {
		t.Fatal(err)
	}

	if _, err := tx.Exec(`UPDATE repositories SET accessible = true WHERE uuid = $1`,
		repositories[2].Uuid); err != nil {
		t.Fatal(err)
	}

	repository, err := store.FindByOldestMetadata()
	if err != nil {
		t.Fatal(err)
	}

	if got := repository; got == nil {
		t.Fatalf("repository is nil")
	}

	if got, want := repository.Uuid, repositories[2].Uuid; got != want {
		t.Errorf(`repository.Uuid = %v; want %v`, got, want)
	}
}

func Test_RepositoryStore_FindByOldestMetadata_ignoresArchivedRepositories(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()
	store := stores.NewDbRepositoryStore(tx)
	world := helpers.MustNewWorld(tx, t)
	project := world.Project("public")
	repositories := []*domain.Repository{
		project.NewRepository("A on GitHub", "git@github.com:example/repository.git"),
		project.NewRepository("A on BitBucket", "https://bitbucket.org/example/repository.git"),
		project.NewRepository("B on GitHub", "git@github.com:example/other-repository.git"),
	}

	for _, repository := range repositories {
		if _, err := store.Create(repository); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := tx.Exec(`UPDATE repositories SET accessible = true, archived_at = NOW()`); err != nil {
		t.Fatal(err)
	}

	if _, err := tx.Exec(`UPDATE repositories SET archived_at = null WHERE uuid = $1`,
		repositories[2].Uuid); err != nil {
		t.Fatal(err)
	}

	repository, err := store.FindByOldestMetadata()
	if err != nil {
		t.Fatal(err)
	}

	if got := repository; got == nil {
		t.Fatalf("repository is nil")
	}

	if got, want := repository.Uuid, repositories[2].Uuid; got != want {
		t.Errorf(`repository.Uuid = %v; want %v`, got, want)
	}
}

func Test_RepositoryStore_FindByOldestMetadata_returnsRepositoryWithOldestMetadata(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()
	store := stores.NewDbRepositoryStore(tx)
	world := helpers.MustNewWorld(tx, t)
	project := world.Project("public")
	repositories := []*domain.Repository{
		project.NewRepository("A on GitHub", "git@github.com:example/repository.git"),
		project.NewRepository("A on BitBucket", "https://bitbucket.org/example/repository.git"),
		project.NewRepository("B on GitHub", "git@github.com:example/other-repository.git"),
	}

	now := time.Date(2015, 10, 27, 8, 43, 0, 0, time.UTC)

	if _, err := tx.Exec(`TRUNCATE repositories CASCADE`); err != nil {
		t.Fatal(err)
	}

	for i, repository := range repositories {
		if _, err := store.Create(repository); err != nil {
			t.Fatal(err)
		}
		q := `UPDATE repositories SET metadata_updated_at = $2 WHERE uuid = $1`
		metadataUpdatedAt := now.Add(-time.Duration(i) * time.Hour)
		if _, err := tx.Exec(q, repository.Uuid, metadataUpdatedAt); err != nil {
			t.Fatal(err)
		}
		// t.Logf("repository %q metadata_updated_at %s", repository.Uuid, metadataUpdatedAt)
	}

	if _, err := tx.Exec(`UPDATE repositories SET accessible = true, archived_at = null`); err != nil {
		t.Fatal(err)
	}

	repository, err := store.FindByOldestMetadata()
	if err != nil {
		t.Fatal(err)
	}

	if got := repository; got == nil {
		t.Fatalf("repository is nil")
	}

	if got := repository.MetadataUpdatedAt; got == nil {
		t.Fatalf("repository.MetadataUpdatedAt is nil")
	}

	if got, want := *repository.MetadataUpdatedAt, now.Add(-2*time.Hour); got != want {
		t.Errorf(`repository.MetadataUpdatedAt = %v; want %v`, got, want)
	}
}
