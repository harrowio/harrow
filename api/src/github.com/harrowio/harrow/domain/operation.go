package domain

import (
	"crypto/sha1"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/harrowio/harrow/bus/logevent"
	"github.com/harrowio/harrow/posix"
)

var (
	// ErrUnknownOperationCategory indicates that no script can be
	// run for operations of the given type.
	ErrUnknownOperationCategory = errors.New("operation: unknown category")

	operationEvents = []string{
		EventOperationScheduled,
		EventOperationStarted,
		EventOperationSucceeded,
		EventOperationFailed,
		EventOperationTimedOut,
	}
)

type OperationTriggerReason string

const (
	OperationTriggeredByWebhook          OperationTriggerReason = "webhook"
	OperationTriggeredBySchedule         OperationTriggerReason = "schedule"
	OperationTriggeredByUser             OperationTriggerReason = "user"
	OperationTriggeredByGitTrigger       OperationTriggerReason = "git-trigger"
	OperationTriggeredByNotificationRule OperationTriggerReason = "notification-rule"
)

func (self OperationTriggerReason) String() string { return string(self) }

// NOTE(dh):  This should probably be split into separate types
// (JobOperation, RepositoryOperation, ...) once a new operation type
// is added.  Distinguishing operations by the value of their `Type` field
// seems to work for now, because there are only two cases.
type Operation struct {
	defaultSubject
	Uuid                   string     `json:"uuid"`
	Type                   string     `json:"type"`
	RepositoryUuid         *string    `json:"repositoryUuid"         db:"repository_uuid"`
	JobUuid                *string    `json:"jobUuid"                db:"job_uuid"`
	NotifierUuid           *string    `json:"notifierUuid"           db:"notifier_uuid"`
	NotifierType           *string    `json:"notifierType"           db:"notifier_type"`
	WorkspaceBaseImageUuid string     `json:"workspaceBaseImageUuid" db:"workspace_base_image_uuid"`
	TimeLimit              int        `json:"timeLimitSecs"          db:"time_limit"`
	VisibleTo              string     `json:"-"                      db:"visible_to"`
	ExitStatus             int        `json:"exitStatus"             db:"exit_status"`
	CreatedAt              *time.Time `json:"createdAt"              db:"created_at"`
	StartedAt              *time.Time `json:"startedAt"              db:"started_at"`
	FinishedAt             *time.Time `json:"finishedAt"             db:"finished_at"`
	ArchivedAt             *time.Time `json:"archivedAt"             db:"archived_at"`
	FailedAt               *time.Time `json:"failedAt"               db:"failed_at"`
	TimedOutAt             *time.Time `json:"timedOutAt"             db:"timed_out_at"`
	CanceledAt             *time.Time `json:"canceledAt"             db:"canceled_at"`
	FatalError             *string    `json:"fatalError"             db:"fatal_error"`

	Parameters *OperationParameters `json:"parameters" db:"parameters"`

	RepositoryCheckouts *RepositoryCheckouts `json:"repositoryCheckouts" db:"repository_refs"`

	GitLogs *GitLogs `json:"gitLogs" db:"git_logs"`

	StatusLogs *StatusLogs `json:"statusLogs" db:"status_logs"`

	LogEvents []*logevent.Message `json:"logEvents" db:"-"`
}

type sshConfig struct {
	Host,
	Port,
	SSHHostAlias,
	User,
	RepoUuid,
	KeyFileName string
}

