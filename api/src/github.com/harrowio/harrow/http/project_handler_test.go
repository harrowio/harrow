package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
	"github.com/harrowio/harrow/uuidhelper"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

func Test_ProjectHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountProjectHandler(r, nil)
	spec := routingSpec{
		{"PUT", "/projects", "project-update"},
		{"POST", "/projects", "project-create"},
		{"GET", "/projects/:uuid", "project-show"},
		{"DELETE", "/projects/:uuid", "project-archive"},

		{"GET", "/projects/:uuid/jobs", "project-jobs"},
		{"GET", "/projects/:uuid/environments", "project-environments"},
		{"GET", "/projects/:uuid/repositories", "project-repositories"},
		{"GET", "/projects/:uuid/tasks", "project-tasks"},
		{"GET", "/projects/:uuid/operations", "project-operations"},
		{"GET", "/projects/:uuid/git-triggers", "project-git-triggers"},
		{"GET", "/projects/:uuid/memberships", "project-memberships"},
		{"GET", "/projects/:uuid/members", "project-members"},
		{"DELETE", "/projects/:uuid/members", "project-leave"},
		{"GET", "/projects/:uuid/webhooks", "project-webhooks"},
		{"GET", "/projects/:uuid/schedules", "project-schedules"},
		{"GET", "/projects/:uuid/scheduled-executions", "project-scheduled-executions"},
		{"GET", "/projects/:uuid/job-notifiers", "project-job-notifiers"},
		{"GET", "/projects/:uuid/slack-notifiers", "project-slack-notifiers"},
		{"GET", "/projects/:uuid/email-notifiers", "project-email-notifiers"},
	}
	spec.run(r, t)
}

