package domain

import (
	"fmt"
	"time"
)

var (
	projectMemberVisitorCapabilities = newCapabilityList().
						reads("project-member", "organization", "operation", "job", "task", "schedule", "repository", "environment", "default-environment", "subscription", "webhook", "delivery", "checks", "git-trigger", "job-notifier", "email-notifier", "notification-rule", "script-card").
						strings()

	projectMemberGuestCapabilities = newCapabilityList().
					add(projectMemberVisitorCapabilities).
					strings()
	projectMemberMemberCapabilities = newCapabilityList().
					add(projectMemberGuestCapabilities).
					does("cancel", "operation").
					reads("project").
					reads("script-card").
					reads("limits").
					does("diff-scripts", "project").
					does("diff-scripts", "public").
					writesFor("schedule").
					writesFor("checks").
					writesFor("subscription").
					writesFor("notification-rule").
					writesFor("email-notifier").
					reads("job-notifier").
					reads("slack-notifier").
					reads("secret").
					reads("credential").
					reads("repository-credential").
					reads("project-card").
					strings()

	projectMemberManagerCapabilities = newCapabilityList().
						add(projectMemberMemberCapabilities).
						updates("project").
						updates("public").
						does("save-scripts", "project").
						does("save-scripts", "public").
						writesFor("environment").
						updates("default-environment").
						writesFor("task").
						writesFor("job").
						writesFor("github-deploy-key").
						writesFor("invitation").
						writesFor("repository").
						writesFor("project-member").
						writesFor("webhook").
						writesFor("git-trigger").
						writesFor("job-notifier").
						writesFor("slack-notifier").
						writesFor("stencil").
						does("create", "secret").
						does("archive", "secret").
						does(CapabilityReadPrivileged, "secret").
						strings()
	projectMemberOwnerCapabilities = newCapabilityList().
					add(projectMemberManagerCapabilities).
					archives("project").
					updates("project").
					archives("public").
					updates("public").
					strings()

	ProjectVisitor = &ProjectMember{}
)

// ProjectMember augments User with project specific data.
type ProjectMember struct {
	defaultSubject
	*User
	CreatedAt      time.Time `json:"createdAt" db:"-"`
	MembershipType string    `json:"type" db:"membership_type"`
	MembershipUuid *string   `json:"membershipUuid" db:"membership_uuid"`
	ProjectUuid    string    `json:"projectUuid" db:"project_uuid"`
}

func (self *ProjectMember) OwnedBy(user *User) bool {
	return self.User.Uuid == user.Uuid
}

func (self *ProjectMember) FindUser(users UserStore) (*User, error) {
	return users.FindByUuid(self.User.Uuid)
}

func (self *ProjectMember) FindProject(projects ProjectStore) (*Project, error) {
	return projects.FindByUuid(self.ProjectUuid)
}

func (self *ProjectMember) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/project-members/%s", requestScheme, requestBaseUri, self.Uuid)
}

func (self *ProjectMember) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["project"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s", requestScheme, requestBaseUri, self.ProjectUuid)}
	response["profilePicture"] = map[string]string{"href": newGravatarUrl(self.Email).String()}
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	return response
}

func NewProjectMember(user *User, project *Project, projectMembership *ProjectMembership, organizationMembership *OrganizationMembership) *ProjectMember {
	result := &ProjectMember{
		User:        user,
		ProjectUuid: project.Uuid,
	}

	if organizationMembership != nil {
		result.MembershipType = organizationMembership.Type
		result.CreatedAt = organizationMembership.CreatedAt
	}

	if projectMembership != nil {
		result.MembershipType = projectMembership.MembershipType
		result.MembershipUuid = &projectMembership.Uuid
		result.CreatedAt = projectMembership.CreatedAt
	}

	if projectMembership == nil && organizationMembership == nil {
		return nil
	}

	if projectMembership != nil && organizationMembership != nil {
		if MembershipTypeHierarchyLevel(organizationMembership.Type) > MembershipTypeHierarchyLevel(projectMembership.MembershipType) {
			result.MembershipType = organizationMembership.Type
		}
	}

	return result
}

type ProjectMembershipCreator interface {
	Create(membership *ProjectMembership) (string, error)
}

func (self *ProjectMember) AddMember(userUuid string, membershipType string, projectMemberships ProjectMembershipCreator) (string, error) {
	if MembershipTypeHierarchyLevel(self.MembershipType) < MembershipTypeHierarchyLevel(MembershipTypeManager) {
		return "", NewValidationError("membershipType", "too_low")
	}

	return projectMemberships.Create(&ProjectMembership{
		ProjectUuid:    self.ProjectUuid,
		MembershipType: membershipType,
		UserUuid:       userUuid,
	})
}

func (self *ProjectMember) Promote(other *ProjectMember) error {
	nextMembership := map[string]string{
		MembershipTypeGuest:   MembershipTypeMember,
		MembershipTypeMember:  MembershipTypeManager,
		MembershipTypeManager: MembershipTypeOwner,
		MembershipTypeOwner:   MembershipTypeOwner,
	}

	next, found := nextMembership[other.MembershipType]
	if !found {
		return NewValidationError("membershipType", "invalid")
	}

	if MembershipTypeHierarchyLevel(self.MembershipType) < MembershipTypeHierarchyLevel(next) {
		return NewValidationError("membershipType", "too_low")
	}

	other.MembershipType = next

	return nil
}

func (self *ProjectMember) ToMembership() *ProjectMembership {
	membership := &ProjectMembership{
		MembershipType: self.MembershipType,
		UserUuid:       self.User.Uuid,
		ProjectUuid:    self.ProjectUuid,
		CreatedAt:      self.CreatedAt,
	}

	if self.MembershipUuid != nil {
		membership.Uuid = *self.MembershipUuid
	}

	return membership
}

func (self *ProjectMember) Remove(other *ProjectMember, projectMemberships Archiver) error {
	if self.User != nil && self.User.Uuid == other.User.Uuid && other.MembershipUuid != nil {
		return projectMemberships.ArchiveByUuid(*other.MembershipUuid)
	}

	switch self.MembershipType {
	case MembershipTypeGuest:
		fallthrough
	case MembershipTypeMember:
		return NewValidationError("membershipType", "too_low")
	}

	if other.MembershipType == self.MembershipType && self.MembershipType != MembershipTypeOwner {
		return NewValidationError("membershipType", "too_low")
	}

	if other.MembershipUuid != nil {
		return projectMemberships.ArchiveByUuid(*other.MembershipUuid)
	} else {
		return NewValidationError("membership", "indirect")
	}
}

// Capabilities satisfies authz.Role by returning a list of all
// capabilities the member has in its project.
func (self *ProjectMember) Capabilities() []string {
	switch self.MembershipType {
	case MembershipTypeGuest:
		return projectMemberGuestCapabilities
	case MembershipTypeMember:
		return projectMemberMemberCapabilities
	case MembershipTypeManager:
		return projectMemberManagerCapabilities
	case MembershipTypeOwner:
		return projectMemberOwnerCapabilities
	default:
		return projectMemberVisitorCapabilities
	}
}

func (self *ProjectMember) AuthorizationName() string { return "project-member" }