type OperationParameters struct {
	// Checkout is a map of repository uuids to git references
	// that should be checked out before running an operation.
	Checkout map[string]string `json:"checkout"`

	// Reason is the reason for why this operation was triggered.
	Reason OperationTriggerReason `json:"reason"`

	// Username is the name of the user who triggered the
	// operation.
	Username string `json:"username,omitempty"`
	// UserUuid is the uuid of the user who triggered the
	// operation.
	UserUuid string `json:"userUuid,omitempty"`

	// ScheduleDescription is the description of the schedule which triggered
	// this operation.
	ScheduleDescription string `json:"scheduleDescription,omitempty"`

	// ScheduleUuid is the uuid of the schedule that triggered the
	// operation.
	ScheduleUuid string `json:"scheduleUuid,omitempty"`

	// TriggeredByDelivery is the uuid of the delivery that
	// triggered this operation.
	TriggeredByDelivery string `json:"triggeredByDelivery"`

	// TriggeredByGitTrigger is the uuid of the Git trigger that
	// triggered this operation.
	TriggeredByGitTrigger string `json:"triggeredByGitTrigger"`

	// GitTriggerName is the name of the Git trigger that
	// triggered this operation.
	GitTriggerName string `json:"gitTriggerName"`

	// TriggeredByNotificationRule is the uuid of the notification
	// rule that triggered this operation.
	TriggeredByNotificationRule string `json:"triggeredByNotificationRule"`

	// TriggeredByActivityId is the id of the activity that
	// triggered this operation.
	TriggeredByActivityId int `json:"triggeredByActivityId"`

	// Environment specifies a new environment that should be used
	// instead of the one specified by the job.
	Environment *Environment `json:"environment"`

	// Task specifies a new task that should be used instead of
	// the one specified by the job.
	Task *Task `json:"task"`

	// Secrets specifies new secrets that should be used instead
	// of the ones specified by the environment.
	Secrets []*OperationSecret `json:"secrets"`
}

type OperationSecret struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	OldValue *string `json:"oldValue,omitempty"`
}

func NewOperationParameters() *OperationParameters {
	params := &OperationParameters{}
	return params.Init()
}

func (self *OperationParameters) AddSecret(name, value string) *OperationParameters {
	self.Secrets = append(self.Secrets, &OperationSecret{
		Name:  name,
		Value: value,
	})
	return self
}

func (self *OperationParameters) Init() *OperationParameters {
	if self == nil {
		*self = OperationParameters{}
	}
	if self.Checkout == nil {
		self.Checkout = map[string]string{}
	}
	return self
}

func (self *OperationParameters) Value() (driver.Value, error) {
	if self == nil {
		self = &OperationParameters{}
	}
	data, err := json.Marshal(self.Init())
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func (self *OperationParameters) Scan(from interface{}) error {
	if from == nil {
		self.Init()
		return nil
	}

	dest := OperationParameters{}
	dest.Init()
	src := []byte{}
	switch data := from.(type) {
	case []byte:
		src = data
	case string:
		src = []byte(data)
	default:
		return fmt.Errorf("OperationParameters: cannot scan from %#v", from)
	}

	if err := json.Unmarshal(src, &dest); err != nil {
		return err
	}

	*self = dest
	self.Init()
	return nil
}

func (self *Operation) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/operations/%s", requestScheme, requestBaseUri, self.Uuid)
}

func (self *Operation) Status() string {
	if self.CanceledAt != nil {
		return "canceled"
	}

	if self.TimedOutAt != nil {
		return "timeout"
	}

	if self.FatalError != nil {
		return "fatal"
	}

	if self.FailedAt != nil {
		return "failure"
	}

	if self.FinishedAt == nil {
		return "active"
	}

	if self.Successful() {
		return "success"
	}

	return "failure"
}

func (self *Operation) ShieldsIoStatusColorAndParams() (string, string) {

	if self.CanceledAt != nil {
		return "blue", ""
	}

	if self.TimedOutAt != nil {
		return "blue", ""
	}

	if self.FatalError != nil {
		return "ignored", "?colorB=000000"
	}

	if self.FailedAt != nil {
		return "red", ""
	}

	if self.Successful() {
		return "green", ""
	}

	return "lightgrey", ""
}

func (self *Operation) Successful() bool {
	return self.ExitStatus == 0
}

func (self *Operation) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	if self.RepositoryUuid != nil {
		response["repository"] = map[string]string{"href": fmt.Sprintf("%s://%s/repositories/%s", requestScheme, requestBaseUri, *self.RepositoryUuid)}
	}
	if self.JobUuid != nil {
		response["job"] = map[string]string{"href": fmt.Sprintf("%s://%s/jobs/%s", requestScheme, requestBaseUri, *self.JobUuid)}
	}
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	return response
}

func (self *Operation) FindJob(store JobStore) (*Job, error) {
	if self.JobUuid == nil {
		return nil, new(NotFoundError)
	}

	return store.FindByUuid(*self.JobUuid)
}

