package stencil

import (
	"fmt"
	"testing"

	"github.com/harrowio/harrow/domain"
)

func TestScriptEditorDefaults_creates_task_with_urn_as_name(t *testing.T) {
	projectUuid := "ded9add4-ab9a-400f-81d6-05b9b51c670d"
	tasks := NewTasksInMemory()
	projects := NewProjectsInMemory().Add(&domain.Project{
		Uuid: projectUuid,
	})
	configuration := &Configuration{
		Environments:      NewEnvironmentsInMemory(),
		Projects:          projects,
		ProjectUuid:       projectUuid,
		Tasks:             tasks,
		Jobs:              NewJobsInMemory(),
		EmailNotifiers:    NewEmailNotifiersInMemory(),
		NotificationRules: NewNotificationRulesInMemory(),
		Schedules:         NewSchedulesInMemory(),
		GitTriggers:       NewGitTriggersInMemory(),
		Secrets:           NewSecretsInMemory(),
		Users:             NewUsersInMemory(),
		Webhooks:          NewWebhooksInMemory(),
	}
	stencil := NewScriptEditorDefaults(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	expectedUrn := fmt.Sprintf("urn:harrow:default-task:%s", projectUuid)

	_, err := tasks.FindTaskByName(expectedUrn)
	if err != nil {
		t.Fatal(err)
	}
}

func TestScriptEditorDefaults_creates_environment_with_urn_as_name(t *testing.T) {
	projectUuid := "ded9add4-ab9a-400f-81d6-05b9b51c670d"
	environments := NewEnvironmentsInMemory()
	projects := NewProjectsInMemory().Add(&domain.Project{
		Uuid: projectUuid,
	})
	configuration := &Configuration{
		Environments:      environments,
		Projects:          projects,
		ProjectUuid:       projectUuid,
		Tasks:             NewTasksInMemory(),
		Jobs:              NewJobsInMemory(),
		EmailNotifiers:    NewEmailNotifiersInMemory(),
		NotificationRules: NewNotificationRulesInMemory(),
		Schedules:         NewSchedulesInMemory(),
		GitTriggers:       NewGitTriggersInMemory(),
		Secrets:           NewSecretsInMemory(),
		Users:             NewUsersInMemory(),
		Webhooks:          NewWebhooksInMemory(),
	}
	stencil := NewScriptEditorDefaults(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	expectedUrn := fmt.Sprintf("urn:harrow:default-environment:%s", projectUuid)

	_, err := environments.FindEnvironmentByName(expectedUrn)
	if err != nil {
		t.Fatal(err)
	}
}

func TestScriptEditorDefaults_creates_job_with_urn_as_name(t *testing.T) {
	projectUuid := "ded9add4-ab9a-400f-81d6-05b9b51c670d"
	jobs := NewJobsInMemory()
	projects := NewProjectsInMemory().Add(&domain.Project{
		Uuid: projectUuid,
	})
	configuration := &Configuration{
		Environments:      NewEnvironmentsInMemory(),
		Projects:          projects,
		ProjectUuid:       projectUuid,
		Tasks:             NewTasksInMemory(),
		Jobs:              jobs,
		EmailNotifiers:    NewEmailNotifiersInMemory(),
		NotificationRules: NewNotificationRulesInMemory(),
		Schedules:         NewSchedulesInMemory(),
		GitTriggers:       NewGitTriggersInMemory(),
		Secrets:           NewSecretsInMemory(),
		Users:             NewUsersInMemory(),
		Webhooks:          NewWebhooksInMemory(),
	}
	stencil := NewScriptEditorDefaults(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	expectedUrn := fmt.Sprintf("urn:harrow:default-job:%s", projectUuid)

	_, err := jobs.FindJobByName(expectedUrn)
	if err != nil {
		t.Fatal(err)
	}
}

func TestScriptEditorDefaults_uses_existing_default_task(t *testing.T) {
	projectUuid := "ded9add4-ab9a-400f-81d6-05b9b51c670d"
	taskUuid := "a49a6d8d-feba-499a-88fa-02c369e41c73"
	tasks := NewTasksInMemory()
	tasks.CreateTask(&domain.Task{
		Uuid: taskUuid,
		Name: fmt.Sprintf("urn:harrow:default-task:%s", projectUuid),
	})
	projects := NewProjectsInMemory().Add(&domain.Project{
		Uuid: projectUuid,
	})
	configuration := &Configuration{
		Environments:      NewEnvironmentsInMemory(),
		Projects:          projects,
		ProjectUuid:       projectUuid,
		Tasks:             tasks,
		Jobs:              NewJobsInMemory(),
		EmailNotifiers:    NewEmailNotifiersInMemory(),
		NotificationRules: NewNotificationRulesInMemory(),
		Schedules:         NewSchedulesInMemory(),
		GitTriggers:       NewGitTriggersInMemory(),
		Secrets:           NewSecretsInMemory(),
		Users:             NewUsersInMemory(),
		Webhooks:          NewWebhooksInMemory(),
	}
	stencil := NewScriptEditorDefaults(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	if got, want := tasks.Count(), 1; got != want {
		t.Errorf(`tasks.Count() = %v; want %v`, got, want)
	}
}

func TestScriptEditorDefaults_uses_existing_default_environment(t *testing.T) {
	projectUuid := "ded9add4-ab9a-400f-81d6-05b9b51c670d"
	environmentUuid := "a49a6d8d-feba-499a-88fa-02c369e41c73"
	environments := NewEnvironmentsInMemory()
	environments.CreateEnvironment(&domain.Environment{
		Uuid: environmentUuid,
		Name: fmt.Sprintf("urn:harrow:default-environment:%s", projectUuid),
	})
	projects := NewProjectsInMemory().Add(&domain.Project{
		Uuid: projectUuid,
	})
	configuration := &Configuration{
		Environments:      environments,
		Projects:          projects,
		ProjectUuid:       projectUuid,
		Tasks:             NewTasksInMemory(),
		Jobs:              NewJobsInMemory(),
		EmailNotifiers:    NewEmailNotifiersInMemory(),
		NotificationRules: NewNotificationRulesInMemory(),
		Schedules:         NewSchedulesInMemory(),
		GitTriggers:       NewGitTriggersInMemory(),
		Secrets:           NewSecretsInMemory(),
		Users:             NewUsersInMemory(),
		Webhooks:          NewWebhooksInMemory(),
	}
	stencil := NewScriptEditorDefaults(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	if got, want := environments.Count(), 1; got != want {
		t.Errorf(`environments.Count() = %v; want %v`, got, want)
	}
}

func TestScriptEditorDefaults_uses_existing_default_job(t *testing.T) {
	projectUuid := "ded9add4-ab9a-400f-81d6-05b9b51c670d"
	jobUuid := "a49a6d8d-feba-499a-88fa-02c369e41c73"
	jobs := NewJobsInMemory()
	jobs.CreateJob(&domain.Job{
		Uuid: jobUuid,
		Name: fmt.Sprintf("urn:harrow:default-job:%s", projectUuid),
	})
	projects := NewProjectsInMemory().Add(&domain.Project{
		Uuid: projectUuid,
	})
	configuration := &Configuration{
		Environments:      NewEnvironmentsInMemory(),
		Projects:          projects,
		ProjectUuid:       projectUuid,
		Tasks:             NewTasksInMemory(),
		Jobs:              jobs,
		EmailNotifiers:    NewEmailNotifiersInMemory(),
		NotificationRules: NewNotificationRulesInMemory(),
		Schedules:         NewSchedulesInMemory(),
		GitTriggers:       NewGitTriggersInMemory(),
		Secrets:           NewSecretsInMemory(),
		Users:             NewUsersInMemory(),
		Webhooks:          NewWebhooksInMemory(),
	}
	stencil := NewScriptEditorDefaults(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	if got, want := jobs.Count(), 1; got != want {
		t.Errorf(`jobs.Count() = %v; want %v`, got, want)
	}
}
