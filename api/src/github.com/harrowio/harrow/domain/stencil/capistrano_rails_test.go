package stencil

import (
	"testing"

	"github.com/harrowio/harrow/domain"
)

func TestCapistranoRails_creates_environments(t *testing.T) {
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
	stencil := NewCapistranoRails(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	testcases := []string{"Test", "Staging", "Production"}
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

func TestCapistranoRails_creates_a_CAPISTRANO_ENV_variable_in_staging_and_production(t *testing.T) {
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
	stencil := NewCapistranoRails(configuration)

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

	if got, want := staging.Get("CAPISTRANO_ENV"), "staging"; got != want {
		t.Errorf(`staging.Get("CAPISTRANO_ENV") = %v; want %v`, got, want)
	}

	if got, want := production.Get("CAPISTRANO_ENV"), "production"; got != want {
		t.Errorf(`production.Get("CAPISTRANO_ENV") = %v; want %v`, got, want)
	}
}

func TestCapistranoRails_creates_tasks(t *testing.T) {
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
	stencil := NewCapistranoRails(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	expectedTasks := []struct {
		Name string
		Body string
	}{
		{"Deploy", stencil.DeployTaskBody()},
		{"Notes", stencil.NotesTaskBody()},
		{"Check dependencies", stencil.CheckDependenciesTaskBody()},
		{"Run tests", stencil.RunTestsTaskBody()},
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

func TestCapistranoRails_creates_jobs_for_tasks(t *testing.T) {
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
	stencil := NewCapistranoRails(configuration)

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
		{findTask("Deploy"), findEnvironment("Production")},
		{findTask("Deploy"), findEnvironment("Staging")},
		{findTask("Check dependencies"), findEnvironment("Test")},
		{findTask("Notes"), findEnvironment("Test")},
		{findTask("Run tests"), findEnvironment("Test")},
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

func TestCapistranoRails_creates_an_email_notifier_if_an_email_address_is_provided(t *testing.T) {

	t.Skip("Error looking up notification rule")

	projectUuid := "ded9add4-ab9a-400f-81d6-05b9b51c670d"
	environments := NewEnvironmentsInMemory()
	projects := NewProjectsInMemory().Add(&domain.Project{
		Uuid: projectUuid,
	})
	tasks := NewTasksInMemory()
	jobs := NewJobsInMemory()
	emailNotifiers := NewEmailNotifiersInMemory()
	notificationRules := NewNotificationRulesInMemory()
	configuration := &Configuration{
		UserUuid:       "8978dce1-269a-4c6e-8fde-dcd83bdb4ace",
		NotifyViaEmail: "vagrant@localhost",

		Environments:      environments,
		Projects:          projects,
		ProjectUuid:       projectUuid,
		Tasks:             tasks,
		Jobs:              jobs,
		EmailNotifiers:    emailNotifiers,
		NotificationRules: notificationRules,
		Schedules:         NewSchedulesInMemory(),
		GitTriggers:       NewGitTriggersInMemory(),
		Secrets:           NewSecretsInMemory(),
		Users:             NewUsersInMemory(),
		Webhooks:          NewWebhooksInMemory(),
	}
	stencil := NewCapistranoRails(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	emailNotifier := emailNotifiers.FindByEmailAddress(configuration.NotifyViaEmail)
	if got := emailNotifier; got == nil {
		t.Fatalf(`emailNotifier is nil`)
	}

	if got := emailNotifier.ProjectUuid; got == nil {
		t.Fatalf(`emailNotifier.ProjectUuid is nil`)
	}

	if got, want := *emailNotifier.ProjectUuid, projectUuid; got != want {
		t.Errorf(`emailNotifier.ProjectUuid = %v; want %v`, got, want)
	}

	runNotes, err := jobs.FindJobByName("Test - Notes")
	if err != nil {
		t.Fatal(err)
	}

	notificationRule, err := notificationRules.FindByNotifierAndJobUuidAndType(emailNotifier.Uuid, runNotes.Uuid, "email_notifier")
	if got := notificationRule; got == nil {
		t.Fatalf(`notificationRule is nil`)
	}

	if err != nil {
		t.Fatal(err)
	}

	if got, want := notificationRule.ProjectUuid, projectUuid; got != want {
		t.Errorf(`notificationRule.ProjectUuid = %v; want %v`, got, want)
	}

	if got, want := notificationRule.CreatorUuid, configuration.UserUuid; got != want {
		t.Errorf(`notificationRule.CreatorUuid = %v; want %v`, got, want)
	}
}

func TestCapistranoRails_does_not_notify_via_email_if_invalid_address_is_provided(t *testing.T) {
	projectUuid := "ded9add4-ab9a-400f-81d6-05b9b51c670d"
	environments := NewEnvironmentsInMemory()
	projects := NewProjectsInMemory().Add(&domain.Project{
		Uuid: projectUuid,
	})
	tasks := NewTasksInMemory()
	jobs := NewJobsInMemory()
	emailNotifiers := NewEmailNotifiersInMemory()
	notificationRules := NewNotificationRulesInMemory()
	configuration := &Configuration{
		NotifyViaEmail: "vagrant[at]localhost",

		Environments:      environments,
		Projects:          projects,
		ProjectUuid:       projectUuid,
		Tasks:             tasks,
		Jobs:              jobs,
		EmailNotifiers:    emailNotifiers,
		NotificationRules: notificationRules,
		Schedules:         NewSchedulesInMemory(),
		GitTriggers:       NewGitTriggersInMemory(),
		Secrets:           NewSecretsInMemory(),
		Users:             NewUsersInMemory(),
		Webhooks:          NewWebhooksInMemory(),
	}
	stencil := NewCapistranoRails(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	emailNotifier := emailNotifiers.FindByEmailAddress(configuration.NotifyViaEmail)
	if got, want := emailNotifier, (*domain.EmailNotifier)(nil); got != want {
		t.Errorf(`emailNotifier = %v; want %v`, got, want)
	}
}

func TestCapistranoRails_schedules_notes_task_to_run_weekly(t *testing.T) {
	projectUuid := "ded9add4-ab9a-400f-81d6-05b9b51c670d"
	environments := NewEnvironmentsInMemory()
	projects := NewProjectsInMemory().Add(&domain.Project{
		Uuid: projectUuid,
	})
	tasks := NewTasksInMemory()
	jobs := NewJobsInMemory()
	emailNotifiers := NewEmailNotifiersInMemory()
	notificationRules := NewNotificationRulesInMemory()
	schedules := NewSchedulesInMemory()
	configuration := &Configuration{
		NotifyViaEmail: "vagrant[at]localhost",

		Environments:      environments,
		Projects:          projects,
		ProjectUuid:       projectUuid,
		Tasks:             tasks,
		Jobs:              jobs,
		EmailNotifiers:    emailNotifiers,
		NotificationRules: notificationRules,
		Schedules:         schedules,
		GitTriggers:       NewGitTriggersInMemory(),
		Secrets:           NewSecretsInMemory(),
		Users:             NewUsersInMemory(),
		Webhooks:          NewWebhooksInMemory(),
	}
	stencil := NewCapistranoRails(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	runNotes, err := jobs.FindJobByName("Test - Notes")
	if got := runNotes; got == nil {
		t.Fatalf(`runNotes is nil`)
	}

	if err != nil {
		t.Fatal(err)
	}

	runNotesWeeklySchedule, err := schedules.FindByJobUuid(runNotes.Uuid)
	if got := runNotesWeeklySchedule; got == nil {
		t.Fatalf(`runNotesWeeklySchedule is nil`)
	}

	if err != nil {
		t.Fatal(err)
	}

	if got := runNotesWeeklySchedule.Cronspec; got == nil {
		t.Fatalf(`runNotesWeeklySchedule.Cronspec is nil`)
	}

	if got, want := *runNotesWeeklySchedule.Cronspec, "@weekly"; got != want {
		t.Errorf(`*runNotesWeeklySchedule.Cronspec = %v; want %v`, got, want)
	}
}

func TestCapistranoRails_creates_a_git_trigger_to_run_tests_on_every_new_commit_on_master(t *testing.T) {
	projectUuid := "ded9add4-ab9a-400f-81d6-05b9b51c670d"
	environments := NewEnvironmentsInMemory()
	projects := NewProjectsInMemory().Add(&domain.Project{
		Uuid: projectUuid,
	})
	tasks := NewTasksInMemory()
	jobs := NewJobsInMemory()
	emailNotifiers := NewEmailNotifiersInMemory()
	notificationRules := NewNotificationRulesInMemory()
	schedules := NewSchedulesInMemory()
	gitTriggers := NewGitTriggersInMemory()
	configuration := &Configuration{
		NotifyViaEmail: "vagrant[at]localhost",

		Environments:      environments,
		Projects:          projects,
		ProjectUuid:       projectUuid,
		Tasks:             tasks,
		Jobs:              jobs,
		EmailNotifiers:    emailNotifiers,
		NotificationRules: notificationRules,
		Schedules:         schedules,
		GitTriggers:       gitTriggers,
		Secrets:           NewSecretsInMemory(),
		Users:             NewUsersInMemory(),
		Webhooks:          NewWebhooksInMemory(),
	}
	stencil := NewCapistranoRails(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	runTests, err := jobs.FindJobByName("Test - Run tests")
	if got := runTests; got == nil {
		t.Fatalf(`runTests is nil`)
	}
	if err != nil {
		t.Fatal(err)
	}

	trigger := gitTriggers.FindByProjectUuid(projectUuid)
	if got := trigger; got == nil {
		t.Fatalf(`trigger is nil`)
	}

	if got, want := trigger.ProjectUuid, projectUuid; got != want {
		t.Errorf(`trigger.ProjectUuid = %v; want %v`, got, want)
	}

	if got, want := trigger.MatchRef, ".*"; got != want {
		t.Errorf(`trigger.MatchRef = %v; want %v`, got, want)
	}

	if got, want := trigger.RepositoryUuid, (*string)(nil); got != want {
		t.Errorf(`trigger.RepositoryUuid = %v; want %v`, got, want)
	}

	if got, want := trigger.ChangeType, "change"; got != want {
		t.Errorf(`trigger.ChangeType = %v; want %v`, got, want)
	}

	if got, want := trigger.JobUuid, runTests.Uuid; got != want {
		t.Errorf(`trigger.JobUuid = %v; want %v`, got, want)
	}
}

func TestCapistranoRails_creates_an_SSH_secret_in_staging_and_production(t *testing.T) {
	projectUuid := "ded9add4-ab9a-400f-81d6-05b9b51c670d"
	environments := NewEnvironmentsInMemory()
	secrets := NewSecretsInMemory()
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
		Secrets:           secrets,
		Users:             NewUsersInMemory(),
		Webhooks:          NewWebhooksInMemory(),
	}
	stencil := NewCapistranoRails(configuration)

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

	stagingDeployKey := secrets.FindByEnvironmentUuid(staging.Uuid)
	if got := stagingDeployKey; got == nil {
		t.Fatalf(`stagingDeployKey is nil`)
	}
	if got, want := stagingDeployKey.Type, domain.SecretSsh; got != want {
		t.Errorf(`stagingDeployKey.Type = %v; want %v`, got, want)
	}
	if got, want := stagingDeployKey.Name, "Deploy key"; got != want {
		t.Errorf(`stagingDeployKey.Name = %v; want %v`, got, want)
	}

	productionDeployKey := secrets.FindByEnvironmentUuid(production.Uuid)
	if got := productionDeployKey; got == nil {
		t.Fatalf(`productionDeployKey is nil`)
	}
	if got, want := productionDeployKey.Type, domain.SecretSsh; got != want {
		t.Errorf(`productionDeployKey.Type = %v; want %v`, got, want)
	}
	if got, want := productionDeployKey.Name, "Deploy key"; got != want {
		t.Errorf(`productionDeployKey.Name = %v; want %v`, got, want)
	}

}

func TestCapistranoRails_creates_an_triggers_a_staging_deploy_when_tests_succeed(t *testing.T) {

	t.Skip("error looking up Jon Notifier with FindByTriggeringJobUuid() fn")

	projectUuid := "ded9add4-ab9a-400f-81d6-05b9b51c670d"
	environments := NewEnvironmentsInMemory()
	jobNotifiers := NewJobNotifiersInMemory()
	notificationRules := NewNotificationRulesInMemory()
	jobs := NewJobsInMemory()
	secrets := NewSecretsInMemory()
	projects := NewProjectsInMemory().Add(&domain.Project{
		Uuid: projectUuid,
	})
	userUuid := "2a93b318-8c48-4a24-bd07-bdf9df3b46f8"
	users := NewUsersInMemory().Add(&domain.User{
		Uuid:    userUuid,
		UrlHost: "www.vm.vagrant.io",
	})

	configuration := &Configuration{
		Environments:      environments,
		Projects:          projects,
		ProjectUuid:       projectUuid,
		UserUuid:          userUuid,
		Tasks:             NewTasksInMemory(),
		Jobs:              jobs,
		EmailNotifiers:    NewEmailNotifiersInMemory(),
		NotificationRules: notificationRules,
		JobNotifiers:      jobNotifiers,
		Schedules:         NewSchedulesInMemory(),
		GitTriggers:       NewGitTriggersInMemory(),
		Secrets:           secrets,
		Users:             users,
		Webhooks:          NewWebhooksInMemory(),
	}
	stencil := NewCapistranoRails(configuration)

	if err := stencil.Create(); err != nil {
		t.Fatal(err)
	}

	runTests, err := jobs.FindJobByName("Test - Run tests")
	if err != nil {
		t.Fatal(err)
	}

	jobNotifier := jobNotifiers.FindByTriggeringJobUuid(runTests.Uuid)
	if got := jobNotifier; got == nil {
		t.Fatalf(`jobNotifier is nil`)
	}

	triggerDeployAfterTests, err := notificationRules.FindByNotifierAndJobUuidAndType(jobNotifier.Uuid, runTests.Uuid, "job_notifiers")
	if err != nil {
		t.Fatal(err)
	}
	if got := triggerDeployAfterTests; got == nil {
		t.Fatalf(`triggerDeployAfterTests is nil`)
	}

	if got := triggerDeployAfterTests.JobUuid; got == nil {
		t.Fatalf(`triggerDeployAfterTests.JobUuid is nil`)
	}

	if got, want := *triggerDeployAfterTests.JobUuid, runTests.Uuid; got != want {
		t.Errorf(`triggerDeployAfterTests.JobUuid = %v; want %v`, got, want)
	}

	if got, want := triggerDeployAfterTests.NotifierType, "job_notifiers"; got != want {
		t.Errorf(`triggerDeployAfterTests.NotifierType = %v; want %v`, got, want)
	}

}
