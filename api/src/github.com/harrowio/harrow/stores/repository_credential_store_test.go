package stores_test

import (
	"fmt"
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	helpers "github.com/harrowio/harrow/test_helpers"
	"github.com/jmoiron/sqlx"
)

func setupRepositoryCredentialStoreTest(t *testing.T) (*sqlx.Tx, *stores.RepositoryCredentialStore, stores.SecretKeyValueStore, string, string) {
	tx := helpers.GetDbTx(t)
	ss := helpers.NewMockSecretKeyValueStore()
	world := helpers.MustNewWorldUsingSecretKeyValueStore(tx, ss, t)
	project := world.Project("public")
	repo := helpers.MustCreateRepository(t, tx, &domain.Repository{
		Name:        "test-repo",
		ProjectUuid: project.Uuid,
	})
	store := stores.NewRepositoryCredentialStore(ss, tx)
	rc := &domain.RepositoryCredential{
		Name:           fmt.Sprintf("repository-%s", repo.Uuid),
		RepositoryUuid: repo.Uuid,
		Type:           domain.RepositoryCredentialSsh,
		Status:         domain.RepositoryCredentialPresent,
		SecretBytes:    []byte(`{"PublicKey":"pub","PrivateKey":"priv"}`),
		Key:            []byte("abcd1"),
	}

	helpers.MustCreateRepositoryCredential(t, tx, ss, rc)

	return tx, store, ss, rc.Uuid, rc.RepositoryUuid
}

// Also tests Create implicitly due to the setup
func Test_RepositoryCredentialStore_Find(t *testing.T) {
	tx, store, _, uuid, _ := setupRepositoryCredentialStoreTest(t)
	defer tx.Rollback()

	repositoryCredential, err := store.FindByUuid(uuid)

	if err != nil {
		t.Fatal(err)
	}

	sshRepositoryCredential, err := domain.AsSshRepositoryCredential(repositoryCredential)
	if err != nil {
		t.Fatal(err)
	}
	if string(sshRepositoryCredential.PublicKey) != "pub" {
		t.Fatal("Find: Restored public key doesn't match")
	}
	if string(sshRepositoryCredential.PrivateKey) != "priv" {
		t.Fatal("Find: Restored private key doesn't match")
	}
}

func Test_RepositoryCredentialStore_FindByRepositoryUuid(t *testing.T) {
	tx, store, _, _, repositoryUuid := setupRepositoryCredentialStoreTest(t)
	defer tx.Rollback()

	repositoryCredential, err := store.FindByRepositoryUuid(repositoryUuid)
	if err != nil {
		t.Fatal(err)
	}

	sshRepositoryCredential, err := domain.AsSshRepositoryCredential(repositoryCredential)
	if err != nil {
		t.Fatal(err)
	}
	if string(sshRepositoryCredential.PublicKey) != "pub" {
		t.Fatal("FindByRepositoryUuid: Restored public key doesn't match")
	}
	if string(sshRepositoryCredential.PrivateKey) != "priv" {
		t.Fatal("FindByRepositoryUuid: Restored private key doesn't match")
	}
}

func Test_RepositoryCredentialStore_Update(t *testing.T) {
	tx, store, _, uuid, _ := setupRepositoryCredentialStoreTest(t)
	defer tx.Rollback()

	rc, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}
	rc.Name = "changed name"
	sshRc, err := domain.AsSshRepositoryCredential(rc)
	if err != nil {
		t.Fatal(err)
	}
	sshRc.PublicKey = "changed"
	rc, err = sshRc.AsRepositoryCredential()
	if err != nil {
		t.Fatal(err)
	}
	err = store.Update(rc)
	if err != nil {
		t.Fatal(err)
	}

	// reload secret
	rc, err = store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}

	if rc.Name != "changed name" {
		t.Fatalf("Update: name status should be %#v but was %#v\n", "changed name", rc.Name)
	}

	sshRc, err = domain.AsSshRepositoryCredential(rc)
	if err != nil {
		t.Fatal(err)
	}
	if string(sshRc.PublicKey) != "changed" {
		t.Fatal("Update: Restored public key doesn't match")
	}
	if string(sshRc.PrivateKey) != "priv" {
		t.Fatal("Update: Restored private key doesn't match")
	}
}

func Test_RepositoryCredentialStore_Update_returnsDomainNotFoundError_whenUpdatingNonExistingRepositoryCredential(t *testing.T) {
	tx, store, _, _, _ := setupRepositoryCredentialStoreTest(t)
	defer tx.Rollback()

	got := store.Update(&domain.RepositoryCredential{})
	want, ok := got.(*domain.NotFoundError)
	if !ok {
		t.Errorf("err.(type) = %T; want %T", got, want)
	}
}