// FindProject satisfies authz.BelongsToProject in order to determine
// authorization.
func (self *Operation) FindProject(store ProjectStore) (*Project, error) {
	if self.JobUuid != nil {
		return store.FindByJobUuid(*self.JobUuid)
	} else if self.RepositoryUuid != nil {
		return store.FindByRepositoryUuid(*self.RepositoryUuid)
	} else if self.NotifierUuid != nil {
		return store.FindByNotifierUuid(*self.NotifierUuid, *self.NotifierType)
	} else {
		return nil, nil
	}
}

// FindWorkspaceBaseImage returns the workspace base image that should
// be used for running this operation.
func (self *Operation) FindWorkspaceBaseImage(store WorkspaceBaseImageStore) (*WorkspaceBaseImage, error) {
	return store.FindByUuid(self.WorkspaceBaseImageUuid)
}

// Environment returns the environment in which the operation runs.
// If the operation is not associated with an environment (such as a Git
// repository access check), nil is returned.  Any errors returned are
// from store.
func (self *Operation) Environment(store EnvironmentStore) (*Environment, error) {
	if self.Parameters != nil && self.Parameters.Environment != nil {
		return self.Parameters.Environment, nil
	}

	if self.JobUuid != nil {
		return store.FindByJobUuid(*self.JobUuid)
	}

	return nil, nil
}

// Task returns the task which is used for running this operation.
//
// If the operation is not associated with a task, nil is returned.
// Any errors returned are from store.
func (self *Operation) Task(store TaskStore) (*Task, error) {
	if self.Parameters != nil && self.Parameters.Task != nil {
		return self.Parameters.Task, nil
	}

	if self.JobUuid != nil {
		return store.FindByJobUuid(*self.JobUuid)
	}

	return nil, nil
}

// Secrets returns the secrets that are associated to this Operation.
func (self *Operation) Secrets(es EnvironmentStore, ss SecretStore) ([]*Secret, error) {
	if self.Parameters != nil && len(self.Parameters.Secrets) > 0 {

		result := []*Secret{}
		for _, secret := range self.Parameters.Secrets {
			result = append(result, &Secret{
				Type:        SecretEnvOverride,
				Name:        secret.Name,
				SecretBytes: []byte(secret.Value),
			})
		}

		env, err := self.Environment(es)
		if err != nil || env == nil {
			return result, nil
		}

		envSecrets, err := ss.FindAllByEnvironmentUuid(env.Uuid)
		if err != nil {
			return result, nil
		}

		for _, envSecret := range envSecrets {
			if envSecret.IsSsh() {
				result = append(result, envSecret)
			}
		}

		return result, nil
	}

	env, err := self.Environment(es)
	if err != nil || env == nil {
		return nil, err
	}

	return ss.FindAllByEnvironmentUuid(env.Uuid)
}

// Category returns the broader category an operation belongs into based
// on its type.  For example, the category of OperationTypeGit* operations is
// "repository".  If the type cannot be mapped to a category, "unknown"
// is returned.
func (self *Operation) Category() string {
	if sep := strings.Index(self.Type, "."); sep == -1 {
		return "unknown"
	} else {
		switch self.Type[0:sep] {
		case "git":
			return "repository"
		case "job":
			return "job"
		case "notifier":
			return "notifier"
		default:
			return "unknown"
		}
	}
}

// Repositories returns the list of repositories that will be used when
// running this operation.  If the operation doesn't require any
// repositories to exist (such as a repository access check operation),
// then an empty slice of repositories and no error is returned.
// Otherwise any error returned originates from store.
func (self *Operation) Repositories(store RepositoryStore) ([]*Repository, error) {
	if self.JobUuid != nil {
		return store.FindAllByJobUuid(*self.JobUuid)
	}
	if self.RepositoryUuid != nil {
		repo, err := store.FindByUuid(*self.RepositoryUuid)
		if err != nil {
			return nil, err
		}
		return []*Repository{repo}, nil
	}
	return []*Repository{}, nil
}

func (self *Operation) IsGitAccessCheck() bool {
	return self.Type == OperationTypeGitAccessCheck
}

func (self *Operation) IsGitMetadataCollect() bool {
	return self.Type == OperationTypeGitEnumerationBranches
}

