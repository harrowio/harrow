package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/limits"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
)

type taskParamsWrapper struct {
	Subject taskParams
}

type taskParams struct {
	Uuid        string `json:"uuid"`
	Body        string `json:"body"`
	Name        string `json:"name"`
	ProjectUuid string `json:"projectUuid"`
	Type        string `json:"type"`
}

func ReadTaskParams(r io.Reader) (*taskParams, error) {
	decoder := json.NewDecoder(r)
	var w taskParamsWrapper
	err := decoder.Decode(&w)
	if err != nil {
		return nil, err
	}
	return &w.Subject, nil
}

func copyTaskParams(p *taskParams, m *domain.Task) {
	m.Uuid = p.Uuid
	m.Body = p.Body
	m.Name = p.Name
	m.ProjectUuid = p.ProjectUuid
	m.Type = p.Type
}

func MountTaskHandler(r *mux.Router, ctxt ServerContext) {

	th := taskHandler{}

	// Collection
	root := r.PathPrefix("/tasks").Subrouter()
	root.Methods("PUT").Handler(HandlerFunc(ctxt, th.CreateUpdate)).
		Name("task-update")
	root.Methods("POST").Handler(HandlerFunc(ctxt, th.CreateUpdate)).
		Name("task-create")

	// Relationships
	related := root.PathPrefix("/{uuid}/").Subrouter()
	related.Methods("GET").Path("/jobs").Handler(HandlerFunc(ctxt, th.Jobs)).
		Name("task-jobs")

	// Item
	item := root.PathPrefix("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, th.Show)).
		Name("task-show")
	item.Methods("DELETE").Handler(HandlerFunc(ctxt, th.Archive)).
		Name("task-archive")

}

type taskHandler struct {
}

func (self taskHandler) Show(ctxt RequestContext) (err error) {

	store := stores.NewDbTaskStore(ctxt.Tx())
	uuid := ctxt.PathParameter("uuid")

	task, err := store.FindByUuid(uuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(task); !allowed {
		return err
	}

	writeAsJson(ctxt, task)

	return err
}

func (self taskHandler) CreateUpdate(ctxt RequestContext) (err error) {

	params, err := ReadTaskParams(ctxt.R().Body)
	if err != nil {
		return err
	}

	store := stores.NewDbTaskStore(ctxt.Tx())

	isNew := true
	var model *domain.Task

	if uuidhelper.IsValid(params.Uuid) {
		model, err = store.FindByUuid(params.Uuid)
		_, notFound := err.(*domain.NotFoundError)
		if err != nil && !notFound {
			return err
		}
		isNew = model == nil
	}

	if isNew {
		model = &domain.Task{}
	}
	copyTaskParams(params, model)

	var uuid string

	if isNew {
		if allowed, err := ctxt.Auth().CanCreate(model); !allowed {
			return err
		}

		if allowed, err := limits.CanCreate(model); !allowed {
			return err
		}
		uuid, err = store.Create(model)
		ctxt.EnqueueActivity(activities.TaskAdded(model), nil)
	} else {
		if allowed, err := ctxt.Auth().CanUpdate(model); !allowed {
			return err
		}
		err = store.Update(model)
		uuid = model.Uuid
		ctxt.EnqueueActivity(activities.TaskEdited(model), nil)
	}
	if err != nil {
		return err
	}

	model, err = store.FindByUuid(uuid)
	if err != nil {
		return err
	}

	if isNew {
		ctxt.W().Header().Add("Location", urlForSubject(ctxt.R(), model))
		ctxt.W().WriteHeader(http.StatusCreated)
	}

	writeAsJson(ctxt, model)

	return err

}

func (self taskHandler) Archive(ctxt RequestContext) (err error) {

	taskUuid := ctxt.PathParameter("uuid")
	taskStore := stores.NewDbTaskStore(ctxt.Tx())

	task, err := taskStore.FindByUuid(taskUuid)
	if err != nil {
		return err
	}

	if allowed, _ := ctxt.Auth().CanArchive(task); !allowed {
		return err
	}

	err = taskStore.ArchiveByUuid(taskUuid)
	if err != nil {
		return err
	}

	ctxt.W().WriteHeader(http.StatusNoContent)

	return err
}

func (self taskHandler) Jobs(ctxt RequestContext) (err error) {

	taskUuid := ctxt.PathParameter("uuid")
	taskStore := stores.NewDbTaskStore(ctxt.Tx())
	jobStore := stores.NewDbJobStore(ctxt.Tx())

	task, err := taskStore.FindByUuid(taskUuid)
	if err != nil {
		return err
	}

	if allowed, _ := ctxt.Auth().CanRead(task); !allowed {
		return err
	}

	jobs, err := jobStore.FindAllByTaskUuid(taskUuid)
	if err != nil {
		return err
	}

	var interfaceJobs []interface{} = make([]interface{}, 0, len(jobs))
	for _, j := range jobs {
		if allowed, _ := ctxt.Auth().CanRead(j); allowed {
			interfaceJobs = append(interfaceJobs, j)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceJobs),
		Count:      len(interfaceJobs),
		Collection: interfaceJobs,
	})

	return err

}
