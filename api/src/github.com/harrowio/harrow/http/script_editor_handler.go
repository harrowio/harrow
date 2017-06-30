package http

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/diff"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/domain/stencil"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
)

type scriptEditParams struct {
	ProjectUuid string                    `json:"projectUuid"`
	Task        *domain.Task              `json:"task"`
	Environment *domain.Environment       `json:"environment"`
	Secrets     []*domain.OperationSecret `json:"secrets"`

	jobUuid string
}

type SecretChanges struct {
	Added   []domain.OperationSecret `json:"added"`
	Removed []domain.OperationSecret `json:"removed"`
	Changed []domain.OperationSecret `json:"changed"`
}

func NewScriptEditParamsFromJSON(in io.Reader) (*scriptEditParams, error) {
	result := &scriptEditParams{}
	err := json.NewDecoder(in).Decode(result)

	return result, err
}

func (self *scriptEditParams) SecretByName(name string) *domain.OperationSecret {
	for _, secret := range self.Secrets {
		if secret.Name == name {
			return secret
		}
	}

	return nil
}

func (self *scriptEditParams) SecretChanges(existingSecrets []*domain.Secret) *SecretChanges {
	changes := &SecretChanges{
		Added:   []domain.OperationSecret{},
		Changed: []domain.OperationSecret{},
		Removed: []domain.OperationSecret{},
	}

	findSecret := func(name string) *domain.Secret {
		for _, secret := range existingSecrets {
			if secret.Name == name {
				return secret
			}
		}

		return nil
	}

	for _, existingSecret := range existingSecrets {
		envSecret, err := domain.AsEnvironmentSecret(existingSecret)
		if err != nil {
			continue
		}

		if newSecret := self.SecretByName(existingSecret.Name); newSecret == nil {
			changes.Removed = append(changes.Removed, domain.OperationSecret{
				Name:  existingSecret.Name,
				Value: string(existingSecret.SecretBytes),
			})
		} else if envSecret.Value != newSecret.Value {
			changes.Changed = append(changes.Changed, domain.OperationSecret{
				Name:     existingSecret.Name,
				Value:    newSecret.Value,
				OldValue: &envSecret.Value,
			})
		}
	}

	for _, newSecret := range self.Secrets {
		if existingSecret := findSecret(newSecret.Name); existingSecret == nil {
			changes.Added = append(changes.Added, domain.OperationSecret{
				Name:  newSecret.Name,
				Value: newSecret.Value,
			})
		}
	}

	return changes
}

func (self *scriptEditParams) Validate() error {
	errors := domain.EmptyValidationError()
	if !uuidhelper.IsValid(self.ProjectUuid) {
		errors.Add("projectUuid", "malformed")
	}

	if self.Task != nil && !uuidhelper.IsValid(self.Task.Uuid) {
		errors.Add("task.uuid", "malformed")
	}

	if self.Environment != nil && !uuidhelper.IsValid(self.Environment.Uuid) {
		errors.Add("environment.uuid", "malformed")
	}

	return errors.ToError()
}

type scriptEditorHandler struct {
	tasks        *stores.DbTaskStore
	environments *stores.DbEnvironmentStore
	jobs         *stores.DbJobStore
	secrets      *stores.SecretStore
	projects     *stores.DbProjectStore
}

func (h *scriptEditorHandler) init(ctxt RequestContext) (*scriptEditorHandler, error) {
	handler := &scriptEditorHandler{}
	handler.tasks = stores.NewDbTaskStore(ctxt.Tx())
	handler.tasks.SetLogger(ctxt.Log())
	handler.environments = stores.NewDbEnvironmentStore(ctxt.Tx())
	handler.environments.SetLogger(ctxt.Log())
	handler.secrets = stores.NewSecretStore(ctxt.SecretKeyValueStore(), ctxt.Tx())
	handler.secrets.SetLogger(ctxt.Log())
	handler.projects = stores.NewDbProjectStore(ctxt.Tx())
	handler.projects.SetLogger(ctxt.Log())
	return handler, nil
}

func MountScriptEditorHandler(r *mux.Router, ctxt ServerContext) {
	h := &scriptEditorHandler{}

	root := r.PathPrefix("/script-editor").Subrouter()
	root.PathPrefix("/diff").Methods("POST").Handler(HandlerFunc(ctxt, h.Diff)).
		Name("script-editor-diff")
	root.PathPrefix("/apply").Methods("POST").Handler(HandlerFunc(ctxt, h.Apply)).
		Name("script-editor-apply")
	root.PathPrefix("/save").Methods("POST").Handler(HandlerFunc(ctxt, h.Save)).
		Name("script-editor-save")

}

