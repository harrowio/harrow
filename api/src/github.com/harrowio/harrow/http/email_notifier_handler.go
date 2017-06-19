package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

type EmailNotifierHandler struct {
	store   *stores.DbEmailNotifierStore
	subject *domain.EmailNotifier
}

func (h *EmailNotifierHandler) init(ctxt RequestContext) (*EmailNotifierHandler, error) {
	handler := &EmailNotifierHandler{}
	handler.store = stores.NewDbEmailNotifierStore(ctxt.Tx())

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

func MountEmailNotifierHandler(r *mux.Router, ctxt ServerContext) {
	h := &EmailNotifierHandler{}

	root := r.PathPrefix("/email-notifiers").Subrouter()

	// Collection
	root.Methods("POST").Handler(HandlerFunc(ctxt, h.Create)).
		Name("email-notifiers-create")
	root.Methods("PUT").Handler(HandlerFunc(ctxt, h.Update)).
		Name("email-notifiers-update")

	// Item
	item := root.PathPrefix("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, h.Show)).
		Name("email-notifiers-show")
	item.Methods("DELETE").Handler(HandlerFunc(ctxt, h.Archive)).
		Name("email-notifiers-archive")

}

func (h *EmailNotifierHandler) Create(ctxt RequestContext) error {

	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	h, err := h.init(ctxt)
	if err != nil {
		return err
	}

	self := new(domain.EmailNotifier)
	if err := json.NewDecoder(ctxt.R().Body).Decode(&halWrapper{Subject: self}); err != nil {
		return err
	}

	self.UrlHost = ctxt.User().UrlHost

	if allowed, err := ctxt.Auth().CanCreate(self); !allowed {
		return err
	}

	if err := self.Validate(); err != nil {
		return err
	}

	if _, err := h.store.Create(self); err != nil {
		return err
	}
	ctxt.EnqueueActivity(activities.EmailNotifierCreated(self), nil)
	ctxt.W().Header().Set("Location", urlForSubject(ctxt.R(), self))
	ctxt.W().WriteHeader(http.StatusCreated)
	writeAsJson(ctxt, self)

	return nil
}

func (h *EmailNotifierHandler) Update(ctxt RequestContext) error {

	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	h, err := h.init(ctxt)
	if err != nil {
		return err
	}

	newVersion := new(domain.EmailNotifier)
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

	h.subject.Recipient = newVersion.Recipient
	h.subject.ProjectUuid = newVersion.ProjectUuid

	if err := h.subject.Validate(); err != nil {
		return err
	}

	if err := h.store.Update(h.subject); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.EmailNotifierEdited(h.subject), nil)
	writeAsJson(ctxt, h.subject)

	return nil
}

func (h *EmailNotifierHandler) Show(ctxt RequestContext) error {

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

func (h *EmailNotifierHandler) Archive(ctxt RequestContext) error {

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

	ctxt.EnqueueActivity(activities.EmailNotifierDeleted(h.subject), nil)

	ctxt.W().WriteHeader(http.StatusNoContent)

	return nil
}
