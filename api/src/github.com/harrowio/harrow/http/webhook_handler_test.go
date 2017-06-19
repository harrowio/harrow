package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
	"github.com/harrowio/harrow/uuidhelper"

	"github.com/gorilla/mux"
)

func Test_WebhookHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountWebhookHandler(r, nil)

	spec := routingSpec{
		{"POST", "/webhooks", "webhook-create"},
		{"PUT", "/webhooks", "webhook-update"},
		{"GET", "/webhooks/:uuid", "webhook-show"},
		{"GET", "/webhooks/:uuid/deliveries", "webhook-deliveries"},
		{"DELETE", "/webhooks/:uuid", "webhook-archive"},
		{"PATCH", "/webhooks/:uuid/slug", "webhook-regenerate-slug"},

		{"GET", "/wh/:slug", "webhook-deliver"},
		{"POST", "/wh/:slug", "webhook-deliver"},
		{"HEAD", "/wh/:slug", "webhook-deliver"},
		{"DELETE", "/wh/:slug", "webhook-deliver"},
		{"PUT", "/wh/:slug", "webhook-deliver"},
		{"PATCH", "/wh/:slug", "webhook-deliver"},
	}

	spec.run(r, t)
}

func Test_WebhookHandler_Create_requiresAuthorization(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountWebhookHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("non-member")
	ctxt.u = user
	project := world.Project("public")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	s := setupTestLoginSession(t, tx, user)

	req, err := newAuthenticatedRequestJSON(s.Uuid, "POST", ts.URL+"/webhooks", &halWrapper{
		Subject: webhook,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	ctxt.authz.Expect(t, "create", 1)
}

func Test_WebhookHandler_Create_respondsWithCreatedWebhook(t *testing.T) {
	h := NewHandlerTest(MountWebhookHandler, t)
	defer h.Cleanup()

	world := h.World()
	user := world.User("default")
	project := world.Project("public")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	webhook.Slug = ""

	created := domain.Webhook{}
	h.LoginAs("default")
	h.ResultTo(&halWrapper{Subject: &created})
	h.Do("POST", h.Url("/webhooks"), &halWrapper{Subject: webhook})

	res := h.Response()
	if res.StatusCode != http.StatusCreated {
		t.Errorf("Expected created, got %d", res.StatusCode)
	}

	if res.Header.Get("Location") == "" {
		t.Errorf("Expected Location header to be set.")
	}

	if created.CreatorUuid != user.Uuid {
		t.Errorf("Webhook not marked as created by current user")
	}

	if created.ProjectUuid != project.Uuid {
		t.Errorf("Webhook associated with wrong project")
	}

	if created.Slug == "" {
		t.Errorf("Expected slug to not be empty")
	}

	if created.Name != "github" {
		t.Errorf("Expected name %q, got %q", "github", created.Slug)
	}
}

func Test_WebhookHandler_Create_validatesWebhook(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountWebhookHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	project := world.Project("public")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	s := setupTestLoginSession(t, tx, user)

	req, err := newAuthenticatedRequestJSON(s.Uuid, "POST", ts.URL+"/webhooks", &halWrapper{
		Subject: webhook,
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != StatusUnprocessableEntity {
		t.Fatalf("Expected unprocessable entity, got %d", res.StatusCode)
	}
}

func Test_WebhookHandler_RegenerateSlug_generatesNewSlug(t *testing.T) {
	h := NewHandlerTest(MountWebhookHandler, t)
	defer h.Cleanup()

	world := h.World()
	user := world.User("default")
	project := world.Project("public")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	webhookStore := stores.NewDbWebhookStore(h.Tx())
	uuid, err := webhookStore.Create(webhook)
	if err != nil {
		t.Fatal(err)
	}

	h.Subject(webhook)
	h.Do("PATCH", h.UrlFor("slug"), nil)

	reloaded, err := webhookStore.FindByUuid(uuid)
	if got, want := reloaded.Slug, webhook.Slug; got == want {
		t.Fatalf("reloaded.Slug = %q; want not %q", got, want)
	}
}

func Test_WebhookHandler_Update_doesNotAllowOverridingTheSlug(t *testing.T) {
	h := NewHandlerTest(MountWebhookHandler, t)
	defer h.Cleanup()

	world := h.World()
	user := world.User("default")
	project := world.Project("public")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	webhookStore := stores.NewDbWebhookStore(h.Tx())
	uuid, err := webhookStore.Create(webhook)
	if err != nil {
		t.Fatal(err)
	}

	slug := webhook.Slug
	webhook.GenerateSlug()
	h.Subject(webhook)
	h.Do("PUT", h.UrlFor("self"), &halWrapper{Subject: webhook})

	reloaded, err := webhookStore.FindByUuid(uuid)
	if got, want := reloaded.Slug, slug; got != want {
		t.Fatalf("reloaded.Slug = %q; want %q", got, want)
	}
}

func Test_WebhookHandler_Update_validatesWebhook(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountWebhookHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	project := world.Project("public")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := stores.NewDbWebhookStore(tx).Create(webhook); err != nil {
		t.Fatal(err)
	}
	s := setupTestLoginSession(t, tx, user)

	webhook.Slug = ""
	webhook.Name = ""
	req, err := newAuthenticatedRequestJSON(s.Uuid, "PUT", ts.URL+"/webhooks/"+webhook.Uuid, &halWrapper{
		Subject: webhook,
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != StatusUnprocessableEntity {
		t.Fatalf("Expected unprocessable entity, got %d", res.StatusCode)
	}
}

func Test_WebhookHandler_Update_updatesJob(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountWebhookHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	project := world.Project("public")
	job := world.Job("default")
	otherJob := world.Job("other")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	webhookStore := stores.NewDbWebhookStore(tx)
	if _, err := webhookStore.Create(webhook); err != nil {
		t.Fatal(err)
	}
	s := setupTestLoginSession(t, tx, user)

	webhook.JobUuid = otherJob.Uuid
	req, err := newAuthenticatedRequestJSON(s.Uuid, "PUT", ts.URL+"/webhooks/"+webhook.Uuid, &halWrapper{
		Subject: webhook,
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Fatalf("res.StatusCode = %d, want 200", res.StatusCode)
	}
	reloaded, err := webhookStore.FindByUuid(webhook.Uuid)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.JobUuid != otherJob.Uuid {
		t.Fatalf("reloaded.JobUuid = %d, want %d", reloaded.JobUuid, otherJob.Uuid)
	}
}

func Test_WebhookHandler_Update_respondsWithNewVersion(t *testing.T) {
	h := NewHandlerTest(MountWebhookHandler, t)
	defer h.Cleanup()

	world := h.World()
	user := world.User("default")
	project := world.Project("public")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := stores.NewDbWebhookStore(h.Tx()).Create(webhook); err != nil {
		t.Fatal(err)
	}

	webhook.Name = "new-name"
	webhook.Slug = "new-slug"

	updated := domain.Webhook{}
	result := halWrapper{Subject: &updated}
	h.ResultTo(&result)
	h.Subject(webhook)
	h.Do("PUT", h.UrlFor("self"), &halWrapper{Subject: webhook})

	if got, want := updated.Name, "new-name"; got != want {
		t.Fatalf("result.Subject.Name = %q; want %q", got, want)
	}
}

func Test_WebhookHandler_Update_requiresAuthorization(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountWebhookHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	project := world.Project("public")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := stores.NewDbWebhookStore(tx).Create(webhook); err != nil {
		t.Fatal(err)
	}
	s := setupTestLoginSession(t, tx, user)

	webhook.Slug = "new-slug"
	webhook.Name = "new-name"
	req, err := newAuthenticatedRequestJSON(s.Uuid, "PUT", ts.URL+"/webhooks/"+webhook.Uuid, &halWrapper{
		Subject: webhook,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	ctxt.authz.Expect(t, "update", 1)
}

func Test_WebhookHandler_Archive_returns404IfWebhookDoesNotExist(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountWebhookHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	s := setupTestLoginSession(t, tx, user)

	doesNotExist := uuidhelper.MustNewV4()
	req, err := newAuthenticatedRequestJSON(s.Uuid, "DELETE", ts.URL+"/webhooks/"+doesNotExist, ``)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected not found, got %d", res.StatusCode)
	}

}

func Test_WebhookHandler_Archive_archivesWebhook(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountWebhookHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	project := world.Project("public")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := stores.NewDbWebhookStore(tx).Create(webhook); err != nil {
		t.Fatal(err)
	}
	s := setupTestLoginSession(t, tx, user)

	req, err := newAuthenticatedRequestJSON(s.Uuid, "DELETE", ts.URL+"/webhooks/"+webhook.Uuid, ``)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNoContent {
		t.Errorf("Expected no content, got %d", res.StatusCode)
	}

	archived := &domain.Webhook{}
	if err := tx.Get(archived, `SELECT * FROM webhooks WHERE uuid = $1`, webhook.Uuid); err != nil {
		t.Fatal(err)
	}

	if archived.ArchivedAt == nil {
		t.Fatalf("Expected webhook %q to be archived.", webhook.Uuid)
	}
}

func Test_WebhookHandler_Archive_usesAuthorization(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountWebhookHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	project := world.Project("public")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := stores.NewDbWebhookStore(tx).Create(webhook); err != nil {
		t.Fatal(err)
	}
	s := setupTestLoginSession(t, tx, user)

	req, err := newAuthenticatedRequestJSON(s.Uuid, "DELETE", ts.URL+"/webhooks/"+webhook.Uuid, ``)
	if err != nil {
		t.Fatal(err)
	}

	_, err = new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	ctxt.authz.Expect(t, "archive", 1)
}

func Test_WebhookHandler_Deliveries_returns404IfWebhookDoesNotExist(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountWebhookHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	s := setupTestLoginSession(t, tx, user)

	doesNotExist := uuidhelper.MustNewV4()
	req, err := newAuthenticatedRequestJSON(s.Uuid, "GET", ts.URL+"/webhooks/"+doesNotExist+"/deliveries", ``)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected not found, got %d", res.StatusCode)
	}

}

func Test_WebhookHandler_Deliveries_returnsDeliveriesForWebhook(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountWebhookHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	project := world.Project("public")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := stores.NewDbWebhookStore(tx).Create(webhook); err != nil {
		t.Fatal(err)
	}
	s := setupTestLoginSession(t, tx, user)

	deliveryStore := stores.NewDbDeliveryStore(tx)
	exampleRequest, err := http.NewRequest("POST", "http://example.com/webhook", nil)
	if err != nil {
		t.Fatal(err)
	}

	expectedDeliveryIds := []string{}
	for i := 0; i < 3; i++ {
		delivery := webhook.NewDelivery(exampleRequest)
		if uuid, err := deliveryStore.Create(delivery); err != nil {
			t.Fatalf("Creating delivery %d: %s", i, err)
		} else {
			expectedDeliveryIds = append(expectedDeliveryIds, uuid)
		}
	}

	result := struct {
		Collection []struct {
			// Unmarshal only the uuid, because unmarshalling
			// an http.Request fails (cannot unmarshal the body
			// into a io.ReadCloser)
			Subject struct {
				Uuid string
			}
		}
	}{}

	req, err := newAuthenticatedRequestJSON(s.Uuid, "GET", ts.URL+"/webhooks/"+webhook.Uuid+"/deliveries", ``)
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

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("%s\n%s\n", err, body)
	}

	if len(result.Collection) != len(expectedDeliveryIds) {
		t.Errorf("Expected %d deliveries, got %d", len(expectedDeliveryIds), len(result.Collection))
	}

	for _, actual := range result.Collection {
		if !test_helpers.InStrings(actual.Subject.Uuid, expectedDeliveryIds) {
			t.Errorf("Unexpected delivery id: %q", actual.Subject.Uuid)
		}
	}

}

func Test_WebhookHandler_Deliveries_returns20MostRecentDeliveries(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountWebhookHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	project := world.Project("public")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := stores.NewDbWebhookStore(tx).Create(webhook); err != nil {
		t.Fatal(err)
	}
	s := setupTestLoginSession(t, tx, user)

	deliveryStore := stores.NewDbDeliveryStore(tx)
	exampleRequest, err := http.NewRequest("POST", "http://example.com/webhook", nil)
	if err != nil {
		t.Fatal(err)
	}

	createdDeliveries := []*domain.Delivery{}
	now := time.Now()
	for i := 0; i < 30; i++ {
		delivery := webhook.NewDelivery(exampleRequest)
		at := now.Add(-(time.Duration(30-i) * time.Minute))
		if _, err := deliveryStore.CreateAt(delivery, at); err != nil {
			t.Fatalf("Creating delivery %d: %s", i, err)
		}
		createdDeliveries = append(createdDeliveries, delivery)
	}

	req, err := newAuthenticatedRequestJSON(s.Uuid, "GET", ts.URL+"/webhooks/"+webhook.Uuid+"/deliveries", ``)
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

	result := struct {
		Collection []struct {
			Subject struct {
				Uuid        string
				DeliveredAt time.Time
			}
		}
	}{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}

	if n := 20; len(result.Collection) != n {
		t.Fatalf("Expected %d deliveries, got %d", n, len(result.Collection))
	}

	for i := 0; i < 20; i += 20 {
		a := result.Collection[i].Subject
		b := result.Collection[i+1].Subject
		if a.DeliveredAt.Before(b.DeliveredAt) {
			t.Fatalf("Wrong order at index %d: %s before %s\n", i, a.DeliveredAt, b.DeliveredAt)
		}
	}
}

func Test_WebhookHandler_Deliveries_usesAuthorization(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountWebhookHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	project := world.Project("public")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := stores.NewDbWebhookStore(tx).Create(webhook); err != nil {
		t.Fatal(err)
	}
	s := setupTestLoginSession(t, tx, user)

	deliveryStore := stores.NewDbDeliveryStore(tx)
	exampleRequest, err := http.NewRequest("POST", "http://example.com/webhook", nil)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		delivery := webhook.NewDelivery(exampleRequest)
		if _, err := deliveryStore.Create(delivery); err != nil {
			t.Fatalf("Creating delivery %d: %s", i, err)
		}
	}

	req, err := newAuthenticatedRequestJSON(s.Uuid, "GET", ts.URL+"/webhooks/"+webhook.Uuid+"/deliveries", ``)
	if err != nil {
		t.Fatal(err)
	}

	_, err = new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	ctxt.authz.Expect(t, "read", 1 /* webhook */ +3 /* deliveries */)
}

func Test_WebhookHandler_Deliver_returns404IfSlugDoesNotMatch(t *testing.T) {
	h := NewHandlerTest(MountWebhookHandler, t)
	defer h.Cleanup()

	world := h.World()
	user := world.User("default")
	project := world.Project("public")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := stores.NewDbWebhookStore(h.Tx()).Create(webhook); err != nil {
		t.Fatal(err)
	}

	expected := "TEST_DELIVERY_BODY\n"

	webhook.Slug = "does-not-exist"
	h.Subject(webhook)
	h.DoString("POST", h.UrlFor("deliver"), expected)
	res := h.Response()

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected to be not found, got %d", res.StatusCode)
	}
}

func Test_WebhookHandler_Deliver_createsDeliveries(t *testing.T) {
	h := NewHandlerTest(MountWebhookHandler, t)
	defer h.Cleanup()

	user := h.World().User("default")
	project := h.World().Project("public")
	job := h.World().Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := stores.NewDbWebhookStore(h.Tx()).Create(webhook); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		method string
		body   string
	}{
		{"GET", ""},
		{"POST", "post-body"},
		{"HEAD", ""},
		{"DELETE", ""},
		{"PUT", "put-body"},
		{"PATCH", "patch-body"},
	}

	h.Subject(webhook)
	for _, testCase := range testCases {
		h.DoString(testCase.method, h.UrlFor("deliver"), testCase.body)
		if got, want := h.Response().StatusCode, http.StatusCreated; got != want {
			t.Errorf("%s %s %d; want %d", testCase.method, h.UrlFor("deliver"), got, want)
		}
	}

	count := 0
	if err := h.Tx().Get(
		&count,
		`SELECT COUNT(*) FROM deliveries WHERE webhook_uuid = $1`,
		webhook.Uuid,
	); err != nil {
		t.Fatal(err)
	}

	if count != len(testCases) {
		t.Errorf("Expected %d deliveries to have been created, got %d", len(testCases), count)
	}
}

func Test_WebhookHandler_Deliver_storesRequestBody(t *testing.T) {
	h := NewHandlerTest(MountWebhookHandler, t)
	defer h.Cleanup()

	user := h.World().User("default")
	project := h.World().Project("public")
	job := h.World().Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := stores.NewDbWebhookStore(h.Tx()).Create(webhook); err != nil {
		t.Fatal(err)
	}

	expected := "TEST_DELIVERY_BODY"
	h.Subject(webhook)
	h.DoString("POST", h.UrlFor("deliver"), expected)

	count := 0
	if err := h.Tx().Get(
		&count,
		`SELECT COUNT(*) FROM deliveries WHERE request LIKE '%TEST_DELIVERY_BODY%'`,
	); err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Errorf("Expected to find one delivery with %s in body, got %d", expected, count)
	}
}

func Test_WebhookHandler_Deliver_createsSchedule(t *testing.T) {
	h := NewHandlerTest(MountWebhookHandler, t)
	defer h.Cleanup()

	user := h.World().User("default")
	project := h.World().Project("public")
	job := test_helpers.MustCreateJob(t, h.Tx(), &domain.Job{
		EnvironmentUuid: h.World().Environment("astley").Uuid,
		TaskUuid:        h.World().Task("default").Uuid,
		Name:            "to be triggered by webhook",
	})
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := stores.NewDbWebhookStore(h.Tx()).Create(webhook); err != nil {
		t.Fatal(err)
	}

	h.Subject(webhook)
	h.DoString("POST", h.UrlFor("deliver"), "")

	scheduleStore := stores.NewDbScheduleStore(h.Tx())

	schedules, err := scheduleStore.FindAllByJobUuid(job.Uuid)
	if err != nil {
		t.Fatal(err)
	}
	if len(schedules) != 1 {
		t.Errorf("len(schedules) = %d, want %d", len(schedules), 1)
	}
}

func Test_WebhookHandler_Deliver_emitsJobScheduledActivity(t *testing.T) {
	h := NewHandlerTest(MountWebhookHandler, t)
	defer h.Cleanup()

	user := h.World().User("default")
	project := h.World().Project("public")
	job := test_helpers.MustCreateJob(t, h.Tx(), &domain.Job{
		EnvironmentUuid: h.World().Environment("astley").Uuid,
		TaskUuid:        h.World().Task("default").Uuid,
		Name:            "to be triggered by webhook",
	})
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := stores.NewDbWebhookStore(h.Tx()).Create(webhook); err != nil {
		t.Fatal(err)
	}

	h.Subject(webhook)
	h.DoString("POST", h.UrlFor("deliver"), "")

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "job.scheduled" {
			if got, want := activity.Extra["trigger"], "webhook"; !reflect.DeepEqual(got, want) {
				t.Errorf(`activity.Extra["trigger"] = %v; want %v`, got, want)
			}
			return
		}
	}

	t.Fatalf("Activity %q not found", "job.scheduled")
}

func Test_WebhookHandler_Show_returns404IfWebhookDoesNotExist(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountWebhookHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	s := setupTestLoginSession(t, tx, user)

	doesNotExist := uuidhelper.MustNewV4()
	req, err := newAuthenticatedRequestJSON(s.Uuid, "GET", ts.URL+"/webhooks/"+doesNotExist, ``)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected not found, got %d", res.StatusCode)
	}

}

