package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/harrowio/harrow/uuidhelper"
)

type Invitation struct {
	defaultSubject
	Uuid             string     `json:"uuid" db:"uuid"`
	RecipientName    string     `json:"recipientName" db:"recipient_name"`
	Email            string     `json:"email" db:"email"`
	OrganizationUuid string     `json:"organizationUuid" db:"organization_uuid"`
	ProjectUuid      string     `json:"projectUuid" db:"project_uuid"`
	MembershipType   string     `json:"membershipType" db:"membership_type"`
	CreatorUuid      string     `json:"creatorUuid" db:"creator_uuid"`
	CreatedAt        time.Time  `json:"createdAt" db:"created_at"`
	SentAt           *time.Time `json:"sentAt" db:"sent_at"`
	AcceptedAt       *time.Time `json:"acceptedAt" db:"accepted_at"`
	RefusedAt        *time.Time `json:"refusedAt" db:"refused_at"`
	InviteeUuid      string     `json:"inviteeUuid" db:"invitee_uuid"`
	Message          string     `json:"message" db:"message"`
	ArchivedAt       *time.Time `json:"archivedAt" db:"archived_at"`
}

func (self *Invitation) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/invitations/%s", requestScheme, requestBaseUri, self.Uuid)
}

func (self *Invitation) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["self"] = map[string]string{
		"href": self.OwnUrl(requestScheme, requestBaseUri),
	}

	project := Project{Uuid: self.ProjectUuid}
	response["project"] = map[string]string{
		"href": project.OwnUrl(requestScheme, requestBaseUri),
	}

	organization := Organization{Uuid: self.OrganizationUuid}
	response["organization"] = map[string]string{
		"href": organization.OwnUrl(requestScheme, requestBaseUri),
	}

	creator := User{Uuid: self.CreatorUuid}
	response["creator"] = map[string]string{
		"href": creator.OwnUrl(requestScheme, requestBaseUri),
	}

	invitee := User{Uuid: self.InviteeUuid}
	response["invitee"] = map[string]string{
		"href": invitee.OwnUrl(requestScheme, requestBaseUri),
	}

	return response
}

// CallToActionPath returns the URL path for the page on which the
// invitee can react to this invitation.
func (self *Invitation) CallToActionPath(users UserStore) string {
	if user, _ := users.FindByUuid(self.InviteeUuid); user == nil {
		return fmt.Sprintf("/a/signup?invitation=%s", self.Uuid)
	} else {
		return fmt.Sprintf("/a/invitations/%s", self.Uuid)
	}
}

// FindProject satifies authz.BelongsToProject to determine access to
// project level invitations.
func (self *Invitation) FindProject(store ProjectStore) (*Project, error) {
	return store.FindByUuid(self.ProjectUuid)
}

// FindUser satisfies authz.BelongsToUser by looking for the invitation's
// recipient (who might not exist yet).
func (self *Invitation) FindUser(store UserStore) (*User, error) {
	return store.FindByUuid(self.InviteeUuid)
}

// OwnedBy satisfies authz.Ownable by looking at the creator
// of the invitation.
func (self *Invitation) OwnedBy(user *User) bool {
	return user.Uuid == self.CreatorUuid
}

// FindOrganization satisfies authz.BelongsToOrganization
func (self *Invitation) FindOrganization(store OrganizationStore) (*Organization, error) {
	return store.FindByUuid(self.OrganizationUuid)
}

func (self *Invitation) Validate() error {
	result := EmptyValidationError()
	uuids := []struct {
		field string
		value string
	}{
		{"organizationUuid", self.OrganizationUuid},
		{"projectUuid", self.ProjectUuid},
		{"creatorUuid", self.CreatorUuid},
		{"inviteeUuid", self.InviteeUuid},
	}

	for _, uuid := range uuids {
		if !uuidhelper.IsValid(uuid.value) {
			result.Add(uuid.field, "malformed")
		}
	}

	if !strings.Contains(self.Email, "@") {
		result.Add("email", "malformed")
	}

	if self.RecipientName == "" {
		result.Add("recipientName", "empty")
	}

	return result.ToError()
}

// Accept marks the invitation as accepted.  The current time is used
// for recording the time of accepting the invitation.
func (self *Invitation) Accept(invitee *User) {
	if !self.IsOpen() {
		return
	}

	t := time.Now()
	self.AcceptedAt = &t
	self.InviteeUuid = invitee.Uuid
}

// IsAccepted returns true if the invitation has been marked as accepted.
func (self *Invitation) IsAccepted() bool {
	return self.AcceptedAt != nil
}

// Refuse marks the invitation as refused.  The current time is usd for
// recording the time of refusing the invitation.
func (self *Invitation) Refuse() {
	if !self.IsOpen() {
		return
	}

	t := time.Now()
	self.RefusedAt = &t
}

// IsRefused returns true if the invitation has been marked as refused.
func (self *Invitation) IsRefused() bool {
	return self.RefusedAt != nil
}

// IsOpen returns true if the invitation can still be responded to,
// i.e. whether it can still be accepted or refused.
func (self *Invitation) IsOpen() bool {
	return !(self.IsAccepted() || self.IsRefused())
}

// NewProjectMembership returns the new project membership that should
// be created when accepting this invitation.
func (self *Invitation) NewProjectMembership() *ProjectMembership {
	return &ProjectMembership{
		UserUuid:       self.InviteeUuid,
		ProjectUuid:    self.ProjectUuid,
		MembershipType: self.MembershipType,
	}
}

func (self *Invitation) AuthorizationName() string { return "invitation" }
