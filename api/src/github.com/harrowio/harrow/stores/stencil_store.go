package stores

import (
	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/domain/stencil"
	"github.com/harrowio/harrow/logger"
	"github.com/jmoiron/sqlx"
)

type ActivitySink interface {
	EnqueueActivity(activity *domain.Activity, userUuid *string)
}

// DbStencilStore is a convenience class for fulfilling many of
// stencil.Configuration's dependencies.
type DbStencilStore struct {
	bus               ActivitySink
	tx                *sqlx.Tx
	log               logger.Logger
	environments      *DbEnvironmentStore
	projects          *DbProjectStore
	tasks             *DbTaskStore
	jobs              *DbJobStore
	emailNotifiers    *DbEmailNotifierStore
	notificationRules *DbNotificationRuleStore
	schedules         *DbScheduleStore
	gitTriggers       *DbGitTriggerStore
	secrets           *SecretStore
	users             *DbUserStore
	webhooks          *DbWebhookStore
	jobNotifiers      *DbJobNotifierStore
}

func NewDbStencilStore(ss SecretKeyValueStore, tx *sqlx.Tx, bus ActivitySink) *DbStencilStore {
	conf := config.GetConfig()
	return &DbStencilStore{
		bus:               bus,
		tx:                tx,
		environments:      NewDbEnvironmentStore(tx),
		projects:          NewDbProjectStore(tx),
		tasks:             NewDbTaskStore(tx),
		jobs:              NewDbJobStore(tx),
		emailNotifiers:    NewDbEmailNotifierStore(tx),
		notificationRules: NewDbNotificationRuleStore(tx),
		schedules:         NewDbScheduleStore(tx),
		gitTriggers:       NewDbGitTriggerStore(tx),
		secrets:           NewSecretStore(ss, tx),
		users:             NewDbUserStore(tx, conf),
		webhooks:          NewDbWebhookStore(tx),
		jobNotifiers:      NewDbJobNotifierStore(tx),
	}
}

func (self *DbStencilStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *DbStencilStore) SetLogger(l logger.Logger) {
	self.log = l
}

func (self *DbStencilStore) Create(subject *domain.Stencil) error {
	configuration := self.ToConfiguration()
	configuration.NotifyViaEmail = subject.NotifyViaEmail
	configuration.UrlHost = subject.UrlHost
	configuration.UserUuid = subject.UserUuid
	configuration.ProjectUuid = subject.ProjectUuid

	switch subject.Id {
	case "capistrano-rails":
		return stencil.NewCapistranoRails(configuration).Create()
	case "bash-linux":
		return stencil.NewBashLinux(configuration).Create()
	default:
		return domain.NewValidationError("id", "unknown")
	}
}

func (self *DbStencilStore) ToConfiguration() *stencil.Configuration {
	return &stencil.Configuration{
		Environments:      self,
		Projects:          self,
		Tasks:             self,
		Jobs:              self,
		EmailNotifiers:    self,
		NotificationRules: self,
		Schedules:         self,
		GitTriggers:       self,
		Secrets:           self,
		Users:             self,
		JobNotifiers:      self,
		Webhooks:          self,
	}
}

func (self *DbStencilStore) CreateEnvironment(environment *domain.Environment) error {
	_, err := self.environments.Create(environment)
	self.bus.EnqueueActivity(activities.EnvironmentAdded(environment), nil)
	return err
}

func (self *DbStencilStore) FindEnvironmentByName(environmentName string) (*domain.Environment, error) {
	return self.environments.FindByName(environmentName)
}

func (self *DbStencilStore) FindProject(uuid string) (*domain.Project, error) {
	return self.projects.FindByUuid(uuid)
}

func (self *DbStencilStore) CreateTask(task *domain.Task) error {
	_, err := self.tasks.Create(task)
	self.bus.EnqueueActivity(activities.TaskAdded(task), nil)
	return err
}

func (self *DbStencilStore) FindTaskByName(taskName string) (*domain.Task, error) {
	return self.tasks.FindByName(taskName)
}

func (self *DbStencilStore) CreateJob(job *domain.Job) error {
	_, err := self.jobs.Create(job)
	self.bus.EnqueueActivity(activities.JobAdded(job), nil)
	return err
}

func (self *DbStencilStore) FindJobByName(jobName string) (*domain.Job, error) {
	return self.jobs.FindByName(jobName)
}

func (self *DbStencilStore) CreateEmailNotifier(notifier *domain.EmailNotifier) error {
	_, err := self.emailNotifiers.Create(notifier)
	self.bus.EnqueueActivity(activities.EmailNotifierCreated(notifier), nil)
	return err
}

func (self *DbStencilStore) CreateNotificationRule(notificationRule *domain.NotificationRule) error {
	_, err := self.notificationRules.Create(notificationRule)
	self.bus.EnqueueActivity(activities.NotificationRuleCreated(notificationRule), nil)
	return err
}

func (self *DbStencilStore) CreateSchedule(schedule *domain.Schedule) error {
	_, err := self.schedules.Create(schedule)
	self.bus.EnqueueActivity(activities.JobScheduled(schedule, "stencil"), nil)
	return err
}

func (self *DbStencilStore) CreateGitTrigger(gitTrigger *domain.GitTrigger) error {
	_, err := self.gitTriggers.Create(gitTrigger)
	self.bus.EnqueueActivity(activities.GitTriggerCreated(gitTrigger), nil)
	return err
}

func (self *DbStencilStore) CreateSecret(secret *domain.Secret) error {
	_, err := self.secrets.Create(secret)
	self.bus.EnqueueActivity(activities.SecretAdded(secret), nil)
	return err
}

func (self *DbStencilStore) FindUser(uuid string) (*domain.User, error) {
	return self.users.FindByUuid(uuid)
}

func (self *DbStencilStore) CreateJobNotifier(jobNotifier *domain.JobNotifier) error {
	_, err := self.jobNotifiers.Create(jobNotifier)
	self.bus.EnqueueActivity(activities.JobNotifierCreated(jobNotifier), nil)
	return err
}

func (self *DbStencilStore) CreateWebhook(webhook *domain.Webhook) error {
	_, err := self.webhooks.Create(webhook)
	self.bus.EnqueueActivity(activities.WebhookCreated(webhook), nil)
	return err
}
