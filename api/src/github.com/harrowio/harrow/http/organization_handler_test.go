package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
)

func Test_OrganizationHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountOrganizationHandler(r, nil)

	spec := routingSpec{
		{"POST", "/organizations", "organization-create"},
		{"PUT", "/organizations", "organization-update"},
		{"GET", "/organizations/:uuid", "organization-show"},
		{"DELETE", "/organizations/:uuid", "organization-archive"},
		{"GET", "/organizations/:uuid/projects", "organization-projects"},
		{"GET", "/organizations/:uuid/memberships", "organization-memberships"},
		{"GET", "/organizations/:uuid/members", "organization-members"},
	}

	spec.run(r, t)
}

func Test_OrganizationCreateHandler_Status400OnMalformedInput(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountOrganizationHandler, t)
	defer ctxt.Tx().Rollback()
	defer ts.Close()

	tx := ctxt.Tx()
	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequest(s.Uuid, "POST", ts.URL+"/organizations/", `{"subject": {"name:Test Org"}}`)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusBadRequest {
		t.Fatal("Expected StatusBadRequest, got:", res.StatusCode, res.Body)
	}

}

func Test_OrganizationCreateHandler_Status422OnValidationError(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountOrganizationHandler, t)
	defer ctxt.Tx().Rollback()
	defer ts.Close()

	tx := ctxt.Tx()
	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequest(s.Uuid, "POST", ts.URL+"/organizations/", `{"subject": {"name": ""}}`)
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

	if res.StatusCode != 422 {
		t.Fatal("Expected Status #422, got:", res.StatusCode, string(body))
	}

	if len(body) == 0 {
		t.Fatal("body should contain a JSON representation of the validation error")
	}

	_, ok := res.Header["Location"]
	if ok {
		t.Fatal("location header should not be set")
	}

}

func Test_OrganizationCreateHandler_Status201OnSuccess(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountOrganizationHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	ctxt.u = u
	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequest(s.Uuid, "POST", ts.URL+"/organizations/", `{"subject": {"name": "Test Org", "public": true}}`)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(res.Body)
		t.Fatalf("Expected StatusCreated, got: %d (body: %s)", res.StatusCode, body)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(body) == 0 {
		t.Fatal("body should be non-empty")
	}

	location, ok := res.Header["Location"]
	if !ok {
		t.Fatal("location header should be set")
	}
	if !strings.Contains(location[0], ts.URL+"/organizations/") {
		t.Fatal(location, "does not match the expected header")
	}

}

func Test_OrganizationProjectsHandler_Status200OkOnFoundOrganization_Member(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountOrganizationHandler, t)
	defer ctxt.Tx().Rollback()
	defer ts.Close()

	tx := ctxt.Tx()
	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	s := setupTestLoginSession(t, tx, u)

	var err error

	orgMembershipStore := stores.NewDbOrganizationMembershipStore(tx)
	orgStore := stores.NewDbOrganizationStore(tx)
	projectStore := stores.NewDbProjectStore(tx)

	oExample := &domain.Organization{Name: "Example"}
	oExample.Uuid, err = orgStore.Create(oExample)
	if err != nil {
		t.Fatal(err)
	}

	pExample := &domain.Project{
		Name:             "Example",
		OrganizationUuid: oExample.Uuid,
	}
	pExample.Uuid, err = projectStore.Create(pExample)
	if err != nil {
		t.Fatal(err)
	}

	oExampleMembership := &domain.OrganizationMembership{
		OrganizationUuid: oExample.Uuid,
		UserUuid:         u.Uuid,
		Type:             domain.MembershipTypeOwner,
	}
	if err = orgMembershipStore.Create(oExampleMembership); err != nil {
		t.Fatal(err)
	}

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/organizations/"+oExample.Uuid+"/projects", "")
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

	type item struct {
		Subject domain.Organization          `json:"subject"`
		Links   map[string]map[string]string `json:"_links"`
	}

	var container struct {
		Collection []item                       `json:"collection"`
		Meta       map[string]string            `json:"_meta"`
		Links      map[string]map[string]string `json:"_links"`
	}

	if err = json.Unmarshal(body, &container); err != nil {
		t.Fatal(err)
	}

	if len(container.Collection) != 1 {
		t.Fatal("Expected a single item in the collection")
	}

	if container.Collection[0].Subject.Uuid != pExample.Uuid {
		t.Fatal("Expected to get the Project's UUID in the collection's first subject's UUID field")
	}

}

func Test_OrganizationHandler_Members_ReturnsOrganizationMembers(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountOrganizationHandler, t)
	defer ctxt.Tx().Rollback()
	defer ts.Close()

	tx := ctxt.Tx()
	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	o := world.Organization("default")
	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/organizations/"+o.Uuid+"/members", "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	result := struct {
		Collection []struct {
			Subject struct {
				Name string `json:"name"`
			} `json:"subject"`
		} `json:"collection"`
	}{}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("%s while unmarshalling:\n%s\n", err, body)
	}

	actual := []string{}
	for _, item := range result.Collection {
		actual = append(actual, item.Subject.Name)
	}

	expected := []string{
		world.User("other").Name,
		world.User("default").Name,
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Expected %#v to equal %#v", actual, expected)
	}
}

func TestOrganizationHandler_Create_enqueuesOrganizationCreatedActivity_forPublicOrganizations(t *testing.T) {
	if !test_helpers.IsBraintreeProxyAvailable() {
		t.Skip("braintree-proxy not running")
		return
	}

	h := NewHandlerTest(MountOrganizationHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")

	h.Do("POST", h.Url("/organizations"), &halWrapper{
		Subject: OrgParams{
			Name:   "test-organization",
			Public: true,
		},
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "organization.created" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "organization.created")
}

func TestOrganizationHandler_Create_enqueuesOrganizationCreatedActivity_forPrivateOrganizations(t *testing.T) {
	if !test_helpers.IsBraintreeProxyAvailable() {
		t.Skip("braintree-proxy not running")
		return
	}

	h := NewHandlerTest(MountOrganizationHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")

	freePlanUuid := "b99a21cc-b108-466e-aa4d-bde10ebbe1f3"
	h.Do("POST", h.Url("/organizations"), &halWrapper{
		Subject: OrgParams{
			Name:     "test-organization",
			PlanUuid: freePlanUuid,
		},
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "organization.created" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "organization.created")
}

func TestOrganizationHandler_Show_returnsPlanUuidForOrganization(t *testing.T) {
	h := NewHandlerTest(MountOrganizationHandler, t)
	defer h.Cleanup()

	found := struct {
		Subject struct {
			BillingPlanUuid string
		}
	}{}
	h.ResultTo(&found)
	h.LoginAs("default")
	h.Subject(h.World().Organization("default"))
	h.Do("GET", h.UrlFor("self"), nil)
	if got, want := found.Subject.BillingPlanUuid, test_helpers.TestPlan.Uuid; got != want {
		t.Errorf(`found.Subject.PlanUuid = %v; want %v`, got, want)
	}
}
