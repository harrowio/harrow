package http

import (
	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/domain"
)

type featuresHandler struct{}

func MountFeaturesHandler(r *mux.Router, ctxt ServerContext) {
	h := &featuresHandler{}

	root := r.PathPrefix("/api-features").Subrouter()
	item := root.PathPrefix("/{name}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, h.Feature))
	root.Methods("GET").Handler(HandlerFunc(ctxt, h.List))
}

func (self *featuresHandler) Feature(ctxt RequestContext) error {
	features := domain.NewFeaturesFromConfig(c.FeaturesConfig())
	name := ctxt.PathParameter("name")

	for _, feature := range features {
		if feature.Name == name {
			writeAsJson(ctxt, feature)
			return nil
		}
	}

	return new(domain.NotFoundError)
}

func (self *featuresHandler) List(ctxt RequestContext) error {
	features := domain.NewFeaturesFromConfig(c.FeaturesConfig())
	result := []interface{}{}

	for _, feature := range features {
		result = append(result, feature)
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Count:      len(result),
		Total:      len(result),
		Collection: result,
	})

	return nil
}
