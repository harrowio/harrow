package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/authz"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

type webhookHandler struct {
	webhooks   *stores.DbWebhookStore
	deliveries *stores.DbDeliveryStore
	subject    *domain.Webhook
}

func (h *webhookHandler) init(ctxt RequestContext) (*webhookHandler, error) {
	handler := &webhookHandler{}
	handler.webhooks = stores.NewDbWebhookStore(ctxt.Tx())
	handler.deliveries = stores.NewDbDeliveryStore(ctxt.Tx())

	uuid := ctxt.PathParameter("uuid")
	if uuid != "" {
		subject, err := handler.webhooks.FindByUuid(uuid)
		if err != nil {
			return nil, err
		}
		handler.subject = subject
	}

	return handler, nil
}

func (h *webhookHandler) subjectFrom(src io.Reader) (*domain.Webhook, error) {
	dec := json.NewDecoder(src)
	result := struct {
		Subject *domain.Webhook `json:"subject"`
	}{}

	if err := dec.Decode(&result); err != nil {
		return nil, err
	}

	return result.Subject, nil
}

func MountWebhookHandler(r *mux.Router, ctxt ServerContext) {
	h := &webhookHandler{}

	r.PathPrefix("/wh/{slug}").Handler(HandlerFunc(ctxt, h.Deliver)).
		Name("webhook-deliver")

	root := r.PathPrefix("/webhooks").Subrouter()
	// Relationships
	related := root.PathPrefix("/{uuid}/").Subrouter()
	related.Path("/deliveries").Handler(HandlerFunc(ctxt, h.Deliveries)).
		Name("webhook-deliveries")
	related.Methods("PATCH").Path("/slug").Handler(HandlerFunc(ctxt, h.RegenerateSlug)).
		Name("webhook-regenerate-slug")

	// Collection
	root.Methods("POST").Handler(HandlerFunc(ctxt, h.Create)).
		Name("webhook-create")
	root.Methods("PUT").Handler(HandlerFunc(ctxt, h.Update)).
		Name("webhook-update")

	// Item
	item := root.PathPrefix("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, h.Show)).
		Name("webhook-show")
	item.Methods("DELETE").Handler(HandlerFunc(ctxt, h.Archive)).
		Name("webhook-archive")

}

func (h *webhookHandler) Create(ctxt RequestContext) error {

	h, err := h.init(ctxt)
	if err != nil {
		return err
	}

	webhook, err := h.subjectFrom(ctxt.R().Body)
	if err != nil {
		return err
	}

	webhook.CreatorUuid = ctxt.User().Uuid
	webhook.GenerateSlug()

	if allowed, err := ctxt.Auth().CanCreate(webhook); !allowed {
		return err
	}

	if err := webhook.Validate(); err != nil {
		return err
	}

	if _, err := h.webhooks.Create(webhook); err != nil {
		return err
	}

	ctxt.W().Header().Set("Location", urlForSubject(ctxt.R(), webhook))
	ctxt.W().WriteHeader(http.StatusCreated)
	writeAsJson(ctxt, webhook)

	return nil
}

func (h *webhookHandler) Update(ctxt RequestContext) error {

	h, err := h.init(ctxt)
	if err != nil {
		return err
	}

	newVersion, err := h.subjectFrom(ctxt.R().Body)
	if err != nil {
		return err
	}
	h.subject, err = h.webhooks.FindByUuid(newVersion.Uuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanUpdate(h.subject); !allowed {
		return err
	}

	h.subject.Name = newVersion.Name
	h.subject.JobUuid = newVersion.JobUuid

	if err := h.subject.Validate(); err != nil {
		return err
	}

	if err := h.webhooks.Update(h.subject); err != nil {
		return err
	}

	writeAsJson(ctxt, h.subject)

	return nil
}

func (h *webhookHandler) Show(ctxt RequestContext) error {

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

func (h *webhookHandler) Archive(ctxt RequestContext) error {

	h, err := h.init(ctxt)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanArchive(h.subject); !allowed {
		return err
	}

	if err := h.webhooks.ArchiveByUuid(h.subject.Uuid); err != nil {
		return err
	}

	ctxt.W().WriteHeader(http.StatusNoContent)

	return nil
}

func (h *webhookHandler) Deliveries(ctxt RequestContext) error {

	h, err := h.init(ctxt)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(h.subject); !allowed {
		return err
	}

	deliveries, err := h.deliveries.FindByWebhookUuid(h.subject.Uuid)
	if err != nil {
		return err
	}

	result := []interface{}{}
	for _, delivery := range deliveries {
		if allowed, err := ctxt.Auth().CanRead(delivery); allowed {
			result = append(result, delivery)
		} else {
			aerr := err.(*authz.Error)
			if internal := aerr.Internal(); internal != nil {
				ctxt.Log().Error().Msgf("canread(delivery=%q): %s", internal)
			}
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(result),
		Count:      len(result),
		Collection: result,
	})

	return nil
}

func (h *webhookHandler) Deliver(ctxt RequestContext) error {

	h, err := h.init(ctxt)
	if err != nil {
		return err
	}

	slug := ctxt.PathParameter("slug")
	webhook, err := h.webhooks.FindBySlug(slug)
	if err != nil {
		return err
	}

	delivery := webhook.NewDelivery(ctxt.R())
	repositories := stores.NewDbRepositoryStore(ctxt.Tx())
	params := delivery.OperationParameters(webhook.ProjectUuid, repositories)
	scheduleStore := stores.NewDbScheduleStore(ctxt.Tx())
	now := "now"
	schedule := &domain.Schedule{
		UserUuid:    webhook.CreatorUuid,
		JobUuid:     webhook.JobUuid,
		Description: fmt.Sprintf("Triggered by Delivery(%s)", delivery.Uuid),
		CreatedAt:   time.Now(),
		Timespec:    &now,
		Parameters:  params,
	}
	scheduleUuid, err := scheduleStore.Create(schedule)
	if err != nil {
		return err
	}
	delivery.ScheduleUuid = &scheduleUuid

	if _, err := h.deliveries.Create(delivery); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.JobScheduled(schedule, "webhook"), nil)
	ctxt.W().WriteHeader(http.StatusCreated)

	if ctxt.R().Header.Get("Accept") == "text/plain" {
		host := ctxt.R().Host
		fmt.Fprintf(ctxt.W(), fmt.Sprintf("https://%s/#/a/schedules/%s\n", host, scheduleUuid))
	}

	return nil
}

func (h *webhookHandler) RegenerateSlug(ctxt RequestContext) error {

	h, err := h.init(ctxt)
	if err != nil {
		return err
	}

	h.subject.GenerateSlug()
	if err := h.webhooks.Update(h.subject); err != nil {
		return err
	}

	writeAsJson(ctxt, h.subject)
	return nil
}
