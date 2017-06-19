package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/harrowio/harrow/uuidhelper"
)

type EmailNotifier struct {
	defaultSubject

	Uuid        string  `json:"uuid" db:"uuid"`
	Recipient   string  `json:"recipient" db:"recipient"`
	ProjectUuid *string `json:"projectUuid" db:"project_uuid"`
	UrlHost     string  `json:"urlHost" db:"url_host"`

	ArchivedAt *time.Time `json:"archivedAt" db:"archived_at"`
}

func (self *EmailNotifier) OwnUrl(requestScheme, requestBase string) string {
	return fmt.Sprintf("%s://%s/email-notifiers/%s", requestScheme, requestBase, self.Uuid)
}

func (self *EmailNotifier) Links(response map[string]map[string]string, requestScheme, requestBase string) map[string]map[string]string {
	response["self"] = map[string]string{
		"href": self.OwnUrl(requestScheme, requestBase),
	}

	if self.ProjectUuid != nil {
		response["project"] = map[string]string{
			"href": fmt.Sprintf("%s://%s/projects/%s", requestScheme, requestBase, *self.ProjectUuid),
		}
	}

	return response
}

func (self *EmailNotifier) AuthorizationName() string { return "email-notifier" }

func (self *EmailNotifier) FindProject(projects ProjectStore) (*Project, error) {
	if self.ProjectUuid != nil {
		return projects.FindByUuid(*self.ProjectUuid)
	}

	return projects.FindByNotificationRule("email_notifiers", self.Uuid)
}

func (self *EmailNotifier) Validate() error {
	result := NewValidationError("", "")

	if !strings.Contains(self.Recipient, "@") {
		result.Add("recipient", "malformed")
	}

	if self.ProjectUuid == nil || !uuidhelper.IsValid(*self.ProjectUuid) {
		result.Add("projectUuid", "malformed")
	}

	return result.ToError()
}
