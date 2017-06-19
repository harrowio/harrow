package domain

import (
	"fmt"
	"time"
)

type ProjectCard struct {
	defaultSubject

	LastActivitySeenAt time.Time

	ProjectUuid    string     `json:"projectUuid" db:"project_uuid"`
	ProjectName    string     `json:"projectName" db:"project_name"`
	LastTaskUuid   string     `json:"lastTaskUuid" db:"last_task_uuid"`
	LastTaskName   string     `json:"lastTaskName" db:"last_task_name"`
	LastTaskRunAt  *time.Time `json:"lastTaskRunAt" db:"last_task_run_at"`
	LastTaskStatus string     `json:"lastTaskStatus" db:"last_task_status"`
}

func NewProjectCard(project *Project, environment *Environment, task *Task, mostRecentOperation *Operation) *ProjectCard {
	return &ProjectCard{
		ProjectUuid:    project.Uuid,
		ProjectName:    project.Name,
		LastTaskName:   fmt.Sprintf("%s - %s", environment.Name, task.Name),
		LastTaskUuid:   mostRecentOperation.Uuid,
		LastTaskRunAt:  mostRecentOperation.StartedAt,
		LastTaskStatus: mostRecentOperation.Status(),
	}
}

func (self *ProjectCard) AuthorizationName() string { return "project-card" }
func (self *ProjectCard) FindProject(projects ProjectStore) (*Project, error) {
	return projects.FindByUuid(self.ProjectUuid)
}

func (self *ProjectCard) FindOrganization(organizations OrganizationStore) (*Organization, error) {
	return organizations.FindByProjectUuid(self.ProjectUuid)
}

func (self *ProjectCard) OwnUrl(requestScheme, requestBase string) string {
	return fmt.Sprintf("%s://%s/projects/%s/card", requestScheme, requestBase, self.ProjectUuid)
}

func (self *ProjectCard) Links(response map[string]map[string]string, requestScheme string, requestBase string) map[string]map[string]string {
	response["self"] = map[string]string{
		"href": self.OwnUrl(requestScheme, requestBase),
	}

	response["project"] = map[string]string{
		"href": fmt.Sprintf("%s://%s/projects/%s", requestScheme, requestBase, self.ProjectUuid),
	}

	return response
}
