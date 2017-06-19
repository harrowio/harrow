package http

import (
	"strconv"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

func MountActivitiesHandler(r *mux.Router, ctxt ServerContext) {

	h := &activitiesHandler{}

	// Collection
	root := r.PathPrefix("/activities").Subrouter()

	// Relationships
	related := root.PathPrefix("/{id}/").Subrouter()
	related.Methods("PUT").Path("/read-status").Handler(HandlerFunc(ctxt, h.ReadStatus)).
		Name("activities-read-status")
}

type activitiesHandler struct {
}

func (self *activitiesHandler) ReadStatus(ctxt RequestContext) error {

	user := ctxt.User()
	if user == nil {
		return ErrLoginRequired
	}

	stream := stores.NewKVActivityStreamStore(ctxt.KeyValueStore())
	stream.SetLogger(ctxt.Log())

	activityId, err := strconv.Atoi(ctxt.PathParameter("id"))
	if err != nil {
		return domain.NewValidationError("id", "malformed")
	}

	if err := stream.MarkAsRead(activityId, user.Uuid); err != nil {
		return err
	}

	return nil
}
