package http

import (
	"testing"

	"github.com/gorilla/mux"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
)

func Test_NotificationRuleHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountNotificationRuleHandler(r, nil)

	spec := routingSpec{
		{"POST", "/notification-rules", "notification-rules-create"},
		{"PUT", "/notification-rules", "notification-rules-update"},
		{"GET", "/notification-rules/:uuid", "notification-rules-show"},
		{"DELETE", "/notification-rules/:uuid", "notification-rules-archive"},
	}

	spec.run(r, t)
}

func createDefaultNotificationRule(t *testing.T, h *httpHandlerTest) *domain.NotificationRule {
	store := stores.NewDbNotificationRuleStore(h.Tx())
	project := h.World().Project("public")
	job := h.World().Job("default")
	subject := &domain.NotificationRule{
		Uuid:          uuidhelper.MustNewV4(),
		ProjectUuid:   project.Uuid,
		JobUuid:       &job.Uuid,
		NotifierType:  "email_notifiers",
		NotifierUuid:  uuidhelper.MustNewV4(),
		MatchActivity: "*",
		CreatorUuid:   h.World().User("default").Uuid,
	}

	if _, err := store.Create(subject); err != nil {
		t.Fatal(err)
	}

	return subject
}

func Test_NotificationRuleHandler_Create_emitsNotificationRuleCreatedActivity(t *testing.T) {
	h := NewHandlerTest(MountNotificationRuleHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	project := h.World().Project("public")
	job := h.World().Job("default")
	subject := &domain.NotificationRule{
		Uuid:          uuidhelper.MustNewV4(),
		ProjectUuid:   project.Uuid,
		JobUuid:       &job.Uuid,
		NotifierType:  "email_notifiers",
		NotifierUuid:  uuidhelper.MustNewV4(),
		MatchActivity: "*",
		CreatorUuid:   h.World().User("default").Uuid,
	}

	h.Do("POST", h.Url("/notification-rules"), &halWrapper{
		Subject: subject,
	})
	t.Logf("Response:\n%s\n", h.ResponseBody())
	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "notification-rules.created" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "notification-rules.created")
}

func Test_NotificationRuleHandler_Update_emitsNotificationRuleEditedActivity(t *testing.T) {
	h := NewHandlerTest(MountNotificationRuleHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	subject := createDefaultNotificationRule(t, h)

	h.Do("PUT", h.Url("/notification-rules"), &halWrapper{
		Subject: subject,
	})
	t.Logf("Response:\n%s\n", h.ResponseBody())
	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "notification-rules.edited" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "notification-rules.edited")
}

func Test_NotificationRuleHandler_Delete_emitsNotificationRuleDeletedActivity(t *testing.T) {
	h := NewHandlerTest(MountNotificationRuleHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	subject := createDefaultNotificationRule(t, h)

	h.Subject(subject)
	h.Do("DELETE", h.UrlFor("self"), nil)
	t.Logf("Response:\n%s\n", h.ResponseBody())
	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "notification-rules.deleted" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "notification-rules.deleted")
}