func Test_WebhookHandler_Show_returnsWebhook(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountWebhookHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	project := world.Project("public")
	ctxt.u = user
	s := setupTestLoginSession(t, tx, user)
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := stores.NewDbWebhookStore(tx).Create(webhook); err != nil {
		t.Fatal(err)
	}

	req, err := newAuthenticatedRequestJSON(s.Uuid, "GET", ts.URL+"/webhooks/"+webhook.Uuid, ``)
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

	result := struct{ Subject *domain.Webhook }{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("%s:\n%s\n", err, body)
	}

	if result.Subject.Uuid != webhook.Uuid {
		t.Fatalf("Wrong webhook returned. Expected %q, got %q", webhook.Uuid, result.Subject.Uuid)
	}

}

func Test_WebhookHandler_Show_usesAuthorization(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountWebhookHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	project := world.Project("public")
	ctxt.u = user
	s := setupTestLoginSession(t, tx, user)
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := stores.NewDbWebhookStore(tx).Create(webhook); err != nil {
		t.Fatal(err)
	}

	req, err := newAuthenticatedRequestJSON(s.Uuid, "GET", ts.URL+"/webhooks/"+webhook.Uuid, ``)
	if err != nil {
		t.Fatal(err)
	}

	_, err = new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	ctxt.authz.Expect(t, "read", 1)
}
