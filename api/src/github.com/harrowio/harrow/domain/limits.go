package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

type Limits struct {
	defaultSubject
	OrganizationUuid    string                `json:"organizationUuid"`
	Projects            int                   `json:"projects"`
	Members             int                   `json:"members"`
	PublicRepositories  int                   `json:"publicRepositories"`
	PrivateRepositories int                   `json:"privateRepositories"`
	TrialDaysLeft       int                   `json:"trialDaysLeft"`
	TrialEnabled        bool                  `json:"trialEnabled"`
	Version             time.Time             `json:"version"`
	Plan                *LimitsComparedToPlan `json:"plan"`
}

type LimitsComparedToPlan struct {
	UsersExceedingLimit           int  `json:"usersExceedingLimit"`
	ProjectsExceedingLimit        int  `json:"projectsExceedingLimit"`
	RequiresUpgradeForPrivateCode bool `json:"requiresUpgradeForPrivateCode"`
	UsersIncluded                 int  `json:"usersIncluded"`
	ProjectsIncluded              int  `json:"projectsIncluded"`
}

func (self *Limits) MarshalJSON() ([]byte, error) {
	result := struct {
		Limits
		Exceeded bool `json:"exceeded"`
	}{*self, self.Exceeded()}

	return json.Marshal(result)
}

func (self *Limits) Exceeded() bool {
	if self.Plan == nil {
		return false
	}

	if self.TrialDaysLeft > 0 {
		return false
	}

	return self.Plan.UsersExceedingLimit > 0 || self.Plan.ProjectsExceedingLimit > 0
}

func (self *Limits) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/organizations/%s/limits", requestScheme, requestBaseUri, self.OrganizationUuid)
}

func (self *Limits) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	response["organization"] = map[string]string{"href": fmt.Sprintf("%s://%s/organizations/%s", requestScheme, requestBaseUri, self.OrganizationUuid)}
	return response
}

func (self *Limits) AuthorizationName() string { return "limits" }
func (self *Limits) FindOrganization(organizations OrganizationStore) (*Organization, error) {
	return organizations.FindByUuid(self.OrganizationUuid)
}

func (self *Limits) FindProject(store ProjectStore) (*Project, error) {
	return store.FindByOrganizationUuid(self.OrganizationUuid)
}
