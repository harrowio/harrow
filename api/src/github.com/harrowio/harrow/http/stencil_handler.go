package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

type stencilHandler struct {
	stencils *stores.DbStencilStore
	subject  *domain.Stencil
}

func (h *stencilHandler) init(ctxt RequestContext) (*stencilHandler, error) {
	handler := &stencilHandler{}
	handler.stencils = stores.NewDbStencilStore(ctxt.SecretKeyValueStore(), ctxt.Tx(), ctxt)

	return handler, nil
}

func (h *stencilHandler) subjectFrom(src io.Reader) (*domain.Stencil, error) {
	dec := json.NewDecoder(src)
	result := struct {
		Subject *domain.Stencil `json:"subject"`
	}{}

	if err := dec.Decode(&result); err != nil {
		return nil, err
	}

	return result.Subject, nil
}

func MountStencilHandler(r *mux.Router, ctxt ServerContext) {
	h := &stencilHandler{}

	root := r.PathPrefix("/stencils").Subrouter()
	// Collection
	root.Methods("POST").Handler(HandlerFunc(ctxt, h.Instantiate)).
		Name("stencil-instantiate")

	root.Methods("GET").Handler(HandlerFunc(ctxt, h.List)).
		Name("stencil-list")
}

func (h *stencilHandler) Instantiate(ctxt RequestContext) error {

	h, err := h.init(ctxt)
	if err != nil {
		return err
	}

	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	stencil, err := h.subjectFrom(ctxt.R().Body)
	if err != nil {
		return err
	}

	stencil.UserUuid = ctxt.User().Uuid
	stencil.UrlHost = ctxt.User().UrlHost
	stencil.NotifyViaEmail = ctxt.User().Email

	if allowed, err := ctxt.Auth().CanCreate(stencil); !allowed {
		return err
	}

	if err := stencil.Validate(); err != nil {
		return err
	}

	if err := h.stencils.Create(stencil); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.StencilInstantiated(stencil), nil)

	ctxt.W().Header().Set("Location", urlForSubject(ctxt.R(), stencil))
	ctxt.W().WriteHeader(http.StatusCreated)
	writeAsJson(ctxt, stencil)

	return nil
}

func (h *stencilHandler) List(ctxt RequestContext) error {
	result := []interface{}{
		&domain.Stencil{
			Id: "capistrano-rails",
		},
		&domain.Stencil{
			Id: "bash-linux",
		},
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Count:      len(result),
		Total:      len(result),
		Collection: result,
	})
	return nil
}
