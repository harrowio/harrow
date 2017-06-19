package projector

import (
	"strings"
	"time"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
)

type Organization struct {
	Uuid         string
	ProjectCards map[string]*ProjectCard
}

type ProjectCard struct {
	OrganizationUuid string     `json:"organizationUuid"`
	ProjectUuid      string     `json:"projectUuid"`
	ProjectName      string     `json:"projectName"`
	LastTaskUuid     string     `json:"lastTaskUuid"`
	LastTaskName     string     `json:"lastTaskName"`
	LastTaskRunAt    *time.Time `json:"lastTaskRunAt"`
	LastTaskStatus   string     `json:"lastTaskStatus"`

	lastActivity time.Time
}

type ProjectCards struct {
}

func NewProjectCards() *ProjectCards {
	return &ProjectCards{}
}

func (self *ProjectCards) SubscribedTo() []string {
	return []string{"project.created", "project.deleted", "operation.started", "operation.failed", "operation.failed-fatally", "operation.succeeded"}
}

func (self *ProjectCards) IndexCard(tx IndexTransaction, card *ProjectCard) error {
	organization := &Organization{}
	if err := tx.Get(card.OrganizationUuid, organization); err != nil {
		organization = &Organization{
			Uuid:         card.OrganizationUuid,
			ProjectCards: map[string]*ProjectCard{},
		}
	}

	organization.ProjectCards[card.ProjectUuid] = card

	if err := tx.Put(organization.Uuid, organization); err != nil {
		return err
	}

	return tx.Put("project-card:"+card.ProjectUuid, card)
}

func (self *ProjectCards) HandleActivity(tx IndexTransaction, activity *domain.Activity) error {
	switch activity.Name {
	case "project.created":
		project, ok := activity.Payload.(*domain.Project)
		if !ok {
			return NewTypeError(activity, project)
		}

		card := &ProjectCard{
			OrganizationUuid: project.OrganizationUuid,
			ProjectName:      project.Name,
			ProjectUuid:      project.Uuid,
			lastActivity:     activity.OccurredOn,
		}

		return self.IndexCard(tx, card)
	case "project.deleted":
		project, ok := activity.Payload.(*domain.Project)
		if !ok {
			return NewTypeError(activity, project)
		}
		organization := &Organization{}
		if err := tx.Get(project.OrganizationUuid, organization); err != nil {
			return err
		}
		delete(organization.ProjectCards, project.Uuid)
		return tx.Put(organization.Uuid, organization)
	case "operation.started":
		operation, ok := activity.Payload.(*domain.Operation)
		if !ok {
			return NewTypeError(activity, operation)
		}
		if operation.JobUuid == nil {
			return nil
		}

		projectedOperation := &Operation{}
		if err := tx.Get(operation.Uuid, projectedOperation); err != nil {
			return err
		}

		projectUuid := projectedOperation.ProjectUuid(tx)
		card := &ProjectCard{}
		if err := tx.Get("project-card:"+projectUuid, card); err != nil {
			return err
		}

		job := &Job{}
		if err := tx.Get(*operation.JobUuid, job); err != nil {
			return err
		}

		if strings.HasPrefix(job.Name(), "urn:") {
			return nil
		}

		card.LastTaskUuid = operation.Uuid
		card.LastTaskRunAt = operation.StartedAt
		card.LastTaskStatus = operation.Status()
		card.LastTaskName = job.Name()
		card.lastActivity = activity.OccurredOn

		return self.IndexCard(tx, card)
	case "operation.canceled-by-user":
		canceled, ok := activity.Payload.(*activities.OperationCanceledByUserPayload)
		if !ok {
			return NewTypeError(activity, canceled)
		}

		projectedOperation := &Operation{}
		if err := tx.Get(canceled.Uuid, projectedOperation); err != nil {
			return err
		}

		projectUuid := projectedOperation.ProjectUuid(tx)
		card := &ProjectCard{}
		if err := tx.Get("project-card:"+projectUuid, card); err != nil {
			return err
		}

		if activity.OccurredOn.Before(card.lastActivity) {
			return nil
		}
		card.lastActivity = activity.OccurredOn
		return self.IndexCard(tx, card)
	case "operation.failed",
		"operation.failed-fatally",
		"operation.succeeded":
		operation, ok := activity.Payload.(*domain.Operation)
		if !ok {
			return NewTypeError(activity, operation)
		}

		if operation.JobUuid == nil {
			return nil
		}

		projectedOperation := &Operation{}
		if err := tx.Get(operation.Uuid, projectedOperation); err != nil {
			return err
		}
		projectUuid := projectedOperation.ProjectUuid(tx)
		card := &ProjectCard{}
		if err := tx.Get("project-card:"+projectUuid, card); err != nil {
			return err
		}

		if activity.OccurredOn.Before(card.lastActivity) {
			return nil
		}
		card.lastActivity = activity.OccurredOn

		if operation.Uuid == card.LastTaskUuid {
			switch activity.Name {
			case "operation.failed":
				card.LastTaskStatus = "failure"
			case "operation.succeeded":
				card.LastTaskStatus = "success"
			case "operation.failed-fatally":
				card.LastTaskStatus = "fatal"
			}
		}

		return self.IndexCard(tx, card)
	}
	return nil
}
