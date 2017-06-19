package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/limits"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
)

type envParamsWrapper struct {
	Subject envParams
}

type envParams struct {
	Uuid        string                      `json:"uuid"`
	Name        string                      `json:"name"`
	ProjectUuid string                      `json:"projectUuid"`
	Variables   domain.EnvironmentVariables `json:"variables"`
}

func copyEnvParams(p *envParams, m *domain.Environment) {
	m.Uuid = p.Uuid
	m.Name = p.Name
	m.ProjectUuid = p.ProjectUuid
	m.Variables = p.Variables
}

func ReadEnvParams(r io.Reader) (*envParams, error) {
	decoder := json.NewDecoder(r)
	var w envParamsWrapper
	err := decoder.Decode(&w)
	if err != nil {
		return nil, err
	}
	return &w.Subject, nil
}

func MountEnvironmentHandler(r *mux.Router, ctxt ServerContext) {

	eh := &envHandler{}

	// Collection
	root := r.PathPrefix("/environments").Subrouter()
	root.Methods("POST").Handler(HandlerFunc(ctxt, eh.CreateUpdate)).
		Name("environment-create")
	root.Methods("PUT").Handler(HandlerFunc(ctxt, eh.CreateUpdate)).
		Name("environment-update")

	// Relationships
	related := root.PathPrefix("/{uuid}/").Subrouter()
	related.Methods("GET").Path("/targets").Handler(HandlerFunc(ctxt, eh.Targets)).
		Name("environment-targets")
	related.Methods("GET").Path("/jobs").Handler(HandlerFunc(ctxt, eh.Jobs)).
		Name("environment-jobs")
	related.Methods("GET").Path("/secrets").Handler(HandlerFunc(ctxt, eh.Secrets)).
		Name("environment-secrets")

	// Item
	item := root.PathPrefix("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, eh.Show)).
		Name("environment-show")
	item.Methods("DELETE").Handler(HandlerFunc(ctxt, eh.Archive)).
		Name("environment-archive")

}

type envHandler struct{}

func (self envHandler) Show(ctxt RequestContext) (err error) {

	store := stores.NewDbEnvironmentStore(ctxt.Tx())

	env, err := store.FindByUuid(ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(env); !allowed {
		return err
	}

	writeAsJson(ctxt, env)

	return err
}

func (self envHandler) CreateUpdate(ctxt RequestContext) (err error) {

	params, err := ReadEnvParams(ctxt.R().Body)
	if err != nil {
		return err
	}

	store := stores.NewDbEnvironmentStore(ctxt.Tx())

	isNew := true
	var env *domain.Environment

	if uuidhelper.IsValid(params.Uuid) {
		env, err = store.FindByUuid(params.Uuid)
		_, notFound := err.(*domain.NotFoundError)
		if err != nil && !notFound {
			return err
		}
		isNew = (env == nil)
	}

	if isNew {
		env = &domain.Environment{}
	}

	copyEnvParams(params, env)

	var uuid string

	if isNew {

		if allowed, err := ctxt.Auth().CanCreate(env); !allowed {
			return err
		}

		if allowed, err := limits.CanCreate(env); !allowed {
			return err
		}

		uuid, err = store.Create(env)
		ctxt.EnqueueActivity(activities.EnvironmentAdded(env), nil)
	} else {
		if allowed, err := ctxt.Auth().CanUpdate(env); !allowed {
			return err
		}
		err = store.Update(env)
		uuid = env.Uuid
		ctxt.EnqueueActivity(activities.EnvironmentEdited(env), nil)
	}

	if err != nil {
		return err
	}

	env, err = store.FindByUuid(uuid)
	if err != nil {
		return err
	}

	if isNew {
		ctxt.W().Header().Add("Location", urlForSubject(ctxt.R(), env))
		ctxt.W().WriteHeader(http.StatusCreated)
	}

	writeAsJson(ctxt, env)

	return err
}

func (self envHandler) Archive(ctxt RequestContext) (err error) {

	envUuid := ctxt.PathParameter("uuid")

	store := stores.NewDbEnvironmentStore(ctxt.Tx())

	env, err := store.FindByUuid(envUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanArchive(env); !allowed {
		return err
	}

	err = store.ArchiveByUuid(envUuid)
	if err != nil {
		return err
	}

	ctxt.W().WriteHeader(http.StatusNoContent)

	return err
}

func (self envHandler) Targets(ctxt RequestContext) (err error) {

	envUuid := ctxt.PathParameter("uuid")

	store := stores.NewDbEnvironmentStore(ctxt.Tx())
	targetStore := stores.NewDbTargetStore(ctxt.Tx())

	env, err := store.FindByUuid(envUuid)
	if err != nil {
		return err
	}

	if allowed, _ := ctxt.Auth().CanRead(env); !allowed {
		return err
	}

	targets, err := targetStore.FindAllByEnvironmentUuid(envUuid)
	if err != nil {
		return err
	}

	var interfaceTargets []interface{} = make([]interface{}, 0, len(targets))
	for _, r := range targets {
		if allowed, _ := ctxt.Auth().CanRead(r); allowed {
			interfaceTargets = append(interfaceTargets, r)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceTargets),
		Count:      len(interfaceTargets),
		Collection: interfaceTargets,
	})

	return err
}

func (self envHandler) Jobs(ctxt RequestContext) (err error) {

	envUuid := ctxt.PathParameter("uuid")

	store := stores.NewDbEnvironmentStore(ctxt.Tx())
	jobStore := stores.NewDbJobStore(ctxt.Tx())

	env, err := store.FindByUuid(envUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(env); !allowed {
		return err
	}

	jobs, err := jobStore.FindAllByEnvironmentUuid(envUuid)
	if err != nil {
		return err
	}

	var interfaceJobs []interface{} = make([]interface{}, 0, len(jobs))
	for _, j := range jobs {
		if allowed, _ := ctxt.Auth().CanRead(j); allowed {
			interfaceJobs = append(interfaceJobs, j)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceJobs),
		Count:      len(interfaceJobs),
		Collection: interfaceJobs,
	})

	return err

}

func (self envHandler) Secrets(ctxt RequestContext) (err error) {

	envUuid := ctxt.PathParameter("uuid")

	store := stores.NewDbEnvironmentStore(ctxt.Tx())
	secretStore := stores.NewSecretStore(ctxt.SecretKeyValueStore(), ctxt.Tx())

	env, err := store.FindByUuid(envUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(env); !allowed {
		return err
	}

	secrets, err := secretStore.FindAllByEnvironmentUuid(envUuid)
	if err != nil {
		return err
	}

	var interfaceSecrets []interface{} = make([]interface{}, 0)
	for _, secret := range secrets {

		allowed, _ := ctxt.Auth().CanRead(secret)
		if allowed {
			if secret.IsPending() {
				interfaceSecrets = append(interfaceSecrets, secret)
			} else {
				if secret.IsSsh() {
					sshSecret, err := domain.AsSshSecret(secret)
					if err != nil {
						return err
					}
					unpriv, err := sshSecret.AsUnprivileged()
					if err != nil {
						return err
					}
					interfaceSecrets = append(interfaceSecrets, unpriv)
				} else {
					envSecret, err := domain.AsEnvironmentSecret(secret)
					if err != nil {
						return err
					}
					if privileged, _ := ctxt.Auth().Can(domain.CapabilityReadPrivileged, secret); privileged {
						priv, err := envSecret.AsPrivileged()
						if err != nil {
							return err
						}
						interfaceSecrets = append(interfaceSecrets, priv)
					} else {
						unpriv, err := envSecret.AsUnprivileged()
						if err != nil {
							return err
						}
						interfaceSecrets = append(interfaceSecrets, unpriv)
					}
				}
			}
		}

	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceSecrets),
		Count:      len(interfaceSecrets),
		Collection: interfaceSecrets,
	})

	return err
}
