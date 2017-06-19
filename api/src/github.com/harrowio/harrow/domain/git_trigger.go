package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/harrowio/harrow/uuidhelper"
)

var (
	GitTriggerChangeTypes = map[string]string{
		"change": "ref-changed",
		"add":    "ref-added",
		"remove": "ref-removed",
	}
)

type GitTrigger struct {
	defaultSubject

	Uuid string `json:"uuid" db:"uuid"`
	Name string `json:"name" db:"name"`

	ProjectUuid    string  `json:"projectUuid" db:"project_uuid"`
	JobUuid        string  `json:"jobUuid" db:"job_uuid"`
	RepositoryUuid *string `json:"repositoryUuid" db:"repository_uuid"`
	ChangeType     string  `json:"changeType" db:"change_type"`
	MatchRef       string  `json:"matchRef" db:"match_ref"`

	CreatorUuid string `json:"creatorUuid" db:"creator_uuid"`

	ArchivedAt *time.Time `json:"archivedAt" db:"archived_at"`
}

func NewGitTrigger(name, creatorUuid string) *GitTrigger {
	return &GitTrigger{
		Name:        name,
		CreatorUuid: creatorUuid,
	}
}

func (self *GitTrigger) OwnUrl(requestScheme, requestBase string) string {
	return fmt.Sprintf("%s://%s/git-triggers/%s", requestScheme, requestBase, self.Uuid)
}

func (self *GitTrigger) Links(response map[string]map[string]string, requestScheme, requestBase string) map[string]map[string]string {
	href := func(format string, args ...interface{}) string {
		return fmt.Sprintf(
			fmt.Sprintf("%s://%s%s", requestScheme, requestBase, format),
			args...,
		)
	}
	response["self"] = map[string]string{
		"href": self.OwnUrl(requestScheme, requestBase),
	}
	response["project"] = map[string]string{
		"href": href("/projects/%s", self.ProjectUuid),
	}
	response["job"] = map[string]string{
		"href": href("/jobs/%s", self.JobUuid),
	}

	return response
}

func (self *GitTrigger) Validate() error {
	result := NewValidationError("", "")
	if strings.TrimSpace(self.Name) == "" {
		result.Add("name", "empty")
	}

	if !uuidhelper.IsValid(self.ProjectUuid) {
		result.Add("projectUuid", "malformed")
	}

	if _, err := regexp.Compile(self.MatchRef); err != nil {
		result.Add("matchRef", "malformed")
	}

	if !uuidhelper.IsValid(self.JobUuid) {
		result.Add("jobUuid", "malformed")
	}

	if self.RepositoryUuid != nil && !uuidhelper.IsValid(*self.RepositoryUuid) {
		result.Add("repositoryUuid", "malformed")
	}

	if _, found := GitTriggerChangeTypes[self.ChangeType]; !found {
		result.Add("changeType", "malformed")
	}

	return result.ToError()
}

func (self *GitTrigger) AuthorizationName() string {
	return "git-trigger"
}

func (self *GitTrigger) FindProject(store ProjectStore) (*Project, error) {
	return store.FindByUuid(self.ProjectUuid)
}

func (self *GitTrigger) ForJob(jobUuid string) *GitTrigger {
	self.JobUuid = jobUuid
	return self
}

func (self *GitTrigger) InRepository(repositoryUuid string) *GitTrigger {
	self.RepositoryUuid = &repositoryUuid
	return self
}

func (self *GitTrigger) InProject(projectUuid string) *GitTrigger {
	self.ProjectUuid = projectUuid
	return self
}

func (self *GitTrigger) ForChangeType(changeType string) *GitTrigger {
	_, found := GitTriggerChangeTypes[changeType]
	if !found {
		panic(fmt.Errorf("Invalid GitTrigger change type: %q", changeType))
	}

	self.ChangeType = changeType

	return self
}

func (self *GitTrigger) MatchingRef(refspec string) *GitTrigger {
	self.MatchRef = refspec
	return self
}

// Match returns true if this trigger should fire for the given
// activity.
func (self *GitTrigger) Match(activity *Activity) bool {
	expectedActivityName := fmt.Sprintf(
		"repository-metadata.%s",
		GitTriggerChangeTypes[self.ChangeType],
	)

	if activity.Name != expectedActivityName {
		return false
	}

	repositoryUuid := ""
	refName := ""
	switch payload := activity.Payload.(type) {
	case *ChangedRepositoryRef:
		repositoryUuid = payload.RepositoryUuid
		refName = payload.Symbolic
	case *RepositoryRef:
		repositoryUuid = payload.RepositoryUuid
		refName = payload.Symbolic
	}

	if self.RepositoryUuid != nil &&
		*self.RepositoryUuid != repositoryUuid &&
		*self.RepositoryUuid != "" {
		return false
	}

	if regexp.MustCompile(self.MatchRef).FindString(refName) == "" {
		return false
	}

	return true
}
