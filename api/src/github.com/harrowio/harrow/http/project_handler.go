package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/authz"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/domain/stencil"
	"github.com/harrowio/harrow/limits"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
)

func ReadProjectParams(r io.Reader) (*ProjectParams, error) {
	decoder := json.NewDecoder(r)
	var w ProjectPararamsWrapper
	err := decoder.Decode(&w)
	if err != nil {
		return nil, err
	}
	return &w.Subject, nil
}

type ProjectPararamsWrapper struct {
	Subject ProjectParams
}

type ProjectParams struct {
	Uuid             string `json:"uuid"`
	Name             string `json:"name"`
	OrganizationUuid string `json:"organizationUuid"`
	Public           bool   `json:"public"`
}

func copyProjParams(p *ProjectParams, m *domain.Project) {
	m.Uuid = p.Uuid
	m.Name = p.Name
	m.OrganizationUuid = p.OrganizationUuid
	m.Public = p.Public
}

func MountProjectHandler(r *mux.Router, ctxt ServerContext) {

	ph := projectHandler{}

	// Collection
	root := r.PathPrefix("/projects").Subrouter()

	// Relationships
	related := root.PathPrefix("/{uuid}/").Subrouter()
	related.Methods("GET").Path("/card").Handler(HandlerFunc(ctxt, ph.Card)).
		Name("project-card")
	related.Methods("GET").Path("/environments").Handler(HandlerFunc(ctxt, ph.Environments)).
		Name("project-environments")
	related.Methods("GET").Path("/repositories").Handler(HandlerFunc(ctxt, ph.Repositories)).
		Name("project-repositories")
	related.Methods("GET").Path("/jobs").Handler(HandlerFunc(ctxt, ph.Jobs)).
		Name("project-jobs")
	related.Methods("GET").Path("/tasks").Handler(HandlerFunc(ctxt, ph.Tasks)).
		Name("project-tasks")
	related.Methods("GET").Path("/operations").Handler(HandlerFunc(ctxt, ph.Operations)).
		Name("project-operations")
	related.Methods("GET").Path("/git-triggers").Handler(HandlerFunc(ctxt, ph.GitTriggers)).
		Name("project-git-triggers")
	related.Methods("GET").Path("/webhooks").Handler(HandlerFunc(ctxt, ph.Webhooks)).
		Name("project-webhooks")
	related.Methods("GET").Path("/memberships").Handler(HandlerFunc(ctxt, ph.Memberships)).
		Name("project-memberships")
	related.Methods("POST").Path("/members").Handler(HandlerFunc(ctxt, ph.AddMember)).
		Name("project-add-member")
	related.Methods("GET").Path("/members").Handler(HandlerFunc(ctxt, ph.Members)).
		Name("project-members")
	related.Methods("DELETE").Path("/members").Handler(HandlerFunc(ctxt, ph.Leave)).
		Name("project-leave")
	related.Methods("GET").Path("/schedules").Handler(HandlerFunc(ctxt, ph.Schedules)).
		Name("project-schedules")
	related.Methods("GET").Path("/scheduled-executions").Handler(HandlerFunc(ctxt, ph.ScheduledExecutions)).
		Name("project-scheduled-executions")
	related.Methods("GET").Path("/job-notifiers").Handler(HandlerFunc(ctxt, ph.JobNotifiers)).
		Name("project-job-notifiers")
	related.Methods("GET").Path("/slack-notifiers").Handler(HandlerFunc(ctxt, ph.SlackNotifiers)).
		Name("project-slack-notifiers")
	related.Methods("GET").Path("/email-notifiers").Handler(HandlerFunc(ctxt, ph.EmailNotifiers)).
		Name("project-email-notifiers")
	related.Methods("GET").Path("/scripts").Handler(HandlerFunc(ctxt, ph.ScriptCards)).
		Name("project-script-cards")
	related.Methods("GET").Path("/notification-rules").Handler(HandlerFunc(ctxt, ph.NotificationRules)).
		Name("project-notification-rules")

	root.Methods("PUT").Handler(HandlerFunc(ctxt, ph.CreateUpdate)).
		Name("project-update")
	root.Methods("POST").Handler(HandlerFunc(ctxt, ph.CreateUpdate)).
		Name("project-create")

	// Item
	item := root.PathPrefix("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, ph.Show)).
		Name("project-show")
	item.Methods("DELETE").Handler(HandlerFunc(ctxt, ph.Archive)).
		Name("project-archive")

}

type projectHandler struct {
}

func (self projectHandler) Show(ctxt RequestContext) (err error) {

	store := stores.NewDbProjectStore(ctxt.Tx())

	proj, err := store.FindByUuid(ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(proj); !allowed {
		return err
	}

	writeAsJson(ctxt, proj)
	return err
}

func (self projectHandler) CreateUpdate(ctxt RequestContext) (err error) {

	store := stores.NewDbProjectStore(ctxt.Tx())
	orgStore := stores.NewDbOrganizationStore(ctxt.Tx())

	params, err := ReadProjectParams(ctxt.R().Body)
	if err != nil {
		return err
	}

	isNew := true
	var proj *domain.Project

	if uuidhelper.IsValid(params.Uuid) {
		proj, err = store.FindByUuid(params.Uuid)
		_, notFound := err.(*domain.NotFoundError)
		if err != nil && !notFound {
			return err
		}
		isNew = (proj == nil)
	}

	if isNew {
		proj = &domain.Project{}
	}

	copyProjParams(params, proj)

	if err := domain.ValidateProject(proj); err != nil {
		return err
	}

	var uuid string

	if isNew {

		if allowed, err := ctxt.Auth().CanCreate(proj); !allowed {
			return err
		}

		if allowed, err := limits.CanCreate(proj); !allowed {
			return err
		}

		uuid, err = store.Create(proj)
		if err != nil {
			return err
		}

		org, err := orgStore.FindByUuid(proj.OrganizationUuid)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("can't lookup org for project uuid %s", proj.Uuid))
		}

		if exceeded, err := NewLimitsFromContext(ctxt).OrganizationLimitsExceeded(org); exceeded && err == nil {
			return ErrLimitsExceeded
		}

		stencils := stores.NewDbStencilStore(ctxt.SecretKeyValueStore(), ctxt.Tx(), ctxt)
		configuration := stencils.ToConfiguration()
		configuration.ProjectUuid = proj.Uuid
		if err := stencil.NewProjectDefaults(configuration).Create(); err != nil {
			return err
		}
	} else {

		if allowed, err := ctxt.Auth().CanUpdate(proj); !allowed {
			return err
		}

		err = store.Update(proj)
		uuid = proj.Uuid

	}

	if err != nil {
		return err
	}

	proj, err = store.FindByUuid(uuid)
	if err != nil {
		return err
	}

	if isNew {
		ctxt.EnqueueActivity(activities.ProjectCreated(proj), nil)
		ctxt.W().Header().Add("Location", urlForSubject(ctxt.R(), proj))
		ctxt.W().WriteHeader(http.StatusCreated)
	}

	writeAsJson(ctxt, proj)

	return err

}

