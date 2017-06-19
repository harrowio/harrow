package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"

	"github.com/gorilla/mux"
)

func Test_JobHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountJobHandler(r, nil)

	spec := routingSpec{
		{"POST", "/jobs", "job-create"},
		{"PUT", "/jobs", "job-update"},
		{"GET", "/jobs/:uuid/watch", "job-watch-status"},
		{"PUT", "/jobs/:uuid/watch", "job-watch"},
		{"GET", "/jobs/:uuid/operations", "job-operations"},
		{"GET", "/jobs/:uuid/scheduled-executions", "job-scheduled-executions"},
		{"GET", "/jobs/:uuid", "job-show"},
		{"DELETE", "/jobs/:uuid", "job-archive"},
		{"PUT", "/jobs/:uuid/subscriptions", "job-subscribe"},
		{"GET", "/jobs/:uuid/subscriptions", "job-subscriptions"},
		{"GET", "/jobs/:uuid/notification-rules", "job-notification-rules"},
	}

	spec.run(r, t)
}

func Test_JobHandler_Subscribe_watch_true_SubscribesToAllEvents(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountJobHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	subscriptions := stores.NewDbSubscriptionStore(tx)
	user := world.User("default")
	ctxt.u = user
	job := world.Job("default")

	s := setupTestLoginSession(t, tx, user)
	url := ts.URL + "/jobs/" + job.Uuid + "/subscriptions"
	req, err := newAuthenticatedRequest(s.Uuid, "PUT", url, `{"watch": true}`)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNoContent {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal(string(body))
	}

	for _, event := range job.WatchableEvents() {
		_, err := subscriptions.Find(job.Uuid, event, user.Uuid)
		if err != nil {
			t.Errorf("subscription for %s: %s", event, err)
		}
	}
}

func Test_JobHandler_Subscribe_watch_false_UnsubscribesFromAllEvents(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountJobHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	subscriptions := stores.NewDbSubscriptionStore(tx)
	user := world.User("default")
	ctxt.u = user
	job := world.Job("default")

	if err := user.Watch(job, subscriptions); err != nil {
		t.Fatal(err)
	}

	s := setupTestLoginSession(t, tx, user)
	url := ts.URL + "/jobs/" + job.Uuid + "/subscriptions"
	req, err := newAuthenticatedRequest(s.Uuid, "PUT", url, `{"watch": false}`)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNoContent {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal(string(body))
	}

	for _, event := range job.WatchableEvents() {
		_, err := subscriptions.Find(job.Uuid, event, user.Uuid)
		if _, ok := err.(*domain.NotFoundError); !ok {
			if err == nil {
				t.Errorf("still subscribed to %s", event)
			} else {
				t.Fatal(err)
			}
		}
	}
}

func Test_JobHandler_Subscribe_events_SubscribesToGivenEvents(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountJobHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	subscriptions := stores.NewDbSubscriptionStore(tx)
	user := world.User("default")
	ctxt.u = user
	job := world.Job("default")

	s := setupTestLoginSession(t, tx, user)
	url := ts.URL + "/jobs/" + job.Uuid + "/subscriptions"
	req, err := newAuthenticatedRequest(s.Uuid, "PUT", url, `{"events": {"operations.scheduled": true, "operations.succeeded": false}}`)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNoContent {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal(string(body))
	}

	_, err = subscriptions.Find(job.Uuid, domain.EventOperationSucceeded, user.Uuid)
	if _, ok := err.(*domain.NotFoundError); !ok {
		if err == nil {
			t.Errorf("still subscribed to %s", domain.EventOperationSucceeded)
		} else {
			t.Fatal(err)
		}
	}

	_, err = subscriptions.Find(job.Uuid, domain.EventOperationScheduled, user.Uuid)
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); ok {
			t.Errorf("not subscribed to %s", domain.EventOperationScheduled)
		} else {
			t.Fatal(err)
		}
	}
}

