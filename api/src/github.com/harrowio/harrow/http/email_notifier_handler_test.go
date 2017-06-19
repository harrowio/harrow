package http

import (
	"testing"

	"github.com/gorilla/mux"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
)

func Test_EmailNotifierHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountEmailNotifierHandler(r, nil)

	spec := routingSpec{
		{"POST", "/email-notifiers", "email-notifiers-create"},
		{"PUT", "/email-notifiers", "email-notifiers-update"},
		{"GET", "/email-notifiers/:uuid", "email-notifiers-show"},
		{"DELETE", "/email-notifiers/:uuid", "email-notifiers-archive"},
	}

	spec.run(r, t)
}

func createDefaultEmailNotifier(t *testing.T, h *httpHandlerTest) *domain.EmailNotifier {
	project := h.World().Project("public")
	store := stores.NewDbEmailNotifierStore(h.Tx())
	subject := &domain.EmailNotifier{
		Uuid:        uuidhelper.MustNewV4(),
		Recipient:   "vagrant@localhost",
		ProjectUuid: &project.Uuid,
	}

	if _, err := store.Create(subject); err != nil {
		t.Fatal(err)
	}

	return subject
}

func Test_EmailNotifierHandler_Create_emitsEmailNotifierCreatedActivity(t *testing.T) {
	h := NewHandlerTest(MountEmailNotifierHandler, t)
	defer h.Cleanup()
	project := h.World().Project("public")
	h.LoginAs("default")
	subject := &domain.EmailNotifier{
		Uuid:        uuidhelper.MustNewV4(),
		Recipient:   "vagrant@localhost",
		ProjectUuid: &project.Uuid,
	}

	h.Do("POST", h.Url("/email-notifiers"), &halWrapper{
		Subject: subject,
	})
	t.Logf("Response:\n%s\n", h.ResponseBody())
	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "email-notifiers.created" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "email-notifiers.created")
}

func Test_EmailNotifierHandler_Update_emitsEmailNotifierEditedActivity(t *testing.T) {
	h := NewHandlerTest(MountEmailNotifierHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	subject := createDefaultEmailNotifier(t, h)

	h.Do("PUT", h.Url("/email-notifiers"), &halWrapper{
		Subject: subject,
	})
	t.Logf("Response:\n%s\n", h.ResponseBody())
	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "email-notifiers.edited" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "email-notifiers.edited")
}

func Test_EmailNotifierHandler_Delete_emitsEmailNotifierDeletedActivity(t *testing.T) {
	h := NewHandlerTest(MountEmailNotifierHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	subject := createDefaultEmailNotifier(t, h)

	h.Subject(subject)
	h.Do("DELETE", h.UrlFor("self"), nil)
	t.Logf("Response:\n%s\n", h.ResponseBody())
	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "email-notifiers.deleted" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "email-notifiers.deleted")
}
