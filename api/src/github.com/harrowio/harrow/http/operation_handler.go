package http

import (
	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

func MountOperationHandler(r *mux.Router, ctxt ServerContext) {
	oh := operationHandler{}

	// Collection
	root := r.PathPrefix("/operations").Subrouter()

	// Item
	item := root.PathPrefix("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, oh.Show)).
		Name("operation-show")

	item.Methods("DELETE").Handler(HandlerFunc(ctxt, oh.Cancel)).
		Name("operation-cancel")
}

type operationHandler struct {
}

func (self operationHandler) Show(ctxt RequestContext) (err error) {

	store := stores.NewDbOperationStore(ctxt.Tx())
	jobStore := stores.NewDbJobStore(ctxt.Tx())

	operation, err := store.FindByUuid(mux.Vars(ctxt.R())["uuid"])
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(operation); !allowed {
		return err
	}

	if operation.JobUuid != nil {
		job, err := jobStore.FindByUuid(*operation.JobUuid)
		if err != nil {
			return err
		}
		if allowed, _ := ctxt.Auth().CanRead(job); allowed {
			operation.Embed("job", job)
		}
	}

	if operation.GitLogs != nil {
		operation.GitLogs.Trim(5)
	}

	result := struct {
		*domain.Operation
		Status string `json:"status"`
	}{operation, operation.Status()}

	writeAsJson(ctxt, result)

	return err
}

func (self operationHandler) Cancel(ctxt RequestContext) error {

	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	store := stores.NewDbOperationStore(ctxt.Tx())

	operationUuid := ctxt.PathParameter("uuid")
	operation, err := store.FindByUuid(operationUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().Can("cancel", operation); !allowed {
		return err
	}

	ctxt.EnqueueActivity(activities.OperationCanceledByUser(operationUuid), nil)

	return nil
}
