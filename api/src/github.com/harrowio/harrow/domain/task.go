package domain

import (
	"fmt"
	"strings"
	"time"
)

type Task struct {
	defaultSubject
	Uuid        string     `json:"uuid"`
	Body        string     `json:"body"`
	CreatedAt   time.Time  `json:"createdAt"   db:"created_at"`
	Name        string     `json:"name"`
	ProjectUuid string     `json:"projectUuid" db:"project_uuid"`
	Type        string     `json:"type"`
	ArchivedAt  *time.Time `json:"archivedAt"  db:"archived_at"`
}

func (self *Task) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/tasks/%s", requestScheme, requestBaseUri, self.Uuid)
}

func (self *Task) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	response["project"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s", requestScheme, requestBaseUri, self.ProjectUuid)}
	response["jobs"] = map[string]string{"href": fmt.Sprintf("%s://%s/tasks/%s/jobs", requestScheme, requestBaseUri, self.Uuid)}
	return response
}

// FindProject satisfies authz.BelongsToProject in order to determine
// authorization.
func (self *Task) FindProject(store ProjectStore) (*Project, error) {
	return store.FindByUuid(self.ProjectUuid)
}

// NewJob constructs a new job for this task in the given environment
func (self *Task) NewJob(environment *Environment) *Job {
	return &Job{
		Name:            fmt.Sprintf("%s - %s", strings.Title(environment.Name), self.Name),
		TaskUuid:        self.Uuid,
		EnvironmentUuid: environment.Uuid,
	}
}

func (self *Task) AuthorizationName() string { return "task" }
func (self *Task) CreationDate() time.Time   { return self.CreatedAt }
func (self *Task) Id() string                { return self.Uuid }
