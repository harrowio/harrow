package limits

import (
	"fmt"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
)

var (
	activityHandlers = map[string]func(*Limits, *domain.Activity) error{
		"repository.added":               handleRepositoryAdded,
		"repository.detected-as-private": handleRepositoryDetectedAsPrivate,
		"repository.detected-as-public":  handleRepositoryDetectedAsPublic,
		"operation.succeeded":            handleOperationSucceeded,
		"user.joined-project":            handleUserJoinedProject,
		"user.added-to-project":          handleUserJoinedProject,
		"user.left-project":              handleUserRemovedFromProject,
		"user.removed-from-project":      handleUserRemovedFromProject,
		"project.created":                handleProjectCreated,
		"project.deleted":                handleProjectDeleted,
		"organization.created":           handleOrganizationCreated,
	}
)

func handleRepositoryAdded(self *Limits, activity *domain.Activity) error {
	repository, ok := activity.Payload.(*domain.Repository)
	if !ok {
		return fmt.Errorf("invalid payload: want %T, got %T", repository, activity.Payload)
	}

	if self.hasProject(repository.ProjectUuid) {
		self.state.PublicRepositories[repository.Uuid] = true
	}

	return nil
}

func handleRepositoryDetectedAsPrivate(self *Limits, activity *domain.Activity) error {
	repository, ok := activity.Payload.(*domain.Repository)
	if !ok {
		return fmt.Errorf("invalid payload: want %T, got %T", repository, activity.Payload)
	}

	if self.hasProject(repository.ProjectUuid) {
		self.state.PrivateRepositories[repository.Uuid] = true
		delete(self.state.PublicRepositories, repository.Uuid)
	}
	return nil
}

func handleRepositoryDetectedAsPublic(self *Limits, activity *domain.Activity) error {
	repository, ok := activity.Payload.(*domain.Repository)
	if !ok {
		return fmt.Errorf("invalid payload: want %T, got %T", repository, activity.Payload)
	}

	if self.hasProject(repository.ProjectUuid) {
		self.state.PublicRepositories[repository.Uuid] = true
		delete(self.state.PrivateRepositories, repository.Uuid)
	}
	return nil
}

func handleOperationSucceeded(self *Limits, activity *domain.Activity) error {
	projectUuid := activity.ProjectUuid()
	if projectUuid == "" {
		return nil
	}

	if self.hasProject(projectUuid) {
		key := activity.OccurredOn.Format("2006-01")
		self.state.NumberOfOperationsByYearAndMonth[key]++
	}

	return nil
}

func handleUserJoinedProject(self *Limits, activity *domain.Activity) error {
	payload, ok := activity.Payload.(*activities.UserProjectPayload)
	if !ok {
		return fmt.Errorf("invalid payload: want %T, got %T", payload, activity.Payload)
	}
	if payload.Project.OrganizationUuid == self.state.OrganizationUuid {
		self.addProjectForUser(payload.User.Uuid, payload.Project.Uuid)
	}
	return nil
}

func handleUserRemovedFromProject(self *Limits, activity *domain.Activity) error {
	payload, ok := activity.Payload.(*activities.UserProjectPayload)
	if !ok {
		return fmt.Errorf("invalid payload: want %T, got %T", payload, activity.Payload)
	}

	delete(self.state.ProjectsForUser[payload.User.Uuid], payload.Project.Uuid)

	if len(self.state.ProjectsForUser[payload.User.Uuid]) == 0 {
		delete(self.state.ProjectsForUser, payload.User.Uuid)
	}
	return nil
}

func handleProjectDeleted(self *Limits, activity *domain.Activity) error {
	project, ok := activity.Payload.(*domain.Project)
	if !ok {
		return fmt.Errorf("invalid payload: want %T, got %T", project, activity.Payload)
	}
	if project.OrganizationUuid == self.state.OrganizationUuid {
		self.state.ProjectCount--
	}
	return nil
}

func handleProjectCreated(self *Limits, activity *domain.Activity) error {
	project, ok := activity.Payload.(*domain.Project)
	if !ok {
		return fmt.Errorf("invalid payload: want %T, got %T", project, activity.Payload)
	}

	if project.OrganizationUuid == self.state.OrganizationUuid {
		self.addProject(project)
		self.state.ProjectCount++
	}
	return nil
}

func handleOrganizationCreated(self *Limits, activity *domain.Activity) error {
	organizationAndPlan, ok := activity.Payload.(*activities.OrganizationWithBillingPlan)
	if !ok {
		return fmt.Errorf("invalid payload: want %T, got %T", organizationAndPlan, activity.Payload)
	}

	organization := organizationAndPlan.Organization
	if organization == nil {
		// old, broken event; can be ignored
		return nil
	}

	if self.state.OrganizationUuid != organization.Uuid {
		return nil
	}

	self.state.OrganizationCreatedAt = activity.OccurredOn
	return nil
}
