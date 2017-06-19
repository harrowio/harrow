package domain

import (
	"fmt"
	"time"
)

type OrganizationMembership struct {
	defaultSubject
	OrganizationUuid string    `json:"organizationUuid" db:"organization_uuid"`
	UserUuid         string    `json:"userUuid"         db:"user_uuid"`
	Type             string    `json:"type"`
	CreatedAt        time.Time `json:"createdAt"        db:"created_at"`
}

func (self *OrganizationMembership) OwnUrl(requestScheme, requestBaseUri string) string {
	return ""
}

func (self *OrganizationMembership) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["organization"] = map[string]string{"href": fmt.Sprintf("%s://%s/organizations/%s", requestScheme, requestBaseUri, self.OrganizationUuid)}
	response["user"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s", requestScheme, requestBaseUri, self.UserUuid)}
	return response
}

func (self *OrganizationMembership) OwnedBy(user *User) bool {
	return self.UserUuid == user.Uuid
}

func (self *OrganizationMembership) FindUser(users UserStore) (*User, error) {
	return users.FindByUuid(self.UserUuid)
}

func (self *OrganizationMembership) FindOrganization(organizations OrganizationStore) (*Organization, error) {
	return organizations.FindByUuid(self.OrganizationUuid)
}
func (self *OrganizationMembership) AuthorizationName() string { return "organization-member" }
