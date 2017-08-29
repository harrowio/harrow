package stores_test

import (
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	helpers "github.com/harrowio/harrow/test_helpers"
	"github.com/jmoiron/sqlx"
)

func setupPresentSecret(t *testing.T) (*sqlx.Tx, *stores.SecretStore, stores.SecretKeyValueStore, string, string, string) {
	tx := helpers.GetDbTx(t)
	sshSecret := &domain.SshSecret{PrivateKey: "priv", PublicKey: "pub"}
	ss := helpers.NewMockSecretKeyValueStore()
	world := helpers.MustNewWorldUsingSecretKeyValueStore(tx, ss, t)
	environmentUuid := world.Environment("default").Uuid
	secret, err := sshSecret.AsSecret()
	if err != nil {
		t.Fatal(err)
	}
	secret.Status = domain.SecretPresent
	secret.EnvironmentUuid = environmentUuid
	store := stores.NewSecretStore(ss, tx)
	sUuid, err := store.Create(secret)
	if err != nil {
		t.Fatal(err)
	}

	envSecret := &domain.EnvironmentSecret{
		Value: "foo",
	}
	secret, err = envSecret.AsSecret()
	if err != nil {
		t.Fatal(err)
	}
	secret.Status = domain.SecretPresent
	secret.EnvironmentUuid = environmentUuid
	eUuid, err := store.Create(secret)
	if err != nil {
		t.Fatal(err)
	}

	return tx, store, ss, sUuid, eUuid, environmentUuid
}

func setupPendingSecret(t *testing.T) (*sqlx.Tx, *stores.SecretStore, stores.SecretKeyValueStore, string, string) {
	tx := helpers.GetDbTx(t)
	ss := helpers.NewMockSecretKeyValueStore()
	world := helpers.MustNewWorldUsingSecretKeyValueStore(tx, ss, t)
	environmentUuid := world.Environment("default").Uuid
	secret := &domain.Secret{
		EnvironmentUuid: environmentUuid,
		Type:            domain.SecretSsh,
	}
	store := stores.NewSecretStore(ss, tx)
	uuid, err := store.Create(secret)
	if err != nil {
		t.Fatal(err)
	}
	return tx, store, ss, uuid, environmentUuid
}

func setupEnvSecret(t *testing.T) (*sqlx.Tx, *stores.SecretStore, stores.SecretKeyValueStore, string, string) {
	tx := helpers.GetDbTx(t)
	ss := helpers.NewMockSecretKeyValueStore()
	world := helpers.MustNewWorldUsingSecretKeyValueStore(tx, ss, t)
	environmentUuid := world.Environment("default").Uuid
	envSecret := &domain.EnvironmentSecret{
		Secret: &domain.Secret{
			EnvironmentUuid: environmentUuid,
			Type:            domain.SecretEnv,
		},
		Value: "foo",
	}
	store := stores.NewSecretStore(ss, tx)
	secret, err := envSecret.AsSecret()
	if err != nil {
		t.Fatal(err)
	}
	uuid, err := store.Create(secret)
	if err != nil {
		t.Fatal(err)
	}
	return tx, store, ss, uuid, environmentUuid
}

// Also tests Create implicitly due to the setup
func Test_SecretStore_FindSsh(t *testing.T) {
	tx, store, _, uuid, _, environmentUuid := setupPresentSecret(t)
	defer tx.Rollback()

	secret, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}
	if secret.Status != domain.SecretPresent {
		t.Fatalf("Find: Secret status should be %#v but was %#v\n", domain.SecretPresent, secret.Status)
	}

	sshSecret, err := domain.AsSshSecret(secret)
	if err != nil {
		t.Fatal(err)
	}
	if string(sshSecret.PublicKey) != "pub" {
		t.Fatal("Find: Restored public key doesn't match")
	}
	if string(sshSecret.PrivateKey) != "priv" {
		t.Fatal("Find: Restored private key doesn't match")
	}

	secrets, err := store.FindAllByEnvironmentUuid(environmentUuid)
	if err != nil {
		t.Fatal(err)
	}
	secret = secrets[0]
	if secret.Status != domain.SecretPresent {
		t.Fatalf("FindByEnvironmentUuid: Secret status should be %#v but was %#v\n", domain.SecretPresent, secret.Status)
	}

	sshSecret, err = domain.AsSshSecret(secret)
	if err != nil {
		t.Fatal(err)
	}
	if string(sshSecret.PublicKey) != "pub" {
		t.Fatal("FindByEnvironmentUuid: Restored public key doesn't match")
	}
	if string(sshSecret.PrivateKey) != "priv" {
		t.Fatal("FindByEnvironmentUuid: Restored private key doesn't match")
	}

	secrets, err = store.FindAll()
	if err != nil {
		t.Fatal(err)
	}

	// There is 1 Secret in world already
	if len(secrets) != 4 {
		t.Fatalf("FindAll: Wanted to find 4 records but found %d\n", len(secrets))
	}
}

