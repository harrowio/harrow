package http

import (
	"testing"

	"github.com/gorilla/mux"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
)

func Test_SlackNotifierHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountSlackNotifierHandler(r, nil)

	spec := routingSpec{
		{"POST", "/slack-notifiers", "slack-notifiers-create"},
		{"PUT", "/slack-notifiers", "slack-notifiers-update"},
		{"GET", "/slack-notifiers/:uuid", "slack-notifiers-show"},
		{"DELETE", "/slack-notifiers/:uuid", "slack-notifiers-archive"},
	}

	spec.run(r, t)
}

func createDefaultSlackNotifier(t *testing.T, h *httpHandlerTest) *domain.SlackNotifier {
	store := stores.NewDbSlackNotifierStore(h.Tx())
	subject := &domain.SlackNotifier{
		Uuid:        uuidhelper.MustNewV4(),
		Name:        "default SlackNotifier",
		WebhookURL:  "https://hooks.slack.com/services/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",
		UrlHost:     "https://www.vm.harrow.io",
		ProjectUuid: h.World().Project("public").Uuid,
	}

	if _, err := store.Create(subject); err != nil {
		t.Fatal(err)
	}

	return subject
}

func Test_SlackNotifierHandler_Create_emitsSlackNotifierCreatedActivity(t *testing.T) {
	h := NewHandlerTest(MountSlackNotifierHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	subject := &domain.SlackNotifier{
		Uuid:        uuidhelper.MustNewV4(),
		Name:        "created SlackNotifier",
		WebhookURL:  "https://hooks.slack.com/services/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",
		UrlHost:     "https://www.vm.harrow.io",
		ProjectUuid: h.World().Project("public").Uuid,
	}

	h.Do("POST", h.Url("/slack-notifiers"), &halWrapper{
		Subject: subject,
	})
	for _, activity := range h.Activities() {
		if activity.Name == "slack-notifiers.created" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "slack-notifiers.created")
}

func Test_SlackNotifierHandler_Update_emitsSlackNotifierEditedActivity(t *testing.T) {
	h := NewHandlerTest(MountSlackNotifierHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	subject := createDefaultSlackNotifier(t, h)

	h.Do("PUT", h.Url("/slack-notifiers"), &halWrapper{
		Subject: subject,
	})
	for _, activity := range h.Activities() {
		if activity.Name == "slack-notifiers.edited" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "slack-notifiers.edited")
}

func Test_SlackNotifierHandler_Delete_emitsSlackNotifierDeletedActivity(t *testing.T) {
	h := NewHandlerTest(MountSlackNotifierHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	subject := createDefaultSlackNotifier(t, h)

	h.Subject(subject)
	h.Do("DELETE", h.UrlFor("self"), nil)
	for _, activity := range h.Activities() {
		if activity.Name == "slack-notifiers.deleted" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "slack-notifiers.deleted")
}
