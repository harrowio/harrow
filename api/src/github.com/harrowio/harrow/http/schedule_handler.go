package http

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/dhamidi/timespec"
	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

func ReadSchedParams(r io.Reader) (*schedParams, error) {
	decoder := json.NewDecoder(r)
	var w schedParamsWrapper
	err := decoder.Decode(&w)
	if err != nil {
		return nil, err
	}
	return &w.Subject, nil
}

type schedParamsWrapper struct {
	Subject schedParams
}

type schedParams struct {
	Uuid        string  `json:"uuid"`
	JobUuid     string  `json:"jobUuid"`
	Cronspec    *string `json:"cronspec"`
	Timespec    *string `json:"timespec"`
	Description string  `json:"description"`
}

func MountScheduleHandler(r *mux.Router, ctxt ServerContext) {

	sh := schedHandler{}

	// Collection
	root := r.PathPrefix("/schedules").Subrouter()
	root.Methods("POST").Handler(HandlerFunc(ctxt, sh.Create)).
		Name("schedule-create")
	root.Methods("PUT").Handler(HandlerFunc(ctxt, sh.Edit)).
		Name("schedule-edit")

	// Item
	item := root.PathPrefix("/{uuid}").Subrouter()
	item.Methods("DELETE").Handler(HandlerFunc(ctxt, sh.Delete)).
		Name("schedule-delete")
	item.Methods("GET").Path("/scheduled-executions").Handler(HandlerFunc(ctxt, sh.ScheduledExecutions)).
		Name("schedule-scheduled-executions")
	item.Methods("GET").Path("/operations").Handler(HandlerFunc(ctxt, sh.Operations)).
		Name("schedule-operations")
	item.Methods("GET").Handler(HandlerFunc(ctxt, sh.Show)).
		Name("schedule-show")

}

type schedHandler struct {
}

func (self schedHandler) Show(ctxt RequestContext) (err error) {

	scheduleUuid := ctxt.PathParameter("uuid")
	scheduleStore := stores.NewDbScheduleStore(ctxt.Tx())
	schedule, err := scheduleStore.FindByUuid(scheduleUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(schedule); !allowed {
		return err
	}

	scheduledExecutions, err := domain.ExecutionsBetween(
		time.Now(),
		time.Now().Add(domain.ScheduledExecutionDefaultInterval),
		domain.ScheduledExecutionDefaultN,
		[]*domain.Schedule{schedule},
	)
	if err != nil {
		return err
	}

	schedule.NextExecutions = scheduledExecutions

	writeAsJson(ctxt, schedule)

	return err
}

func (self schedHandler) Edit(ctxt RequestContext) error {

	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	params, err := ReadSchedParams(ctxt.R().Body)
	if err != nil {
		return err
	}

	schedules := stores.NewDbScheduleStore(ctxt.Tx())
	schedule, err := schedules.FindByUuid(params.Uuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanUpdate(schedule); !allowed {
		return err
	}

	schedule.Cronspec = params.Cronspec
	schedule.Timespec = params.Timespec
	schedule.Description = params.Description

	if err := schedules.Update(schedule); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.ScheduleEdited(schedule), nil)

	return nil
}

func (self schedHandler) Delete(ctxt RequestContext) error {

	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	scheduleUuid := ctxt.PathParameter("uuid")
	schedules := stores.NewDbScheduleStore(ctxt.Tx())

	schedule, err := schedules.FindByUuid(scheduleUuid)
	if err != nil {
		return err
	}
	if err := schedules.ArchiveByUuid(scheduleUuid); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.ScheduleDeleted(schedule), nil)

	return nil
}

func (self schedHandler) Create(ctxt RequestContext) (err error) {

	params, err := ReadSchedParams(ctxt.R().Body)
	if err != nil {
		return err
	}

	operationParameters := domain.NewOperationParameters()
	if user := ctxt.User(); user != nil {
		operationParameters.Username = ctxt.User().Name
		operationParameters.UserUuid = ctxt.User().Uuid
	}

	schedule := &domain.Schedule{
		Uuid:        params.Uuid,
		JobUuid:     params.JobUuid,
		UserUuid:    ctxt.User().Uuid,
		Cronspec:    params.Cronspec,
		Timespec:    params.Timespec,
		Description: params.Description,
		Parameters:  operationParameters,
	}

	if err := schedule.Validate(); err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanCreate(schedule); !allowed {
		return err
	}

	theLimits, err := NewLimitsFromContext(ctxt)
	if err != nil {
		return err
	}

	if exceeded, err := theLimits.Exceeded(schedule); exceeded {
		return ErrLimitsExceeded
	} else if err != nil {
		ctxt.Log().Warn().Msgf("error calculating limits: %s", err)
	}

	scheduleStore := stores.NewDbScheduleStore(ctxt.Tx())
	scheduleUuid, err := scheduleStore.Create(schedule)
	if err != nil {
		return err
	}

	schedule, err = scheduleStore.FindByUuid(scheduleUuid)
	if err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.JobScheduled(schedule, "user"), nil)

	ctxt.W().Header().Add("Location", urlForSubject(ctxt.R(), schedule))
	ctxt.W().WriteHeader(http.StatusCreated)

	writeAsJson(ctxt, schedule)

	return err

}

func (self schedHandler) ScheduledExecutions(ctxt RequestContext) error {

	scheduleUuid := ctxt.PathParameter("uuid")
	scheduleStore := stores.NewDbScheduleStore(ctxt.Tx())
	schedule, err := scheduleStore.FindByUuid(scheduleUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(schedule); !allowed {
		return err
	}

	scheduledExecutions := []*domain.ScheduledExecution{}
	if schedule.Timespec != nil {
		ts, _ := timespec.Parse(*schedule.Timespec)
		runAt := ts.Resolve(domain.Clock.Now())
		scheduledExecutions = append(scheduledExecutions, &domain.ScheduledExecution{
			Time:        runAt,
			JobUuid:     schedule.JobUuid,
			Spec:        *schedule.Timespec,
			Description: schedule.Description,
		})
	} else {

		params, err := NewScheduledExecutionsParams(ctxt.R().URL.Query())
		if err != nil {
			return err
		}

		scheduledExecutions, err = domain.ExecutionsBetween(params.from, params.to, domain.ScheduledExecutionDefaultN, []*domain.Schedule{schedule})
		if err != nil {
			return err
		}
	}

	result := make([]interface{}, 0, len(scheduledExecutions))
	for _, execution := range scheduledExecutions {
		if allowed, _ := ctxt.Auth().CanRead(execution); allowed {
			result = append(result, execution)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(result),
		Count:      len(result),
		Collection: result,
	})

	return err
}

func (self schedHandler) Operations(ctxt RequestContext) error {

	scheduleUuid := ctxt.PathParameter("uuid")
	scheduleStore := stores.NewDbScheduleStore(ctxt.Tx())
	schedule, err := scheduleStore.FindByUuid(scheduleUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(schedule); !allowed {
		return err
	}

	operationStore := stores.NewDbOperationStore(ctxt.Tx())
	operations, err := operationStore.FindAllByScheduleUuid(scheduleUuid)
	if err != nil {
		return err
	}

	result := make([]interface{}, 0, len(operations))
	for _, operation := range operations {
		if allowed, _ := ctxt.Auth().CanRead(operation); allowed {
			result = append(result, operation)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(result),
		Count:      len(result),
		Collection: result,
	})

	return err
}