func (self projectHandler) Archive(ctxt RequestContext) (err error) {

	projUuid := ctxt.PathParameter("uuid")
	store := stores.NewDbProjectStore(ctxt.Tx())

	proj, err := store.FindByUuid(projUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanArchive(proj); !allowed {
		return err
	}

	err = store.ArchiveByUuid(projUuid)
	if err != nil {
		return err
	}
	ctxt.EnqueueActivity(activities.ProjectDeleted(proj), nil)
	ctxt.W().WriteHeader(http.StatusNoContent)

	return err

}

func (self projectHandler) Repositories(ctxt RequestContext) (err error) {

	projUuid := ctxt.PathParameter("uuid")
	repoStore := stores.NewDbRepositoryStore(ctxt.Tx())

	if err := self.requireProject(ctxt); err != nil {
		return err
	}

	repos, err := repoStore.FindAllByProjectUuid(projUuid)
	if err != nil {
		return err
	}

	var interfaceRepos []interface{} = make([]interface{}, 0, len(repos))
	for _, r := range repos {
		if allowed, _ := ctxt.Auth().CanRead(r); allowed {
			interfaceRepos = append(interfaceRepos, r)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceRepos),
		Count:      len(interfaceRepos),
		Collection: interfaceRepos,
	})

	return err

}

