package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"

	"github.com/gorilla/mux"
)

func Test_SecretHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountSecretHandler(r, nil)

	spec := routingSpec{
		{"POST", "/secrets", "secret-create"},
		{"GET", "/secrets/:uuid", "secret-show"},
		{"DELETE", "/secrets/:uuid", "secret-archive"},
	}

	spec.run(r, t)
}

func Test_SecretHandler_Create_AsOwner(t *testing.T) {
	ts, ctxt := setupAuthzHandlerTestServer(MountSecretHandler, t)
	tx := ctxt.Tx()

	world := test_helpers.MustNewWorldUsingSecretKeyValueStore(tx, ctxt.SecretKeyValueStore(), t)
	// project member
	u := world.User("project-owner")
	e := world.Environment("private")
	s := setupTestLoginSession(t, tx, u)

	defer tx.Rollback()
	defer ts.Close()
	req, err := newAuthenticatedRequestJSON(s.Uuid, "POST", ts.URL+"/secrets",
		&halWrapper{
			Subject: &secretParams{
				Name:            "testik",
				EnvironmentUuid: e.Uuid,
				Type:            domain.SecretSsh,
			},
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 201 {
		t.Fatalf("res.StatusCode=%d, want %d", res.StatusCode, 201)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	result := struct {
		Subject *domain.Secret `json:"subject"`
	}{}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}

	secretStore := stores.NewSecretStore(ctxt.SecretKeyValueStore(), tx)
	secret, err := secretStore.FindByUuid(result.Subject.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if secret.Name != "testik" {
		t.Fatalf("secret.Name=%s, want %s", secret.Name, "testik")
	}

	if !secret.IsSsh() {
		t.Fatal("secret should be a ssh secret")
	}

	if !secret.IsPending() {
		t.Fatal("secret should be pending")
	}
}

func Test_SecretHandler_Create_AsNonMember(t *testing.T) {
	ts, ctxt := setupAuthzHandlerTestServer(MountSecretHandler, t)
	tx := ctxt.Tx()

	world := test_helpers.MustNewWorldUsingSecretKeyValueStore(tx, ctxt.SecretKeyValueStore(), t)
	// not a project member
	u := world.User("non-member")
	e := world.Environment("private")
	s := setupTestLoginSession(t, tx, u)

	defer tx.Rollback()
	defer ts.Close()
	req, err := newAuthenticatedRequestJSON(s.Uuid, "POST", ts.URL+"/secrets",
		&halWrapper{
			Subject: &secretParams{
				Name:            "testik",
				EnvironmentUuid: e.Uuid,
				Type:            domain.SecretSsh,
			},
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 403 {
		t.Fatalf("res.StatusCode=%d, want %d", res.StatusCode, 403)
	}
}

func Test_SecretHandler_Show_AsNonOwner(t *testing.T) {
	ts, ctxt := setupAuthzHandlerTestServer(MountSecretHandler, t)
	tx := ctxt.Tx()

	world := test_helpers.MustNewWorldUsingSecretKeyValueStore(tx, ctxt.SecretKeyValueStore(), t)
	// should not be able to read it
	u := world.User("project-member")
	s := setupTestLoginSession(t, tx, u)
	secret := world.Secret("default")

	defer tx.Rollback()
	defer ts.Close()
	req, err := newAuthenticatedRequestJSON(s.Uuid, "GET", ts.URL+"/secrets/"+secret.Uuid, "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Fatalf("res.StatusCode=%d, want %d", res.StatusCode, 200)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	result := struct {
		Subject *domain.PrivilegedSshSecret `json:"subject"`
	}{}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Subject.PrivateKey) != 0 {
		t.Fatal("Gave private key to a non-owner, but it should not.")
	}
	if result.Subject.PublicKey != "public key" {
		t.Fatalf("PublicKey=%s, want %s", result.Subject.PublicKey, "public key")
	}
}

func Test_SecretHandler_Show_EnvironmentSecretAsNonOwner(t *testing.T) {
	ts, ctxt := setupAuthzHandlerTestServer(MountSecretHandler, t)
	tx := ctxt.Tx()

	world := test_helpers.MustNewWorldUsingSecretKeyValueStore(tx, ctxt.SecretKeyValueStore(), t)
	// should not be able to read it
	u := world.User("project-member")
	s := setupTestLoginSession(t, tx, u)
	secret := world.Secret("env")

	defer tx.Rollback()
	defer ts.Close()
	req, err := newAuthenticatedRequestJSON(s.Uuid, "GET", ts.URL+"/secrets/"+secret.Uuid, "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Fatalf("res.StatusCode=%d, want %d", res.StatusCode, 200)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	result := struct {
		Subject *domain.PrivilegedEnvironmentSecret
	}{}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}

	if len(result.Subject.Value) != 0 {
		t.Fatal("Gave secret environment variables key to a non-owner, but it should not.")
	}
}

func Test_SecretHandler_Show_AsOwner(t *testing.T) {
	ts, ctxt := setupAuthzHandlerTestServer(MountSecretHandler, t)
	tx := ctxt.Tx()

	world := test_helpers.MustNewWorldUsingSecretKeyValueStore(tx, ctxt.SecretKeyValueStore(), t)
	// project owner
	u := world.User("project-owner")
	s := setupTestLoginSession(t, tx, u)
	secret := world.Secret("default")

	defer tx.Rollback()
	defer ts.Close()
	req, err := newAuthenticatedRequestJSON(s.Uuid, "GET", ts.URL+"/secrets/"+secret.Uuid, "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Fatalf("res.StatusCode=%d, want %d", res.StatusCode, 201)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	result := struct {
		Subject *domain.PrivilegedSshSecret `json:"subject"`
	}{}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Subject.PrivateKey) == 0 {
		t.Fatal("Did not give private key to  the owner, but it should")
	}
	if result.Subject.PublicKey != "public key" {
		t.Fatalf("PublicKey=%s, want %s", result.Subject.PublicKey, "public key")
	}
}

func Test_SecretHandler_Show_AsNonMember(t *testing.T) {
	ts, ctxt := setupAuthzHandlerTestServer(MountSecretHandler, t)
	tx := ctxt.Tx()

	world := test_helpers.MustNewWorldUsingSecretKeyValueStore(tx, ctxt.SecretKeyValueStore(), t)
	// has no relationship to project
	u := world.User("non-member")
	s := setupTestLoginSession(t, tx, u)
	secret := world.Secret("default")

	defer tx.Rollback()
	defer ts.Close()
	req, err := newAuthenticatedRequestJSON(s.Uuid, "GET", ts.URL+"/secrets/"+secret.Uuid, "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 403 {
		t.Fatalf("res.StatusCode=%d, want %d", res.StatusCode, 403)
	}
}

func Test_SecretHandler_Archive_AsOwner(t *testing.T) {
	ts, ctxt := setupAuthzHandlerTestServer(MountSecretHandler, t)
	tx := ctxt.Tx()

	world := test_helpers.MustNewWorldUsingSecretKeyValueStore(tx, ctxt.SecretKeyValueStore(), t)
	// project owner
	u := world.User("project-owner")
	s := setupTestLoginSession(t, tx, u)
	secret := world.Secret("default")

	defer tx.Rollback()
	defer ts.Close()
	req, err := newAuthenticatedRequestJSON(s.Uuid, "DELETE", ts.URL+"/secrets/"+secret.Uuid, "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 204 {
		t.Fatalf("res.StatusCode=%d, want %d", res.StatusCode, 204)
	}

	store := stores.NewSecretStore(ctxt.SecretKeyValueStore(), tx)
	reloadedSecret, err := store.FindByUuid(secret.Uuid)
	if _, notFound := err.(*domain.NotFoundError); !notFound {
		t.Fatalf("err=%T, expect *domain.NotFoundError", err)
	}
	if reloadedSecret != nil {
		t.Fatalf("secret=%#v, expect %#v", secret, nil)
	}

	secretBytes, err := ctxt.SecretKeyValueStore().Get(secret.Uuid, []byte{})
	if err != stores.ErrKeyNotFound {
		t.Fatalf("err=%s, expect %s", err, stores.ErrKeyNotFound)
	}
	if len(secretBytes) != 0 {
		t.Fatalf("len(secretBytes)=%d, expect %d", len(secretBytes), 0)
	}

}

func Test_SecretHandler_Archive_AsMember(t *testing.T) {
	ts, ctxt := setupAuthzHandlerTestServer(MountSecretHandler, t)
	tx := ctxt.Tx()

	world := test_helpers.MustNewWorldUsingSecretKeyValueStore(tx, ctxt.SecretKeyValueStore(), t)
	// project member
	u := world.User("project-member")
	s := setupTestLoginSession(t, tx, u)
	secret := world.Secret("default")

	defer tx.Rollback()
	defer ts.Close()
	req, err := newAuthenticatedRequestJSON(s.Uuid, "DELETE", ts.URL+"/secrets/"+secret.Uuid, "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 403 {
		t.Fatalf("res.StatusCode=%d, want %d", res.StatusCode, 403)
	}
}

func Test_SecretHandler_Create_emitsSecretAddedActivity(t *testing.T) {
	h := NewHandlerTest(MountSecretHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	env := h.World().Environment("default")

	h.Do("POST", h.Url("/secrets"), &halWrapper{
		Subject: secretParams{
			Name:            "environment secret",
			EnvironmentUuid: env.Uuid,
			Type:            domain.SecretSsh,
		},
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "secret.added" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "secret.added")
}