func (self *Operation) IsUserJob() bool {
	return self.Category() == "job"
}

type OperationSetupScriptCtxt struct {
	// Operation is the operation for which this script runs
	Operation *Operation

	// Keys is a list of SSH key pairs that need to be present in the
	// workspace for cloning any repositories before running an operation.
	Keys []*keyPair

	// Repositories is the list of repositories that are related to this
	// operation.
	Repositories []*Repository
	SshConfigs   []*sshConfig

	// Parameters are operation-specific parameters that change
	// the state in which an operation runs.
	Parameters *OperationParameters

	// PreviousOperation is the last completed operation that ran
	// before this operation.  This can be nil if there is no such
	// operation.
	PreviousOperation *Operation

	// WebhookBody is the body of the request that triggered this
	// operation via a webhook.  If this operation has not been
	// triggered through a webhook, WebhookBody is nil.
	WebhookBody []byte

	// If the related repositories should be cloned.  If false, only the
	// deploy keys will be written.
	ShouldCloneRepos bool

	// Environment is the environment in which the operation runs.  This
	// value is used to define and export any necessary environment
	// variables before running the operation's user script.
	Environment *Environment

	// Secrets holds the EnvironmentSecrets for this job. They are being
	// exported as environment variables as well.
	Secrets []*EnvironmentSecret

	// WsbiName is the name of the docker workspace base image.
	// Operations are run in a docker container using this image.
	WsbiName string

	// WsbiRepository is the url of the git repository from where
	// the workspace base image can be obtained with "git clone".
	WsbiRepository string

	// WsbiPath is the path to change to within the repository for
	// workspace base images before invoking "make" to build the image.
	WsbiPath string
}

func (self *OperationSetupScriptCtxt) LoadPreviousOperation(operations OperationStore) error {
	previous, err := operations.FindPreviousOperation(self.Operation.Uuid)
	if err != nil {
		return err
	}

	if previous == nil {
		return nil
	}

	self.PreviousOperation = previous

	if self.PreviousOperation.RepositoryCheckouts == nil {
		self.PreviousOperation.RepositoryCheckouts = NewRepositoryCheckouts()
	}

	return nil
}

func (self *OperationSetupScriptCtxt) LoadWebhookBody(deliveries DeliveryStore) error {
	deliveryUuid := self.Parameters.TriggeredByDelivery
	if deliveryUuid == "" {
		return nil
	}
	delivery, err := deliveries.FindByUuid(deliveryUuid)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(delivery.Request.Body)
	if err != nil {
		return err
	}

	self.WebhookBody = body
	return nil
}

func (self *Operation) HandleEvent(payload EventPayload) {
	if self.RepositoryCheckouts == nil {
		self.RepositoryCheckouts = NewRepositoryCheckouts()
	}
	if self.GitLogs == nil {
		self.GitLogs = NewGitLogs()
	}
	if self.StatusLogs == nil {
		self.StatusLogs = NewStatusLogs()
	}

	self.RepositoryCheckouts.HandleEvent(payload)
	self.GitLogs.HandleEvent(payload)
	self.StatusLogs.HandleEvent(payload)
}

func (self *Operation) IsReady(repos RepositoryStore, credentials RepositoryCredentialStore, envs EnvironmentStore, secrets SecretStore) (bool, error) {

	env, err := self.Environment(envs)
	if err != nil {
		return false, err
	}

	// If there is no env (because it is a repository operation, for example),
	// just skip the secret check. Readyness is then solely determined by the
	// presence of RepositoryCredentials
	if env != nil {
		ses, err := secrets.FindAllByEnvironmentUuid(env.Uuid)
		if err != nil {
			return false, err
		}
		for _, secret := range ses {
			// if any Secret is pending, this Operation is not ready yet
			if secret.IsPending() {
				return false, nil
			}
		}
	}

	return true, err
}