func Test_ProjectHandler_401Unauthorized_WhenUnauthenticated(t *testing.T) {
	t.Skip("Enable once authz is implemented")

	ts, ctxt := setupHandlerTestServer(MountProjectHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)

	o := world.Organization("default")

	req, err := newRequestJSON("POST", ts.URL+"/projects", &ProjectPararamsWrapper{
		Subject: ProjectParams{
			Name:             "Test Project",
			OrganizationUuid: o.Uuid,
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if expected, status := 401, res.StatusCode; status != expected {
		t.Fatalf("Expected status %v, got %v", expected, status)
	}
	ctxt.authz.Expect(t, "create", 1)
}

func Test_ProjectHandler_201Created_OnSuccess(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountProjectHandler, t)
	tx := ctxt.Tx()
	defer func(t *testing.T, tx *sqlx.Tx) { tx.Rollback() }(t, tx)
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	o := world.Organization("default")
	u := world.User("default")
	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequestJSON(
		s.Uuid,
		"POST",
		ts.URL+"/projects",
		&ProjectPararamsWrapper{
			Subject: ProjectParams{
				Name:             "Test Project",
				OrganizationUuid: o.Uuid,
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if expected, status := http.StatusCreated, res.StatusCode; status != expected {
		t.Fatalf("Expected status %v, got %v", expected, status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	var item struct {
		Subject domain.Project               `json:"subject"`
		Links   map[string]map[string]string `json:"_links"`
	}

	if err = json.Unmarshal(body, &item); err != nil {
		t.Fatal(err)
	}

	if item.Subject.OrganizationUuid != o.Uuid {
		t.Fatal("Expected item.Subject.OrganizationUuid to equal", o.Uuid)
	}
}

func Test_ProjectHandler_401Unauthorized_WhenNonMemberManagerOfOrganization(t *testing.T) {
	t.Skip("Enable once authz is implemented")

	ts, ctxt := setupHandlerTestServer(MountProjectHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	o := world.Organization("default")
	nonMember := world.User("non-member")
	s := setupTestLoginSession(t, tx, nonMember)

	req, err := newAuthenticatedRequest(s.Uuid, "POST", ts.URL+"/projects", `{"subject": {"name":"Test Project", "organizationUuid":"`+o.Uuid+`"}}`)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if expected, status := http.StatusUnauthorized, res.StatusCode; status != expected {
		t.Fatalf("Expected status %v, got %v", expected, status)
	}
	ctxt.authz.Expect(t, "create", 1)
}

func Test_ProjectShowHandler_Status200OkOnUnauthenticatedPublicProject(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountProjectHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)

	p := world.Project("public")

	req, err := newRequest("GET", ts.URL+"/projects/"+p.Uuid, "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if expected, status := http.StatusOK, res.StatusCode; status != expected {
		t.Fatalf("Expected status %v, got %v", expected, status)
	}
	ctxt.authz.Expect(t, "read", 1)
}

func Test_ProjectShowHandler_Status200OkOnFoundProject(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountProjectHandler, t)
	tx := ctxt.Tx()
	defer func(t *testing.T, tx *sqlx.Tx) { tx.Rollback() }(t, tx)
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	p := world.Project("public")
	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/projects/"+p.Uuid, "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if expected, status := http.StatusOK, res.StatusCode; status != expected {
		t.Fatalf("Expected status %v, got %v", expected, status)
	}
}

func setupProjectRepositoryHandlerTest(t *testing.T, tx *sqlx.Tx) (*domain.User, *domain.Project, *domain.Repository) {

	u := createTestUser(t, tx)

	o := test_helpers.MustCreateOrganization(t, tx, &domain.Organization{Name: "Test Organization"})
	p := test_helpers.MustCreateProject(t, tx, &domain.Project{Name: "Test Project", OrganizationUuid: o.Uuid})

	r := test_helpers.MustCreateRepository(t, tx, &domain.Repository{
		Uuid:        "33333333-3333-4333-a333-333333333333",
		ProjectUuid: p.Uuid,
		Name:        "Git Core Tools",
		Url:         "git://git.kernel.org/pub/scm/git/git.git",
	})

	_ = test_helpers.MustCreateOrganizationMembership(t, tx, &domain.OrganizationMembership{
		UserUuid:         u.Uuid,
		OrganizationUuid: o.Uuid,
		Type:             domain.MembershipTypeMember,
	})

	return u, p, r

}

func Test_ProjectRepositoriesHandler_Status200OkOnFoundProject_Member(t *testing.T) {

	var err error

	ts, ctxt := setupHandlerTestServer(MountProjectHandler, t)
	tx := ctxt.Tx()
	u, p, r := setupProjectRepositoryHandlerTest(t, tx)
	defer tx.Rollback()
	defer ts.Close()

	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/projects/"+p.Uuid+"/repositories", "")
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
		Subject domain.Repository            `json:"subject"`
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

	if container.Collection[0].Subject.Uuid != r.Uuid {
		t.Fatal("Expected to get the Repository's UUID in the collection's first subject's UUID field")
	}

}

func Test_ProjectRepositoriesHandler_Status403ForbiddenOnFoundProject_ForeignOrUnauthenticated(t *testing.T) {
	t.Skip("Enable once authz is implemented")

	ts, ctxt := setupHandlerTestServer(MountProjectHandler, t)
	tx := ctxt.Tx()
	setupProjectRepositoryHandlerTest(t, tx)
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)

	p := world.Project("public")
	s := setupTestLoginSession(t, tx, setupOtherTestUser(tx, t))

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/projects/"+p.Uuid+"/repositories", "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if expected, status := http.StatusForbidden, res.StatusCode; status != expected {
		t.Fatalf("Expected status %v, got %v", expected, status)
	}
}

func Test_ProjectRepositoriesHandler_Status404NonExistentProject(t *testing.T) {
	var err error

	ts, ctxt := setupHandlerTestServer(MountProjectHandler, t)
	tx := ctxt.Tx()
	u, _, _ := setupProjectRepositoryHandlerTest(t, tx)
	defer tx.Rollback()
	defer ts.Close()

	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/projects/0b8d0449-080e-4b45-aebc-937d6b810dfd/repositories", "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if expected, status := http.StatusNotFound, res.StatusCode; status != expected {
		t.Fatalf("Expected status %v, got %v", expected, status)
	}
}

func Test_ProjectHandler_ProjectMemberships_ReturnsExistingMemberships(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountProjectHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	project := world.Project("public")
	membership := world.ProjectMembership("member")

	s := setupTestLoginSession(t, tx, user)
	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/projects/"+project.Uuid+"/memberships", "")
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
				Uuid string `json:"uuid"`
			} `json:"subject"`
		} `json:"collection"`
	}{}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err, string(body))
	}

	for _, m := range result.Collection {
		if m.Subject.Uuid == membership.Uuid {
			return
		}
	}

	t.Fatalf("Membership with uuid %q not found in result:\n%s\n", membership.Uuid, string(body))
}

