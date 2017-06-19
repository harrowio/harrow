package stencil

import (
	"testing"

	"github.com/harrowio/harrow/domain"
)

func TestBashLinux_creates_environments(t *testing.T) {
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
	stencil := NewBashLinux(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	testcases := []string{"Sandbox", "Staging", "Production"}
	for _, testcase := range testcases {
		t.Logf("Environment: %s", testcase)
		environment, err := environments.FindEnvironmentByName(testcase)
		if err != nil {
			t.Fatal(err)
		}

		if got := environment; got == nil {
			t.Fatalf(`environment is nil`)
		}
	}
}

func TestBashLinux_creates_DOCKERHUB_variables_in_production(t *testing.T) {
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
	stencil := NewBashLinux(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	production, err := environments.FindEnvironmentByName("Production")
	if err != nil {
		t.Fatal(err)
	}

	if got, want := production.Get("DOCKERHUB_USER"), "user"; got != want {
		t.Errorf(`production.Get("DOCKERHUB_USER") = %v; want %v`, got, want)
	}

	if got, want := production.Get("DOCKERHUB_PASS"), "Correct Horse Battery Staple"; got != want {
		t.Errorf(`production.Get("DOCKERHUB_PASS") = %v; want %v`, got, want)
	}

	if got, want := production.Get("DOCKERHUB_EMAIL"), "user@example.org"; got != want {
		t.Errorf(`production.Get("DOCKERHUB_EMAIL") = %v; want %v`, got, want)
	}

}

func TestBashLinux_creates_REDIS_VERSION_variable_in_staging_and_production(t *testing.T) {
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
	stencil := NewBashLinux(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	staging, err := environments.FindEnvironmentByName("Staging")
	if err != nil {
		t.Fatal(err)
	}

	production, err := environments.FindEnvironmentByName("Production")
	if err != nil {
		t.Fatal(err)
	}

	if got, want := staging.Get("REDIS_VERSION"), "unstable"; got != want {
		t.Errorf(`staging.Get("REDIS_VERSION") = %v; want %v`, got, want)
	}

	if got, want := production.Get("REDIS_VERSION"), "stable"; got != want {
		t.Errorf(`production.Get("REDIS_VERSION") = %v; want %v`, got, want)
	}

}

func TestBashLinux_creates_tasks(t *testing.T) {
	projectUuid := "ded9add4-ab9a-400f-81d6-05b9b51c670d"
	environments := NewEnvironmentsInMemory()
	projects := NewProjectsInMemory().Add(&domain.Project{
		Uuid: projectUuid,
	})
	tasks := NewTasksInMemory()
	configuration := &Configuration{
		Environments:      environments,
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
	stencil := NewBashLinux(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	expectedTasks := []struct {
		Name string
		Body string
	}{
		{"Build & Push Docker Container", stencil.BuildAndPushDockerContainerTaskBody()},
		{"Cowsay Greeting", stencil.CowsayTaskBody()},
		{"Deploy Files With Rsync", stencil.DeployFilesRsyncTaskBody()},
		{"Download & Compile Redis", stencil.DownloadCompileRedisBody()},
		{"Lint Scripts", stencil.LintScriptsTaskBody()},
	}

	for _, expectedTask := range expectedTasks {
		t.Logf("Task: %s", expectedTask.Name)
		task, err := tasks.FindTaskByName(expectedTask.Name)
		if err != nil {
			t.Fatal(err)
		}

		if got, want := task.ProjectUuid, projectUuid; got != want {
			t.Errorf(`task.ProjectUuid = %v; want %v`, got, want)
		}

		if got, want := task.Body, expectedTask.Body; got != want {
			t.Errorf(`task.Body = %v; want %v`, got, want)
		}
	}
}

func TestBashLinux_creates_jobs_for_tasks(t *testing.T) {
	projectUuid := "ded9add4-ab9a-400f-81d6-05b9b51c670d"
	environments := NewEnvironmentsInMemory()
	projects := NewProjectsInMemory().Add(&domain.Project{
		Uuid: projectUuid,
	})
	tasks := NewTasksInMemory()
	jobs := NewJobsInMemory()
	configuration := &Configuration{
		Environments:      environments,
		Projects:          projects,
		ProjectUuid:       projectUuid,
		Tasks:             tasks,
		Jobs:              jobs,
		EmailNotifiers:    NewEmailNotifiersInMemory(),
		NotificationRules: NewNotificationRulesInMemory(),
		Schedules:         NewSchedulesInMemory(),
		GitTriggers:       NewGitTriggersInMemory(),
		Secrets:           NewSecretsInMemory(),
		Users:             NewUsersInMemory(),
		Webhooks:          NewWebhooksInMemory(),
	}
	stencil := NewBashLinux(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	findTask := func(name string) *domain.Task {
		task, err := tasks.FindTaskByName(name)
		if err != nil {
			t.Fatal(err)
		}
		return task
	}

	findEnvironment := func(name string) *domain.Environment {
		environment, err := environments.FindEnvironmentByName(name)
		if err != nil {
			t.Fatal(err)
		}

		return environment
	}

	expectedJobs := []struct {
		Task        *domain.Task
		Environment *domain.Environment
	}{
		{findTask("Cowsay Greeting"), findEnvironment("Sandbox")},
		{findTask("Lint Scripts"), findEnvironment("Staging")},
		{findTask("Deploy Files With Rsync"), findEnvironment("Staging")},
		{findTask("Deploy Files With Rsync"), findEnvironment("Production")},
		{findTask("Build & Push Docker Container"), findEnvironment("Production")},
		{findTask("Download & Compile Redis"), findEnvironment("Staging")},
		{findTask("Download & Compile Redis"), findEnvironment("Production")},
	}

	for _, expectedJob := range expectedJobs {
		t.Logf("Job: %q in %q ", expectedJob.Task.Name, expectedJob.Environment.Name)
		job, err := jobs.FindByTaskAndEnvironmentUuid(expectedJob.Task.Uuid, expectedJob.Environment.Uuid)
		if err != nil {
			t.Fatal(err)
		}

		if got, want := job.TaskUuid, expectedJob.Task.Uuid; got != want {
			t.Errorf(`job.TaskUuid = %v; want %v`, got, want)
		}

		if got, want := job.ProjectUuid, projectUuid; got != want {
			t.Errorf(`job.ProjectUuid = %v; want %v`, got, want)
		}

		if got, want := job.EnvironmentUuid, expectedJob.Environment.Uuid; got != want {
			t.Errorf(`job.EnvironmentUuid = %v; want %v`, got, want)
		}
	}
}