func (self *scriptEditorHandler) Apply(ctxt RequestContext) error {

	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	h, err := self.init(ctxt)
	if err != nil {
		return err
	}

	params, err := NewScriptEditParamsFromJSON(ctxt.R().Body)
	if err != nil {
		return err
	}

	if err := params.Validate(); err != nil {
		return err
	}

	if err := h.ensureDefaults(params, ctxt); err != nil {
		return err
	}

	operationParameters := domain.NewOperationParameters()
	if user := ctxt.User(); user != nil {
		operationParameters.Username = ctxt.User().Name
		operationParameters.UserUuid = ctxt.User().Uuid
	}
	for _, secret := range params.Secrets {
		operationParameters.AddSecret(secret.Name, secret.Value)
	}
	operationParameters.Task = params.Task
	operationParameters.Environment = params.Environment

	now := "now"
	schedule := &domain.Schedule{
		Uuid:        uuidhelper.MustNewV4(),
		JobUuid:     params.jobUuid,
		UserUuid:    ctxt.User().Uuid,
		Cronspec:    nil,
		Timespec:    &now,
		Description: "script-editor",
		Parameters:  operationParameters,
	}

	if err := schedule.Validate(); err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanCreate(schedule); !allowed {
		return err
	}

	theLimits, err := NewLimitsFromContext(ctxt)
	if err != nil {
		return err
	}

	if exceeded, err := theLimits.Exceeded(schedule); exceeded {
		return ErrLimitsExceeded
	} else if err != nil {
		ctxt.Log().Warn().Msgf("error calculating limits: %s", err)
	}

	scheduleStore := stores.NewDbScheduleStore(ctxt.Tx())
	scheduleUuid, err := scheduleStore.Create(schedule)
	if err != nil {
		return err
	}

	schedule, err = scheduleStore.FindByUuid(scheduleUuid)
	if err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.JobScheduled(schedule, "user.script-editor"), nil)
	ctxt.EnqueueActivity(activities.ScriptEditorTested(&activities.ScriptEditorPayload{
		TaskUuid:        params.Task.Uuid,
		EnvironmentUuid: params.Environment.Uuid,
		ProjectUuid:     params.ProjectUuid,
	}), nil)

	writeAsJson(ctxt, schedule)

	return nil
}

