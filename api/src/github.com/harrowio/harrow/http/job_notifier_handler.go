package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
)

type JobNotifierHandler struct {
	store             *stores.DbJobNotifierStore
	projects          *stores.DbProjectStore
	webhooks          *stores.DbWebhookStore
	notificationRules *stores.DbNotificationRuleStore
	jobs              *stores.DbJobStore
	subject           *domain.JobNotifier
}

func (h *JobNotifierHandler) init(ctxt RequestContext) (*JobNotifierHandler, error) {
	handler := &JobNotifierHandler{}
	handler.store = stores.NewDbJobNotifierStore(ctxt.Tx())
	handler.projects = stores.NewDbProjectStore(ctxt.Tx())
	handler.webhooks = stores.NewDbWebhookStore(ctxt.Tx())
	handler.notificationRules = stores.NewDbNotificationRuleStore(ctxt.Tx())
	handler.jobs = stores.NewDbJobStore(ctxt.Tx())
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

func MountJobNotifierHandler(r *mux.Router, ctxt ServerContext) {
	h := &JobNotifierHandler{}

	root := r.PathPrefix("/job-notifiers").Subrouter()

	// Collection
	root.Methods("POST").Handler(HandlerFunc(ctxt, h.Create)).
		Name("job-notifiers-create")

	// Item
	item := root.PathPrefix("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, h.Show)).
		Name("job-notifiers-show")
	item.Methods("DELETE").Handler(HandlerFunc(ctxt, h.Archive)).
		Name("job-notifiers-archive")

}

func (h *JobNotifierHandler) Create(ctxt RequestContext) error {

	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	h, err := h.init(ctxt)
	if err != nil {
		return err
	}

	self := new(domain.JobNotifier)
	if err := json.NewDecoder(ctxt.R().Body).Decode(&halWrapper{Subject: self}); err != nil {
		return err
	}

	if self.Uuid == "" {
		self.Uuid = uuidhelper.MustNewV4()
	}

	if _, err := h.jobs.FindByUuid(self.JobUuid); err != nil {
		return domain.NewValidationError("jobUuid", "not_found")
	}

	project, err := h.projects.FindByUuid(self.ProjectUuid)
	if err != nil {
		return domain.NewValidationError("projectUuid", "not_found")
	}

	webhook := domain.NewWebhook(
		project.Uuid,
		ctxt.User().Uuid,
		self.JobUuid,
		fmt.Sprintf("urn:harrow:job-notifier:%s", self.Uuid),
	)

	if allowed, err := ctxt.Auth().CanCreate(self); !allowed {
		return err
	}

	if _, err := h.webhooks.Create(webhook); err != nil {
		return err
	}

	self.WebhookURL = webhook.Links(map[string]map[string]string{}, "https", ctxt.User().UrlHost+"/api")["deliver"]["href"]

	if err := self.Validate(); err != nil {
		return err
	}

	if _, err := h.store.Create(self); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.JobNotifierCreated(self), nil)
	ctxt.W().Header().Set("Location", urlForSubject(ctxt.R(), self))
	ctxt.W().WriteHeader(http.StatusCreated)
	writeAsJson(ctxt, self)

	return nil
}

func (h *JobNotifierHandler) Show(ctxt RequestContext) error {

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

func (h *JobNotifierHandler) Archive(ctxt RequestContext) error {

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

	webhook, err := h.webhooks.FindBySlug(h.subject.WebhookSlug())
	if err != nil && !domain.IsNotFound(err) {
		return err
	}
	if webhook != nil {

		if err := h.webhooks.ArchiveByUuid(webhook.Uuid); err != nil && !domain.IsNotFound(err) {
			return err
		}
	}

	if err := h.store.ArchiveByUuid(h.subject.Uuid); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.JobNotifierDeleted(h.subject), nil)

	ctxt.W().WriteHeader(http.StatusNoContent)

	return nil
}
