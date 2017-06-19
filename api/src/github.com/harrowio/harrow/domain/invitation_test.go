package domain

import (
	"reflect"
	"strings"
	"testing"
)

type mockUserStore struct {
	user *User
	err  error
}

func (store *mockUserStore) FindByUuid(uuid string) (*User, error) {
	return store.user, store.err
}

func (store *mockUserStore) FindAllSubscribers(watchableUuid, event string) ([]*User, error) {
	return []*User{}, nil
}

func TestInvitation_CallToActionPath_linksToSignUpForNonExistingUser(t *testing.T) {
	invitation := &Invitation{
		Uuid:        "15460734-2bca-47c3-9d8c-ade0aa16afef",
		ProjectUuid: "98db058d-9d97-487d-9aa8-049ce116121d",
		InviteeUuid: "476dd4d0-5aa9-4749-bbd9-f841584d242e",
	}
	users := &mockUserStore{nil, nil}

	expected := "signup?invitation=" + invitation.Uuid
	link := invitation.CallToActionPath(users)
	if !strings.Contains(link, expected) {
		t.Fatalf("Expected %q to contain %q", link, expected)
	}
}

func TestInvitation_CallToActionPath_linksToInvitationForExistingUser(t *testing.T) {
	invitation := &Invitation{
		Uuid:        "15460734-2bca-47c3-9d8c-ade0aa16afef",
		ProjectUuid: "98db058d-9d97-487d-9aa8-049ce116121d",
		InviteeUuid: "476dd4d0-5aa9-4749-bbd9-f841584d242e",
	}
	users := &mockUserStore{&User{}, nil}

	expected := "invitations/" + invitation.Uuid
	link := invitation.CallToActionPath(users)
	if !strings.Contains(link, expected) {
		t.Fatalf("Expected %q to contain %q", link, expected)
	}
}

func TestInvitation_Links_ContainsLinkToSelf(t *testing.T) {
	invitation := &Invitation{
		Uuid: "d9760edf-28ba-4788-9930-16839a566a33",
	}

	links := map[string]map[string]string{}
	invitation.Links(links, "http", "example.com")

	expected := "invitations/" + invitation.Uuid
	if selfLink := links["self"]["href"]; !strings.Contains(selfLink, expected) {
		t.Fatalf("Expected %q to contain %q", selfLink, expected)
	}
}

func TestInvitation_Links_ContainsLinkToProject(t *testing.T) {
	invitation := &Invitation{
		ProjectUuid: "49a2bc61-4c72-4817-9337-6191c5631139",
	}

	links := map[string]map[string]string{}
	invitation.Links(links, "http", "example.com")

	expected := "projects/" + invitation.ProjectUuid
	if projectLink := links["project"]["href"]; !strings.Contains(projectLink, expected) {
		t.Fatalf("Expected %q to contain %q", projectLink, expected)
	}
}

func TestInvitation_Links_ContainsLinkToOrganization(t *testing.T) {
	invitation := &Invitation{
		OrganizationUuid: "97208447-b174-40fb-90f2-44e8f91bdd56",
	}

	links := map[string]map[string]string{}
	invitation.Links(links, "http", "example.com")

	expected := "organizations/" + invitation.OrganizationUuid
	if orgLink := links["organization"]["href"]; !strings.Contains(orgLink, expected) {
		t.Fatalf("Expected %q to contain %q", orgLink, expected)
	}
}

func TestInvitation_Links_ContainsLinkToCreator(t *testing.T) {
	invitation := &Invitation{
		CreatorUuid: "7d1fb2fc-a2ed-4fb9-9666-b63f50ea5e38",
	}

	links := map[string]map[string]string{}
	invitation.Links(links, "http", "example.com")

	expected := "users/" + invitation.InviteeUuid
	if creatorLink := links["creator"]["href"]; !strings.Contains(creatorLink, expected) {
		t.Fatalf("Expected %q to contain %q", creatorLink, expected)
	}
}

func TestInvitation_Links_ContainsLinkToInvitee(t *testing.T) {
	invitation := &Invitation{
		InviteeUuid: "7d1fb2fc-a2ed-4fb9-9666-b63f50ea5e38",
	}

	links := map[string]map[string]string{}
	invitation.Links(links, "http", "example.com")

	expected := "users/" + invitation.InviteeUuid
	if inviteeLink := links["invitee"]["href"]; !strings.Contains(inviteeLink, expected) {
		t.Fatalf("Expected %q to contain %q", inviteeLink, expected)
	}
}

