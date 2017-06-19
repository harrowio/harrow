package stencil

import (
	"fmt"
	"strings"

	"github.com/harrowio/harrow/domain"
)

// CapistranoRails sets up default tasks and environments for working
// with a RubyOnRails application using capistrano.
type CapistranoRails struct {
	conf *Configuration
}

func NewCapistranoRails(configuration *Configuration) *CapistranoRails {
	return &CapistranoRails{
		conf: configuration,
	}
}

// Create creates the following objects for this stencil:
//
// Environments: test, staging, production
//
// Tasks (in test): "bundle outdated", "rake notes", "rake test"
//
// Tasks (in staging, production): "cap $env deploy"
func (self *CapistranoRails) Create() error {
	errors := NewError()
	project, err := self.conf.Projects.FindProject(self.conf.ProjectUuid)
	if err != nil {
		return errors.Add("FindProject", project, err)
	}

	testEnvironment := project.NewEnvironment("Test")
	if err := self.conf.Environments.CreateEnvironment(testEnvironment); err != nil {
		errors.Add("CreateEnvironment", testEnvironment, err)
	}
	stagingEnvironment := project.NewEnvironment("Staging").
		Set("CAPISTRANO_ENV", "staging")
	if err := self.conf.Environments.CreateEnvironment(stagingEnvironment); err != nil {
		errors.Add("CreateEnvironment", stagingEnvironment, err)
	}
	productionEnvironment := project.NewEnvironment("Production").
		Set("CAPISTRANO_ENV", "production")
	if err := self.conf.Environments.CreateEnvironment(productionEnvironment); err != nil {
		errors.Add("CreateEnvironment", productionEnvironment, err)
	}

	deployTask := project.NewTask("Deploy", self.DeployTaskBody())
	if err := self.conf.Tasks.CreateTask(deployTask); err != nil {
		errors.Add("CreateTask", deployTask, err)
	}

	notesTask := project.NewTask("Notes", self.NotesTaskBody())
	if err := self.conf.Tasks.CreateTask(notesTask); err != nil {
		errors.Add("CreateTask", notesTask, err)
	}

	checkDependenciesTask := project.NewTask("Check dependencies", self.CheckDependenciesTaskBody())
	if err := self.conf.Tasks.CreateTask(checkDependenciesTask); err != nil {
		errors.Add("CreateTask", checkDependenciesTask, err)
	}

	runTestsTask := project.NewTask("Run tests", self.RunTestsTaskBody())
	if err := self.conf.Tasks.CreateTask(runTestsTask); err != nil {
		errors.Add("CreateTask", runTestsTask, err)
	}

	notesJob := &domain.Job{}
	runTestsJob := &domain.Job{}
	deployToStagingJob := &domain.Job{}
	jobsToCreate := []struct {
		Task        *domain.Task
		Environment *domain.Environment
		SaveAs      **domain.Job
	}{
		{runTestsTask, testEnvironment, &runTestsJob},
		{notesTask, testEnvironment, &notesJob},
		{checkDependenciesTask, testEnvironment, nil},
		{deployTask, stagingEnvironment, &deployToStagingJob},
		{deployTask, productionEnvironment, nil},
	}

	for _, jobToCreate := range jobsToCreate {
		jobName := fmt.Sprintf("%s - %s", jobToCreate.Environment.Name, jobToCreate.Task.Name)
		job := project.NewJob(jobName, jobToCreate.Task.Uuid, jobToCreate.Environment.Uuid)
		if err := self.conf.Jobs.CreateJob(job); err != nil {
			errors.Add("CreateJob", job, err)
		}
		if jobToCreate.SaveAs != nil {
			*jobToCreate.SaveAs = job
		}
	}

	if strings.Contains(self.conf.NotifyViaEmail, "@") {
		emailNotifier := project.NewEmailNotifier(self.conf.NotifyViaEmail, self.conf.UrlHost)
		if err := self.conf.EmailNotifiers.CreateEmailNotifier(emailNotifier); err != nil {
			errors.Add("CreateEmailNotifier", emailNotifier, err)
		}

		notifyAboutNotes := project.NewNotificationRule(
			"email_notifiers",
			emailNotifier.Uuid,
			notesJob.Uuid,
			"operation.succeeded",
		)
		notifyAboutNotes.CreatorUuid = self.conf.UserUuid

		if err := self.conf.NotificationRules.CreateNotificationRule(notifyAboutNotes); err != nil {
			errors.Add("CreateNotificationRule", notifyAboutNotes, err)
		}
	}

	runNotesWeekly := notesJob.NewRecurringSchedule(self.conf.UserUuid, "@weekly")
	if err := self.conf.Schedules.CreateSchedule(runNotesWeekly); err != nil {
		errors.Add("CreateSchedule", runNotesWeekly, err)
	}

	triggerTestsOnNewCommits := project.NewGitTrigger(
		"run tests",
		self.conf.UserUuid,
		runTestsJob.Uuid,
	)

	if err := self.conf.GitTriggers.CreateGitTrigger(triggerTestsOnNewCommits); err != nil {
		errors.Add("CreateGitTrigger", triggerTestsOnNewCommits, err)
	}

	stagingDeployKey := stagingEnvironment.NewSecret("Deploy key", domain.SecretSsh)
	if err := self.conf.Secrets.CreateSecret(stagingDeployKey); err != nil {
		errors.Add("CreateSecret", stagingDeployKey, err)
	}

	productionDeployKey := productionEnvironment.NewSecret("Deploy key", domain.SecretSsh)
	if err := self.conf.Secrets.CreateSecret(productionDeployKey); err != nil {
		errors.Add("CreateSecret", productionDeployKey, err)
	}

	currentUser, err := self.conf.Users.FindUser(self.conf.UserUuid)
	if err != nil && !domain.IsNotFound(err) {
		errors.Add("FindUser", self.conf.UserUuid, err)
	}
	if currentUser != nil {
		self.setUpJobNotifierForDeployingToStaging(errors, currentUser, project, runTestsJob, deployToStagingJob)
	}

	return errors.ToError()
}

