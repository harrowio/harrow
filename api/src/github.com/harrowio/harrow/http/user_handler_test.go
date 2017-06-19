package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/harrowio/harrow/clock"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
	"github.com/harrowio/harrow/uuidhelper"
)

func Test_UserHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountUserHandler(r, nil)

	spec := routingSpec{
		{"POST", "/users", "user-create"},
		{"PUT", "/users", "user-update"},
		{"GET", "/users/:uuid/oauth-tokens", "user-oauth-tokens"},
		{"GET", "/users/:uuid/organizations", "user-organizations"},
		{"GET", "/users/:uuid/sessions", "user-sessions"},
		{"GET", "/users/:uuid/projects", "user-projects"},
		{"GET", "/users/:uuid/jobs", "user-jobs"},
		{"GET", "/users/:uuid/blocks", "user-blocks"},
		{"GET", "/users/:uuid", "user-show"},
		{"POST", "/users/:uuid/verify-email", "user-verify-email"},
		{"PATCH", "/users/:uuid/mfa", "user-change-mfa"},
	}

	spec.run(r, t)
}

func Test_UserHandler_Projects_returnsProjectsThroughProjectMemberships(t *testing.T) {
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	world := h.World()
	tx := h.Tx()
	h.LoginAs("non-member")
	project := world.Project("private")
	user := h.User()
	test_helpers.MustCreateProjectMembership(t, tx, project.NewMembership(user, domain.MembershipTypeMember))

	response := struct {
		Collection []struct {
			Subject struct {
				Uuid string
			}
		}
	}{}

	h.Subject(user)
	h.ResultTo(&response)
	h.Do("GET", h.UrlFor("projects"), nil)

	want := project.Uuid
	found := []string{}
	for _, item := range response.Collection {
		got := item.Subject.Uuid
		found = append(found, got)
		if got == want {
			return
		}
	}

	t.Errorf("response.Collection = %v; does not contain %s", found, want)
}

func Test_UserHandler_Projects_returnsOnlyProjectsThroughProjectMemberships_whenMembershipOnly_is_Set(t *testing.T) {
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	world := h.World()
	tx := h.Tx()
	h.LoginAs("non-member")
	project := world.Project("private")
	user := h.User()
	test_helpers.MustCreateProjectMembership(t, tx, project.NewMembership(user, domain.MembershipTypeMember))

	response := struct {
		Collection []struct {
			Subject struct {
				Uuid string
			}
		}
	}{}

	h.Subject(user)
	h.ResultTo(&response)
	h.Do("GET", h.UrlFor("projects"), url.Values{"membershipOnly": []string{"yes"}})

	projectMemberships, err := stores.NewDbProjectStore(tx).FindAllByUserUuidOnlyThroughProjectMemberships(user.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(response.Collection), len(projectMemberships); got != want {
		t.Errorf(`len(response.Collection) = %v; want %v`, got, want)
	}
}

func Test_UserProjectsHandler_ReturnsPublicProjectsOnly_WhenForeignUser(t *testing.T) {
	t.Skip("enable once authz is implemented")

	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	defer ctxt.Tx().Rollback()
	defer ts.Close()

	tx := ctxt.Tx()
	world := test_helpers.MustNewWorld(tx, t)
	u1 := world.User("default")
	u2 := world.User("non-member")
	project := world.Project("public")
	s := setupTestLoginSession(t, tx, u2)

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/users/"+u1.Uuid+"/projects", "")
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

	var container struct {
		Collection []struct {
			Subject *domain.Project              `json:"subject"`
			Links   map[string]map[string]string `json:"_links"`
		} `json:"collection"`
		Meta  map[string]string            `json:"_meta"`
		Links map[string]map[string]string `json:"_links"`
	}

	if err = json.Unmarshal(body, &container); err != nil {
		t.Fatalf("%s in:\n%s\n", err, string(body))
	}

	if len(container.Collection) != 1 {
		t.Fatalf("Expected a single item in the collection, got %d", len(container.Collection))
	}

	if container.Collection[0].Subject.Uuid != project.Uuid {
		t.Fatal("Expected to get the Project's UUID in the collection's first subject's UUID field")
	}

}

func Test_UserShowHandler_Status200OkOnFoundUser_ForeignOrUnauthenticated(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/users/"+u.Uuid, "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Fatal("Expected Status 200, got:", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	var response = struct {
		Subject domain.User
	}{}

	err = json.Unmarshal(body, &response)
	if err != nil {
		t.Fatal(err)
	}

	if response.Subject.Uuid != u.Uuid {
		t.Fatal("API UUID Incorrect")
	}
	if response.Subject.Email != u.Email {
		t.Fatal("API Email Incorrect")
	}
	if response.Subject.Name != u.Name {
		t.Fatal("API Name Incorrect")
	}
	if response.Subject.CreatedAt.IsZero() {
		t.Fatal("Expected CreatedAt to be set")
	}
	if response.Subject.TotpSecret != "" {
		t.Fatal("Expected TotpSecret not to be set")
	}
	if response.Subject.PasswordResetToken != "" {
		t.Fatal("Expected PasswordResetToken not to be set")
	}

}

func Test_UserShowHandler_Status404NotFoundOnNotFoundUser(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)

	u := world.User("default")
	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/users/4632e5da-081a-4cbc-baba-4a2b84a5d200", "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 404 {
		t.Fatal("Expected Status 404, got:", res.StatusCode)
	}

}

func Test_UserCreateHandler_Status400OnMalformedInput(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	req, err := newRequest("POST", ts.URL+"/users", `{"subject": {"max@musterma.nn", "password":"changeme123"}}`)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusBadRequest {
		t.Fatal("Expected StatusBadRequest, got:", res.StatusCode)
	}

}

func Test_UserCreateHandler_Status201OnSuccess(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	req, err := newRequest("POST", ts.URL+"/users", `{"subject": {"email": "max@musterma.nn", "password":"changeme123"}}`)
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
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected StatusCreated, got: %d\nBody:\n%s\n", res.StatusCode, body)
	}

	if len(body) == 0 {
		t.Fatal("body should be non-empty")
	}

	location, ok := res.Header["Location"]
	if !ok {
		t.Fatal("location header should be set")
	}
	if !strings.Contains(location[0], ts.URL+"/users/") {
		t.Fatal(location, "does not match the expected header")
	}

}

