package domain

import (
	"fmt"
	"time"

	"github.com/harrowio/harrow/uuidhelper"
)

type Job struct {
	defaultSubject
	Uuid            string     `json:"uuid"`
	CreatedAt       time.Time  `json:"createdAt"       db:"created_at"`
	Name            string     `json:"name"`
	Description     *string    `json:"description"     db:"description"`
	TaskUuid        string     `json:"taskUuid"        db:"task_uuid"`
	EnvironmentUuid string     `json:"environmentUuid" db:"environment_uuid"`
	ArchivedAt      *time.Time `json:"archivedAt"      db:"archived_at"`

	// makes the job widget much easier to implement
	ProjectUuid      string     `json:"projectUuid" db:"project_uuid"`
	ProjectName      string     `json:"projectName" db:"project_name"`
	LastRunStartedAt *time.Time `json:"lastRunStartedAt"`
	LastRunCreatedAt *time.Time `json:"lastRunCreatedAt"`
	Runs             []string   `json:"runs"`
}

func (self *Job) CreationDate() time.Time { return self.CreatedAt }

func (self *Job) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/jobs/%s", requestScheme, requestBaseUri, self.Uuid)
}

func (self *Job) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	response["task"] = map[string]string{"href": fmt.Sprintf("%s://%s/tasks/%s", requestScheme, requestBaseUri, self.TaskUuid)}
	response["environment"] = map[string]string{"href": fmt.Sprintf("%s://%s/environments/%s", requestScheme, requestBaseUri, self.EnvironmentUuid)}
	response["scheduled-executions"] = map[string]string{"href": fmt.Sprintf("%s://%s/jobs/%s/scheduled-executions", requestScheme, requestBaseUri, self.Uuid)}
	response["scheduledExecutions"] = map[string]string{"href": fmt.Sprintf("%s://%s/jobs/%s/scheduled-executions", requestScheme, requestBaseUri, self.Uuid)}
	response["subscriptions"] = map[string]string{"href": fmt.Sprintf("%s://%s/jobs/%s/watch", requestScheme, requestBaseUri, self.Uuid)}
	response["schedule"] = map[string]string{"href": fmt.Sprintf("%s://%s/schedules", requestScheme, requestBaseUri)}
	response["operations"] = map[string]string{"href": fmt.Sprintf("%s://%s/jobs/%s/operations", requestScheme, requestBaseUri, self.Uuid)}
	response["notification-rules"] = map[string]string{"href": fmt.Sprintf("%s://%s/jobs/%s/notification-rules", requestScheme, requestBaseUri, self.Uuid)}
	response["job-notifiers"] = map[string]string{"href": fmt.Sprintf("%s://%s/jobs/%s/job-notifiers", requestScheme, requestBaseUri, self.Uuid)}
	response["project"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s", requestScheme, requestBaseUri, self.ProjectUuid)}
	response["triggers-schedules"] = map[string]string{"href": fmt.Sprintf("%s://%s/jobs/%s/triggers/schedules", requestScheme, requestBaseUri, self.Uuid)}
	response["triggers-webhooks"] = map[string]string{"href": fmt.Sprintf("%s://%s/jobs/%s/triggers/webhooks", requestScheme, requestBaseUri, self.Uuid)}
	response["triggers-git"] = map[string]string{"href": fmt.Sprintf("%s://%s/jobs/%s/triggers/git", requestScheme, requestBaseUri, self.Uuid)}
	response["triggers-jobs"] = map[string]string{"href": fmt.Sprintf("%s://%s/jobs/%s/triggers/jobs", requestScheme, requestBaseUri, self.Uuid)}
	response["build-badge-simple"] = map[string]string{"href": fmt.Sprintf("%s://%s/jobs/%s/build-badges/simple.svg", requestScheme, requestBaseUri, self.Uuid)}
	return response
}

// FindProject satisfies authz.BelongsToProject in order to determine
// authorization.
func (self *Job) FindProject(store ProjectStore) (*Project, error) {
	return store.FindByTaskUuid(self.TaskUuid)
}

func (self *Job) Id() string                { return self.Uuid }
func (self *Job) WatchableType() string     { return "job" }
func (self *Job) WatchableEvents() []string { return operationEvents }

// NewOperation constructs a new operation for running this job.
func (self *Job) NewOperation(wsbiUuid string) *Operation {
	// prevent accidental modification of self.Uuid
	jobUuid := self.Uuid

	return &Operation{
		Type:                   OperationTypeJobScheduled,
		JobUuid:                &jobUuid,
		WorkspaceBaseImageUuid: wsbiUuid,
		TimeLimit:              900,
		ExitStatus:             256,
	}
}

// NewRecurringSchedule returns a recurring schedule for this job,
// using the given cronexpr for scheduling runs of this job.
func (self *Job) NewRecurringSchedule(initiatorUuid string, cronexpr string) *Schedule {
	return &Schedule{
		UserUuid:   initiatorUuid,
		JobUuid:    self.Uuid,
		Cronspec:   &cronexpr,
		Parameters: NewOperationParameters(),
	}
}

func (self *Job) AuthorizationName() string { return "job" }

func (self *Job) FindRecentOperations(operations RecentOperations) error {
	recent, err := operations.FindRecentByJobUuid(5, self.Uuid)
	if err != nil {
		return err
	} else {
		for i, recentOperation := range recent {
			if i == 0 {
				self.LastRunCreatedAt = recentOperation.CreatedAt
				self.LastRunStartedAt = recentOperation.StartedAt
			}
			if recentOperation.GitLogs != nil {
				recentOperation.GitLogs.Trim(5)
			}
			self.Runs = append(self.Runs, recentOperation.Status())
		}
	}

	return nil
}

// NewJobNotifier returns a new job notifier which triggers this job.
func (self *Job) NewJobNotifier() *JobNotifier {
	return &JobNotifier{
		Uuid:    uuidhelper.MustNewV4(),
		JobUuid: self.Uuid,
		JobName: self.Name,
	}
}
