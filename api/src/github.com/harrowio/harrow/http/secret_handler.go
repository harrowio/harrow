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
)

type secretParamsWrapper struct {
	Subject secretParams
}

type secretParams struct {
	Uuid            string              `json:"uuid"`
	Name            string              `json:"name"`
	EnvironmentUuid string              `json:"environmentUuid"`
	Type            domain.SecretType   `json:"type"`
	Status          domain.SecretStatus `json:"status"`
	Value           string              `json:"value"`
}

func copySecretParams(p *secretParams, m *domain.Secret) {
	m.Uuid = p.Uuid
	m.Name = p.Name
	m.EnvironmentUuid = p.EnvironmentUuid
	m.Type = p.Type
	m.Status = p.Status
}

func copyEnvSecretParams(p *secretParams, m *domain.EnvironmentSecret) {
	m.Uuid = p.Uuid
	m.Name = p.Name
	m.EnvironmentUuid = p.EnvironmentUuid
	m.Type = p.Type
	m.Status = p.Status
	m.Value = p.Value
}

func ReadSecretParams(r io.Reader) (*secretParams, error) {
	decoder := json.NewDecoder(r)
	var w secretParamsWrapper
	err := decoder.Decode(&w)
	if err != nil {
		return nil, err
	}
	return &w.Subject, nil
}

func MountSecretHandler(r *mux.Router, ctxt ServerContext) {

	h := &secretHandler{}

	// Collection
	root := r.PathPrefix("/secrets").Subrouter()
	root.Methods("POST").Handler(HandlerFunc(ctxt, h.Create)).
		Name("secret-create")

	// Item
	item := root.PathPrefix("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, h.Show)).
		Name("secret-show")
	item.Methods("DELETE").Handler(HandlerFunc(ctxt, h.Archive)).
		Name("secret-archive")
}

type secretHandler struct{}

func (self secretHandler) Create(ctxt RequestContext) (err error) {

	params, err := ReadSecretParams(ctxt.R().Body)
	if err != nil {
		return err
	}

	if params.Type == domain.SecretEnv {
		return self.CreateEnv(ctxt, params)
	}
	return self.CreateSsh(ctxt, params)
}

func (self secretHandler) CreateSsh(ctxt RequestContext, params *secretParams) (err error) {

	store := stores.NewSecretStore(ctxt.SecretKeyValueStore(), ctxt.Tx())

	secret := &domain.Secret{}

	copySecretParams(params, secret)

	if allowed, err := ctxt.Auth().CanCreate(secret); !allowed {
		return err
	}

	if allowed, err := limits.CanCreate(secret); !allowed {
		return err
	}

	uuid, err := store.Create(secret)

	if err != nil {
		return err
	}

	secret, err = store.FindByUuid(uuid)
	if err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.SecretAdded(secret), nil)

	ctxt.W().Header().Add("Location", urlForSubject(ctxt.R(), secret))
	ctxt.W().WriteHeader(http.StatusCreated)

	writeAsJson(ctxt, secret)

	return nil
}

func (self secretHandler) CreateEnv(ctxt RequestContext, params *secretParams) (err error) {

	store := stores.NewSecretStore(ctxt.SecretKeyValueStore(), ctxt.Tx())

	envSecret := &domain.EnvironmentSecret{Secret: &domain.Secret{}}

	copyEnvSecretParams(params, envSecret)
	secret, err := envSecret.AsSecret()
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanCreate(secret); !allowed {
		return err
	}

	if allowed, err := limits.CanCreate(secret); !allowed {
		return err
	}

	uuid, err := store.Create(secret)

	if err != nil {
		return err
	}

	secret, err = store.FindByUuid(uuid)
	if err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.SecretAdded(secret), nil)

	ctxt.W().Header().Add("Location", urlForSubject(ctxt.R(), secret))
	ctxt.W().WriteHeader(http.StatusCreated)

	writeAsJson(ctxt, secret)

	return nil
}

func (self secretHandler) Show(ctxt RequestContext) error {

	store := stores.NewSecretStore(ctxt.SecretKeyValueStore(), ctxt.Tx())

	secret, err := store.FindByUuid(ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(secret); !allowed {
		return err
	}
	if secret.IsSsh() {
		if secret.IsPending() {
			writeAsJson(ctxt, secret)
			return nil
		}
		sshSecret, err := domain.AsSshSecret(secret)
		if err != nil {
			return err
		}
		// check if we can read the privileged variant of the secret
		// ignore errors
		if privileged, _ := ctxt.Auth().Can(domain.CapabilityReadPrivileged, secret); privileged {
			priv, _ := sshSecret.AsPrivileged()
			writeAsJson(ctxt, priv)
			return nil
		}
		// no, render the unprivileged variant
		unpriv, err := sshSecret.AsUnprivileged()
		if err != nil {
			return err
		}
		writeAsJson(ctxt, unpriv)
		return nil
	} else {
		envSecret, err := domain.AsEnvironmentSecret(secret)
		if err != nil {
			return err
		}
		// check if we can read the privileged variant of the secret
		// ignore errors
		if privileged, _ := ctxt.Auth().Can(domain.CapabilityReadPrivileged, secret); privileged {
			priv, _ := envSecret.AsPrivileged()
			writeAsJson(ctxt, priv)
			return nil
		}
		// no, render the unprivileged variant
		unpriv, err := envSecret.AsUnprivileged()
		if err != nil {
			return err
		}
		writeAsJson(ctxt, unpriv)
		return nil
	}
}

func (self secretHandler) Archive(ctxt RequestContext) (err error) {

	uuid := ctxt.PathParameter("uuid")

	store := stores.NewSecretStore(ctxt.SecretKeyValueStore(), ctxt.Tx())

	secret, err := store.FindByUuid(uuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanArchive(secret); !allowed {
		return err
	}

	err = store.ArchiveByUuid(uuid)
	if err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.SecretDeleted(secret), nil)
	ctxt.W().WriteHeader(http.StatusNoContent)

	return err
}
