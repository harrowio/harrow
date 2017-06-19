package domain

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/harrowio/harrow/uuidhelper"
)

type JobNotifier struct {
	defaultSubject

	Uuid       string `json:"uuid" db:"uuid"`
	WebhookURL string `json:"webhookURL" db:"webhook_url"`

	ArchivedAt *time.Time `json:"archivedAt" db:"archived_at"`

	ProjectUuid string `json:"projectUuid" db:"project_uuid"`
	JobUuid     string `json:"jobUuid" db:"job_uuid"`
	JobName     string `json:"jobName" db:"job_name"`
}

func (self *JobNotifier) OwnUrl(requestScheme, requestBase string) string {
	return fmt.Sprintf("%s://%s/job-notifiers/%s", requestScheme, requestBase, self.Uuid)
}

func (self *JobNotifier) Links(response map[string]map[string]string, requestScheme, requestBase string) map[string]map[string]string {
	response["self"] = map[string]string{
		"href": self.OwnUrl(requestScheme, requestBase),
	}
	return response
}

func (self *JobNotifier) Validate() error {
	result := NewValidationError("", "")

	if strings.TrimSpace(self.WebhookURL) == "" {
		result.Add("webhookURL", "empty")
	}

	_, err := url.Parse(self.WebhookURL)
	if err != nil {
		result.Add("webhookURL", "malformed")
	}

	if !uuidhelper.IsValid(self.JobUuid) {
		result.Add("jobUuid", "malformed")
	}

	if !uuidhelper.IsValid(self.ProjectUuid) {
		result.Add("projectUuid", "malformed")
	}

	return result.ToError()
}

func (self *JobNotifier) FindProject(projects ProjectStore) (*Project, error) {
	return projects.FindByJobUuid(self.JobUuid)
}

func (self *JobNotifier) AuthorizationName() string { return "job-notifier" }

func (self *JobNotifier) WebhookSlug() string {
	webhookURL, err := url.Parse(self.WebhookURL)
	if err != nil {
		return ""
	}

	parts := strings.Split(webhookURL.Path, "/")
	if len(parts) == 0 {
		return ""
	}

	return parts[len(parts)-1]
}
