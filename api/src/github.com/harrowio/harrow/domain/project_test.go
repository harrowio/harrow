package domain

import (
	"testing"

	"github.com/harrowio/harrow/uuidhelper"
)

type mockOrganizationStore map[string]*Organization

func (store mockOrganizationStore) FindByUuid(uuid string) (*Organization, error) {
	return store[uuid], nil
}

func (store mockOrganizationStore) FindByProjectUuid(uuid string) (*Organization, error) {
	return store[uuid], nil
}

func Test_Project_Links_linksToWebhooks(t *testing.T) {
	project := &Project{Uuid: "e25b1a98-1a34-442c-be25-9c41ed21dd13"}
	links := map[string]map[string]string{}
	project.Links(links, "http", "example.com")

	actual := links["webhooks"]["href"]
	expected := "http://example.com/projects/e25b1a98-1a34-442c-be25-9c41ed21dd13/webhooks"

	if actual != expected {
		t.Fatalf("Expected %q to be %q", actual, expected)
	}
}

func Test_Project_Links_linksToMembers(t *testing.T) {
	project := &Project{Uuid: "e25b1a98-1a34-442c-be25-9c41ed21dd13"}
	links := map[string]map[string]string{}
	project.Links(links, "http", "example.com")

	actual := links["project-members"]["href"]
	expected := "http://example.com/projects/e25b1a98-1a34-442c-be25-9c41ed21dd13/members"

	if actual != expected {
		t.Fatalf("Expected %q to be %q", actual, expected)
	}
}

func Test_Project_Links_linksToScheduledExecutions(t *testing.T) {
	project := &Project{Uuid: "e25b1a98-1a34-442c-be25-9c41ed21dd13"}
	links := map[string]map[string]string{}
	project.Links(links, "http", "example.com")

	actual := links["scheduled-executions"]["href"]
	expected := "http://example.com/projects/e25b1a98-1a34-442c-be25-9c41ed21dd13/scheduled-executions"

	if actual != expected {
		t.Fatalf("Expected %q to be %q", actual, expected)
	}
}

func Test_Project_NewInvitationToUser_CreatesInvitation(t *testing.T) {
	project := &Project{
		Uuid:             "5859270c-0f8f-46af-b82a-0993b7a799a2",
		OrganizationUuid: "a34260c1-f516-48ed-9c25-5ecf86966a1b",
		Name:             "The Project",
		Public:           true,
	}

	invitee := &User{
		Uuid:  "4ac9ead8-14b0-4f49-817d-b26648d82e36",
		Email: "vagrant+test-user-invitation@localhost",
		Name:  "Vagrant",
	}

	message := "invitation-message"
	invitation := project.NewInvitationToUser(message, MembershipTypeMember, invitee)

	if invitation.InviteeUuid != invitee.Uuid {
		t.Fatal("Wrong invitee uuid")
	}

	if invitation.ProjectUuid != project.Uuid {
		t.Fatal("Wrong project uuid")
	}

	if invitation.OrganizationUuid != project.OrganizationUuid {
		t.Fatal("Wrong organization uuid")
	}

	if invitation.Message != message {
		t.Fatal("Wrong message: ", invitation.Message)
	}

	if invitation.RecipientName != invitee.Name {
		t.Fatal("Wrong recipient name: ", invitation.RecipientName)
	}

	if invitation.Email != invitee.Email {
		t.Fatal("Wrong email: ", invitation.Email)
	}

	if invitation.MembershipType != MembershipTypeMember {
		t.Fatal("Wrong membership type: ", invitation.MembershipType)
	}
}

func Test_Project_NewInvitationToHarrow_GeneratesInviteeUuid(t *testing.T) {
	project := &Project{
		Uuid:             "5859270c-0f8f-46af-b82a-0993b7a799a2",
		OrganizationUuid: "a34260c1-f516-48ed-9c25-5ecf86966a1b",
		Name:             "The Project",
		Public:           true,
	}

	name := "Invitee"
	email := "vagrant+test-harrow-invitation"
	message := "invitation-message"
	invitation := project.NewInvitationToHarrow(name, email, message, MembershipTypeMember)

	if !uuidhelper.IsValid(invitation.InviteeUuid) {
		t.Fatal("Invalid InviteeUuid: ", invitation.InviteeUuid)
	}

	if invitation.ProjectUuid != project.Uuid {
		t.Fatal("Wrong project uuid")
	}

	if invitation.OrganizationUuid != project.OrganizationUuid {
		t.Fatal("Wrong organization uuid")
	}

	if invitation.Message != message {
		t.Fatal("Wrong message: ", invitation.Message)
	}

	if invitation.RecipientName != name {
		t.Fatal("Wrong recipient name: ", invitation.RecipientName)
	}

	if invitation.Email != email {
		t.Fatal("Wrong email: ", invitation.Email)
	}

	if invitation.MembershipType != MembershipTypeMember {
		t.Fatal("Wrong membership type: ", invitation.MembershipType)
	}
}

func Test_Project_FindOrganization_ReturnsAssociatedOrganization(t *testing.T) {
	organization := &Organization{
		Uuid: "b44f7519-8539-4870-8bd7-3e0f6773d339",
	}

	project := &Project{
		Uuid:             "b69c0056-f5de-4223-a04b-622eaa44c22d",
		OrganizationUuid: organization.Uuid,
		Name:             "The Project",
	}

	store := mockOrganizationStore{organization.Uuid: organization}

	if found, _ := project.FindOrganization(store); found == nil {
		t.Fatalf("Expected organization to be found.")
	}
}
