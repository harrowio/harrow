package domain

import (
	"fmt"
	"time"
)

type Organization struct {
	defaultSubject
	Uuid        string     `json:"uuid"`
	Name        string     `json:"name"`
	GithubLogin string     `json:"githubLogin"  db:"github_login"`
	Public      bool       `json:"public"`
	CreatedAt   time.Time  `json:"createdAt"    db:"created_at"`
	ArchivedAt  *time.Time `json:"archivedAt"   db:"archived_at"`
}

// TODO: Lh/DH Move this to a struct method
func ValidateOrganization(u *Organization) error {

	if len(u.Name) == 0 {
		return NewValidationError("name", "too_short")
	}

	return nil
}

func (self *Organization) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/organizations/%s", requestScheme, requestBaseUri, self.Uuid)
}

func (self *Organization) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["memberships"] = map[string]string{"href": fmt.Sprintf("%s://%s/organizations/%s/memberships", requestScheme, requestBaseUri, self.Uuid)}
	response["members"] = map[string]string{"href": fmt.Sprintf("%s://%s/organizations/%s/members", requestScheme, requestBaseUri, self.Uuid)}
	response["projects"] = map[string]string{"href": fmt.Sprintf("%s://%s/organizations/%s/projects", requestScheme, requestBaseUri, self.Uuid)}
	response["project-cards"] = map[string]string{"href": fmt.Sprintf("%s://%s/organizations/%s/project-cards", requestScheme, requestBaseUri, self.Uuid)}
	response["add-credit-card"] = map[string]string{"href": fmt.Sprintf("%s://%s/billing-plans/braintree/credit-cards", requestScheme, requestBaseUri)}
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	response["limits"] = map[string]string{"href": fmt.Sprintf("%s://%s/organizations/%s/limits", requestScheme, requestBaseUri, self.Uuid)}
	return response
}

func (self *Organization) FindOrganization(store OrganizationStore) (*Organization, error) {
	return self, nil
}

func (self *Organization) FindProject(store ProjectStore) (*Project, error) {
	return store.FindByOrganizationUuid(self.Uuid)
}

func (self *Organization) AuthorizationName() string { return "organization" }

func (self *Organization) NewBillingEvent(data BillingEventData) *BillingEvent {
	return &BillingEvent{
		EventName:        data.BillingEventName(),
		Data:             data,
		OccurredOn:       time.Now(),
		OrganizationUuid: self.Uuid,
	}
}

// CreationDate returns the date on which this organization was
// created.
func (self *Organization) CreationDate() time.Time {
	return self.CreatedAt
}

func (self *Organization) DeletionDate() time.Time {
	if self.ArchivedAt == nil {
		return time.Time{}
	}

	return *self.ArchivedAt
}

func (self *Organization) Id() string {
	return self.Uuid
}