func (self *scriptEditorHandler) Save(ctxt RequestContext) error {
	h, err := self.init(ctxt)
	errors := domain.EmptyValidationError()
	if err != nil {
		return err
	}

	params, err := NewScriptEditParamsFromJSON(ctxt.R().Body)
	if err != nil {
		return err
	}

	if err := params.Validate(); err != nil {
		return err
	}

	project, err := h.projects.FindByUuid(params.ProjectUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().Can("save-scripts", project); !allowed {
		return err
	}

	if params.Task != nil {
		task, err := h.tasks.FindByUuid(params.Task.Uuid)
		if err != nil && domain.IsNotFound(err) {
			if _, err := h.tasks.Create(params.Task); err != nil {
				errors.Add("task", err.Error())
			}
			activity := activities.TaskAdded(params.Task)
			activity.Extra["script-editor"] = true
			ctxt.EnqueueActivity(activity, nil)
		} else {
			task.Body = params.Task.Body
			task.Name = params.Task.Name
			if err := h.tasks.Update(task); err != nil {
				errors.Add("task", err.Error())
			} else {
				activity := activities.TaskEdited(task)
				activity.Extra["script-editor"] = true
				ctxt.EnqueueActivity(activity, nil)
			}
		}
	}

	if params.Environment != nil {
		environment, err := h.environments.FindByUuid(params.Environment.Uuid)
		if err != nil && domain.IsNotFound(err) {
			if envUuid, err := h.environments.Create(params.Environment); err != nil {
				errors.Add("environment", err.Error())
			} else {
				params.Environment.Uuid = envUuid
				activity := activities.EnvironmentAdded(params.Environment)
				activity.Extra["script-editor"] = true
				ctxt.EnqueueActivity(activity, nil)
			}
		} else {
			environment.Variables = params.Environment.Variables
			if err := h.environments.Update(environment); err != nil {
				errors.Add("environment", err.Error())
			} else {
				activity := activities.EnvironmentAdded(environment)
				activity.Extra["script-editor"] = true
				ctxt.EnqueueActivity(activity, nil)
			}
		}

		if environment != nil {
			existingSecrets, err := h.secrets.FindAllByEnvironmentUuid(params.Environment.Uuid)
			if err != nil {
				errors.Add("secrets", err.Error())
			} else {
				h.saveSecrets(params, existingSecrets, environment, errors, ctxt)
			}

		}

	}

	ctxt.EnqueueActivity(activities.ScriptEditorSaved(&activities.ScriptEditorPayload{
		TaskUuid:        params.Task.Uuid,
		EnvironmentUuid: params.Environment.Uuid,
		ProjectUuid:     params.ProjectUuid,
	}), nil)

	return errors.ToError()
}

func (self *scriptEditorHandler) saveSecrets(params *scriptEditParams, existingSecrets []*domain.Secret, environment *domain.Environment, errors *domain.ValidationError, ctxt interface {
	EnqueueActivity(*domain.Activity, *string)
}) {
	for _, existingSecret := range existingSecrets {
		found := params.SecretByName(existingSecret.Name)
		if found == nil && !existingSecret.IsSsh() {
			if err := self.secrets.ArchiveByUuid(existingSecret.Uuid); err != nil {
				errors.Add(fmt.Sprintf("secrets[%s]", existingSecret.Name), err.Error())
			} else {
				activity := activities.SecretDeleted(existingSecret)
				activity.Extra["script-editor"] = true
				ctxt.EnqueueActivity(activity, nil)

			}
		} else if existingSecret.IsEnv() {
			envSecret, err := domain.AsEnvironmentSecret(existingSecret)
			if err != nil {
				errors.Add(fmt.Sprintf("secrets[%s]", existingSecret.Name), err.Error())
				continue
			}
			envSecret.Value = found.Value
			toSave, err := envSecret.AsSecret()
			if err != nil {
				errors.Add(fmt.Sprintf("secrets[%s]", existingSecret.Name), err.Error())
				continue
			}
			if err := self.secrets.Update(toSave); err != nil {
				errors.Add(fmt.Sprintf("secrets[%s]", existingSecret.Name), err.Error())
			} else {
				activity := activities.SecretEdited(toSave)
				activity.Extra["script-editor"] = true
				ctxt.EnqueueActivity(activity, nil)
			}
		}
	}

	findExistingSecret := func(name string) *domain.Secret {
		for _, existingSecret := range existingSecrets {
			if existingSecret.Name == name {
				return existingSecret
			}
		}

		return nil
	}

	for _, secret := range params.Secrets {
		if findExistingSecret(secret.Name) != nil {
			continue
		}

		envSecret := &domain.EnvironmentSecret{
			Secret: &domain.Secret{
				Name:            secret.Name,
				Type:            domain.SecretEnv,
				EnvironmentUuid: environment.Uuid,
				Status:          domain.SecretPresent,
			},
			Value: secret.Value,
		}

		toSave, err := envSecret.AsSecret()
		if err != nil {
			errors.Add(fmt.Sprintf("secrets[%s]", secret.Name), err.Error())
			continue
		}

		if _, err := self.secrets.Create(toSave); err != nil {
			errors.Add(fmt.Sprintf("secrets[%s]", secret.Name), err.Error())
		} else {
			activity := activities.SecretAdded(toSave)
			activity.Extra["script-editor"] = true
			ctxt.EnqueueActivity(activity, nil)
		}
	}
}

func (self *scriptEditorHandler) Diff(ctxt RequestContext) error {
	h, err := self.init(ctxt)
	if err != nil {
		return err
	}

	result := struct {
		Secrets      *SecretChanges                     `json:"secrets"`
		Environment  *domain.EnvironmentVariableChanges `json:"environment"`
		TaskBodyDiff []*diff.Change                     `json:"taskBodyDiff"`
	}{}

	params, err := NewScriptEditParamsFromJSON(ctxt.R().Body)
	if err != nil {
		return err
	}

	if err := params.Validate(); err != nil {
		return err
	}

	project, err := h.projects.FindByUuid(params.ProjectUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().Can("diff-scripts", project); !allowed {
		return err
	}

	if params.Secrets != nil {
		existingSecrets := []*domain.Secret{}
		if params.Environment != nil {
			if !uuidhelper.IsValid(params.Environment.Uuid) {
				return domain.NewValidationError("environment.uuid", "malformed")
			}

			secrets, err := h.secrets.FindAllByEnvironmentUuid(params.Environment.Uuid)
			if err != nil {
				return err
			}
			existingSecrets = secrets
		}

		result.Secrets = params.SecretChanges(existingSecrets)
	}

	if params.Environment != nil {
		if !uuidhelper.IsValid(params.Environment.Uuid) {
			return domain.NewValidationError("environment.uuid", "malformed")
		}

		originalEnvironment, err := h.environments.FindByUuid(params.Environment.Uuid)
		if err != nil {
			return err
		}

		changes := originalEnvironment.Variables.Diff(params.Environment.Variables)
		result.Environment = changes
	}

	if params.Task != nil {
		if !uuidhelper.IsValid(params.Task.Uuid) {
			return domain.NewValidationError("task.uuid", "malformed")
		}

		originalTask, err := h.tasks.FindByUuid(params.Task.Uuid)
		if err != nil {
			return err
		}

		bodyDiff, err := diff.Changes([]byte(originalTask.Body), []byte(params.Task.Body))
		if err != nil {
			return NewInternalError(err)
		}
		result.TaskBodyDiff = bodyDiff
	}

	return json.NewEncoder(ctxt.W()).Encode(result)
}

func (self *scriptEditorHandler) ensureDefaults(params *scriptEditParams, ctxt RequestContext) error {
	stencilStore := stores.NewDbStencilStore(ctxt.SecretKeyValueStore(), ctxt.Tx(), ctxt)
	configuration := stencilStore.ToConfiguration()
	configuration.ProjectUuid = params.ProjectUuid
	defaults := stencil.NewScriptEditorDefaults(configuration)

	if err := defaults.Create(); err != nil {
		return err
	}
	params.jobUuid = defaults.Job().Uuid

	return nil
}
