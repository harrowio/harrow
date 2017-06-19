package stencil

import (
	"fmt"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/uuidhelper"
)

type EnvironmentsInMemory struct {
	environments map[string]*domain.Environment
}

func NewEnvironmentsInMemory() *EnvironmentsInMemory {
	return &EnvironmentsInMemory{
		environments: map[string]*domain.Environment{},
	}
}

// Count returns the number environments stored in this instance
func (self *EnvironmentsInMemory) Count() int {
	return len(self.environments)
}

// FindEnvironmentByName returns the first environment stored by this instance
// with a name equaling the provided name.
func (self *EnvironmentsInMemory) FindEnvironmentByName(name string) (*domain.Environment, error) {
	for _, environment := range self.environments {
		if environment.Name == name {
			return environment, nil
		}
	}

	return nil, new(domain.NotFoundError)
}

// CreateEnvironment creates a new environment and stores it in memory.
func (self *EnvironmentsInMemory) CreateEnvironment(environment *domain.Environment) error {
	if !uuidhelper.IsValid(environment.Uuid) {
		environment.Uuid = uuidhelper.MustNewV4()
	}

	self.environments[environment.Uuid] = environment
	return nil
}

type ProjectsInMemory struct {
	projects map[string]*domain.Project
}

func NewProjectsInMemory() *ProjectsInMemory {
	return &ProjectsInMemory{
		projects: map[string]*domain.Project{},
	}
}

// Add adds project to this store.  If project has no uuid, a new uuid
// is generated.
func (self *ProjectsInMemory) Add(project *domain.Project) *ProjectsInMemory {
	if !uuidhelper.IsValid(project.Uuid) {
		project.Uuid = uuidhelper.MustNewV4()
	}

	self.projects[project.Uuid] = project
	return self
}

func (self *ProjectsInMemory) FindProject(uuid string) (*domain.Project, error) {
	if project, found := self.projects[uuid]; found {
		return project, nil
	} else {
		return nil, new(domain.NotFoundError)
	}
}

type TasksInMemory struct {
	tasks map[string]*domain.Task
}

func NewTasksInMemory() *TasksInMemory {
	return &TasksInMemory{
		tasks: map[string]*domain.Task{},
	}
}

// Count returns the number of tasks stored in this instance.
func (self *TasksInMemory) Count() int {
	return len(self.tasks)
}

// FindByName returns the first task stored by this instance
// with a name equaling the provided name.
func (self *TasksInMemory) FindTaskByName(name string) (*domain.Task, error) {
	for _, task := range self.tasks {
		if task.Name == name {
			return task, nil
		}
	}

	return nil, new(domain.NotFoundError)
}

// CreateTask creates a new task and stores it in memory.
func (self *TasksInMemory) CreateTask(task *domain.Task) error {
	if !uuidhelper.IsValid(task.Uuid) {
		task.Uuid = uuidhelper.MustNewV4()
	}

	self.tasks[task.Uuid] = task
	return nil
}

type JobsInMemory struct {
	jobs map[string]*domain.Job
}

func NewJobsInMemory() *JobsInMemory {
	return &JobsInMemory{
		jobs: map[string]*domain.Job{},
	}
}

// FindByTaskAndEnvironmentUuid returns the first job stored by this
// instance for the task identified by task uuid.
func (self *JobsInMemory) FindByTaskAndEnvironmentUuid(taskUuid, environmentUuid string) (*domain.Job, error) {
	for _, job := range self.jobs {
		if job.TaskUuid == taskUuid && job.EnvironmentUuid == environmentUuid {
			return job, nil
		}
	}

	return nil, new(domain.NotFoundError)
}

// FindJobByName returns a stored job with the given name or an error if
// no such job can be found.
func (self *JobsInMemory) FindJobByName(name string) (*domain.Job, error) {
	for _, job := range self.jobs {
		if job.Name == name {
			return job, nil
		}
	}

	return nil, new(domain.NotFoundError)
}

// CreateJob creates a new job and stores it in memory.
func (self *JobsInMemory) CreateJob(job *domain.Job) error {
	if !uuidhelper.IsValid(job.Uuid) {
		job.Uuid = uuidhelper.MustNewV4()
	}

	self.jobs[job.Uuid] = job
	return nil
}

