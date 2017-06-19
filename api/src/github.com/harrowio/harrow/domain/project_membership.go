package domain

import (
	"fmt"
	"time"
)

type ProjectMembership struct {
	defaultSubject
	Uuid           string     `json:"uuid" db:"uuid"`
	ProjectUuid    string     `json:"projectUuid" db:"project_uuid"`
	UserUuid       string     `json:"userUuid" db:"user_uuid"`
	MembershipType string     `json:"membershipType" db:"membership_type"`
	CreatedAt      time.Time  `json:"createdAt" db:"created_at"`
	ArchivedAt     *time.Time `json:"archivedAt" db:"archived_at"`
}

func (self *ProjectMembership) OwnUrl(requestScheme, requestUri string) string {
	return fmt.Sprintf("%s://%s/project-memberships/%s", requestScheme, requestUri, self.Uuid)
}

func (self *ProjectMembership) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["project"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s", requestScheme, requestBaseUri, self.ProjectUuid)}
	response["user"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s", requestScheme, requestBaseUri, self.UserUuid)}
	return response
}

func (self *ProjectMembership) OwnedBy(user *User) bool {
	return user.Uuid == self.UserUuid
}

func (self *ProjectMembership) FindUser(users UserStore) (*User, error) {
	return users.FindByUuid(self.UserUuid)
}

func (self *ProjectMembership) FindProject(project ProjectStore) (*Project, error) {
	return project.FindByUuid(self.ProjectUuid)
}

func (self *ProjectMembership) AuthorizationName() string { return "project-member" }
