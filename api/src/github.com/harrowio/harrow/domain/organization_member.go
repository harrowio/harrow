package domain

import "fmt"

var (
	organizationMemberGuestCapabilities = newCapabilityList().
						reads("organization", "project-member", "organization-member").
						strings()

	organizationMemberMemberCapabilities = newCapabilityList().
						add(organizationMemberGuestCapabilities).
						reads("limits").
						creates("project").
						creates("public").
						strings()

	organizationMemberManagerCapabilities = newCapabilityList().
						add(organizationMemberMemberCapabilities).
						archives("project").
						updates("project").
						updates("public").
						writesFor("organization-member").
						strings()

	organizationMemberOwnerCapabilities = newCapabilityList().
						add(organizationMemberManagerCapabilities).
						does("braintree", "purchase").
						writesFor("organization").
						strings()
)

// OrganizationMember augments the data from user with information about
// a user's membership with a given organization.
type OrganizationMember struct {
	defaultSubject
	*User
	MembershipType   string `json:"type" db:"membership_type"`
	OrganizationUuid string `json:"organizationUuid" db:"organization_uuid"`
}

func (self *OrganizationMember) OwnedBy(user *User) bool {
	return self.User.Uuid == user.Uuid
}

func (self *OrganizationMember) FindUser(users UserStore) (*User, error) {
	return users.FindByUuid(self.User.Uuid)
}

// FindOrganization satisfies authz.BelongsToOrganization by finding
// the organizing this member belongs to.
func (self *OrganizationMember) FindOrganization(organizations OrganizationStore) (*Organization, error) {
	return organizations.FindByUuid(self.OrganizationUuid)
}

func NewOrganizationMember(user *User, membership *OrganizationMembership) *OrganizationMember {
	return &OrganizationMember{
		User:             user,
		MembershipType:   membership.Type,
		OrganizationUuid: membership.OrganizationUuid,
	}
}

func (self *OrganizationMember) OwnUrl(requestScheme, requestBase string) string {
	return fmt.Sprintf("%s://%s/organization-members/%s", requestScheme, requestBase, self.Uuid)
}

func (self *OrganizationMember) Links(response map[string]map[string]string, requestScheme, requestBase string) map[string]map[string]string {
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBase)}
	response["organization"] = map[string]string{"href": fmt.Sprintf("%s://%s/organizations/%s", requestScheme, requestBase, self.OrganizationUuid)}
	response["user"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s", requestScheme, requestBase, self.User.Uuid)}
	response["profilePicture"] = map[string]string{"href": newGravatarUrl(self.User.Email).String()}

	return response
}

func (self *OrganizationMember) Capabilities() []string {
	switch self.MembershipType {
	case MembershipTypeGuest:
		return organizationMemberGuestCapabilities
	case MembershipTypeMember:
		return organizationMemberMemberCapabilities
	case MembershipTypeManager:
		return organizationMemberManagerCapabilities
	case MembershipTypeOwner:
		return organizationMemberOwnerCapabilities
	default:
		return nil
	}
}

func (self *OrganizationMember) AuthorizationName() string { return "organization-member" }
