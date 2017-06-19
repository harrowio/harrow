package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
	"github.com/harrowio/harrow/uuidhelper"

	"github.com/gorilla/mux"
)

func Test_InvitationHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountInvitationHandler(r, nil)

	spec := routingSpec{
		{"POST", "/invitations", "invitation-create"},
		{"GET", "/invitations/:uuid", "invitation-show"},
		{"PATCH", "/invitations/:uuid", "invitation-accept"},
	}

	spec.run(r, t)

}

func Test_InvitationHandler_Create_createsInvitation(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountInvitationHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	ctxt.u = u
	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequestJSON(s.Uuid, "POST", ts.URL+"/invitations", &halWrapper{
		Subject: &CreateInvitationParams{
			Email:          "vagrant+test-invitations@localhost",
			RecipientName:  "Vagrant",
			ProjectUuid:    world.Project("public").Uuid,
			MembershipType: domain.MembershipTypeMember,
			Message:        "Join the dark side.",
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, _ := ioutil.ReadAll(res.Body)
	if expected, status := http.StatusCreated, res.StatusCode; status != expected {
		t.Fatalf("Expected status %v, got %v\n---\n%s\n---", expected, status, body)
	}

	result := struct {
		Subject *domain.Invitation `json:"subject"`
		Links   Links              `json:"_links"`
	}{}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Error: %s\nWhile unmarshaling:\n%s\n", err, body)
	}

	id := result.Subject.Uuid
	if id == "" {
		t.Fatal("Empty uuid")
	}

	if result.Subject.Email == "" {
		t.Fatalf("Empty email, expected %q", "vagrant+test-invitations@localhost")
	}

	store := stores.NewDbInvitationStore(tx)
	if _, err := store.FindByUuid(id); err != nil {
		t.Fatal(err)
	}
}

func Test_InvitationHandler_Create_acceptsInviteeUuid(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountInvitationHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	invitee := world.User("non-member")
	ctxt.u = u
	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequestJSON(s.Uuid, "POST", ts.URL+"/invitations", &halWrapper{
		Subject: &CreateInvitationParams{
			Email:          "vagrant+test-invitations@localhost",
			RecipientName:  "Vagrant",
			ProjectUuid:    world.Project("public").Uuid,
			InviteeUuid:    invitee.Uuid,
			MembershipType: domain.MembershipTypeMember,
			Message:        "Join the dark side.",
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, _ := ioutil.ReadAll(res.Body)
	if expected, status := http.StatusCreated, res.StatusCode; status != expected {
		t.Fatalf("Expected status %v, got %v\n---\n%s\n---", expected, status, body)
	}

	result := struct {
		Subject *domain.Invitation `json:"subject"`
		Links   Links              `json:"_links"`
	}{}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Error: %s\nWhile unmarshaling:\n%s\n", err, body)
	}

	id := result.Subject.Uuid
	if id == "" {
		t.Fatal("Empty uuid")
	}

	store := stores.NewDbInvitationStore(tx)
	invitation, err := store.FindByUuid(id)
	if err != nil {
		t.Fatal(err)
	}

	if invitation.InviteeUuid != invitee.Uuid {
		t.Fatalf("Expected %s to be %s\n", invitation.InviteeUuid, invitee.Uuid)
	}
}

func Test_InvitationHandler_Create_requiresProjectToExist(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountInvitationHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	ctxt.u = u
	s := setupTestLoginSession(t, tx, u)
	uuid := uuidhelper.MustNewV4()

	req, err := newAuthenticatedRequestJSON(s.Uuid, "POST", ts.URL+"/invitations", &halWrapper{
		Subject: &CreateInvitationParams{
			Email:          "vagrant+test-invitations@localhost",
			RecipientName:  "Vagrant",
			ProjectUuid:    uuid,
			MembershipType: domain.MembershipTypeMember,
			Message:        "Join the dark side.",
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, _ := ioutil.ReadAll(res.Body)
	if expected, status := StatusUnprocessableEntity, res.StatusCode; status != expected {
		t.Fatalf("Expected status %v, got %v\n---\n%s\n---", expected, status, body)
	}
}

func Test_InvitationHandler_Create_422IfNoProjectUuidProvided(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountInvitationHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	ctxt.u = u
	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequestJSON(s.Uuid, "POST", ts.URL+"/invitations", &halWrapper{
		Subject: &CreateInvitationParams{
			Email:          "vagrant+test-invitations@localhost",
			RecipientName:  "Vagrant",
			MembershipType: domain.MembershipTypeMember,
			Message:        "Join the dark side.",
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, _ := ioutil.ReadAll(res.Body)
	if expected, status := StatusUnprocessableEntity, res.StatusCode; status != expected {
		t.Fatalf("Expected status %v, got %v\n---\n%s\n---", expected, status, body)
	}
}

func Test_InvitationHandler_Create_RequiresAuthorization(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountInvitationHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	u := world.User("default")
	ctxt.u = u
	s := setupTestLoginSession(t, tx, u)

	req, err := newAuthenticatedRequestJSON(s.Uuid, "POST", ts.URL+"/invitations", &halWrapper{
		Subject: &CreateInvitationParams{
			Email:          "vagrant+test-invitations@localhost",
			RecipientName:  "Vagrant",
			ProjectUuid:    world.Project("public").Uuid,
			MembershipType: domain.MembershipTypeMember,
			Message:        "Join the dark side.",
		},
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

func Test_InvitationHandler_Accept_storesIdOfCurrentUser(t *testing.T) {
	h := NewHandlerTest(MountInvitationHandler, t)
	defer h.Cleanup()

	tx := h.Tx()
	world := h.World()
	project := world.Project("public")
	invitee := world.User("non-member")
	inviter := world.User("default")
	invitationStore := stores.NewDbInvitationStore(tx)
	h.LoginAs("non-member")
	invitation := test_helpers.MustCreateInvitation(t, tx, &domain.Invitation{
		Email:            "vagrant+test-invitations@localhost",
		RecipientName:    "vagrant",
		OrganizationUuid: project.OrganizationUuid,
		ProjectUuid:      project.Uuid,
		InviteeUuid:      uuidhelper.MustNewV4(),
		CreatorUuid:      inviter.Uuid,
		MembershipType:   domain.MembershipTypeMember,
		Message:          "Join us!",
	})

	h.Subject(invitation)
	h.Do("PATCH", h.UrlFor("self"), &PatchInvitationParams{
		Accept: "accept",
	})

	found, err := invitationStore.FindByUuid(invitation.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := found.InviteeUuid, invitee.Uuid; got != want {
		t.Errorf("found.InviteeUuid = %q; want %q", got, want)
	}
}

func Test_InvitationHandler_Accept_CreatesProjectMembership(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountInvitationHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	inviter := world.User("default")

	// Invitee is a new user to ensure that there are no existing
	// project / organization memberships.
	invitee := test_helpers.MustCreateUser(t, tx, &domain.User{
		Name:     "Test user",
		Email:    "vagrant+test-user-email@localhost",
		Password: "password-is-long-and-secure",
		UrlHost:  "localdomain",
	})

	ctxt.u = invitee
	s := setupTestLoginSession(t, tx, invitee)
	invitation := test_helpers.MustCreateInvitation(t, tx, &domain.Invitation{
		Email:            "vagrant+test-invitations@localhost",
		RecipientName:    "Vagrant",
		OrganizationUuid: world.Project("public").OrganizationUuid,
		ProjectUuid:      world.Project("public").Uuid,
		CreatorUuid:      inviter.Uuid,
		InviteeUuid:      invitee.Uuid,
		MembershipType:   domain.MembershipTypeMember,
		Message:          "Join us!",
	})

	url := ts.URL + "/invitations/" + invitation.Uuid
	req, err := newAuthenticatedRequestJSON(s.Uuid, "PATCH", url, &PatchInvitationParams{
		Accept: "accept",
	})

	_, err = new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	projectMembership, err := stores.NewDbProjectMembershipStore(tx).FindByUserAndProjectUuid(invitee.Uuid, invitation.ProjectUuid)
	if err != nil {
		t.Fatal(err)
	}
	if projectMembership.MembershipType != domain.MembershipTypeMember {
		t.Fatalf("Expected project membership type to be %q, got %q", domain.MembershipTypeMember, projectMembership.MembershipType)
	}
}

// NOTE(dh): added because expected behavior changed from creating an
// organization AND project membership to creating ONLY a project membership.
func Test_InvitationHandler_Accept_DoesNotCreateOrganizationMembership(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountInvitationHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	inviter := world.User("default")

	// Invitee is a new user to ensure that there are no existing
	// project / organization memberships.
	invitee := test_helpers.MustCreateUser(t, tx, &domain.User{
		Name:     "Test user",
		Email:    "vagrant+test-user-email@localhost",
		Password: "password-is-long-and-secure",
		UrlHost:  "localdomain",
	})

	ctxt.u = invitee
	s := setupTestLoginSession(t, tx, invitee)
	invitation := test_helpers.MustCreateInvitation(t, tx, &domain.Invitation{
		Email:            "vagrant+test-invitations@localhost",
		RecipientName:    "Vagrant",
		OrganizationUuid: world.Project("public").OrganizationUuid,
		ProjectUuid:      world.Project("public").Uuid,
		CreatorUuid:      inviter.Uuid,
		InviteeUuid:      invitee.Uuid,
		MembershipType:   domain.MembershipTypeMember,
		Message:          "Join us!",
	})

	url := ts.URL + "/invitations/" + invitation.Uuid
	req, err := newAuthenticatedRequestJSON(s.Uuid, "PATCH", url, &PatchInvitationParams{
		Accept: "accept",
	})

	_, err = new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	_, err = stores.NewDbOrganizationMembershipStore(tx).FindByOrganizationAndUserUuids(invitation.OrganizationUuid, invitee.Uuid)
	if err == nil {
		t.Fatal("Expected an error")
	}
	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatalf("Expected a *domain.NotFoundError, got: %s", err)
	}
}

func Test_InvitationHandler_Accept_DoesNotFail_WhenOrganizationMembershipExists(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountInvitationHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	inviter := world.User("default")
	project := world.Project("public")
	invitee := world.User("other")

	ctxt.u = invitee
	s := setupTestLoginSession(t, tx, invitee)
	invitation := test_helpers.MustCreateInvitation(t, tx, &domain.Invitation{
		Email:            "vagrant+test-invitations@localhost",
		RecipientName:    "Vagrant",
		OrganizationUuid: project.OrganizationUuid,
		ProjectUuid:      project.Uuid,
		CreatorUuid:      inviter.Uuid,
		InviteeUuid:      invitee.Uuid,
		MembershipType:   domain.MembershipTypeMember,
		Message:          "Join us!",
	})

	url := ts.URL + "/invitations/" + invitation.Uuid
	req, err := newAuthenticatedRequestJSON(s.Uuid, "PATCH", url, &PatchInvitationParams{
		Accept: "accept",
	})

	_, err = new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	projectMembership, err := stores.NewDbProjectMembershipStore(tx).FindByUserAndProjectUuid(invitee.Uuid, invitation.ProjectUuid)
	if err != nil {
		t.Fatal(err)
	}
	if projectMembership.MembershipType != domain.MembershipTypeMember {
		t.Fatalf("Expected project membership type to be %q, got %q", domain.MembershipTypeMember, projectMembership.MembershipType)
	}

	organizationMembership, err := stores.NewDbOrganizationMembershipStore(tx).FindByOrganizationAndUserUuids(invitation.OrganizationUuid, invitee.Uuid)
	if err != nil {
		t.Fatal(err)
	}
	if organizationMembership.Type != domain.MembershipTypeMember {
		t.Fatalf("Expected organization membership type to be %q, got %q", domain.MembershipTypeMember, organizationMembership.Type)
	}
}

func Test_InvitationHandler_Accept_MarksInvitationAsAccepted(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountInvitationHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	inviter := world.User("default")
	invitee := world.User("non-member")
	ctxt.u = invitee
	invitation := test_helpers.MustCreateInvitation(t, tx, &domain.Invitation{
		Email:            "vagrant+test-invitations@localhost",
		RecipientName:    "Vagrant",
		OrganizationUuid: world.Project("public").OrganizationUuid,
		ProjectUuid:      world.Project("public").Uuid,
		CreatorUuid:      inviter.Uuid,
		InviteeUuid:      invitee.Uuid,
		MembershipType:   domain.MembershipTypeMember,
		Message:          "Join us!",
	})

	s := setupTestLoginSession(t, tx, invitee)
	url := ts.URL + "/invitations/" + invitation.Uuid
	req, err := newAuthenticatedRequestJSON(s.Uuid, "PATCH", url, &PatchInvitationParams{
		Accept: "accept",
	})

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, _ := ioutil.ReadAll(res.Body)
	if expected, status := http.StatusOK, res.StatusCode; status != expected {
		t.Fatalf("Expected status %v, got %v\n---\n%s\n---", expected, status, body)
	}

	invitation, err = stores.NewDbInvitationStore(tx).FindByUuid(invitation.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if !invitation.IsAccepted() {
		t.Fatal("Expected IsAccepted to return true")
	}

	if invitation.InviteeUuid != invitee.Uuid {
		t.Fatalf("Expected InviteeUuid to be %s, got %s\n", invitee.Uuid, invitation.InviteeUuid)
	}
}

func Test_InvitationHandler_Accept_DoesNotRequireAuthorization(t *testing.T) {
	h := NewHandlerTest(MountInvitationHandler, t)
	defer h.Cleanup()
	world := h.World()
	inviter := world.User("default")
	h.LoginAs("non-member")
	invitation := test_helpers.MustCreateInvitation(t, h.Tx(), &domain.Invitation{
		Email:            "vagrant+test-invitations@localhost",
		RecipientName:    "Vagrant",
		OrganizationUuid: world.Project("public").OrganizationUuid,
		ProjectUuid:      world.Project("public").Uuid,
		CreatorUuid:      inviter.Uuid,
		MembershipType:   domain.MembershipTypeMember,
		Message:          "Join us!",
	})

	h.Subject(invitation)
	h.Do("PATCH", h.UrlFor("self"), &PatchInvitationParams{
		Accept: "accept",
	})
	h.context.authz.Expect(t, "update", 0)
}

func Test_InvitationHandler_Accept_RefusesInvitation(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountInvitationHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	inviter := world.User("default")
	invitee := world.User("non-member")
	project := world.Project("public")
	ctxt.u = invitee
	invitation := test_helpers.MustCreateInvitation(t, tx, &domain.Invitation{
		Email:            "vagrant+test-invitations@localhost",
		RecipientName:    "Vagrant",
		OrganizationUuid: project.OrganizationUuid,
		ProjectUuid:      project.Uuid,
		CreatorUuid:      inviter.Uuid,
		MembershipType:   domain.MembershipTypeMember,
		Message:          "Join us!",
	})

	s := setupTestLoginSession(t, tx, invitee)
	url := ts.URL + "/invitations/" + invitation.Uuid
	req, err := newAuthenticatedRequestJSON(s.Uuid, "PATCH", url, &PatchInvitationParams{
		Accept: "refuse",
	})

	_, err = new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	reloadedInvitation, err := stores.NewDbInvitationStore(tx).FindByUuid(invitation.Uuid)
	if !reloadedInvitation.IsRefused() {
		t.Fatal("Expected invitation to be refused")
	}

}

func Test_InvitationHandler_Show_ReturnsInvitation(t *testing.T) {
	ts, ctxt := setupHandlerTestServer(MountInvitationHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	world := test_helpers.MustNewWorld(tx, t)
	inviter := world.User("default")
	invitee := world.User("non-member")
	organization := world.Organization("default")
	project := world.Project("public")
	ctxt.u = invitee
	invitation := test_helpers.MustCreateInvitation(t, tx, &domain.Invitation{
		Email:            "vagrant+test-invitations@localhost",
		InviteeUuid:      invitee.Uuid,
		RecipientName:    "Vagrant",
		OrganizationUuid: project.OrganizationUuid,
		ProjectUuid:      project.Uuid,
		CreatorUuid:      inviter.Uuid,
		MembershipType:   domain.MembershipTypeMember,
		Message:          "Join us!",
	})

	s := setupTestLoginSession(t, tx, invitee)
	url := ts.URL + "/invitations/" + invitation.Uuid
	req, err := newAuthenticatedRequest(s.Uuid, "GET", url, "")

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	response := struct {
		Subject struct {
			*domain.Invitation
			CreatorName      string
			ProjectName      string
			OrganizationName string
		}
	}{}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}

	inv := response.Subject
	expectations := []struct {
		field            string
		expected, actual interface{}
	}{
		{"Email", "vagrant+test-invitations@localhost", inv.Email},
		{"RecipientName", "Vagrant", inv.RecipientName},
		{"OrganizationUuid", project.OrganizationUuid, inv.OrganizationUuid},
		{"OrganizationName", organization.Name, inv.OrganizationName},
		{"ProjectUuid", project.Uuid, inv.ProjectUuid},
		{"ProjectName", project.Name, inv.ProjectName},
		{"CreatorUuid", inviter.Uuid, inv.CreatorUuid},
		{"CreatorName", inviter.Name, inv.CreatorName},
		{"RefusedAt", (*time.Time)(nil), inv.RefusedAt},
		{"AcceptedAt", (*time.Time)(nil), inv.AcceptedAt},
		{"SentAt", (*time.Time)(nil), inv.SentAt},
		{"Message", "Join us!", inv.Message},
		{"MembershipType", domain.MembershipTypeMember, inv.MembershipType},
		{"InviteeUuid", invitee.Uuid, inv.InviteeUuid},
	}

	for _, expectation := range expectations {
		if !reflect.DeepEqual(expectation.actual, expectation.expected) {
			t.Errorf("Expected %q to be %#v, got %#v", expectation.field, expectation.expected, expectation.actual)
		}
	}
}

func Test_InvitationHandler_Show_DoesNotUseAuthorization(t *testing.T) {
	h := NewHandlerTest(MountInvitationHandler, t)
	defer h.Cleanup()

	world := h.World()
	inviter := world.User("default")
	invitee := world.User("non-member")
	project := world.Project("public")
	h.LoginAs("non-member")
	invitation := test_helpers.MustCreateInvitation(t, h.Tx(), &domain.Invitation{
		Email:            "vagrant+test-invitations@localhost",
		InviteeUuid:      invitee.Uuid,
		RecipientName:    "Vagrant",
		OrganizationUuid: project.OrganizationUuid,
		ProjectUuid:      project.Uuid,
		CreatorUuid:      inviter.Uuid,
		MembershipType:   domain.MembershipTypeMember,
		Message:          "Join us!",
	})

	h.Subject(invitation)
	h.Do("GET", h.UrlFor("self"), nil)
	h.context.authz.Expect(t, "read", 0)
}

func Test_InvitationHandler_Create_emitsInvitationCreatedActivity(t *testing.T) {
	h := NewHandlerTest(MountInvitationHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	project := h.World().Project("public")

	h.Do("POST", h.Url("/invitations"), &halWrapper{
		Subject: CreateInvitationParams{
			RecipientName:  "deploy key",
			Message:        "Join us!",
			MembershipType: domain.MembershipTypeMember,
			Email:          "foo@localhost",
			ProjectUuid:    project.Uuid,
		},
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "invitation.created" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "invitation.created")
}

func Test_InvitationHandler_Accept_RefusesInvitation_andEmitsInvitationRefusedActivity(t *testing.T) {
	h := NewHandlerTest(MountInvitationHandler, t)
	defer h.Cleanup()

	world := h.World()
	inviter := world.User("default")
	project := world.Project("public")
	h.LoginAs("non-member")
	invitation := test_helpers.MustCreateInvitation(t, h.Tx(), &domain.Invitation{
		Email:            "vagrant+test-invitations@localhost",
		RecipientName:    "Vagrant",
		OrganizationUuid: project.OrganizationUuid,
		ProjectUuid:      project.Uuid,
		CreatorUuid:      inviter.Uuid,
		MembershipType:   domain.MembershipTypeMember,
		Message:          "Join us!",
	})

	h.Subject(invitation)
	h.Do("PATCH", h.UrlFor("self"), &PatchInvitationParams{
		Accept: "refuse",
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "invitation.refused" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "invitation.refused")
}

func Test_InvitationHandler_Accept_AcceptsInvitation_andEmitsInvitationAcceptedActivity(t *testing.T) {
	h := NewHandlerTest(MountInvitationHandler, t)
	defer h.Cleanup()

	world := h.World()
	inviter := world.User("default")
	project := world.Project("public")
	h.LoginAs("non-member")
	invitation := test_helpers.MustCreateInvitation(t, h.Tx(), &domain.Invitation{
		Email:            "vagrant+test-invitations@localhost",
		RecipientName:    "Vagrant",
		OrganizationUuid: project.OrganizationUuid,
		ProjectUuid:      project.Uuid,
		CreatorUuid:      inviter.Uuid,
		MembershipType:   domain.MembershipTypeMember,
		Message:          "Join us!",
	})

	h.Subject(invitation)
	h.Do("PATCH", h.UrlFor("self"), &PatchInvitationParams{
		Accept: "accept",
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "invitation.accepted" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "invitation.accepted")
}

func Test_InvitationHandler_Accept_AcceptsInvitation_andEmitsUserJoinedProjectActivity(t *testing.T) {
	h := NewHandlerTest(MountInvitationHandler, t)
	defer h.Cleanup()

	world := h.World()
	inviter := world.User("default")
	project := world.Project("public")
	h.LoginAs("non-member")
	invitation := test_helpers.MustCreateInvitation(t, h.Tx(), &domain.Invitation{
		Email:            "vagrant+test-invitations@localhost",
		RecipientName:    "Vagrant",
		OrganizationUuid: project.OrganizationUuid,
		ProjectUuid:      project.Uuid,
		CreatorUuid:      inviter.Uuid,
		MembershipType:   domain.MembershipTypeMember,
		Message:          "Join us!",
	})

	h.Subject(invitation)
	h.Do("PATCH", h.UrlFor("self"), &PatchInvitationParams{
		Accept: "accept",
	})

	check := func(activity *domain.Activity) {
		payload, ok := activity.Payload.(*activities.UserProjectPayload)
		if !ok {
			t.Errorf("payload.(type) = %T; want %T", activity.Payload, payload)
		}

		if got, want := payload.User.Name, h.World().User("non-member").Name; got != want {
			t.Errorf(`payload.User.Name = %v; want %v`, got, want)
		}

		if got, want := payload.Project.Uuid, project.Uuid; got != want {
			t.Errorf(`payload.Project.Uuid = %v; want %v`, got, want)
		}
	}

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "user.joined-project" {
			check(activity)
			return
		}
	}

	t.Fatalf("Activity %q not found", "user.joined-project")
}