// Also tests Create implicitly due to the setup
func Test_SecretStore_FindEnv(t *testing.T) {
	tx, store, _, _, uuid, _ := setupPresentSecret(t)
	defer tx.Rollback()

	secret, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}
	if secret.Status != domain.SecretPresent {
		t.Fatalf("Find: Secret status should be %#v but was %#v\n", domain.SecretPresent, secret.Status)
	}

	envSecret, err := domain.AsEnvironmentSecret(secret)
	if err != nil {
		t.Fatal(err)
	}

	if have, want := envSecret.Value, "foo"; have != want {
		t.Fatalf(`envSecret.Value == "%s", want "%s"`, have, want)
	}

}

// Also tests Create implicitly due to the setup
func Test_SecretStore_SshSecret_DefaultsToPending(t *testing.T) {
	tx, store, _, uuid, _ := setupPendingSecret(t)
	defer tx.Rollback()

	// should not fail although no secret bytes are present
	// because the status is pending
	secret, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}
	if !secret.IsPending() {
		t.Fatal("Find: Secret should be pending")
	}
}

// Also tests Create implicitly due to the setup
func Test_SecretStore_EnvSecret_DefaultsToPresent(t *testing.T) {
	tx, store, _, uuid, _ := setupEnvSecret(t)
	defer tx.Rollback()

	secret, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}
	if secret.IsPending() {
		t.Fatal("Find: EnvironmentSecret should be present")
	}
}

func Test_SecretStore_ArchiveByUuid(t *testing.T) {
	tx, store, ss, uuid, _, _ := setupPresentSecret(t)
	defer tx.Rollback()

	err := store.ArchiveByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.FindByUuid(uuid)
	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatalf("Expected to get a *domain.NotFoundError but got a %#v", err)
	}
	_, err = ss.Get(uuid, []byte{})
	if err != stores.ErrKeyNotFound {
		t.Fatalf("Expected to get stores.ErrKeyNotFound but got %#v", err)
	}
}

func Test_SecretStore_Update(t *testing.T) {
	tx, store, _, uuid, _, _ := setupPresentSecret(t)
	defer tx.Rollback()

	secret, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}
	secret.Name = "changed name"
	sshSecret, err := domain.AsSshSecret(secret)
	if err != nil {
		t.Fatal(err)
	}
	sshSecret.PublicKey = "changed"
	secret, err = sshSecret.AsSecret()
	if err != nil {
		t.Fatal(err)
	}
	err = store.Update(secret)
	if err != nil {
		t.Fatal(err)
	}

	// reload secret
	secret, err = store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}

	if secret.Name != "changed name" {
		t.Fatalf("Update: Secret name should be %#v but was %#v\n", "changed name", secret.Status)
	}

	sshSecret, err = domain.AsSshSecret(secret)
	if err != nil {
		t.Fatal(err)
	}
	if string(sshSecret.PublicKey) != "changed" {
		t.Fatal("Update: Restored public key doesn't match")
	}
	if string(sshSecret.PrivateKey) != "priv" {
		t.Fatal("Update: Restored private key doesn't match")
	}
}

func Test_SecretStore_Update_returnsDomainNotFoundError_whenUpdatingNonExistingSecret(t *testing.T) {
	tx, store, _, _, _, _ := setupPresentSecret(t)
	defer tx.Rollback()

	got := store.Update(&domain.Secret{})
	want, ok := got.(*domain.NotFoundError)
	if !ok {
		t.Errorf("err.(type) = %T; want %T", got, want)
	}
}
