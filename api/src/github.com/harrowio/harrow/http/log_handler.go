package http

import (
	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

type logHandler struct {
	logStores []stores.LogStore
}

func MountLogHandler(r *mux.Router, ctxt ServerContext) {
	config := ctxt.Config()
	lh := logHandler{
		logStores: []stores.LogStore{ // the order is important
			stores.NewRedisLogStore(ctxt.KeyValueStore()),
			stores.NewDiskLogStore(config.FilesystemConfig().LogDir),
		},
	}

	// Collection
	root := r.PathPrefix("/logs").Subrouter()

	// Item
	item := root.PathPrefix("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, lh.Show)).
		Name("log-show")
}

func (self logHandler) Show(ctxt RequestContext) error {

	operationStore := stores.NewDbOperationStore(ctxt.Tx())
	jobStore := stores.NewDbJobStore(ctxt.Tx())

	operationUuid := mux.Vars(ctxt.R())["uuid"]

	operation, err := operationStore.FindByUuid(operationUuid)
	if err != nil {
		return err
	}

	job, err := jobStore.FindByUuid(*operation.JobUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(job); !allowed {
		return err
	}

	var log *domain.Loggable

	for _, logStore := range self.logStores {
		log, err = logStore.FindByOperationUuid(operationUuid, domain.LoggableWorkspace)
		if log == (*domain.Loggable)(nil) {
			// if a log could not be found at the first (redis) store, just continue to the next
			err = nil
			continue
		} else {
			break
		}
	}

	//
	// Handle the case where neither store found a log because
	// the operation might not have started yet.
	//
	if log == (*domain.Loggable)(nil) {
		return &domain.NotFoundError{}
	}

	if err != nil {
		return err
	}

	writeAsJson(ctxt, log)

	return nil
}
