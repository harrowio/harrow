package http

import (
	"testing"

	"github.com/gorilla/mux"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
)

func Test_JobNotifierHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountJobNotifierHandler(r, nil)

	spec := routingSpec{
		{"POST", "/job-notifiers", "job-notifiers-create"},
		{"GET", "/job-notifiers/:uuid", "job-notifiers-show"},
		{"DELETE", "/job-notifiers/:uuid", "job-notifiers-archive"},
	}

	spec.run(r, t)
}

func createDefaultJobNotifier(t *testing.T, h *httpHandlerTest) *domain.JobNotifier {
	store := stores.NewDbJobNotifierStore(h.Tx())
	subject := &domain.JobNotifier{
		ProjectUuid: h.World().Project("private").Uuid,
		Uuid:        uuidhelper.MustNewV4(),
		JobUuid:     h.World().Job("default").Uuid,
	}

	if _, err := store.Create(subject); err != nil {
		t.Logf("erred trying to make %#v", subject)
		t.Fatal(err)
	}

	return subject
}

func Test_JobNotifierHandler_Create_emitsJobNotifierCreatedActivity(t *testing.T) {
	h := NewHandlerTest(MountJobNotifierHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	subject := &domain.JobNotifier{
		ProjectUuid: h.World().Project("private").Uuid,
		Uuid:        uuidhelper.MustNewV4(),
		JobUuid:     h.World().Job("default").Uuid,
	}

	h.Do("POST", h.Url("/job-notifiers"), &halWrapper{
		Subject: subject,
	})
	t.Logf("Response:\n%s\n", h.ResponseBody())
	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "job-notifiers.created" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "job-notifiers.created")
}

func Test_JobNotifierHandler_Delete_emitsJobNotifierDeletedActivity(t *testing.T) {
	h := NewHandlerTest(MountJobNotifierHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	subject := createDefaultJobNotifier(t, h)

	h.Subject(subject)
	h.Do("DELETE", h.UrlFor("self"), nil)
	t.Logf("Response:\n%s\n", h.ResponseBody())
	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "job-notifiers.deleted" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "job-notifiers.deleted")
}