// Count returns the number of jobs stored in this instance.
func (self *JobsInMemory) Count() int {
	return len(self.jobs)
}

type EmailNotifiersInMemory struct {
	emailNotifiers map[string]*domain.EmailNotifier
}

func NewEmailNotifiersInMemory() *EmailNotifiersInMemory {
	return &EmailNotifiersInMemory{
		emailNotifiers: map[string]*domain.EmailNotifier{},
	}
}

// CreateEmailNotifier creates a new emailNotifier and stores it in memory.
func (self *EmailNotifiersInMemory) CreateEmailNotifier(emailNotifier *domain.EmailNotifier) error {
	if !uuidhelper.IsValid(emailNotifier.Uuid) {
		emailNotifier.Uuid = uuidhelper.MustNewV4()
	}

	self.emailNotifiers[emailNotifier.Uuid] = emailNotifier
	return nil
}

// FindByEmailAddress returns an email notifier for address or nil if
// no such notifier has been stored.
func (self *EmailNotifiersInMemory) FindByEmailAddress(address string) *domain.EmailNotifier {
	for _, notifier := range self.emailNotifiers {
		if notifier.Recipient == address {
			return notifier
		}
	}

	return nil
}

type NotificationRulesInMemory struct {
	notificationRules map[string]*domain.NotificationRule
}

func NewNotificationRulesInMemory() *NotificationRulesInMemory {
	return &NotificationRulesInMemory{
		notificationRules: map[string]*domain.NotificationRule{},
	}
}

// CreateNotificationRule creates a new notificationRule and stores it in memory.
func (self *NotificationRulesInMemory) CreateNotificationRule(notificationRule *domain.NotificationRule) error {
	if !uuidhelper.IsValid(notificationRule.Uuid) {
		notificationRule.Uuid = uuidhelper.MustNewV4()
	}

	self.notificationRules[notificationRule.Uuid] = notificationRule
	return nil
}

// FindByNotifierAndJobUuidAndType returns the first notification rule
// which matches the given parameters.  Matching is performed with an
// equality test.
func (self *NotificationRulesInMemory) FindByNotifierAndJobUuidAndType(notifierUuid, jobUuid, notifierType string) (*domain.NotificationRule, error) {
	for _, notificationRule := range self.notificationRules {
		if notificationRule.NotifierType == notifierType &&
			notificationRule.JobUuid != nil && *notificationRule.JobUuid == jobUuid &&
			notificationRule.NotifierUuid == notifierUuid {
			return notificationRule, nil
		}
	}

	return nil, new(domain.NotFoundError)
}

type SchedulesInMemory struct {
	schedules map[string]*domain.Schedule
}

func NewSchedulesInMemory() *SchedulesInMemory {
	return &SchedulesInMemory{
		schedules: map[string]*domain.Schedule{},
	}
}

// CreateSchedule creates a new schedule and stores it in memory.
func (self *SchedulesInMemory) CreateSchedule(schedule *domain.Schedule) error {
	if !uuidhelper.IsValid(schedule.Uuid) {
		schedule.Uuid = uuidhelper.MustNewV4()
	}

	self.schedules[schedule.Uuid] = schedule
	return nil
}

// FindByJobUuid returns the first schedule encountered for the given
// jobUuid.
func (self *SchedulesInMemory) FindByJobUuid(jobUuid string) (*domain.Schedule, error) {
	for _, schedule := range self.schedules {
		if schedule.JobUuid == jobUuid {
			return schedule, nil
		}
	}

	return nil, new(domain.NotFoundError)
}

type GitTriggersInMemory struct {
	gitTriggers map[string]*domain.GitTrigger
}

func NewGitTriggersInMemory() *GitTriggersInMemory {
	return &GitTriggersInMemory{
		gitTriggers: map[string]*domain.GitTrigger{},
	}
}

// CreateGitTrigger creates a new gitTrigger and stores it in memory.
func (self *GitTriggersInMemory) CreateGitTrigger(gitTrigger *domain.GitTrigger) error {
	if !uuidhelper.IsValid(gitTrigger.Uuid) {
		gitTrigger.Uuid = uuidhelper.MustNewV4()
	}

	self.gitTriggers[gitTrigger.Uuid] = gitTrigger
	return nil
}