func (self projectHandler) Jobs(ctxt RequestContext) (err error) {

	projUuid := ctxt.PathParameter("uuid")

	jobStore := stores.NewDbJobStore(ctxt.Tx())
	operationStore := stores.NewDbOperationStore(ctxt.Tx())

	if err := self.requireProject(ctxt); err != nil {
		return err
	}

	jobs, err := jobStore.FindAllByProjectUuid(projUuid)
	if err != nil {
		return err
	}

	var interfaceJobs []interface{} = make([]interface{}, 0, len(jobs))
	for _, j := range jobs {
		if strings.HasPrefix(j.Name, "urn:harrow:default-job:") {
			continue
		}

		if allowed, _ := ctxt.Auth().CanRead(j); allowed {
			if err := j.FindRecentOperations(operationStore); err != nil {
				return err
			}
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

func (self projectHandler) Environments(ctxt RequestContext) (err error) {

	envStore := stores.NewDbEnvironmentStore(ctxt.Tx())

	if err := self.requireProject(ctxt); err != nil {
		return err
	}

	envs, err := envStore.FindAllByProjectUuid(ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}

	var interfaceEnvs []interface{} = make([]interface{}, 0, len(envs))
	for _, e := range envs {
		if strings.HasPrefix(e.Name, "urn:harrow:default-environment:") {
			continue
		}
		if allowed, _ := ctxt.Auth().CanRead(e); allowed {
			interfaceEnvs = append(interfaceEnvs, e)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceEnvs),
		Count:      len(interfaceEnvs),
		Collection: interfaceEnvs,
	})

	return err

}

func (self projectHandler) Tasks(ctxt RequestContext) (err error) {

	if err := self.requireProject(ctxt); err != nil {
		return err
	}

	taskStore := stores.NewDbTaskStore(ctxt.Tx())

	tasks, err := taskStore.FindAllByProjectUuid(ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}

	var interfaceTasks []interface{} = make([]interface{}, 0, len(tasks))
	for _, t := range tasks {
		if strings.HasPrefix(t.Name, "urn:harrow:default-task:") {
			continue
		}

		if allowed, _ := ctxt.Auth().CanRead(t); allowed {
			interfaceTasks = append(interfaceTasks, t)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceTasks),
		Count:      len(interfaceTasks),
		Collection: interfaceTasks,
	})

	return err

}

func (self projectHandler) Operations(ctxt RequestContext) (err error) {

	operationStore := stores.NewDbOperationStore(ctxt.Tx())
	jobStore := stores.NewDbJobStore(ctxt.Tx())

	if err := self.requireProject(ctxt); err != nil {
		return err
	}

	operations, err := operationStore.FindRecentByProjectUuid(ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}
	operations, err = embedJobs(jobStore, ctxt.Auth(), operations)
	if err != nil {
		return err
	}
	var interfaceOperations []interface{} = make([]interface{}, 0, len(operations))
	for _, o := range operations {
		if allowed, _ := ctxt.Auth().CanRead(o); allowed {
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

func (self projectHandler) Memberships(ctxt RequestContext) (err error) {

	projUuid := ctxt.PathParameter("uuid")

	if err := self.requireProject(ctxt); err != nil {
		return err
	}
	membershipStore := stores.NewDbProjectMembershipStore(ctxt.Tx())
	memberships, err := membershipStore.FindAllByProjectUuid(projUuid)
	if err != nil {
		return err
	}

	interfaceProjectMemberships := []interface{}{}
	for _, membership := range memberships {
		if allowed, _ := ctxt.Auth().CanRead(membership); allowed {
			interfaceProjectMemberships = append(interfaceProjectMemberships, membership)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceProjectMemberships),
		Count:      len(interfaceProjectMemberships),
		Collection: interfaceProjectMemberships,
	})

	return nil
}

func (self projectHandler) Members(ctxt RequestContext) error {

	projUuid := ctxt.PathParameter("uuid")

	if err := self.requireProject(ctxt); err != nil {
		return err
	}

	memberStore := stores.NewDbProjectMemberStore(ctxt.Tx())
	members, err := memberStore.FindAllByProjectUuid(projUuid)
	if err != nil {
		return err
	}
	interfaceMembers := []interface{}{}
	for _, member := range members {
		if allowed, _ := ctxt.Auth().CanRead(member); allowed {
			interfaceMembers = append(interfaceMembers, member)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceMembers),
		Count:      len(interfaceMembers),
		Collection: interfaceMembers,
	})

	return nil
}

func (self projectHandler) Schedules(ctxt RequestContext) (err error) {

	uuid := ctxt.PathParameter("uuid")

	projectStore := stores.NewDbProjectStore(ctxt.Tx())
	scheduleStore := stores.NewDbScheduleStore(ctxt.Tx())

	project, err := projectStore.FindByUuid(uuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(project); !allowed {
		return err
	}

	schedules, err := scheduleStore.FindAllFutureSchedulesByProjectUuid(uuid)
	if err != nil {
		return err
	}

	var interfaceSchedules []interface{} = make([]interface{}, 0, len(schedules))
	for _, schedule := range schedules {
		if allowed, _ := ctxt.Auth().CanRead(schedule); allowed {
			interfaceSchedules = append(interfaceSchedules, schedule)
		}
	}
	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceSchedules),
		Count:      len(interfaceSchedules),
		Collection: interfaceSchedules,
	})

	return err
}

