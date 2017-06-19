package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

type NotificationRuleHandler struct {
	store   *stores.DbNotificationRuleStore
	subject *domain.NotificationRule
}

func (h *NotificationRuleHandler) init(ctxt RequestContext) (*NotificationRuleHandler, error) {
	handler := &NotificationRuleHandler{}
	handler.store = stores.NewDbNotificationRuleStore(ctxt.Tx())

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

func MountNotificationRuleHandler(r *mux.Router, ctxt ServerContext) {
	h := &NotificationRuleHandler{}

	root := r.PathPrefix("/notification-rules").Subrouter()

	// Collection
	root.Methods("POST").Handler(HandlerFunc(ctxt, h.Create)).
		Name("notification-rules-create")
	root.Methods("PUT").Handler(HandlerFunc(ctxt, h.Update)).
		Name("notification-rules-update")

	// Item
	item := root.PathPrefix("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, h.Show)).
		Name("notification-rules-show")
	item.Methods("DELETE").Handler(HandlerFunc(ctxt, h.Archive)).
		Name("notification-rules-archive")

}

func (h *NotificationRuleHandler) Create(ctxt RequestContext) error {

	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	h, err := h.init(ctxt)
	if err != nil {
		return err
	}

	self := new(domain.NotificationRule)
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
	ctxt.EnqueueActivity(activities.NotificationRuleCreated(self), nil)
	ctxt.W().Header().Set("Location", urlForSubject(ctxt.R(), self))
	ctxt.W().WriteHeader(http.StatusCreated)
	writeAsJson(ctxt, self)

	return nil
}

func (h *NotificationRuleHandler) Update(ctxt RequestContext) error {

	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	h, err := h.init(ctxt)
	if err != nil {
		return err
	}

	newVersion := new(domain.NotificationRule)
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

	h.subject.NotifierUuid = newVersion.NotifierUuid
	h.subject.NotifierType = newVersion.NotifierType
	h.subject.MatchActivity = newVersion.MatchActivity
	h.subject.JobUuid = newVersion.JobUuid
	h.subject.CreatorUuid = ctxt.User().Uuid

	if err := h.subject.Validate(); err != nil {
		return err
	}

	if err := h.store.Update(h.subject); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.NotificationRuleEdited(h.subject), nil)
	writeAsJson(ctxt, h.subject)

	return nil
}

func (h *NotificationRuleHandler) Show(ctxt RequestContext) error {

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

func (h *NotificationRuleHandler) Archive(ctxt RequestContext) error {

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

	ctxt.EnqueueActivity(activities.NotificationRuleDeleted(h.subject), nil)

	ctxt.W().WriteHeader(http.StatusNoContent)

	return nil
}
