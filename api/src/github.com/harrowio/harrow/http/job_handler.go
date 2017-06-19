package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/limits"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
)

type jobParamsWrapper struct {
	Subject jobParams
}

type jobParams struct {
	Uuid            string  `json:"uuid"`
	Name            string  `json:"name"`
	Description     *string `json:"description"`
	TaskUuid        string  `json:"taskUuid"`
	EnvironmentUuid string  `json:"environmentUuid"`
}

func ReadJobParams(r io.Reader) (*jobParams, error) {
	var w jobParamsWrapper
	if err := json.NewDecoder(r).Decode(&w); err != nil {
		return nil, err
	}
	return &w.Subject, nil
}

func copyJobParams(j *jobParams, m *domain.Job) {
	m.Uuid = j.Uuid
	m.Name = j.Name
	m.Description = j.Description
	m.TaskUuid = j.TaskUuid
	m.EnvironmentUuid = j.EnvironmentUuid
}

func MountJobHandler(r *mux.Router, ctxt ServerContext) {

	jh := &jobHandler{}
	jbh := &jobBadgeHandler{}

	// Collection

	root := r.PathPrefix("/jobs").Subrouter()
	root.Methods("POST").Handler(HandlerFunc(ctxt, jh.Create)).
		Name("job-create")

	// Relationships
	related := root.PathPrefix("/{uuid}/").Subrouter()
	related.Methods("GET").Path("/operations").Handler(HandlerFunc(ctxt, jh.Operations)).
		Name("job-operations")
	related.Methods("GET").Path("/scheduled-executions").Handler(HandlerFunc(ctxt, jh.ScheduledExecutions)).
		Name("job-scheduled-executions")
	related.Methods("GET").Path("/subscriptions").Handler(HandlerFunc(ctxt, jh.Subscriptions)).
		Name("job-subscriptions")
	related.Methods("GET").Path("/job-notifiers").Handler(HandlerFunc(ctxt, jh.JobNotifiers)).
		Name("job-notifiers")
	related.Methods("GET").Path("/notification-rules").Handler(HandlerFunc(ctxt, jh.NotificationRules)).
		Name("job-notification-rules")
	related.Methods("GET").Path("/watch").Handler(HandlerFunc(ctxt, jh.WatchStatus)).
		Name("job-watch-status")
	related.Methods("PUT").Path("/watch").Handler(HandlerFunc(ctxt, jh.Watch)).
		Name("job-watch")

	related.Methods("PUT").Path("/subscriptions").Handler(HandlerFunc(ctxt, jh.Subscribe)).
		Name("job-subscribe")

	// BuildBadges
	badges := related.PathPrefix("/build-badges/").Subrouter()
	badges.Methods("GET").Path("/simple.svg").Handler(HandlerFunc(ctxt, jbh.Simple)).
		Name("job-build-badge-simple")

	// Triggers
	triggers := related.PathPrefix("/triggers/").Subrouter()
	triggers.Methods("GET").Path("/schedules").Handler(HandlerFunc(ctxt, jh.ScheduleTriggers)).
		Name("job-triggers-schedules")
	triggers.Methods("GET").Path("/webhooks").Handler(HandlerFunc(ctxt, jh.WebhookTriggers)).
		Name("job-triggers-webhooks")
	triggers.Methods("GET").Path("/git").Handler(HandlerFunc(ctxt, jh.GitTriggers)).
		Name("job-triggers-git")
	triggers.Methods("GET").Path("/jobs").Handler(HandlerFunc(ctxt, jh.JobTriggers)).
		Name("job-triggers-jobs")

	// Item
	item := root.PathPrefix("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, jh.Show)).
		Name("job-show")
	item.Methods("DELETE").Handler(HandlerFunc(ctxt, jh.Archive)).
		Name("job-archive")

	root.Methods("PUT").Handler(HandlerFunc(ctxt, jh.Update)).
		Name("job-update")
}

type jobBadgeHandler struct {
}

func (self jobBadgeHandler) Simple(ctxt RequestContext) (err error) {

	uuid := ctxt.PathParameter("uuid")

	jobStore := stores.NewDbJobStore(ctxt.Tx())
	opStore := stores.NewDbOperationStore(ctxt.Tx())

	job, err := jobStore.FindByUuid(uuid)
	if err != nil {
		return err
	}

	mostRecentOps, err := opStore.FindRecentByJobUuid(1, job.Uuid)
	if err != nil {
		return err
	}

	imgPathPattern := "https://img.shields.io/badge/%s-%s-%s.svg%s"
	statusText := "not run yet"
	color, params := "lightgrey", ""

	if len(mostRecentOps) > 0 {
		color, params = mostRecentOps[0].ShieldsIoStatusColorAndParams()
		statusText = mostRecentOps[0].Status()
	}

	shieldsIoUrl := url.URL{
		Scheme: "https",
		Host:   "img.shields.io",
		Path:   fmt.Sprintf(imgPathPattern, job.Name, strings.Title(statusText), color, params),
	}

	res, err := http.Get(shieldsIoUrl.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()

	ctxt.W().Header().Set("Content-Type", "image/svg+xml")
	if _, err := io.Copy(ctxt.W(), res.Body); err != nil {
		return err
	}

	return err
}

func (self jobBadgeHandler) Detailed(ctxt RequestContext) (err error) {
	return err
}

type jobHandler struct {
}

func (self jobHandler) Show(ctxt RequestContext) (err error) {

	store := stores.NewDbJobStore(ctxt.Tx())
	job, err := store.FindByUuid(ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(job); !allowed {
		return err
	}

	writeAsJson(ctxt, job)

	return err
}

func (self jobHandler) Create(ctxt RequestContext) (err error) {

	store := stores.NewDbJobStore(ctxt.Tx())

	params, err := ReadJobParams(ctxt.R().Body)
	if err != nil {
		return err
	}

	job := &domain.Job{}

	copyJobParams(params, job)

	var uuid string

	if allowed, err := ctxt.Auth().CanCreate(job); !allowed {
		return err
	}

	if allowed, err := limits.CanCreate(job); !allowed {
		return err
	}

	uuid, err = store.Create(job)

	if err != nil {
		return err
	}

	job, err = store.FindByUuid(uuid)
	if err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.JobAdded(job), nil)

	ctxt.W().Header().Add("Location", urlForSubject(ctxt.R(), job))
	ctxt.W().WriteHeader(http.StatusCreated)

	writeAsJson(ctxt, job)

	return err

}

func (self jobHandler) Update(ctxt RequestContext) (err error) {

	store := stores.NewDbJobStore(ctxt.Tx())

	params, err := ReadJobParams(ctxt.R().Body)
	if err != nil {
		return err
	}
	job, err := store.FindByUuid(params.Uuid)
	if err != nil {
		return err
	}

	copyJobParams(params, job)

	if allowed, err := ctxt.Auth().CanUpdate(job); !allowed {
		return err
	}

	if err := store.Update(job); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.JobEdited(job), nil)

	writeAsJson(ctxt, job)

	return nil
}

func (self jobHandler) Archive(ctxt RequestContext) (err error) {

	uuid := ctxt.PathParameter("uuid")

	store := stores.NewDbJobStore(ctxt.Tx())

	job, err := store.FindByUuid(uuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanArchive(job); !allowed {
		return err
	}

	if err = store.ArchiveByUuid(uuid); err != nil {
		return err
	}

	ctxt.W().WriteHeader(http.StatusNoContent)

	return err
}

func (self jobHandler) Operations(ctxt RequestContext) (err error) {

	uuid := ctxt.PathParameter("uuid")

	jobStore := stores.NewDbJobStore(ctxt.Tx())
	opStore := stores.NewDbOperationStore(ctxt.Tx())

	job, err := jobStore.FindByUuid(uuid)
	if err != nil {
		return err
	}

	if allowed, _ := ctxt.Auth().CanRead(job); !allowed {
		return err
	}

	operations, err := opStore.FindRecentByJobUuid(20, job.Uuid)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	var interfaceOperations []interface{} = make([]interface{}, 0, len(operations))
	for _, o := range operations {
		if allowed, _ := ctxt.Auth().CanRead(o); allowed {
			o.Embed("job", job)
			if o.GitLogs != nil {
				o.GitLogs.Trim(5)
			}
			interfaceOperations = append(interfaceOperations, o)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceOperations),
		Count:      len(interfaceOperations),
		Collection: interfaceOperations,
	})

	return err
}

func (self jobHandler) ScheduledExecutions(ctxt RequestContext) (err error) {

	uuid := ctxt.PathParameter("uuid")

	params, err := NewScheduledExecutionsParams(ctxt.R().URL.Query())
	if err != nil {
		return err
	}

	jobStore := stores.NewDbJobStore(ctxt.Tx())
	scheduleStore := stores.NewDbScheduleStore(ctxt.Tx())

	job, err := jobStore.FindByUuid(uuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(job); !allowed {
		return err
	}

	schedules, err := scheduleStore.FindAllByJobUuid(uuid)
	if err != nil {
		return err
	}

	executions, err := domain.ExecutionsBetween(params.from, params.to, domain.ScheduledExecutionDefaultN, schedules)
	if err != nil {
		return err
	}

	var interfaceExecutions []interface{} = make([]interface{}, 0, len(executions))
	for _, e := range executions {
		if allowed, _ := ctxt.Auth().CanRead(e); allowed {
			interfaceExecutions = append(interfaceExecutions, e)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceExecutions),
		Count:      len(interfaceExecutions),
		Collection: interfaceExecutions,
	})

	return err

}

type subscribeParams struct {
	Watch  bool            `json:"watch"`
	Events map[string]bool `json:"events"`
}

func (params *subscribeParams) handle(user *domain.User, watchable domain.Watchable, subscriptions domain.SubscriptionStore) (err error) {

	if params.Watch {
		return user.Watch(watchable, subscriptions)
	}

	if params.Events == nil {
		return user.Unwatch(watchable, subscriptions)
	}

	for event, subscribe := range params.Events {
		if subscribe {
			err = user.SubscribeTo(watchable, event, subscriptions)
		} else {
			err = user.UnsubscribeFrom(watchable, event, subscriptions)
		}

		if err != nil {
			if _, ok := err.(*domain.NotFoundError); !ok {
				return err
			}
		}
	}

	return nil
}

func NewSubscribeParams(src io.Reader) (*subscribeParams, error) {
	params := &subscribeParams{}
	dec := json.NewDecoder(src)
	if err := dec.Decode(params); err != nil {
		return nil, err
	} else {
		return params, nil
	}
}

func (self jobHandler) Subscribe(ctxt RequestContext) error {

	params, err := NewSubscribeParams(ctxt.R().Body)
	if err != nil {
		return err
	}

	store := stores.NewDbJobStore(ctxt.Tx())
	subscriptions := stores.NewDbSubscriptionStore(ctxt.Tx())
	job, err := store.FindByUuid(ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(job); !allowed {
		return err
	}

	if err := params.handle(ctxt.User(), job, subscriptions); err != nil {
		return err
	}

	ctxt.W().WriteHeader(http.StatusNoContent)
	return nil
}

func (self jobHandler) Subscriptions(ctxt RequestContext) error {

	store := stores.NewDbJobStore(ctxt.Tx())
	subscriptions := stores.NewDbSubscriptionStore(ctxt.Tx())
	job, err := store.FindByUuid(ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(job); !allowed {
		return err
	}

	userSubscriptions, err := ctxt.User().SubscriptionsFor(job, subscriptions)
	if err != nil {
		return err
	}

	writeAsJson(ctxt, userSubscriptions)
	return nil
}

func (self jobHandler) Watch(ctxt RequestContext) error {

	params, err := NewSubscribeParams(ctxt.R().Body)
	if err != nil {
		return err
	}

	user := ctxt.User()
	if user == nil {
		return ErrLoginRequired
	}

	jobs := stores.NewDbJobStore(ctxt.Tx())
	job, err := jobs.FindByUuid(ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}

	notifier := &domain.EmailNotifier{
		Uuid:      uuidhelper.MustNewV4(),
		Recipient: user.Email,
		UrlHost:   user.UrlHost,
	}
	notifiers := stores.NewDbEmailNotifierStore(ctxt.Tx())
	if existingNotifier, err := notifiers.FindByRecipient(user.Email); err == nil {
		notifier = existingNotifier
	} else {
		_, err := notifiers.Create(notifier)
		if err != nil {
			return err
		}
	}

	project, err := stores.NewDbProjectStore(ctxt.Tx()).FindByJobUuid(job.Uuid)
	if err != nil {
		return err
	}

	rules := stores.NewDbNotificationRuleStore(ctxt.Tx())
	rule, err := rules.FindByNotifierAndJobUuidAndType(notifier.Uuid, job.Uuid, "email_notifiers")
	if err != nil {
		if _, notFoundError := err.(*domain.NotFoundError); !notFoundError {
			return err
		}
	}

	if params.Watch {
		if rule == nil {
			rule = &domain.NotificationRule{
				Uuid:          uuidhelper.MustNewV4(),
				ProjectUuid:   project.Uuid,
				NotifierUuid:  notifier.Uuid,
				NotifierType:  "email_notifiers",
				MatchActivity: "operation.*",
				JobUuid:       &job.Uuid,
				CreatorUuid:   user.Uuid,
			}

			if err := rule.Validate(); err != nil {
				return err
			}

			_, err := rules.Create(rule)
			if err != nil {
				return err
			}
			writeAsJson(ctxt, rule)
		}
	} else {
		if rule != nil {
			if err := rules.ArchiveByUuid(rule.Uuid); err != nil {
				return err
			}
			writeAsJson(ctxt, rule)
		}
	}

	return nil
}

func (self jobHandler) WatchStatus(ctxt RequestContext) error {
	user := ctxt.User()
	if user == nil {
		return ErrLoginRequired
	}

	jobs := stores.NewDbJobStore(ctxt.Tx())
	job, err := jobs.FindByUuid(ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}

	result := []interface{}{}
	notifiers := stores.NewDbEmailNotifierStore(ctxt.Tx())
	existingNotifier, err := notifiers.FindByRecipient(user.Email)
	if err != nil {
		if _, isNotFound := err.(*domain.NotFoundError); !isNotFound {
			return err
		}
	}
	if existingNotifier != nil {
		rules := stores.NewDbNotificationRuleStore(ctxt.Tx())
		rule, err := rules.FindByNotifierAndJobUuidAndType(existingNotifier.Uuid, job.Uuid, "email_notifiers")
		if err != nil {
			if _, isNotFound := err.(*domain.NotFoundError); !isNotFound {
				return err
			}
		} else {
			result = append(result, rule)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(result),
		Count:      len(result),
		Collection: result,
	})

	return nil
}

func (self *jobHandler) JobNotifiers(ctxt RequestContext) error {

	jobs := stores.NewDbJobStore(ctxt.Tx())
	job, err := jobs.FindByUuid(ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}

	notifiers, err := stores.NewDbJobNotifierStore(ctxt.Tx()).FindAllByJobUuid(job.Uuid)
	if err != nil {
		return err
	}

	result := []interface{}{}
	for _, notifier := range notifiers {
		if allowed, _ := ctxt.Auth().CanRead(notifier); allowed {
			result = append(result, notifier)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(result),
		Count:      len(result),
		Collection: result,
	})

	return nil
}

func (self *jobHandler) ScheduleTriggers(ctxt RequestContext) error {

	jobUuid := ctxt.PathParameter("uuid")
	schedules := stores.NewDbScheduleStore(ctxt.Tx())
	triggers, err := schedules.FindAllRecurringByJobUuid(jobUuid)
	if err != nil {
		return err
	}

	result := []interface{}{}
	for _, trigger := range triggers {
		if allowed, _ := ctxt.Auth().CanRead(trigger); allowed {
			result = append(result, trigger)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(result),
		Count:      len(result),
		Collection: result,
	})

	return nil
}

func (self *jobHandler) WebhookTriggers(ctxt RequestContext) error {

	jobUuid := ctxt.PathParameter("uuid")
	webhooks := stores.NewDbWebhookStore(ctxt.Tx())
	triggers, err := webhooks.FindAllByJobUuid(jobUuid)
	if err != nil {
		return err
	}

	result := []interface{}{}
	for _, trigger := range triggers {
		if trigger.IsInternal() {
			continue
		}

		if allowed, _ := ctxt.Auth().CanRead(trigger); allowed {
			result = append(result, trigger)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(result),
		Count:      len(result),
		Collection: result,
	})

	return nil
}

func (self *jobHandler) GitTriggers(ctxt RequestContext) error {

	jobUuid := ctxt.PathParameter("uuid")
	gitTriggers := stores.NewDbGitTriggerStore(ctxt.Tx())
	triggers, err := gitTriggers.FindAllByJobUuid(jobUuid)
	if err != nil {
		return err
	}

	result := []interface{}{}
	for _, trigger := range triggers {
		if allowed, _ := ctxt.Auth().CanRead(trigger); allowed {
			result = append(result, trigger)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(result),
		Count:      len(result),
		Collection: result,
	})

	return nil
}

func (self *jobHandler) JobTriggers(ctxt RequestContext) error {

	jobUuid := ctxt.PathParameter("uuid")
	jobStore := stores.NewDbJobStore(ctxt.Tx())
	jobTriggers := stores.NewDbJobNotifierStore(ctxt.Tx())
	triggers, err := jobTriggers.FindAllByTriggeredJobUuid(jobUuid)
	if err != nil {
		return err
	}

	notificationRuleStore := stores.NewDbNotificationRuleStore(ctxt.Tx())
	result := []interface{}{}
	for _, trigger := range triggers {
		if allowed, _ := ctxt.Auth().CanRead(trigger); allowed {
			triggeringRules, err := notificationRuleStore.FindAllByNotifierUuid(trigger.Uuid)
			if err != nil {
				ctxt.Log().Warn().Msgf("failed to fetch triggering rules for job notifier %q", trigger.Uuid)
			} else {
				for _, rule := range triggeringRules {
					if rule.JobUuid != nil {
						job, err := jobStore.FindByUuid(*rule.JobUuid)
						if err != nil {
							ctxt.Log().Warn().Msgf("failed to fetch triggering job %q", rule.JobUuid)
						} else {
							rule.Embed("job", job)
						}
					}
					trigger.Embedded()["notificationRules"] = append(trigger.Embedded()["notificationRules"], rule)
				}
			}
			result = append(result, trigger)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(result),
		Count:      len(result),
		Collection: result,
	})

	return nil
}

func (self *jobHandler) NotificationRules(ctxt RequestContext) error {

	jobs := stores.NewDbJobStore(ctxt.Tx())
	job, err := jobs.FindByUuid(ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}

	notificationRules, err := stores.NewDbNotificationRuleStore(ctxt.Tx()).FindAllByJobUuid(job.Uuid)
	if err != nil {
		return err
	}

	notifiers := stores.NewDbNotifierStore(ctxt.Tx())
	result := []interface{}{}
	for _, notificationRule := range notificationRules {
		if allowed, _ := ctxt.Auth().CanRead(notificationRule); allowed {
			notifier, err := notifiers.FindByUuidAndType(notificationRule.NotifierUuid, notificationRule.NotifierType)
			if err == nil {
				subject, ok := notifier.(domain.Subject)
				if ok {
					notificationRule.Embed("notifier", subject)
				}
			} else {
				ctxt.Log().Warn().Msgf("fetching notifier(%s=%s) for job %s: %s", notificationRule.NotifierType, notificationRule.NotifierUuid, job.Uuid, err)
			}
			result = append(result, notificationRule)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(result),
		Count:      len(result),
		Collection: result,
	})

	return nil
}