func (self *CapistranoRails) setUpJobNotifierForDeployingToStaging(errors *Error, currentUser *domain.User, project *domain.Project, triggeringJob *domain.Job, jobToTrigger *domain.Job) {
	deployToStagingAfterTestsRan := jobToTrigger.NewJobNotifier()
	deployToStagingAfterTestsRan.ProjectUuid = project.Uuid
	webhook := domain.NewWebhook(
		self.conf.ProjectUuid,
		currentUser.Uuid,
		jobToTrigger.Uuid,
		fmt.Sprintf("urn:harrow:job-notifier:%s", deployToStagingAfterTestsRan.Uuid),
	)

	if err := self.conf.Webhooks.CreateWebhook(webhook); err != nil {
		errors.Add("CreateWebhook", webhook, err)
		return
	}

	deployToStagingAfterTestsRan.WebhookURL = webhook.Links(map[string]map[string]string{}, "https", currentUser.UrlHost+"/api")["deliver"]["href"]

	if err := self.conf.JobNotifiers.CreateJobNotifier(deployToStagingAfterTestsRan); err != nil {
		errors.Add("CreateJobNotifier", deployToStagingAfterTestsRan, err)
	}

	notificationRule := project.NewNotificationRule(
		"job_notifiers",
		deployToStagingAfterTestsRan.Uuid,
		triggeringJob.Uuid,
		"operation.succeeded",
	)
	notificationRule.CreatorUuid = currentUser.Uuid

	if err := notificationRule.Validate(); err != nil {
		errors.Add("notificationRule.Validate", notificationRule, err)
	}

	if err := self.conf.NotificationRules.CreateNotificationRule(notificationRule); err != nil {
		errors.Add("CreateNotificationRule", notificationRule, err)
	}
}

// DeployTaskBody returns the body of the "deploy" task.
func (self *CapistranoRails) DeployTaskBody() string {
	return `#!/bin/bash -e

find ~/repositories -type f -name 'Capfile' -print0 |
  xargs -L 1 -0 -I% /bin/bash -c '
cd $(dirname $1)
hfold bundling bundle install
bundle exec cap $CAPISTRANO_ENV deploy
' _ %
`
}

// CheckDependenciesTaskBody returns the body of the "check
// dependencies" task.
func (self *CapistranoRails) CheckDependenciesTaskBody() string {
	return `#!/bin/bash -e

find ~/repositories -type f -name 'Gemfile' -print0 |
  xargs -L 1 -0 -I% /bin/bash -c '
cd $(dirname $1)
hfold bundling bundle install
bundle outdated
' _ %
`
}

// NotesTaskBody returns the body of the task for displaying project
// notes.
func (self *CapistranoRails) NotesTaskBody() string {
	return `#!/bin/bash -e

find ~/repositories -type f -name 'Rakefile' -print0 |
  xargs -L 1 -0 -I% /bin/bash -c '
cd $(dirname $1)
hfold bundling bundle install
bundle exec rake notes
' _ %
`
}

// RunTestsTaskBody returns the body of the task for running the
// project's unit tests.
func (self *CapistranoRails) RunTestsTaskBody() string {
	return `#!/bin/bash -e

find ~/repositories -type f -name 'Rakefile' -print0 |
  xargs -L 1 -0 -I% /bin/bash -c '
cd $(dirname $1)
hfold bundling bundle install
bundle exec rake test
' _ %
`
}
