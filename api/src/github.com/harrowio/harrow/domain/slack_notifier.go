package domain

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/harrowio/harrow/uuidhelper"
)

type SlackNotifier struct {
	defaultSubject

	Uuid        string `json:"uuid" db:"uuid"`
	Name        string `json:"name" db:"name"`
	WebhookURL  string `json:"webhookURL" db:"webhook_url"`
	UrlHost     string `json:"urlHost" db:"url_host"`
	ProjectUuid string `json:"projectUuid" db:"project_uuid"`

	ArchivedAt *time.Time `json:"archivedAt" db:"archived_at"`
}

func (self *SlackNotifier) OwnUrl(requestScheme, requestBase string) string {
	return fmt.Sprintf("%s://%s/slack-notifiers/%s", requestScheme, requestBase, self.Uuid)
}

func (self *SlackNotifier) Links(response map[string]map[string]string, requestScheme, requestBase string) map[string]map[string]string {
	response["self"] = map[string]string{
		"href": self.OwnUrl(requestScheme, requestBase),
	}

	return response
}

func (self *SlackNotifier) Validate() error {
	result := NewValidationError("", "")
	if strings.TrimSpace(self.Name) == "" {
		result.Add("name", "empty")
	}

	if strings.TrimSpace(self.WebhookURL) == "" {
		result.Add("webhookURL", "empty")
	}

	if _, err := url.Parse(self.WebhookURL); err != nil {
		result.Add("webhookURL", "malformed")
	}

	if strings.TrimSpace(self.UrlHost) == "" {
		result.Add("urlHost", "empty")
	}

	if !uuidhelper.IsValid(self.ProjectUuid) {
		result.Add("projectUuid", "malformed")
	}

	return result.ToError()
}

func (self *SlackNotifier) FindProject(projects ProjectStore) (*Project, error) {
	return projects.FindByUuid(self.ProjectUuid)
}

func (self *SlackNotifier) AuthorizationName() string {
	return "slack-notifier"
}
