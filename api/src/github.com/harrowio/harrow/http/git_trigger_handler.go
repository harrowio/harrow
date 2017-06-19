package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

type GitTriggerHandler struct {
	store   *stores.DbGitTriggerStore
	subject *domain.GitTrigger
}

func (h *GitTriggerHandler) init(ctxt RequestContext) (*GitTriggerHandler, error) {
	handler := &GitTriggerHandler{}
	handler.store = stores.NewDbGitTriggerStore(ctxt.Tx())

	uuid := ctxt.PathParameter("uuid")
	if uuid != "" {
		subject, err := handler.store.FindByUuid(uuid)
		if err != nil {
			return nil, err
		}
		handler.subject = subject
	}

	return handler, nil
}

func MountGitTriggerHandler(r *mux.Router, ctxt ServerContext) {
	h := &GitTriggerHandler{}

	root := r.PathPrefix("/git-triggers").Subrouter()

	// Collection
	root.Methods("POST").Handler(HandlerFunc(ctxt, h.Create)).
		Name("git-triggers-create")
	root.Methods("PUT").Handler(HandlerFunc(ctxt, h.Update)).
		Name("git-triggers-update")

	// Item
	item := root.PathPrefix("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, h.Show)).
		Name("git-triggers-show")
	item.Methods("DELETE").Handler(HandlerFunc(ctxt, h.Archive)).
		Name("git-triggers-archive")

}

func (h *GitTriggerHandler) Create(ctxt RequestContext) error {

	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	h, err := h.init(ctxt)
	if err != nil {
		return err
	}

	self := new(domain.GitTrigger)
	if err := json.NewDecoder(ctxt.R().Body).Decode(&halWrapper{Subject: self}); err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanCreate(self); !allowed {
		return err
	}

	self.CreatorUuid = ctxt.User().Uuid

	if err := self.Validate(); err != nil {
		return err
	}

	if _, err := h.store.Create(self); err != nil {
		return err
	}
	ctxt.EnqueueActivity(activities.GitTriggerCreated(self), nil)
	ctxt.W().Header().Set("Location", urlForSubject(ctxt.R(), self))
	ctxt.W().WriteHeader(http.StatusCreated)
	writeAsJson(ctxt, self)

	return nil
}

func (h *GitTriggerHandler) Update(ctxt RequestContext) error {

	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	h, err := h.init(ctxt)
	if err != nil {
		return err
	}

	newVersion := new(domain.GitTrigger)
	if err := json.NewDecoder(ctxt.R().Body).Decode(&halWrapper{Subject: newVersion}); err != nil {
		return err
	}

	h.subject, err = h.store.FindByUuid(newVersion.Uuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanUpdate(h.subject); !allowed {
		return err
	}

	h.subject.Name = newVersion.Name
	h.subject.JobUuid = newVersion.JobUuid
	h.subject.RepositoryUuid = newVersion.RepositoryUuid
	h.subject.ChangeType = newVersion.ChangeType
	h.subject.MatchRef = newVersion.MatchRef

	if err := h.subject.Validate(); err != nil {
		return err
	}

	if err := h.store.Update(h.subject); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.GitTriggerEdited(h.subject), nil)
	writeAsJson(ctxt, h.subject)

	return nil
}

func (h *GitTriggerHandler) Show(ctxt RequestContext) error {

	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	h, err := h.init(ctxt)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(h.subject); !allowed {
		return err
	}

	writeAsJson(ctxt, h.subject)

	return nil
}

func (h *GitTriggerHandler) Archive(ctxt RequestContext) error {

	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	h, err := h.init(ctxt)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanArchive(h.subject); !allowed {
		return err
	}

	if err := h.store.ArchiveByUuid(h.subject.Uuid); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.GitTriggerDeleted(h.subject), nil)

	ctxt.W().WriteHeader(http.StatusNoContent)

	return nil
}
