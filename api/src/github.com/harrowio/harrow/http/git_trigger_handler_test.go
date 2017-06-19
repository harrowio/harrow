package http

import (
	"testing"

	"github.com/gorilla/mux"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
	"github.com/harrowio/harrow/uuidhelper"
	"github.com/jmoiron/sqlx"
)

func Test_GitTriggerHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountGitTriggerHandler(r, nil)

	spec := routingSpec{
		{"POST", "/git-triggers", "git-triggers-create"},
		{"PUT", "/git-triggers", "git-triggers-update"},
		{"GET", "/git-triggers/:uuid", "git-triggers-show"},
		{"DELETE", "/git-triggers/:uuid", "git-triggers-archive"},
	}

	spec.run(r, t)
}

func createDefaultGitTrigger(t *testing.T, tx *sqlx.Tx, world *test_helpers.World) *domain.GitTrigger {
	project := world.Project("public")
	job := world.Job("default")
	creator := world.User("default")

	store := stores.NewDbGitTriggerStore(tx)
	subject := &domain.GitTrigger{
		Uuid:        uuidhelper.MustNewV4(),
		Name:        "default GitTrigger",
		ChangeType:  "change",
		ProjectUuid: project.Uuid,
		JobUuid:     job.Uuid,
		CreatorUuid: creator.Uuid,
	}

	if _, err := store.Create(subject); err != nil {
		t.Fatal(err)
	}

	return subject
}

func Test_GitTriggerHandler_Create_emitsGitTriggerCreatedActivity(t *testing.T) {
	h := NewHandlerTest(MountGitTriggerHandler, t)
	defer h.Cleanup()

	project := h.World().Project("public")
	job := h.World().Job("default")

	h.LoginAs("default")
	subject := &domain.GitTrigger{
		Uuid:        uuidhelper.MustNewV4(),
		Name:        "created GitTrigger",
		JobUuid:     job.Uuid,
		ChangeType:  "change",
		ProjectUuid: project.Uuid,
	}

	h.Do("POST", h.Url("/git-triggers"), &halWrapper{
		Subject: subject,
	})
	t.Logf("Response:\n%s\n", h.ResponseBody())
	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "git-triggers.created" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "git-triggers.created")
}

func Test_GitTriggerHandler_Update_emitsGitTriggerEditedActivity(t *testing.T) {
	h := NewHandlerTest(MountGitTriggerHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	subject := createDefaultGitTrigger(t, h.Tx(), h.World())

	h.Do("PUT", h.Url("/git-triggers"), &halWrapper{
		Subject: subject,
	})
	t.Logf("Response:\n%s\n", h.ResponseBody())
	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "git-triggers.edited" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "git-triggers.edited")
}

func Test_GitTriggerHandler_Delete_emitsGitTriggerDeletedActivity(t *testing.T) {
	h := NewHandlerTest(MountGitTriggerHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	subject := createDefaultGitTrigger(t, h.Tx(), h.World())

	h.Subject(subject)
	h.Do("DELETE", h.UrlFor("self"), nil)
	t.Logf("Response:\n%s\n", h.ResponseBody())
	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "git-triggers.deleted" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "git-triggers.deleted")
}
