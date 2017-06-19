package stores

import (
	"database/sql"
	"time"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/jmoiron/sqlx"
)

type DbProjectCardStore struct {
	tx           *sqlx.Tx
	log          logger.Logger
	projects     *DbProjectStore
	tasks        *DbTaskStore
	environments *DbEnvironmentStore
	operations   *DbOperationStore
	activities   *DbActivityStore
}

func NewDbProjectCardStore(tx *sqlx.Tx) *DbProjectCardStore {
	return &DbProjectCardStore{
		tx:           tx,
		projects:     NewDbProjectStore(tx),
		tasks:        NewDbTaskStore(tx),
		environments: NewDbEnvironmentStore(tx),
		operations:   NewDbOperationStore(tx),
		activities:   NewDbActivityStore(tx),
	}
}

func (self *DbProjectCardStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *DbProjectCardStore) SetLogger(l logger.Logger) {
	self.log = l
}

func (self *DbProjectCardStore) FindByProjectUuid(projectUuid string) (*domain.ProjectCard, error) {
	project, err := self.projects.FindByUuid(projectUuid)
	if err != nil {
		return nil, err
	}
	mostRecentOperation, err := self.operations.FindMostRecentUserOperationByProjectUuid(projectUuid)
	if err == sql.ErrNoRows {
		return &domain.ProjectCard{
			ProjectName:        project.Name,
			ProjectUuid:        project.Uuid,
			LastActivitySeenAt: time.Now(),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	if mostRecentOperation.JobUuid == nil {
		return nil, domain.NewValidationError("operation.jobUuid", "null")
	}

	task, err := self.tasks.FindByJobUuid(*mostRecentOperation.JobUuid)
	if err != nil {
		return nil, err
	}

	environment, err := self.environments.FindByJobUuid(*mostRecentOperation.JobUuid)
	if err != nil {
		return nil, err
	}

	return domain.NewProjectCard(project, environment, task, mostRecentOperation), nil
}

func (self *DbProjectCardStore) FindAllByOrganizationUuid(organizationUuid string) ([]*domain.ProjectCard, error) {
	projects, err := self.projects.FindAllByOrganizationUuid(organizationUuid)
	if err != nil {
		return nil, err
	}

	result := []*domain.ProjectCard{}
	for _, project := range projects {
		card, err := self.FindByProjectUuid(project.Uuid)
		if err != nil {
			return nil, err
		}
		if card != nil {
			result = append(result, card)
		}
	}

	return result, nil
}