// FindByProjectUuid returns the first git trigger with the given
// project uuid, or nil of no such trigger exists.
func (self *GitTriggersInMemory) FindByProjectUuid(uuid string) *domain.GitTrigger {
	for _, trigger := range self.gitTriggers {
		if trigger.ProjectUuid == uuid {
			return trigger
		}
	}

	return nil
}

type SecretsInMemory struct {
	secrets map[string]*domain.Secret
}

func NewSecretsInMemory() *SecretsInMemory {
	return &SecretsInMemory{
		secrets: map[string]*domain.Secret{},
	}
}

// CreateSecret creates a new secret and stores it in memory.
func (self *SecretsInMemory) CreateSecret(secret *domain.Secret) error {
	if !uuidhelper.IsValid(secret.Uuid) {
		secret.Uuid = uuidhelper.MustNewV4()
	}

	self.secrets[secret.Uuid] = secret
	return nil
}

// FindByEnvironmentUuid returns the first secret with the given
// environment uuid, or nil of no such trigger exists.
func (self *SecretsInMemory) FindByEnvironmentUuid(uuid string) *domain.Secret {
	for _, secret := range self.secrets {
		if secret.EnvironmentUuid == uuid {
			return secret
		}
	}

	return nil
}

type JobNotifiersInMemory struct {
	jobNotifiers map[string]*domain.JobNotifier
}

func NewJobNotifiersInMemory() *JobNotifiersInMemory {
	return &JobNotifiersInMemory{
		jobNotifiers: map[string]*domain.JobNotifier{},
	}
}

func (self *JobNotifiersInMemory) FindByTriggeringJobUuid(triggeringJobUuid string) *domain.JobNotifier {
	fmt.Println("looking for", triggeringJobUuid)
	for k, v := range self.jobNotifiers {
		fmt.Printf("%s: %#v\n", k, v)
		if v.Uuid == triggeringJobUuid {
			return v
		}
	}
	return nil
}

func (self *JobNotifiersInMemory) CreateJobNotifier(jobNotifier *domain.JobNotifier) error {
	if !uuidhelper.IsValid(jobNotifier.Uuid) {
		jobNotifier.Uuid = uuidhelper.MustNewV4()
	}
	self.jobNotifiers[jobNotifier.Uuid] = jobNotifier
	return nil
}

type UsersInMemory struct {
	users map[string]*domain.User
}

func NewUsersInMemory() *UsersInMemory {
	return &UsersInMemory{
		users: map[string]*domain.User{},
	}
}

// Add adds user to this store.  If user has no uuid, a new uuid
// is generated.
func (self *UsersInMemory) Add(user *domain.User) *UsersInMemory {
	if !uuidhelper.IsValid(user.Uuid) {
		user.Uuid = uuidhelper.MustNewV4()
	}

	self.users[user.Uuid] = user
	return self
}

func (self *UsersInMemory) FindUser(uuid string) (*domain.User, error) {
	if user, found := self.users[uuid]; found {
		return user, nil
	} else {
		return nil, new(domain.NotFoundError)
	}
}

type WebhooksInMemory struct {
	webhooks map[string]*domain.Webhook
}

func NewWebhooksInMemory() *WebhooksInMemory {
	return &WebhooksInMemory{
		webhooks: map[string]*domain.Webhook{},
	}
}

// FindByName returns the first webhook stored by this instance
// with a name equaling the provided name.
func (self *WebhooksInMemory) FindByName(name string) (*domain.Webhook, error) {
	for _, webhook := range self.webhooks {
		if webhook.Name == name {
			return webhook, nil
		}
	}

	return nil, new(domain.NotFoundError)
}

// CreateWebhook creates a new webhook and stores it in memory.
func (self *WebhooksInMemory) CreateWebhook(webhook *domain.Webhook) error {
	if !uuidhelper.IsValid(webhook.Uuid) {
		webhook.Uuid = uuidhelper.MustNewV4()
	}

	self.webhooks[webhook.Uuid] = webhook
	return nil
}