func Test_ProjectHandler_ProjectMemberships_RequiresAuthorization(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountProjectHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	project := world.Project("public")

	s := setupTestLoginSession(t, tx, user)
	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/projects/"+project.Uuid+"/memberships", "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	ctxt.authz.Expect(
		t, "read",
		1+1, // project + number of memberships
	)
}

func Test_ProjectHandler_ProjectMembers_ReturnsExistingMemberships(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountProjectHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	project := world.Project("public")

	s := setupTestLoginSession(t, tx, user)
	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/projects/"+project.Uuid+"/members", "")
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
				Uuid string `json:"uuid"`
			} `json:"subject"`
		} `json:"collection"`
	}{}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err, string(body))
	}

	for _, m := range result.Collection {
		if m.Subject.Uuid == user.Uuid {
			return
		}
	}

	t.Fatalf("User with uuid %q not found in result:\n%s\n", user.Uuid, string(body))
}

func Test_ProjectHandler_ProjectMembers_RequiresAuthorization(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountProjectHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	project := world.Project("public")

	s := setupTestLoginSession(t, tx, user)
	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/projects/"+project.Uuid+"/members", "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	ctxt.authz.Expect(
		t, "read",
		1+3, // project + number of members
	)
}

func Test_ProjectHandler_ScheduledExecutions(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountProjectHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	project := world.Project("public")

	s := setupTestLoginSession(t, tx, user)

	url := fmt.Sprintf("%s/projects/%s/scheduled-executions?from=%s&to=%s", ts.URL, project.Uuid, "2000-01-01T00:00:00Z", "2000-01-01T23:59:59Z")
	req, err := newAuthenticatedRequest(s.Uuid, "GET", url, "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Fatalf("res.StatusCode=%d, want %d", res.StatusCode, 200)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Collection []struct {
			Subject domain.ScheduledExecution `json:"subject"`
		} `json:"collection"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err, string(body))
	}
	if len(result.Collection) != 2 {
		t.Fatalf("len(result.Collection)=%d, want %d", len(result.Collection), 2)
	}
	if result.Collection[0].Subject.JobUuid != world.Job("default").Uuid {
		t.Errorf("result.Collection[0].Subject.JobUuid=%s, want %s", result.Collection[0].Subject.JobUuid, world.Job("default").Uuid)
	}
	if result.Collection[1].Subject.JobUuid != world.Job("other").Uuid {
		t.Errorf("result.Collection[1].Subject.JobUuid=%s, want %s", result.Collection[1].Subject.JobUuid, world.Job("other").Uuid)
	}
}
func Test_ProjectHandler_Webhooks_returns404IfProjectNotFound(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountProjectHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")

	doesNotExist := uuidhelper.MustNewV4()

	s := setupTestLoginSession(t, tx, user)
	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/projects/"+doesNotExist+"/webhooks", "")
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

func Test_ProjectHandler_Webhooks_returnsWebhooksForProject(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountProjectHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	ctxt.u = user
	project := world.Project("public")
	job := world.Job("default")
	webhookStore := stores.NewDbWebhookStore(tx)
	expectedWebhookIds := []string{}
	for i := 0; i < 3; i++ {

		webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, fmt.Sprintf("test-hook-%d", i))
		if _, err := webhookStore.Create(webhook); err != nil {
			t.Fatal(err)
		}
		expectedWebhookIds = append(expectedWebhookIds, webhook.Uuid)
	}

	s := setupTestLoginSession(t, tx, user)
	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/projects/"+project.Uuid+"/webhooks", "")
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
				Uuid string
			}
		}
	}{}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatal(err)
	}

	for _, actual := range result.Collection {
		if uuid := actual.Subject.Uuid; !test_helpers.InStrings(uuid, expectedWebhookIds) {
			t.Errorf("unexpected webhook: %q", uuid)
		}
	}
}

func Test_ProjectHandler_Schedules_returnsSchedules(t *testing.T) {
	h := NewHandlerTest(MountProjectHandler, t)
	defer h.Cleanup()
	project := h.World().Project("public")
	h.Subject(project)
	result := struct {
		Collection []struct {
			Subject *domain.Schedule
		}
	}{}
	h.ResultTo(&result)
	h.Do("GET", h.UrlFor("schedules"), nil)
	if have, want := len(result.Collection), 2; have != want {
		t.Fatalf("len(result.Collection); have=%d, want=%d", have, want)
	}
	schedules := []*domain.Schedule{h.World().Schedule("default"), h.World().Schedule("other")}
	if result.Collection[0].Subject.Uuid != schedules[0].Uuid && result.Collection[1].Subject.Uuid != schedules[0].Uuid {
		t.Errorf("Could not find %s in the returned Schedules", schedules[0].Uuid)
	}
	if result.Collection[0].Subject.Uuid != schedules[1].Uuid && result.Collection[1].Subject.Uuid != schedules[1].Uuid {
		t.Errorf("Could not find %s in the returned Schedules", schedules[1].Uuid)
	}
}

func Test_ProjectHandler_ScheduledExecutions_returnsProperErrorForMalformedTimestamp(t *testing.T) {
	h := NewHandlerTest(MountProjectHandler, t)
	defer h.Cleanup()
	project := h.World().Project("public")
	h.Subject(project)
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

func TestProjectHandler_Create_enqueuesProjectCreatedActivity(t *testing.T) {
	h := NewHandlerTest(MountProjectHandler, t)
	defer h.Cleanup()

	organization := h.World().Organization("default")

	h.LoginAs("default")
	h.Do("POST", h.Url("/projects"), &ProjectPararamsWrapper{
		Subject: ProjectParams{
			Public:           false,
			Name:             "test project",
			OrganizationUuid: organization.Uuid,
		},
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "project.created" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "project.created")
}

func Test_ProjectHandler_Create_creates_default_environment_for_project(t *testing.T) {
	h := NewHandlerTest(MountProjectHandler, t)
	defer h.Cleanup()

	organization := h.World().Organization("default")
	project := &domain.Project{}
	h.LoginAs("default")
	h.ResultTo(&struct {
		Subject domain.Subject `json:"subject"`
	}{
		Subject: project,
	})
	h.Do("POST", h.Url("/projects"), &ProjectPararamsWrapper{
		Subject: ProjectParams{
			Public:           false,
			Name:             "test project",
			OrganizationUuid: organization.Uuid,
		},
	})

	environments, err := stores.NewDbEnvironmentStore(h.Tx()).FindAllByProjectUuid(project.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(environments), 1; got != want {
		t.Fatalf(`len(environments) = %v; want %v`, got, want)
	}

	if got, want := environments[0].Name, "Default"; got != want {
		t.Errorf(`environments[0].Name = %v; want %v`, got, want)
	}

	if got, want := environments[0].IsDefault, true; got != want {
		t.Errorf(`environments[0].IsDefault = %v; want %v`, got, want)
	}
}
