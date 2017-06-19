package http

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/test_helpers"
)

func Test_RepoHandler_Routing(t *testing.T) {

	r := mux.NewRouter()
	kv := test_helpers.NewMockKeyValueStore()
	ss := test_helpers.NewMockSecretKeyValueStore()
	db := test_helpers.GetDbConnection(t)
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	ctxt := NewTestContext(db, tx, kv, ss, c)

	MountRepoHandler(r, ctxt)

	spec := routingSpec{
		{"PUT", "/repositories", "repository-update"},
		{"POST", "/repositories", "repository-create"},

		{"GET", "/repositories/:uuid/operations", "repository-operations"},
		{"POST", "/repositories/:uuid/checks", "repository-checks"},
		{"POST", "/repositories/:uuid/metadata", "repository-metadata"},
		{"GET", "/repositories/:uuid/credential", "repository-credential"},
		{"GET", "/repositories/:uuid", "repository-show"},
		{"DELETE", "/repositories/:uuid", "repository-archive"},
	}

	spec.run(r, t)
}

func Test_RepoHandler_Credential_returnsRC(t *testing.T) {
	h := NewHandlerTest(MountRepoHandler, t)
	defer h.Cleanup()

	world := h.World()
	repo := world.Repository("other")
	defaultRc := world.RepositoryCredential("present")

	rc := domain.RepositoryCredential{}
	h.LoginAs("default")
	h.ResultTo(&halWrapper{Subject: &rc})
	h.Subject(repo)

	h.Do("GET", h.UrlFor("credential"), nil)

	res := h.Response()
	if res.StatusCode != http.StatusOK {
		t.Errorf("res.StatusCode = %d, want %d", res.StatusCode, http.StatusOK)
	}
	if rc.Uuid != defaultRc.Uuid {
		t.Fatalf("have=%#v, want=%#v", rc.Uuid, defaultRc.Uuid)
	}
}

func Test_RepoHandler_CreateUpdate_emitsRepositoryAddedActivity_whenCreatingARepository(t *testing.T) {
	h := NewHandlerTest(MountRepoHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")

	project := h.World().Project("public")

	h.Do("POST", h.Url("/repositories"), &repoParamsWrapper{
		Subject: repoParams{
			Url:         "https://github.com/my/repo",
			Name:        "created",
			ProjectUuid: project.Uuid,
		},
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "repository.added" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "repository.added")
}

func Test_RepoHandler_CreateUpdate_emitsRepositoryDetectedAsPrivate_whenCreatingAPrivateRepository(t *testing.T) {
	h := NewHandlerTest(MountRepoHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")

	project := h.World().Project("public")

	h.Do("POST", h.Url("/repositories"), &repoParamsWrapper{
		Subject: repoParams{
			Url:         "https://github.com/harrowio/private-private.git",
			Name:        "created",
			ProjectUuid: project.Uuid,
		},
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "repository.detected-as-private" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "repository.detected-as-private")
}

func Test_RepoHandler_CreateUpdate_doesNotEmitRepositoryDetectedAsPrivate_whenCreatingAPublicRepository(t *testing.T) {
	h := NewHandlerTest(MountRepoHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")

	project := h.World().Project("public")

	h.Do("POST", h.Url("/repositories"), &repoParamsWrapper{
		Subject: repoParams{
			Url:         "https://github.com/capistrano/capistrano.git",
			Name:        "created",
			ProjectUuid: project.Uuid,
		},
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "repository.detected-as-private" {
			t.Fatalf("Activity %q found", "repository.detected-as-private")
		}
	}
}

func Test_RepoHandler_CreateUpdate_emitsRepositoryEditedActivity_whenUpdatingARepository(t *testing.T) {
	h := NewHandlerTest(MountRepoHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")

	repo := h.World().Repository("default")

	h.Do("PUT", h.Url("/repositories"), &repoParamsWrapper{
		Subject: repoParams{
			Uuid:        repo.Uuid,
			Url:         repo.Url,
			Name:        "edited",
			ProjectUuid: repo.ProjectUuid,
		},
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "repository.edited" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "repository.edited")
}
