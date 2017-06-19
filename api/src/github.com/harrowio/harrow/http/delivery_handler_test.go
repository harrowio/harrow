package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
	"github.com/harrowio/harrow/uuidhelper"

	"github.com/gorilla/mux"
)

func Test_DeliveryHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountDeliveryHandler(r, nil)

	spec := routingSpec{
		{"GET", "/deliveries/:uuid", "delivery-show"},
	}

	spec.run(r, t)
}

func Test_DeliveryHandler_Show_returns404IfDeliveryDoesNotExist(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountDeliveryHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	ctxt.u = user
	s := setupTestLoginSession(t, tx, user)

	doesNotExist := uuidhelper.MustNewV4()
	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/deliveries/"+doesNotExist, ``)
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

func Test_DeliveryHandler_Show_returnsDelivery(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountDeliveryHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()

	webhookStore := stores.NewDbWebhookStore(tx)
	deliveryStore := stores.NewDbDeliveryStore(tx)
	world := test_helpers.MustNewWorld(tx, t)

	user := world.User("default")
	ctxt.u = user
	project := world.Project("public")
	job := world.Job("default")

	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "webhook-name")
	if _, err := webhookStore.Create(webhook); err != nil {
		t.Fatal(err)
	}

	deliveredRequest, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	delivery := webhook.NewDelivery(deliveredRequest)
	if _, err := deliveryStore.Create(delivery); err != nil {
		t.Fatal(err)
	}

	s := setupTestLoginSession(t, tx, user)

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/deliveries/"+delivery.Uuid, ``)
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

	result := struct{ Subject struct{ Uuid string } }{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}

	if result.Subject.Uuid != delivery.Uuid {
		t.Fatalf("Expected to retrieve delivery %q, got %q", delivery.Uuid, result.Subject.Uuid)
	}

}

func Test_DeliveryHandler_Show_usesAuthorization(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountDeliveryHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()

	webhookStore := stores.NewDbWebhookStore(tx)
	deliveryStore := stores.NewDbDeliveryStore(tx)
	world := test_helpers.MustNewWorld(tx, t)

	user := world.User("default")
	ctxt.u = user
	project := world.Project("public")
	job := world.Job("default")

	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "webhook-name")
	if _, err := webhookStore.Create(webhook); err != nil {
		t.Fatal(err)
	}

	deliveredRequest, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	delivery := webhook.NewDelivery(deliveredRequest)
	if _, err := deliveryStore.Create(delivery); err != nil {
		t.Fatal(err)
	}

	s := setupTestLoginSession(t, tx, user)

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/deliveries/"+delivery.Uuid, ``)
	if err != nil {
		t.Fatal(err)
	}

	_, err = new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	ctxt.authz.Expect(t, "read", 1)
}