func Test_UserCreateHandler_Status422OnValidationError(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	params := createUserParams{Email: u.Email, Password: "changeme123"}

	req, err := newRequestJSON("POST", ts.URL+"/users/", &createUserParamsWrapper{params})
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 422 {
		t.Fatal("Expected Status #422, got:", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(body) == 0 {
		t.Fatal("body should contain a JSON representation of the validation error")
	}

	_, ok := res.Header["Location"]
	if ok {
		t.Fatal("location header should not be set")
	}

}

func Test_UserPatchHandler_Status403OnSessionAndUserIdMismatch(t *testing.T) {
	t.Skip("enable once authz is implemented")

	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u1 := world.User("default")
	u2 := world.User("other")
	s := setupTestLoginSession(t, tx, u1)

	req, err := newAuthenticatedRequest(s.Uuid, "PATCH", ts.URL+"/users/"+u2.Uuid, `{"twoFactorAuthEnabled": true}`)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusForbidden {
		t.Fatal("Expected StatusForbidden, got:", res.StatusCode)
	}
	ctxt.authz.Expect(t, "update", 1)

}

func Test_UserPatchHandler_Status400OnMalformedInput(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequest(s.Uuid, "PATCH", ts.URL+"/users/"+u.Uuid+"/mfa", `{"twoFactorAutrue}`)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusBadRequest {
		t.Fatal("Expected StatusBadRequest, got:", res.StatusCode)
	}

}

func Test_UserHandler_Show_responseContainsTotpEnabled(t *testing.T) {
	test_helpers.Flaky(t)
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()
	h.LoginAs("default")
	now := time.Now()
	u := h.User()
	h.Subject(u)
	h.Tx().MustExec(`UPDATE users SET totp_enabled_at = $1 WHERE uuid = $2`, now, u.Uuid)

	result := struct {
		Subject struct {
			TotpEnabledAt *time.Time `json:"totpEnabledAt"`
		}
	}{}
	h.ResultTo(&result)

	h.Do("GET", h.UrlFor("self"), nil)

	if got := result.Subject.TotpEnabledAt; got == nil {
		t.Fatalf("User(%q): not present", "TotpEnabled")
	}

	// Convert to Unix time because otherwise the clock resolution
	// in Go is higher than in Postgres.
	if got := result.Subject.TotpEnabledAt.Unix(); got != now.Unix() {
		t.Errorf("User(%q) = %s; want: %s", "TotpEnabled", got, now.Unix())
	}
}

func Test_UserHandler_Show_doesNotExposeTotpSecret(t *testing.T) {
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	u := h.User()
	h.Subject(u)
	h.Tx().MustExec(`UPDATE users SET totp_secret = $1 WHERE uuid = $2`, `abc`, u.Uuid)

	result := struct {
		Subject struct {
			TotpSecret *string `json:"totpSecret"`
		}
	}{}
	h.ResultTo(&result)

	h.Do("GET", h.UrlFor("self"), nil)

	if got := result.Subject.TotpSecret; got != nil {
		t.Fatalf("User(%q) = %q; want %q", "TotpSecret", got, "")
	}
}

func Test_UserHandler_Patch_generatesNewTotpSecret(t *testing.T) {
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	u := h.User()
	want := domain.RandomTotpSecret()
	u.TotpSecret = want
	if err := stores.NewDbUserStore(h.Tx(), h.Config()).Update(u); err != nil {
		t.Fatal(err)
	}
	h.Subject(u)

	result := struct {
		Subject struct {
			TotpSecret string `json:"totpSecret"`
		}
	}{}
	h.ResultTo(&result)

	h.Do("PATCH", h.UrlFor("mfa"), struct {
		TwoFactorAuthEnabled bool `json:"twoFactorAuthEnabled"`
	}{true})

	if got := result.Subject.TotpSecret; got == want {
		t.Fatalf("User(%q) = %q; want not %q", "TotpSecret", got, want)
	}
}

func Test_UserHandler_Patch_Returns200IfTotpTokenIsValid(t *testing.T) {
	test_helpers.Flaky(t)
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()
	now := time.Now()
	domain.Clock = clock.At(now)
	defer func() { domain.Clock = clock.System }()
	world := h.World()
	userStore := stores.NewDbUserStore(h.Tx(), h.Config())
	u := world.User("default")
	u.GenerateTotpSecret()
	if err := userStore.Update(u); err != nil {
		t.Fatal(err)
	}

	h.LoginAs("default")
	h.Subject(u)
	token := u.CurrentTotpToken()
	h.Do("PATCH", h.UrlFor("mfa"), &patchUserParams{
		TwoFactorAuthEnabled: true,
		TotpToken:            &token,
	})

	if got, want := h.Response().StatusCode, http.StatusOK; got != want {
		t.Errorf("h.Response().StatusCode = %d; want %d", got, want)
		t.Logf("Response body:\n%s\n", h.ResponseBody())
	}
}

func Test_UserHandler_Patch_SetsTotpEnabled_IfTotpTokenIsValid(t *testing.T) {
	test_helpers.Flaky(t)
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()
	now := time.Now()
	domain.Clock = clock.At(now)
	defer func() { domain.Clock = clock.System }()

	world := h.World()
	userStore := stores.NewDbUserStore(h.Tx(), h.Config())
	u := world.User("default")
	u.GenerateTotpSecret()
	if err := userStore.Update(u); err != nil {
		t.Fatal(err)
	}

	h.LoginAs("default")
	h.Subject(u)
	token := u.CurrentTotpToken()

	h.Do("PATCH", h.UrlFor("mfa"), &patchUserParams{
		TwoFactorAuthEnabled: true,
		TotpToken:            &token,
	})

	reloaded, err := userStore.FindByUuid(u.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := reloaded.TotpEnabled(), true; got != want {
		t.Errorf("reloaded.TotpEnabled() = %v; want %v", got, want)
	}
}

func Test_UserHandler_Patch_ReturnsSecretInBody_IfTotpTokenIsValid(t *testing.T) {
	test_helpers.Flaky(t)
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()
	h.LoginAs("default")
	now := time.Now()
	domain.Clock = clock.At(now)
	defer func() { domain.Clock = clock.System }()

	u := h.User()
	u.GenerateTotpSecret()
	userStore := stores.NewDbUserStore(h.Tx(), h.Config())

	if err := userStore.Update(u); err != nil {
		t.Fatal(err)
	}

	result := struct{ Subject struct{ TotpSecret string } }{}
	h.ResultTo(&result)
	token := u.CurrentTotpToken()
	h.Subject(u)
	h.Do(
		"PATCH",
		h.UrlFor("mfa"),
		&patchUserParams{
			TwoFactorAuthEnabled: true,
			TotpToken:            &token,
		},
	)

	reloaded, err := userStore.FindByUuid(u.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := result.Subject.TotpSecret, reloaded.TotpSecret; got != want {
		t.Errorf("result.Subject.TotpSecret = %q; want %q", got, want)
	}
}

func Test_UserHandler_Patch_Returns422IfTotpTokenIsInvalid(t *testing.T) {
	test_helpers.Flaky(t)
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()
	h.LoginAs("default")
	u := h.User()
	h.Subject(u)
	now := time.Now()
	domain.Clock = clock.At(now)
	defer func() { domain.Clock = clock.System }()

	u.GenerateTotpSecret()

	h.Do("PATCH", h.UrlFor("mfa"), struct {
		TwoFactorAuthEnabled bool  `json:"twoFactorAuthEnabled"`
		Totp                 int32 `json:"totp"`
	}{
		true,
		123456,
	})

	if got, want := h.Response().StatusCode, StatusUnprocessableEntity; got != want {
		t.Fatalf("response.StatusCode = %d; want %d", got, want)
	}
}

func Test_UserHandler_Patch_DisablesTotp(t *testing.T) {
	test_helpers.Flaky(t)
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()
	h.LoginAs("default")
	now := time.Now()
	domain.Clock = clock.At(now)
	defer func() { domain.Clock = clock.System }()

	u := h.User()
	u.GenerateTotpSecret()

	userStore := stores.NewDbUserStore(h.Tx(), h.Config())

	u.EnableTotp(u.CurrentTotpToken())
	if err := userStore.Update(u); err != nil {
		t.Fatal(err)
	}

	h.Subject(u)
	token := u.CurrentTotpToken()
	h.Do("PATCH", h.UrlFor("mfa"), &patchUserParams{
		TwoFactorAuthEnabled: false,
		TotpToken:            &token,
	})

	changedUser, err := userStore.FindByUuid(u.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := changedUser.TotpEnabled(), false; got != want {
		t.Fatalf("User(%q) = %v; want %v", "TotpEnabled", got, want)
	}

}

func Test_UserOrganizationsHandler_Status200OkOnFoundUser_Owner(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	o := world.Organization("default")
	s := setupTestLoginSession(t, tx, u)

	var err error

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/users/"+u.Uuid+"/organizations", "")
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
		t.Fatal("Expected a single item in the collection, got %d", len(container.Collection))
	}

	if container.Collection[0].Subject.Uuid != o.Uuid {
		t.Fatal("Expected to get the Organization's UUID in the collection's first subject's UUID field")
	}

}

func Test_UserOrganizationsHandler_Status200OkOnFoundUser_ForeignOrUnauthenticated(t *testing.T) {
	// NOTE(dh): removed because I don't know what this should test.
}

func Test_UserHandler_Organizations_returns_all_organization_the_user_is_a_member_of(t *testing.T) {
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	tx := h.Tx()
	user := test_helpers.MustCreateUser(t, tx, &domain.User{
		Name:     "John Doe",
		Email:    "john-doe@example.com",
		Password: "long-password",
	})

	organizations := []*domain.Organization{
		test_helpers.MustCreateOrganization(t, tx, &domain.Organization{Name: "A"}),
		test_helpers.MustCreateOrganization(t, tx, &domain.Organization{Name: "B"}),
	}

	project := test_helpers.MustCreateProject(t, tx, &domain.Project{
		Name:             "B Project",
		OrganizationUuid: organizations[1].Uuid,
	})

	test_helpers.MustCreateOrganizationMembership(t, tx, &domain.OrganizationMembership{
		UserUuid:         user.Uuid,
		OrganizationUuid: organizations[0].Uuid,
		Type:             domain.MembershipTypeOwner,
	})

	test_helpers.MustCreateProjectMembership(t, tx, &domain.ProjectMembership{
		ProjectUuid:    project.Uuid,
		UserUuid:       user.Uuid,
		MembershipType: domain.MembershipTypeManager,
	})

	result := []struct {
		Subject struct{ Uuid string }
	}{}
	wrapper := &struct {
		Collection *[]struct{ Subject struct{ Uuid string } }
	}{&result}
	h.Subject(user)
	h.ResultTo(wrapper)
	h.Do("GET", h.UrlFor("organizations"), nil)

	if got, want := len(result), 2; got != want {
		t.Fatalf(`len(result) = %v; want %v`, got, want)
	}

}

func Test_UserOrganizationsHandler_Status404_NotFoundNonExistentUuid(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/users/0d56e3b5-fdcd-424a-bfa4-7da814339f91/organizations", "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNotFound {
		t.Fatal("Expected StatusNotFound, got:", res.StatusCode)
	}

}

func Test_UserProjectsHandler_Status200OkOnFoundUser_Owner(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	pPublic := world.Project("public")
	pPrivate := world.Project("private")
	s := setupTestLoginSession(t, tx, u)

	var err error

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/users/"+u.Uuid+"/projects", "")
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
		Subject domain.Project               `json:"subject"`
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

	expected := map[string]*domain.Project{
		pPublic.Uuid:  pPublic,
		pPrivate.Uuid: pPrivate,
	}
	if n := len(container.Collection); n != len(expected) {
		t.Fatalf("Expected collection to have %d elements, got %d", len(expected), n)
	}
	for _, p := range container.Collection {
		if expected[p.Subject.Uuid] == nil {
			t.Fatalf("Unexpected project returned: %s\n", p.Subject.Uuid)
		}
	}
}

// TODO(dh): this test is actually testing that only public projects
// are returned.  What is the correct behaviour here?
func Test_UserProjectsHandler_Status403Forbidden_ForeignOrUnauthenticated(t *testing.T) {
	t.Skip("enable once authz is implemented")

	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	foreign := world.User("non-member")
	pPublic := world.Project("public")
	s := setupTestLoginSession(t, tx, foreign)

	var err error

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/users/"+u.Uuid+"/projects", "")
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
		Subject domain.Project               `json:"subject"`
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

	expected := map[string]*domain.Project{
		pPublic.Uuid: pPublic,
	}
	if n := len(container.Collection); n != len(expected) {
		t.Fatalf("Expected collection to have %d elements, got %d", len(expected), n)
	}
	for _, p := range container.Collection {
		if expected[p.Subject.Uuid] == nil {
			t.Fatalf("Unexpected project returned: %s\n", p.Subject.Uuid)
		}
	}
}
func Test_UserProjectsHandler_Status404_NotFoundNonExistentUuid(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/users/0d56e3b5-fdcd-424a-bfa4-7da814339f91/projects", "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNotFound {
		t.Fatal("Expected StatusNotFound, got:", res.StatusCode)
	}

}

func Test_UserSessionsHandler_Status200OkOnFoundUser_Owner(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	s := setupTestLoginSession(t, tx, u)

	var err error

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/users/"+u.Uuid+"/sessions", "")
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
		Collection []item            `json:"collection"`
		Meta       map[string]string `json:"_meta"`
		Links      map[string]string `json:"_links"`
	}

	if err = json.Unmarshal(body, &container); err != nil {
		t.Fatal(err)
	}

	if len(container.Collection) != 1 {
		t.Fatal("Expected a single item in the collection")
	}

	if container.Collection[0].Subject.Uuid != s.Uuid {
		t.Fatal("Expected to get the Sessions's UUID in the collection's first subject's UUID field")
	}

}

func Test_UserSessionsHandler_Status403ForbiddenOnFoundUser_ForeignOrUnauthenticated(t *testing.T) {
	t.Skip("enable once authz is implemented")

	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u1 := world.User("default")
	u2 := world.User("other")
	s := setupTestLoginSession(t, tx, u2)

	var err error

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/users/"+u1.Uuid+"/sessions", "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusForbidden {
		t.Fatal("Expected StatusForbidden, got:", res.StatusCode)
	}

}

func Test_UserSessionsHandler_Status404NotFound_NonExistentUser(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	s := setupTestLoginSession(t, tx, u)

	var err error

	req, err := newAuthenticatedRequest(s.Uuid, "GET", ts.URL+"/users/766c3cbc-4141-4377-ac7c-82b5dbf38205/sessions", "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNotFound {
		t.Fatal("Expected StatusNotFound got:", res.StatusCode)
	}

}

func Test_UserHandler_Create_UsesUuidFromInvitation_IfInvitationPresent(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountUserHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	inviter := world.User("default")
	project := world.Project("public")

	userId := uuidhelper.MustNewV4()
	invitation := test_helpers.MustCreateInvitation(t, tx, &domain.Invitation{
		Email:            "vagrant+test-user-creation@localhost",
		RecipientName:    "Vagrant Test User",
		Message:          "Join the dark side!",
		OrganizationUuid: project.OrganizationUuid,
		ProjectUuid:      project.Uuid,
		InviteeUuid:      userId,
		CreatorUuid:      inviter.Uuid,
		MembershipType:   domain.MembershipTypeMember,
	})

	params := createUserParams{
		Email:          "vagrant+test-user-creation@localhost",
		Password:       "changeme123",
		InvitationUuid: invitation.Uuid,
	}

	req, err := newRequestJSON("POST", ts.URL+"/users/", &createUserParamsWrapper{params})
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected Status %d, got: %d", http.StatusCreated, res.StatusCode)
	}

	config := ctxt.Config()
	userStore := stores.NewDbUserStore(tx, &config)
	_, err = userStore.FindByUuid(userId)
	if _, ok := err.(*domain.NotFoundError); ok {
		t.Fatalf("Expected to find user with uuid %s\n", userId)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func Test_UserHandler_Update_changesAttributes(t *testing.T) {
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	user := h.User()

	h.Subject(user)
	h.Do("PUT", h.UrlFor("self"), &halWrapper{Subject: &updateUserParams{
		Uuid:        user.Uuid,
		Password:    "password-is-long-enough",
		NewPassword: "new-password",
		Email:       "new-email@example.com",
		Name:        "new-name",
	}})

	if got, want := h.Response().StatusCode, http.StatusOK; got != want {
		t.Errorf("h.Response().StatusCode = %d; want %d", got, want)
	}

	reloaded, err := stores.NewDbUserStore(h.Tx(), h.Config()).FindByUuid(user.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := reloaded.Email, "new-email@example.com"; got != want {
		t.Errorf("reloaded.Email = %q; want %q", got, want)
	}

	if got, want := reloaded.Name, "new-name"; got != want {
		t.Errorf("reloaded.Name = %q; want %q", got, want)
	}

	if got, want := reloaded.PasswordHash, user.PasswordHash; got == want {
		t.Errorf("reloaded.PasswordHash = %q; want not %q", got, want)
	}
}

func Test_UserHandler_Update_returnsUprocessableEntity_ifPasswordNotProvided(t *testing.T) {
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	user := h.User()
	h.Subject(user)
	h.Do("PUT", h.UrlFor("self"), &halWrapper{Subject: &updateUserParams{
		Uuid:  user.Uuid,
		Email: "new-email@example.com",
	}})

	if got, want := h.Response().StatusCode, StatusUnprocessableEntity; got != want {
		t.Errorf("h.Response().StatusCode = %d; want %d", got, want)
	}
}

func Test_UserHandler_Update_DoesntRequirePasswordsForPasswordlessUsers(t *testing.T) {
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	h.LoginAs("without_password")
	user := h.User()
	h.Subject(user)
	h.Do("PUT", h.UrlFor("self"), &halWrapper{Subject: &updateUserParams{
		Uuid:  user.Uuid,
		Email: "new-email@example.com",
	}})

	if got, want := h.Response().StatusCode, http.StatusOK; got != want {
		t.Errorf("h.Response().StatusCode = %d; want %d", got, want)
	}

	reloaded, err := stores.NewDbUserStore(h.Tx(), h.Config()).FindByUuid(user.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := reloaded.Email, "new-email@example.com"; got != want {
		t.Errorf("reloaded.Email = %q; want %q", got, want)
	}
}

func Test_UserHandler_Update_failsIfNewPasswordIsLessThan10CharactersLong(t *testing.T) {
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	user := h.User()
	h.Subject(user)
	result := &ErrorJSON{}
	h.ResultTo(result)
	h.Do("PUT", h.UrlFor("self"), &halWrapper{Subject: &updateUserParams{
		Uuid:        user.Uuid,
		Password:    "password-is-long-enough",
		Email:       "new-email@example.com",
		NewPassword: "1",
	}})

	if got, want := h.Response().StatusCode, StatusUnprocessableEntity; got != want {
		t.Errorf("h.Response().StatusCode = %d; want %d", got, want)
	}

	if _, found := result.Errors["password"]; !found {
		t.Fatalf("result.Errors[%q] unset", "password")
	}

	if got, want := result.Errors["password"][0], "too_short"; got != want {
		t.Errorf("result.Errors[%q] = %q; want %q", "password", got, want)
	}
}

func Test_UserHandler_Update_doesNotChangePassword_ifNewPasswordIsEmpty(t *testing.T) {
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	user := h.User()

	h.Subject(user)
	h.Do("PUT", h.UrlFor("self"), &halWrapper{Subject: &updateUserParams{
		Uuid:        user.Uuid,
		Password:    "password-is-long-enough",
		Email:       "new-email@example.com",
		NewPassword: "",
	}})

	if got, want := h.Response().StatusCode, http.StatusOK; got != want {
		t.Errorf("h.Response().StatusCode = %d; want %d", got, want)
	}

	reloaded, err := stores.NewDbUserStore(h.Tx(), h.Config()).FindByUuid(user.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := reloaded.PasswordHash, user.PasswordHash; got != want {
		t.Errorf("reloaded.PasswordHash = %q; want %q", got, want)
	}

}

func Test_UserHandler_Blocks_returnsAllBlocksForTheUser(t *testing.T) {
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	user := h.User()
	reasonForBlock := "testing"
	block, err := user.NewBlock(reasonForBlock)
	if err != nil {
		t.Fatal(err)
	}
	block.BlockForever(time.Now().Add(-24 * time.Hour))
	if err := stores.NewDbUserBlockStore(h.Tx()).Create(block); err != nil {
		t.Fatal(err)
	}

	h.Subject(user)
	result := struct {
		Collection []*struct{ Subject *domain.UserBlock }
	}{}
	h.ResultTo(&result)
	h.Do("GET", h.UrlFor("blocks"), nil)

	if got, want := h.Response().StatusCode, http.StatusOK; got != want {
		t.Errorf("h.Response().StatusCode = %d; want %d", got, want)
		t.Errorf("Response body:\n%s\n", h.ResponseBody())
	}

	if got, want := len(result.Collection), 1; got != want {
		t.Fatalf("len(result.Collection) = %d; want %d", got, want)
	}

	if got, want := result.Collection[0].Subject.Reason, reasonForBlock; got != want {
		t.Errorf("result.Collection[0].Subject.Reason = %q; want %q", got, want)
	}
}

func Test_UserHandler_Create_blocksUserWithReasonAlpha(t *testing.T) {
	t.Skip("Disabled as we are finishing the Alpha phase, code kept to guard against rot, it might be useful")

	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	email := "alpha-user@example.com"
	password := "the-alpha-user-password"
	result := struct{ Subject struct{ Uuid string } }{}
	h.ResultTo(&result)
	h.Do("POST", h.Url("/users"), &createUserParamsWrapper{
		Subject: createUserParams{
			Email:    email,
			Password: password,
			Name:     "α user",
		},
	})

	if got, want := h.Response().StatusCode, http.StatusCreated; got != want {
		t.Errorf("h.Response().StatusCode = %d; want %d", got, want)
		t.Fatalf("h.ResponseBody():\n%s\n", h.ResponseBody())
	}

	blocks, err := stores.NewDbUserBlockStore(h.Tx()).FindAllByUserUuid(result.Subject.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(blocks), 1; got != want {
		t.Fatalf("len(blocks) = %d; want %d", got, want)
	}

	if got, want := blocks[0].Reason, "alpha"; got != want {
		t.Fatalf("blocks[0].Reason = %q; want %q", got, want)
	}
}

func Test_UserHandler_Create_emitsUserSignedUpActivity(t *testing.T) {
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	email := "alpha-user@example.com"
	password := "the-alpha-user-password"
	result := struct{ Subject struct{ Uuid string } }{}
	h.ResultTo(&result)
	h.Do("POST", h.Url("/users"), &createUserParamsWrapper{
		Subject: createUserParams{
			Email:    email,
			Password: password,
			Name:     "α user",
		},
	})

	check := func(activity *domain.Activity) {
		if got, want := activity.ContextUserUuid, (*string)(nil); got == want {
			t.Errorf(`activity.ContextUserUuid = %v; want not %v`, got, want)
			return
		}
		if got, want := *activity.ContextUserUuid, result.Subject.Uuid; got != want {
			t.Errorf(`activity.ContextUserUuid = %v; want %v`, got, want)
		}
	}

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "user.signed-up" {
			check(activity)
			return
		}
	}

	t.Fatalf("Activity %q not found", "user.signed-up")
}

func Test_UserHandler_Create_blocksUsersInTwelveHours_afterTheySignUpViaTheSignupForm(t *testing.T) {
	now := time.Date(2015, 10, 30, 16, 13, 0, 0, time.UTC)
	domain.Clock = clock.At(now)
	defer func() { domain.Clock = clock.System }()

	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	email := "alpha-user@example.com"
	password := "the-alpha-user-password"
	result := struct{ Subject struct{ Uuid string } }{}
	h.ResultTo(&result)
	h.Do("POST", h.Url("/users"), &createUserParamsWrapper{
		Subject: createUserParams{
			Email:    email,
			Password: password,
			Name:     "α user",
		},
	})

	blocks, err := stores.NewDbUserBlockStore(h.Tx()).FindAllByUserUuidValidFrom(result.Subject.Uuid, now.Add(12*time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(blocks), 1; got != want {
		t.Fatalf(`len(blocks) = %v; want %v`, got, want)
	}

	if got, want := blocks[0].Reason, "email_unverified"; got != want {
		t.Errorf(`blocks[0].Reason = %v; want %v`, got, want)
	}
}

func Test_UserHandler_Create_doesNotBlockUsers_whoSignUpViaInvitation(t *testing.T) {
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	email := "alpha-user@example.com"
	password := "the-alpha-user-password"
	project := h.World().Project("public")
	invitation := project.NewInvitationToHarrow("α user", email, "hello", domain.MembershipTypeMember)
	invitation.CreatorUuid = h.World().User("default").Uuid
	invitationUuid, err := stores.NewDbInvitationStore(h.Tx()).Create(invitation)
	if err != nil {
		t.Fatalf("Failed to create invitation: %s", err)
	}

	result := struct{ Subject struct{ Uuid string } }{}
	h.ResultTo(&result)
	h.Do("POST", h.Url("/users"), &createUserParamsWrapper{
		Subject: createUserParams{
			Email:          email,
			Password:       password,
			Name:           "α user",
			InvitationUuid: invitationUuid,
		},
	})

	if got, want := h.Response().StatusCode, http.StatusCreated; got != want {
		t.Errorf("h.Response().StatusCode = %d; want %d", got, want)
		t.Fatalf("h.ResponseBody():\n%s\n", h.ResponseBody())
	}

	blocks, err := stores.NewDbUserBlockStore(h.Tx()).FindAllByUserUuid(result.Subject.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(blocks), 0; got != want {
		t.Fatalf("len(blocks) = %d; want %d", got, want)
	}
}

func Test_UserHandler_VerifyEmail_removesAnUnverifiedEmailBlockForTheUser(t *testing.T) {
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	now := time.Date(2015, 10, 30, 20, 39, 0, 0, time.UTC)
	blockActiveFrom := now.Add(12 * time.Hour)
	user := h.World().User("default")
	block, err := user.NewBlock("email_unverified")
	if err != nil {
		t.Fatal(err)
	}
	block.BlockForever(blockActiveFrom)
	blocks := stores.NewDbUserBlockStore(h.Tx())
	if err := blocks.Create(block); err != nil {
		t.Fatal(err)
	}

	h.Subject(user)
	h.Do("POST", h.UrlFor("verify-email"), struct {
		Token string `json:"token"`
	}{user.Token})
	t.Logf("Response body:\n%s\n", h.ResponseBody())
	if got, want := h.Response().StatusCode, http.StatusOK; got != want {
		t.Errorf(`h.StatusCode = %v; want %v`, got, want)
	}

	foundBlocks, err := blocks.FindAllByUserUuidValidFrom(user.Uuid, blockActiveFrom)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(foundBlocks), 0; got != want {
		t.Errorf(`len(foundBlocks) = %v; want %v`, got, want)
	}
}

func Test_UserHandler_VerifyEmail_emitsUserEmailVerifiedActivity(t *testing.T) {
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	user := h.World().User("default")

	h.Subject(user)
	h.Do("POST", h.UrlFor("verify-email"), struct {
		Token string `json:"token"`
	}{user.Token})

	check := func(activity *domain.Activity) {
		if got, want := activity.ContextUserUuid, (*string)(nil); got == want {
			t.Errorf(`activity.ContextUserUuid = %v; want not %v`, got, want)
			return
		}
		if got, want := *activity.ContextUserUuid, user.Uuid; got != want {
			t.Errorf(`activity.ContextUserUuid = %v; want %v`, got, want)
		}
	}

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "user.email-verified" {
			check(activity)
			return
		}
	}

	t.Fatalf("Activity %q not found", "user.email-verified")

}

func Test_UserHandler_VerifyEmail_doesNotRemoveAnUnverifiedEmailBlockForTheUser_ifTokenDoesNotMatch(t *testing.T) {
	h := NewHandlerTest(MountUserHandler, t)
	defer h.Cleanup()

	now := time.Date(2015, 10, 30, 20, 39, 0, 0, time.UTC)
	blockActiveFrom := now.Add(12 * time.Hour)
	user := h.World().User("default")
	block, err := user.NewBlock("email_unverified")
	if err != nil {
		t.Fatal(err)
	}
	block.BlockForever(blockActiveFrom)
	blocks := stores.NewDbUserBlockStore(h.Tx())
	if err := blocks.Create(block); err != nil {
		t.Fatal(err)
	}

	url := fmt.Sprintf("/users/%s/verify-email?token=%s", user.Uuid, "does-not-match")
	h.Do("POST", h.Url(url), nil)
	if got, want := h.Response().StatusCode, StatusUnprocessableEntity; got != want {
		t.Errorf(`h.StatusCode = %v; want %v`, got, want)
	}

	foundBlocks, err := blocks.FindAllByUserUuidValidFrom(user.Uuid, blockActiveFrom)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(foundBlocks), 1; got != want {
		t.Errorf(`len(foundBlocks) = %v; want %v`, got, want)
	}
}
