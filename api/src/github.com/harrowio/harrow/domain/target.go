package domain

import (
	"fmt"
	"time"
)

type Target struct {
	defaultSubject
	Uuid            string     `json:"uuid"`
	CreatedAt       time.Time  `json:"createdAt" db:"created_at"`
	ArchivedAt      *time.Time `json:"archivedAt" db:"archived_at"`
	AccessibeSince  *time.Time `json:"accessibeSince" db:"accessibe_since"`
	EnvironmentUuid string     `json:"EnvironmentUuid" db:"environment_uuid"`
	Identifier      string     `json:"identifier"`
	ProjectUuid     string     `json:"projectUuid" db:"project_uuid"`
	Secret          string     `json:"secret"`
	Type            string     `json:"type"`
	Url             string     `json:"url"`
}

func (self *Target) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/targets/%s", requestScheme, requestBaseUri, self.Uuid)
}

func (self *Target) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["environment"] = map[string]string{"href": fmt.Sprintf("%s://%s/environments/%s", requestScheme, requestBaseUri, self.EnvironmentUuid)}
	response["project"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s", requestScheme, requestBaseUri, self.ProjectUuid)}
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	return response
}

func ValidateTarget(t *Target) error {

	if len(t.Url) == 0 {
		return NewValidationError("url", "too_short")
	}

	// TODO: Better validtions here (valid type, etc)

	return nil
}

func (self *Target) AuthorizationName() string { return "target" }

func (self *Target) FindProject(projects ProjectStore) (*Project, error) {
	return projects.FindByUuid(self.ProjectUuid)
}
