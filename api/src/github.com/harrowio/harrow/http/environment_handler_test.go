package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/test_helpers"
)

func Test_EnvironmentHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountEnvironmentHandler(r, nil)

	spec := routingSpec{
		{"GET", "/environments/:uuid", "environment-show"},
		{"DELETE", "/environments/:uuid", "environment-archive"},
		{"PUT", "/environments", "environment-update"},
		{"POST", "/environments", "environment-create"},
		{"GET", "/environments/:uuid/targets", "environment-targets"},
		{"GET", "/environments/:uuid/jobs", "environment-jobs"},
		{"GET", "/environments/:uuid/secrets", "environment-secrets"},
	}

	spec.run(r, t)
}

func Test_EnvironmentHandler_Show_usesChecks_ReadPermissions(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountEnvironmentHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	env := world.Environment("default")

	req, err := newRequest("GET", ts.URL+"/environments/"+env.Uuid, ``)
	if err != nil {
		t.Fatal(err)
	}

	_, err = new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	ctxt.authz.Expect(t, "read", 1)
}

func Test_EnvironmentHandler_Secrets_AsMember(t *testing.T) {
	ts, ctxt := setupAuthzHandlerTestServer(MountEnvironmentHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorldUsingSecretKeyValueStore(tx, ctxt.SecretKeyValueStore(), t)
	u := world.User("project-member")
	s := setupTestLoginSession(t, tx, u)
	e := world.Environment("private")

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/environments/"+e.Uuid+"/secrets", "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	var container struct {
		Collection []struct {
			Subject *domain.PrivilegedSshSecret  `json:"subject"`
			Links   map[string]map[string]string `json:"_links"`
		} `json:"collection"`
		Meta  map[string]string            `json:"_meta"`
		Links map[string]map[string]string `json:"_links"`
	}

	if err = json.Unmarshal(body, &container); err != nil {
		t.Fatalf("%s in:\n%s\n", err, string(body))
	}

	if len(container.Collection) != 2 {
		t.Fatalf("len(container.Collection) = %d, want %d", len(container.Collection), 2)
	}
	var envIdx, sshIdx int
	if container.Collection[0].Subject.Type == domain.SecretSsh {
		envIdx = 1
	} else {
		sshIdx = 1
	}

	// We can parse the result into a PrivilegedSshSecret struct, but the private key will be empty
	// (because PrivilegedSshSecret is a superset of UnprivilegedSshSecret)
	privilegedSecret := mustParsePrivilegedSshSecrets(body, t)[sshIdx]
	if privilegedSecret.PrivateKey != "" {
		t.Fatalf("privilegedSecret.PrivateKey=%s, want empty string", privilegedSecret.PrivateKey)
	}
	if privilegedSecret.PublicKey != "public key" {
		t.Fatalf("privilegedSecret.PublicKey=%s, want %s", privilegedSecret.PublicKey, "pub")
	}

	privilegedEnvSecret := mustParsePrivilegedEnvironmentSecrets(body, t)[envIdx]
	if privilegedEnvSecret.Value != "" {
		t.Fatalf("privilegedEnvSecret.Value=%s, want empty string", privilegedEnvSecret.Value)
	}
}

func Test_EnvironmentHandler_Secrets_AsNonMember(t *testing.T) {
	ts, ctxt := setupAuthzHandlerTestServer(MountEnvironmentHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorldUsingSecretKeyValueStore(tx, ctxt.SecretKeyValueStore(), t)
	u := world.User("non-member")
	s := setupTestLoginSession(t, tx, u)
	e := world.Environment("private")

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/environments/"+e.Uuid+"/secrets", "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	// the user is not allow to read the env, therefore we expect 403
	// if they would be allowed to read the env but not individual secrets, we
	// would expect 200, but a smaller list of secrets accordingly.
	// It is currently not possible to have permissions like this, so this can't
	// be tested.
	if res.StatusCode != 403 {
		t.Fatalf("res.StatusCode=%d, want %d", res.StatusCode, 403)
	}
}

func mustParsePrivilegedSshSecrets(body []byte, t *testing.T) []*domain.PrivilegedSshSecret {
	var container struct {
		Collection []struct {
			Subject *domain.PrivilegedSshSecret  `json:"subject"`
			Links   map[string]map[string]string `json:"_links"`
		} `json:"collection"`
		Meta  map[string]string            `json:"_meta"`
		Links map[string]map[string]string `json:"_links"`
	}

	if err := json.Unmarshal(body, &container); err != nil {
		t.Fatalf("%s in:\n%s\n", err, string(body))
	}
	res := []*domain.PrivilegedSshSecret{}
	for _, item := range container.Collection {
		res = append(res, item.Subject)
	}
	return res
}

func mustParsePrivilegedEnvironmentSecrets(body []byte, t *testing.T) []*domain.PrivilegedEnvironmentSecret {
	var container struct {
		Collection []struct {
			Subject *domain.PrivilegedEnvironmentSecret `json:"subject"`
			Links   map[string]map[string]string        `json:"_links"`
		} `json:"collection"`
		Meta  map[string]string            `json:"_meta"`
		Links map[string]map[string]string `json:"_links"`
	}

	if err := json.Unmarshal(body, &container); err != nil {
		t.Fatalf("%s in:\n%s\n", err, string(body))
	}
	res := []*domain.PrivilegedEnvironmentSecret{}
	for _, item := range container.Collection {
		res = append(res, item.Subject)
	}
	return res
}

func Test_EnvironmentHandler_CreateUpdate_emitsEnvironmentsitoryAddedActivity_whenCreatingAEnvironmentsitory(t *testing.T) {
	h := NewHandlerTest(MountEnvironmentHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")

	project := h.World().Project("public")

	h.Do("POST", h.Url("/environments"), &envParamsWrapper{
		Subject: envParams{
			Name:        "created",
			ProjectUuid: project.Uuid,
			Variables: domain.EnvironmentVariables{
				M: map[string]string{
					"A": "B",
				},
			},
		},
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "environment.added" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "environment.added")
}

func Test_EnvironmentHandler_CreateUpdate_emitsEnvironmentEditedActivity_whenUpdatingAEnvironment(t *testing.T) {
	h := NewHandlerTest(MountEnvironmentHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")

	env := h.World().Environment("default")

	h.Do("PUT", h.Url("/environments"), &envParamsWrapper{
		Subject: envParams{
			Uuid:        env.Uuid,
			Name:        "edited",
			ProjectUuid: env.ProjectUuid,
			Variables:   env.Variables,
		},
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "environment.edited" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "environment.edited")
}