// NewSetupScriptCtxt returns the setup script context necessary for running
// this operation.
func (self *Operation) NewSetupScriptCtxt(wsbi *WorkspaceBaseImage, project *Project) (*OperationSetupScriptCtxt, error) {
	ctxt := &OperationSetupScriptCtxt{
		Operation:      self,
		WsbiName:       wsbi.Name,
		WsbiRepository: wsbi.Repository,
		WsbiPath:       wsbi.Path,
		Parameters:     self.Parameters,
	}
	switch self.Category() {
	case "repository":
		ctxt.ShouldCloneRepos = false
	case "job":
		ctxt.ShouldCloneRepos = true
	case "notifier":
		ctxt.ShouldCloneRepos = false
	default:
		return nil, ErrUnknownOperationCategory
	}
	return ctxt, nil
}

// JobOperationVars holds the necessary information for the template
// producing the shell script to run a job operation.
type JobOperationVars struct {
	// Body contains the body of the script to execute.  It is
	// expected to be directly executable using `exec`.
	Body string
}

func (self *Operation) newJobOperationVars(store TaskStore) (*JobOperationVars, error) {
	if self.JobUuid == nil {
		panic("operation: invalid state")
	}
	task, err := store.FindByJobUuid(*self.JobUuid)
	if err != nil {
		return nil, err
	}
	return &JobOperationVars{
		Body: task.Body,
	}, nil
}

// RepositoryOperationVars hold the necessary information for the template
// production the shell script to run a repository operation.
type RepositoryOperationVars struct {
	// Url is the URL pointing to the repository to which this
	// operation applies.
	Url string
}

func (self *Operation) newRepositoryOperationVars(store RepositoryStore) (*RepositoryOperationVars, error) {
	if self.RepositoryUuid == nil {
		panic("operation: invalid state")
	}

	repository, err := store.FindByUuid(*self.RepositoryUuid)
	if err != nil {
		return nil, err
	}

	return &RepositoryOperationVars{
		Url: repository.Url,
	}, nil
}

func (self *Operation) NewScriptVars(tasks TaskStore, repositories RepositoryStore) (interface{}, error) {
	switch self.Category() {
	case "repository":
		return self.newRepositoryOperationVars(repositories)
	case "job":
		return self.newJobOperationVars(tasks)
	default:
		return nil, ErrUnknownOperationCategory
	}
}

func (vars *OperationSetupScriptCtxt) AddKey(belongsTo, name, private, public string) *OperationSetupScriptCtxt {
	if belongsTo != "repository" {
		belongsTo = "environment"
	}

	vars.Keys = append(vars.Keys, &keyPair{
		BelongsTo: belongsTo,
		Name:      name,
		Public:    public,
		Private:   private,
	})

	return vars
}

func (vars *OperationSetupScriptCtxt) AddSshConfig(repository *Repository, repostioryCredential *RepositoryCredential) *OperationSetupScriptCtxt {

	gitRepo, err := repository.Git()
	if err != nil {
		return vars
	}

	if !gitRepo.UsesSSH() {
		return vars
	}

	key := keyPair{
		Name: repostioryCredential.Name,
	}

	vars.SshConfigs = append(vars.SshConfigs, &sshConfig{
		Host:         gitRepo.SSHHost(),
		Port:         gitRepo.SSHPort(),
		SSHHostAlias: gitRepo.SSHHostAlias(),
		User:         gitRepo.Username(),
		RepoUuid:     repository.Uuid,
		KeyFileName:  key.Filename(),
	})

	return vars
}

type keyPair struct {
	BelongsTo, Name, Private, Public string
}

func (self *keyPair) BelongsToRepository() bool {
	return self.BelongsTo == "repository"
}

// Filename returns a safe file name for the key pair, derived from name.
// The result is guaranteed to only contain characters from the POSIX
// portable filename character set.
func (self *keyPair) Filename() string {
	// The hash of the filename is needed to disambiguate inputs,
	// because the character mapping is never one-to-one.
	h := sha1.New()
	h.Write([]byte(self.Name))
	sum := hex.EncodeToString(h.Sum(nil))

	base := filepath.Base(self.Name)
	base = strings.TrimSpace(base)
	base = strings.ToLower(base)
	base = strings.Map(posix.SafeStr, base)

	return fmt.Sprintf("%s-%s", base, sum[0:7])
}

func (self *Operation) AuthorizationName() string { return "operation" }

func (self *Operation) AddLogEvent(e *logevent.Message) {
	self.LogEvents = append(self.LogEvents, e)
}