func (self projectHandler) ScheduledExecutions(ctxt RequestContext) (err error) {

	uuid := ctxt.PathParameter("uuid")

	params, err := NewScheduledExecutionsParams(ctxt.R().URL.Query())
	if err != nil {
		return err
	}

	projectStore := stores.NewDbProjectStore(ctxt.Tx())
	jobStore := stores.NewDbJobStore(ctxt.Tx())
	scheduleStore := stores.NewDbScheduleStore(ctxt.Tx())

	project, err := projectStore.FindByUuid(uuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(project); !allowed {
		return err
	}

	jobs, err := jobStore.FindAllByProjectUuid(project.Uuid)
	if err != nil {
		return err
	}

	schedules := make([]*domain.Schedule, 0)
	for _, job := range jobs {
		if allowed, _ := ctxt.Auth().CanRead(job); !allowed {
			// ignore error, just skip this job
			continue
		}

		ss, err := scheduleStore.FindAllByJobUuid(job.Uuid)
		if err != nil {
			return err
		}
		schedules = append(schedules, ss...)
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

func (self projectHandler) Webhooks(ctxt RequestContext) error {

	projUuid := ctxt.PathParameter("uuid")
	if err := self.requireProject(ctxt); err != nil {
		return err
	}

	webhookStore := stores.NewDbWebhookStore(ctxt.Tx())
	webhooks, err := webhookStore.FindByProjectUuid(projUuid)
	if err != nil {
		return err
	}

	result := []interface{}{}
	for _, webhook := range webhooks {
		if allowed, _ := ctxt.Auth().CanRead(webhook); allowed {
			result = append(result, webhook)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(result),
		Count:      len(result),
		Collection: result,
	})

	return nil
}

func (self projectHandler) ScriptCards(ctxt RequestContext) error {

	projectUuid := ctxt.PathParameter("uuid")

	if err := self.requireProject(ctxt); err != nil {
		return err
	}

	scriptCards := stores.NewDbScriptCardStore(ctxt.Tx())
	result, err := scriptCards.FindAllByProjectUuid(projectUuid)
	if err != nil {
		return err
	}

	items := []interface{}{}
	for _, card := range result {
		if allowed, err := ctxt.Auth().CanRead(card); allowed {
			items = append(items, card)
		} else {
			return err
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Count:      len(items),
		Total:      len(items),
		Collection: items,
	})

	return nil
}

func (self projectHandler) GitTriggers(ctxt RequestContext) error {

	projUuid := ctxt.PathParameter("uuid")
	projectStore := stores.NewDbProjectStore(ctxt.Tx())
	proj, err := projectStore.FindByUuid(projUuid)
	if err != nil {
		return err
	}

	gitTriggerStore := stores.NewDbGitTriggerStore(ctxt.Tx())
	gitTriggers, err := gitTriggerStore.FindByProjectUuid(proj.Uuid)
	if err != nil {
		return err
	}

	result := []interface{}{}
	for _, gitTrigger := range gitTriggers {
		if allowed, _ := ctxt.Auth().CanRead(gitTrigger); allowed {
			result = append(result, gitTrigger)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(result),
		Count:      len(result),
		Collection: result,
	})

	return nil
}

func (self projectHandler) JobNotifiers(ctxt RequestContext) error {

	projectUuid := ctxt.PathParameter("uuid")
	jobStore := stores.NewDbJobStore(ctxt.Tx())
	allJobNotifiers := stores.NewDbJobNotifierStore(ctxt.Tx())
	notificationRuleStore := stores.NewDbNotificationRuleStore(ctxt.Tx())
	jobNotifiers, err := allJobNotifiers.FindAllByProjectUuid(projectUuid)
	if err != nil {
		return err
	}

	result := []interface{}{}
	for _, jobNotifier := range jobNotifiers {
		if allowed, _ := ctxt.Auth().CanRead(jobNotifier); allowed {
			triggeringRules, err := notificationRuleStore.FindAllByNotifierUuid(jobNotifier.Uuid)
			if err != nil {
				ctxt.Log().Error().Msgf("failed to fetch triggering rules for job notifier %q", jobNotifier.Uuid)
			} else {
				for _, rule := range triggeringRules {
					if rule.JobUuid != nil {
						job, err := jobStore.FindByUuid(*rule.JobUuid)
						if err != nil {
							ctxt.Log().Error().Msgf("failed to fetch triggering job %q", rule.JobUuid)
						} else {
							rule.Embed("job", job)
						}
					}
					jobNotifier.Embedded()["notificationRules"] = append(jobNotifier.Embedded()["notificationRules"], rule)
				}
			}
			result = append(result, jobNotifier)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(result),
		Count:      len(result),
		Collection: result,
	})

	return nil
}

func (self projectHandler) SlackNotifiers(ctxt RequestContext) error {

	projectUuid := ctxt.PathParameter("uuid")
	slackNotifiers := stores.NewDbSlackNotifierStore(ctxt.Tx())

	result, err := slackNotifiers.FindByProjectUuid(projectUuid)
	if err != nil {
		return err
	}

	items := []interface{}{}
	for _, notifier := range result {
		if allowed, _ := ctxt.Auth().CanRead(notifier); allowed {
			items = append(items, notifier)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(items),
		Count:      len(items),
		Collection: items,
	})

	return nil
}

func (self projectHandler) EmailNotifiers(ctxt RequestContext) error {

	projectUuid := ctxt.PathParameter("uuid")
	emailNotifiers := stores.NewDbEmailNotifierStore(ctxt.Tx())

	result, err := emailNotifiers.FindAllByProjectUuid(projectUuid)
	if err != nil {
		return err
	}

	items := []interface{}{}
	for _, notifier := range result {
		if allowed, _ := ctxt.Auth().CanRead(notifier); allowed {
			items = append(items, notifier)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(items),
		Count:      len(items),
		Collection: items,
	})

	return nil
}

func (self projectHandler) AddMember(ctxt RequestContext) error {

	params := struct {
		Subject struct {
			UserUuid       string
			MembershipType string
		}
	}{}

	projectUuid := ctxt.PathParameter("uuid")
	projects := stores.NewDbProjectStore(ctxt.Tx())
	project, err := projects.FindByUuid(projectUuid)
	if err != nil {
		return err
	}

	organizationMemberships := stores.NewDbOrganizationMembershipStore(ctxt.Tx())
	projectMemberships := stores.NewDbProjectMembershipStore(ctxt.Tx())
	user := ctxt.User()
	if user == nil {
		return ErrSessionNotValid
	}

	organizationMembership, err := organizationMemberships.FindByOrganizationAndUserUuids(project.OrganizationUuid, user.Uuid)
	if err != nil && !domain.IsNotFound(err) {
		return err
	}

	projectMembership, err := projectMemberships.FindByUserAndProjectUuid(user.Uuid, projectUuid)
	if err != nil && !domain.IsNotFound(err) {
		return err
	}

	projectMember := domain.NewProjectMember(user, project, projectMembership, organizationMembership)
	if projectMember == nil {
		return domain.NewValidationError("projectUuid", "current_user_not_a_member")
	}

	if err := json.NewDecoder(ctxt.R().Body).Decode(&params); err != nil {
		return err
	}

	if domain.MembershipTypeHierarchyLevel(params.Subject.MembershipType) == 0 {
		return domain.NewValidationError("membershipType", "invalid")
	}

	c := ctxt.Config()
	otherUser, err := stores.NewDbUserStore(ctxt.Tx(), &c).FindByUuid(params.Subject.UserUuid)
	if err != nil {
		return domain.NewValidationError("userUuid", err.Error())
	}

	if _, err := projectMember.AddMember(otherUser.Uuid, params.Subject.MembershipType, projectMemberships); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.UserAddedToProject(otherUser, project), &user.Uuid)

	return nil
}

func (self projectHandler) Leave(ctxt RequestContext) error {

	projectUuid := ctxt.PathParameter("uuid")
	store := stores.NewDbProjectMembershipStore(ctxt.Tx())
	user := ctxt.User()
	if user == nil {
		return ErrSessionNotValid
	}
	membership, err := store.FindByUserAndProjectUuid(user.Uuid, projectUuid)
	if err != nil {
		return err
	}

	if err := store.ArchiveByUuid(membership.Uuid); err != nil {
		return err
	}

	return nil
}

func (self projectHandler) NotificationRules(ctxt RequestContext) error {

	projectUuid := ctxt.PathParameter("uuid")
	notificationRules := stores.NewDbNotificationRuleStore(ctxt.Tx())
	projectStore := stores.NewDbProjectStore(ctxt.Tx())
	project, err := projectStore.FindByUuid(projectUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(project); !allowed {
		return err
	}
	notifiers := stores.NewDbNotifierStore(ctxt.Tx())

	rules, err := notificationRules.FindByProjectUuid(projectUuid)
	result := []interface{}{}
	for _, rule := range rules {
		if allowed, _ := ctxt.Auth().CanRead(rule); allowed {
			notifier, err := notifiers.FindByUuidAndType(rule.NotifierUuid, rule.NotifierType)
			if err == nil {
				subject, ok := notifier.(domain.Subject)
				if ok {
					rule.Embed("notifier", subject)
				}
			} else {
				ctxt.Log().Warn().Msgf("fetching notifier(%s=%s) for project %s: %s", rule.NotifierType, rule.NotifierUuid, projectUuid, err)
			}
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

func (self projectHandler) Card(ctxt RequestContext) error {

	projectUuid := ctxt.PathParameter("uuid")
	projectCards := stores.NewDbProjectCardStore(ctxt.Tx())
	card, err := projectCards.FindByProjectUuid(projectUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(card); !allowed {
		return err
	}

	writeAsJson(ctxt, card)
	return nil
}

func embedJobs(jobStore *stores.DbJobStore, az authz.Service, operations []*domain.Operation) ([]*domain.Operation, error) {
	operationJobUuids := make(map[string]string)

	for _, o := range operations {
		if o.JobUuid != nil {
			operationJobUuids[o.Uuid] = *o.JobUuid
		}
	}

	jobUuids := make([]string, 0, len(operationJobUuids))
	for _, jobUuid := range operationJobUuids {
		jobUuids = append(jobUuids, jobUuid)
	}
	jobs, err := jobStore.FindAllByUuids(jobUuids)
	if err != nil {
		return nil, err
	}
	for _, j := range jobs {
		if allowed, _ := az.CanRead(j); allowed {
			for _, o := range operations {
				if operationJobUuids[o.Uuid] == j.Uuid {
					o.Embed("job", j)
				}
			}
		}
	}
	return operations, nil
}

func (self projectHandler) requireProject(ctxt RequestContext) error {
	projectUuid := ctxt.PathParameter("uuid")
	project, err := stores.NewDbProjectStore(ctxt.Tx()).FindByUuid(projectUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(project); !allowed {
		return err
	}

	return nil
}