func TestInvitation_IsOpen_returnsFalseIfInvitationAccepted(t *testing.T) {
	invitation := &Invitation{}
	invitee := &User{Uuid: "2c56eaab-176b-4cfd-9935-4cddf8ae25eb"}
	invitation.Accept(invitee)

	if invitation.IsOpen() {
		t.Fatal("Expected invitation to not be open")
	}
}

func TestInvitation_IsOpen_returnsTrueIfInvitationHasNotBeenAccepted(t *testing.T) {
	invitation := &Invitation{}

	if !invitation.IsOpen() {
		t.Fatal("Expected invitation to be open")
	}
}

func TestInvitation_IsOpen_returnsFalseIfInvitationRefused(t *testing.T) {
	invitation := &Invitation{}
	invitation.Refuse()

	if invitation.IsOpen() {
		t.Fatal("Expected invitation to not be open")
	}
}

func TestInvitation_IsOpen_returnsTrueIfInvitationHasNotBeenRefused(t *testing.T) {
	invitation := &Invitation{}

	if !invitation.IsOpen() {
		t.Fatal("Expected invitation to be open")
	}
}

func TestInvitation_Accept_setsInviteeUuidToAcceptingUser(t *testing.T) {
	invitation := &Invitation{}
	invitee := &User{Uuid: "2c56eaab-176b-4cfd-9935-4cddf8ae25eb"}
	invitation.Accept(invitee)

	if got, want := invitation.InviteeUuid, invitee.Uuid; got != want {
		t.Errorf("invitation.InviteeUuid = %q; want %q", got, want)
	}
}

func TestInvitation_Accept_DoesNothingIfInvitationIsRefused(t *testing.T) {
	invitation := &Invitation{}
	invitation.Refuse()
	invitee := &User{Uuid: "2c56eaab-176b-4cfd-9935-4cddf8ae25eb"}
	invitation.Accept(invitee)

	if invitation.AcceptedAt != nil {
		t.Fatal("Expected AcceptedAt to be nil")
	}
}

func TestInvitation_Refuse_DoesNothingIfInvitationIsAccepted(t *testing.T) {
	invitation := &Invitation{}
	invitee := &User{Uuid: "2c56eaab-176b-4cfd-9935-4cddf8ae25eb"}
	invitation.Accept(invitee)
	invitation.Refuse()

	if invitation.RefusedAt != nil {
		t.Fatal("Expected RefusedAt to be nil")
	}
}

func TestInvitation_IsRefused_ReturnsTrueAfterCallingRefuse(t *testing.T) {
	invitation := &Invitation{}
	invitation.Refuse()

	if !invitation.IsRefused() {
		t.Fatal("Expected IsRefused to return true")
	}
}

func TestInvitation_IsRefused_ReturnsFalseByDefault(t *testing.T) {
	invitation := &Invitation{}

	if invitation.IsRefused() {
		t.Fatal("Expected IsRefused to return false")
	}
}

func TestInvitation_IsAccepted_ReturnsTrueAfterCallingAccept(t *testing.T) {
	invitation := &Invitation{}
	invitee := &User{Uuid: "2c56eaab-176b-4cfd-9935-4cddf8ae25eb"}
	invitation.Accept(invitee)

	if !invitation.IsAccepted() {
		t.Fatal("Expected IsAccepted to return true")
	}
}

func TestInvitation_IsAccepted_ReturnsFalseByDefault(t *testing.T) {
	invitation := &Invitation{}

	if invitation.IsAccepted() {
		t.Fatal("Expected IsAccepted to return false")
	}
}

func TestInvitation_NewProjectMembership(t *testing.T) {
	invitation := &Invitation{
		InviteeUuid:    "68438b44-19d1-4860-be09-20da1f569455",
		ProjectUuid:    "9cde59fe-72d7-4cd2-8eb7-b051611302c8",
		MembershipType: MembershipTypeManager,
	}

	membership := invitation.NewProjectMembership()
	expected := &ProjectMembership{
		UserUuid:       "68438b44-19d1-4860-be09-20da1f569455",
		ProjectUuid:    "9cde59fe-72d7-4cd2-8eb7-b051611302c8",
		MembershipType: MembershipTypeManager,
	}

	if !reflect.DeepEqual(membership, expected) {
		t.Fatalf("Expected %v, got %v", expected, membership)
	}
}
