package stencil

import "github.com/harrowio/harrow/domain"

// Configuration collects all the necessary components for creating
// stencils.
type Configuration struct {
	// UserUuid identifies the user who requested instantiation of
	// a stencil.
	UserUuid string

	// ProjectUuid identifies the project in which to create this
	// stencil.
	ProjectUuid string

	// NotifyViaEmail causes the stencil to set up an email
	// notifier, if it is set to any string containing an "@".
	NotifyViaEmail string

	// UrlHost is the host name to be used when generating URLs in
	// notifiers.
	UrlHost string

	Environments      EnvironmentStore
	Projects          ProjectStore
	Tasks             TaskStore
	Jobs              JobStore
	EmailNotifiers    EmailNotifierStore
	JobNotifiers      JobNotifierStore
	NotificationRules NotificationRuleStore
	Schedules         ScheduleStore
	GitTriggers       GitTriggerStore
	Secrets           SecretStore
	Users             UserStore
	Webhooks          WebhookStore
}

// EnvironmentStore allows a stencil to create environments.
type EnvironmentStore interface {
	CreateEnvironment(environment *domain.Environment) error
	FindEnvironmentByName(environmentName string) (*domain.Environment, error)
}

// ProjectStore allows a stencil to load a project and provide the
// necessary context for creating other objects.
type ProjectStore interface {
	FindProject(uuid string) (*domain.Project, error)
}

// UserStore allows a stencil to load a user and provide the
// necessary context for creating other objects.
type UserStore interface {
	FindUser(uuid string) (*domain.User, error)
}

// TaskStore allows a stencil to create tasks for a project.
type TaskStore interface {
	CreateTask(task *domain.Task) error
	FindTaskByName(taskName string) (*domain.Task, error)
}

// JobStore allows a stencil to create jobs for a project.
type JobStore interface {
	CreateJob(job *domain.Job) error
	FindJobByName(jobName string) (*domain.Job, error)
}

// EmailNotifierStore allows a stencil to create email notifiers for a
// project.
type EmailNotifierStore interface {
	CreateEmailNotifier(notifier *domain.EmailNotifier) error
}

// NotificationRuleStore allows a stencil to create notification rules
// for wiring up notifiers with jobs.
type NotificationRuleStore interface {
	CreateNotificationRule(notificationRule *domain.NotificationRule) error
}

// ScheduleStore allows a stencil to create a schedule for scheduling
// recurring jobs.
type ScheduleStore interface {
	CreateSchedule(schedule *domain.Schedule) error
}

// GitTriggerStore allows a stencil to create git triggers.
type GitTriggerStore interface {
	CreateGitTrigger(gitTrigger *domain.GitTrigger) error
}

// SecretStore allows a stencil to create SSH secrets.
type SecretStore interface {
	CreateSecret(secret *domain.Secret) error
}

// JobNotifierStore allows a stencil to create job notifiers for a
// project.
type JobNotifierStore interface {
	CreateJobNotifier(notifier *domain.JobNotifier) error
}

// WebhookStore allows a stencil to create webhooks.
type WebhookStore interface {
	CreateWebhook(webhook *domain.Webhook) error
}
