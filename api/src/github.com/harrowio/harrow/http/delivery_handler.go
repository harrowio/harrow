package http

import (
	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/stores"
)

type deliveryHandler struct{}

func MountDeliveryHandler(r *mux.Router, ctxt ServerContext) {
	h := &deliveryHandler{}
	root := r.PathPrefix("/deliveries")

	item := root.Subrouter()
	item.Methods("GET").Path("/{uuid}").Handler(HandlerFunc(ctxt, h.Show)).
		Name("delivery-show")
}

func (h *deliveryHandler) Show(ctxt RequestContext) error {

	deliveryStore := stores.NewDbDeliveryStore(ctxt.Tx())
	delivery, err := deliveryStore.FindByUuid(ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(delivery); !allowed {
		return err
	}

	writeAsJson(ctxt, delivery)

	return nil
}
