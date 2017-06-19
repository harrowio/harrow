package domain

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/harrowio/harrow/uuidhelper"
)

type NotificationRule struct {
	defaultSubject

	Uuid          string     `json:"uuid" db:"uuid"`
	ProjectUuid   string     `json:"projectUuid" db:"project_uuid"`
	NotifierUuid  string     `json:"notifierUuid" db:"notifier_uuid"`
	NotifierType  string     `json:"notifierType" db:"notifier_type"`
	MatchActivity string     `json:"matchActivity" db:"match_activity"`
	CreatorUuid   string     `json:"creatorUuid" db:"creator_uuid"`
	JobUuid       *string    `json:"jobUuid" db:"job_uuid"`
	ArchivedAt    *time.Time `json:"archivedAt" db:"archived_at"`
}

func (self *NotificationRule) OwnUrl(requestScheme, requestBase string) string {
	return fmt.Sprintf("%s://%s/notification-rules/%s", requestScheme, requestBase, self.Uuid)
}

func (self *NotificationRule) Links(response map[string]map[string]string, requestScheme, requestBase string) map[string]map[string]string {
	response["self"] = map[string]string{
		"href": self.OwnUrl(requestScheme, requestBase),
	}
	response["project"] = map[string]string{
		"href": fmt.Sprintf("%s://%s/projects/%s", requestScheme, requestBase, self.ProjectUuid),
	}

	if self.JobUuid != nil {
		response["job"] = map[string]string{
			"href": fmt.Sprintf("%s://%s/jobs/%s", requestScheme, requestBase, *self.JobUuid),
		}
	}

	return response
}

func (self *NotificationRule) Validate() error {
	result := NewValidationError("", "")

	if !uuidhelper.IsValid(self.ProjectUuid) {
		result.Add("projectUuid", "malformed")
	}

	if !uuidhelper.IsValid(self.NotifierUuid) {
		result.Add("notifierUuid", "malformed")
	}

	if strings.TrimSpace(self.NotifierType) == "" {
		result.Add("notifierType", "empty")
	}
	if !strings.HasSuffix(self.NotifierType, "_notifiers") {
		result.Add("notifierType", "malformed")
	}

	if strings.TrimSpace(self.MatchActivity) == "" {
		result.Add("matchActivity", "empty")
	}

	if _, err := path.Match(self.MatchActivity, "example.activity"); err != nil {
		result.Add("matchActivity", "malformed")
	}

	if !uuidhelper.IsValid(self.CreatorUuid) {
		result.Add("creatorUuid", "malformed")
	}

	return result.ToError()
}

func (self *NotificationRule) Matches(activity *Activity) bool {
	if self.ProjectUuid != activity.ProjectUuid() {
		return false
	}

	if !self.AppliesToJob(activity.JobUuid()) {
		return false
	}

	matches, err := path.Match(self.MatchActivity, activity.Name)
	if err != nil {
		return false
	}
	return matches
}

func (self *NotificationRule) AppliesToJob(jobUuid string) bool {
	if self.JobUuid == nil {
		return true
	}

	return *self.JobUuid == jobUuid
}

func (self *NotificationRule) AuthorizationName() string { return "notification-rule" }
func (self *NotificationRule) FindProject(projects ProjectStore) (*Project, error) {
	return projects.FindByUuid(self.ProjectUuid)
}