func Test_JobHandler_Subscribe_events_supersededByWatch(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountJobHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	subscriptions := stores.NewDbSubscriptionStore(tx)
	user := world.User("default")
	ctxt.u = user
	job := world.Job("default")

	s := setupTestLoginSession(t, tx, user)
	url := ts.URL + "/jobs/" + job.Uuid + "/subscriptions"
	req, err := newAuthenticatedRequest(s.Uuid, "PUT", url, `{"watch": true, "events": {"operations.scheduled": true, "operations.succeeded": false}}`)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNoContent {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal(string(body))
	}

	for _, event := range job.WatchableEvents() {
		_, err = subscriptions.Find(job.Uuid, event, user.Uuid)
		if err != nil {
			if _, ok := err.(*domain.NotFoundError); ok {
				t.Errorf("not subscribed to %s", event)
			} else {
				t.Fatal(err)
			}
		}
	}
}

func Test_JobHandler_Subscriptions_listsAllSubscriptionsOfTheCurrentUser(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountJobHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	subscriptions := stores.NewDbSubscriptionStore(tx)
	user := world.User("default")
	ctxt.u = user
	job := world.Job("default")

	if err := user.Watch(job, subscriptions); err != nil {
		t.Fatal(err)
	}

	s := setupTestLoginSession(t, tx, user)
	url := ts.URL + "/jobs/" + job.Uuid + "/subscriptions"
	req, err := newAuthenticatedRequest(s.Uuid, "GET", url, "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode >= 400 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Fatal(string(body))
	}

	result := struct {
		Subject *domain.Subscriptions `json:"subject"`
	}{}

	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&result); err != nil {
		t.Fatal(err)
	}

	subscribed := map[string]bool{}
	for _, event := range job.WatchableEvents() {
		subscribed[event] = true
	}

	if !reflect.DeepEqual(result.Subject.Subscribed, subscribed) {
		t.Fatalf("Expected result.Subject.Subscribed to equal %#v, got %#v\n", subscribed, result.Subject.Subscribed)
	}
}

func Test_JobHandler_ScheduledExecutions_returnsProperErrorForMalformedTimestamp(t *testing.T) {
	h := NewHandlerTest(MountJobHandler, t)
	defer h.Cleanup()
	job := h.World().Job("default")
	h.Subject(job)
	for _, param := range []string{"to", "from"} {
		result := struct{ Reason string }{}
		h.ResultTo(&result)
		t.Logf("param = %q", param)
		h.Do("GET", h.UrlFor("scheduled-executions"), url.Values{
			param: []string{"now"},
		})
		response := h.Response()
		if got, want := response.StatusCode, http.StatusBadRequest; got != want {
			t.Errorf("response.StatusCode = %d; want %d", got, want)
		}
		if got, want := result.Reason, "malformed_parameter"; got != want {
			t.Errorf("result.Reason = %q; want %q", got, want)
		}
	}
}

func Test_JobHandler_Update_emits_JobEditedActivity(t *testing.T) {
	h := NewHandlerTest(MountJobHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")

	job := h.World().Job("default")

	h.LoginAs("default")
	h.Do("PUT", h.Url("/jobs"), &jobParamsWrapper{
		Subject: jobParams{
			Uuid:            job.Uuid,
			Name:            "changed",
			TaskUuid:        job.TaskUuid,
			EnvironmentUuid: job.EnvironmentUuid,
		},
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "job.edited" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "job.edited")
}

func Test_JobHandler_Create_emits_JobAddedActivity(t *testing.T) {
	h := NewHandlerTest(MountJobHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")

	task := h.World().Task("default")
	env := h.World().Environment("astley")

	h.Do("POST", h.Url("/jobs"), &jobParamsWrapper{
		Subject: jobParams{
			Name:            "created",
			TaskUuid:        task.Uuid,
			EnvironmentUuid: env.Uuid,
		},
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "job.added" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "job.added")
}
