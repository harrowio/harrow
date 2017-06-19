package harrowArchivist

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/harrowio/harrow/activities"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"

	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

type Command func() error
type ActivityBus interface {
	Publish(activity *domain.Activity) error
}

type Errors []error

func NewErrors() Errors {
	return Errors{}
}

func (self Errors) Error() string {
	result := []string{}

	for _, err := range self {
		result = append(result, err.Error())
	}

	return fmt.Sprintf("errors:[%s]\n", strings.Join(result, ","))
}

func (self Errors) ToError() error {
	if len(self) == 0 {
		return nil
	} else {
		return self
	}
}

type DryRunActivityBus struct{}

func NewDryRunActivityBus() *DryRunActivityBus { return &DryRunActivityBus{} }

func (self *DryRunActivityBus) Publish(activity *domain.Activity) error {
	return nil
}

const ProgramName = "harrow-archivist"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {
	dryRun := flag.Bool("n", false, "dry run: do not make actual changes")
	flag.Parse()
	c := config.GetConfig()
	db, err := c.DB()
	if err != nil {
		log.Fatal().Err(err)
	}

	var activityBus ActivityBus
	if *dryRun {
		activityBus = NewDryRunActivityBus()
	} else {
		amqpActivities := activity.NewAMQPTransport(c.AmqpConnectionString(), "harrow-archivist")
		if err != nil {
			log.Fatal().Err(err)
		}
		defer amqpActivities.Close()
		activityBus = amqpActivities
	}

	command := DocumentOrganizationCreations(db, c, activityBus)
	if err := command(); err != nil {
		log.Info().Msgf("DocumentOrganizationCreations: %s", err)
	}

	command = DocumentProjectCreations(db, c, activityBus)
	if err := command(); err != nil {
		log.Info().Msgf("DocumentProjectCreations: %s", err)
	}

	command = DocumentEnvironmentCreations(db, c, activityBus)
	if err := command(); err != nil {
		log.Info().Msgf("DocumentEnvironmentCreations: %s", err)
	}

	command = DocumentTaskCreations(db, c, activityBus)
	if err := command(); err != nil {
		log.Info().Msgf("DocumentTaskCreations: %s", err)
	}

	command = DocumentJobCreations(db, c, activityBus)
	if err := command(); err != nil {
		log.Info().Msgf("DocumentJobCreations: %s", err)
	}

	command = DocumentOrganizationDeletions(db, c, activityBus)
	if err := command(); err != nil {
		log.Info().Msgf("DocumentOrganizationDeletions: %s", err)
	}

	command = DocumentProjectDeletions(db, c, activityBus)
	if err := command(); err != nil {
		log.Info().Msgf("DocumentProjectDeletions: %s", err)
	}

}

func DocumentOrganizationCreations(db *sqlx.DB, c *config.Config, activityBus ActivityBus) Command {
	return func() error {
		tx, err := db.Beginx()
		if err != nil {
			return err
		}
		allOrganizations := stores.NewDbOrganizationStore(tx)
		organizations, err := allOrganizations.FindAll()
		if err != nil {
			return err
		}

		errors := NewErrors()
		allActivities := stores.NewDbActivityStore(tx)
		for _, organization := range organizations {
			command := NewDocumentCreation(organization, allActivities)
			activity, err := command.Activity(activities.OrganizationCreated(organization, domain.FreePlan))
			if err != nil {
				errors = append(errors, err)
			}
			if activity == nil {
				continue
			}
			log.Info().Msgf("%s uuid=%s", activity.Name, organization.Uuid)
			if err := activityBus.Publish(activity); err != nil {
				errors = append(errors, err)
			}
		}

		return errors.ToError()
	}
}

func DocumentProjectCreations(db *sqlx.DB, c *config.Config, activityBus ActivityBus) Command {
	return func() error {
		tx, err := db.Beginx()
		if err != nil {
			return err
		}
		allProjects := stores.NewDbProjectStore(tx)
		projects, err := allProjects.FindAllIncludingArchived()
		if err != nil {
			return err
		}

		errors := NewErrors()
		allActivities := stores.NewDbActivityStore(tx)
		for _, project := range projects {
			command := NewDocumentCreation(project, allActivities)
			activity, err := command.Activity(activities.ProjectCreated(project))
			if err != nil {
				errors = append(errors, err)
			}
			if activity == nil {
				continue
			}
			log.Info().Msgf("%s uuid=%s", activity.Name, project.Uuid)
			if err := activityBus.Publish(activity); err != nil {
				errors = append(errors, err)
			}
		}

		return errors.ToError()
	}
}

func DocumentTaskCreations(db *sqlx.DB, c *config.Config, activityBus ActivityBus) Command {
	return func() error {
		tx, err := db.Beginx()
		if err != nil {
			return err
		}
		allTasks := stores.NewDbTaskStore(tx)
		tasks, err := allTasks.FindAll()
		if err != nil {
			return err
		}

		errors := NewErrors()
		allActivities := stores.NewDbActivityStore(tx)
		for _, task := range tasks {
			command := NewDocumentCreation(task, allActivities)
			activity, err := command.Activity(activities.TaskAdded(task))
			if err != nil {
				errors = append(errors, err)
			}
			if activity == nil {
				continue
			}
			log.Info().Msgf("%s uuid=%s", activity.Name, task.Uuid)
			if err := activityBus.Publish(activity); err != nil {
				errors = append(errors, err)
			}
		}

		return errors.ToError()
	}
}

func DocumentJobCreations(db *sqlx.DB, c *config.Config, activityBus ActivityBus) Command {
	return func() error {
		tx, err := db.Beginx()
		if err != nil {
			return err
		}
		allJobs := stores.NewDbJobStore(tx)
		jobs, err := allJobs.FindAll()
		if err != nil {
			return err
		}

		errors := NewErrors()
		allActivities := stores.NewDbActivityStore(tx)
		for _, job := range jobs {
			command := NewDocumentCreation(job, allActivities)
			activity, err := command.Activity(activities.JobAdded(job))
			if err != nil {
				errors = append(errors, err)
			}
			if activity == nil {
				continue
			}
			log.Info().Msgf("%s uuid=%s", activity.Name, job.Uuid)
			if err := activityBus.Publish(activity); err != nil {
				errors = append(errors, err)
			}
		}

		return errors.ToError()
	}
}

func DocumentEnvironmentCreations(db *sqlx.DB, c *config.Config, activityBus ActivityBus) Command {
	return func() error {
		tx, err := db.Beginx()
		if err != nil {
			return err
		}
		allEnvironments := stores.NewDbEnvironmentStore(tx)
		allProjects := stores.NewDbProjectStore(tx)
		environments, err := allEnvironments.FindAll()
		if err != nil {
			return err
		}

		errors := NewErrors()
		allActivities := stores.NewDbActivityStore(tx)
		for _, environment := range environments {
			project, err := allProjects.FindByUuidWithDeleted(environment.ProjectUuid)
			if err != nil {
				log.Info().Msgf("orphaned environment %s", environment.Uuid)
				continue
			}
			environment.CreatedAt = project.CreatedAt

			command := NewDocumentCreation(environment, allActivities)
			activity, err := command.Activity(activities.EnvironmentAdded(environment))
			if err != nil {
				errors = append(errors, err)
			}
			if activity == nil {
				continue
			}
			log.Info().Msgf("%s uuid=%s", activity.Name, environment.Uuid)
			if err := activityBus.Publish(activity); err != nil {
				errors = append(errors, err)
			}
		}

		return errors.ToError()
	}
}

func DocumentOrganizationDeletions(db *sqlx.DB, c *config.Config, activityBus ActivityBus) Command {
	return func() error {
		tx, err := db.Beginx()
		if err != nil {
			return err
		}
		allOrganizations := stores.NewDbOrganizationStore(tx)
		organizations, err := allOrganizations.FindAllArchived()
		if err != nil {
			return err
		}

		errors := NewErrors()
		allActivities := stores.NewDbActivityStore(tx)
		for _, organization := range organizations {
			command := NewDocumentDeletion(organization, allActivities)
			activity, err := command.Activity(activities.OrganizationArchived(organization))
			if err != nil {
				errors = append(errors, err)
			}
			if activity == nil {
				continue
			}
			log.Info().Msgf("%s uuid=%s", activity.Name, organization.Uuid)
			if err := activityBus.Publish(activity); err != nil {
				errors = append(errors, err)
			}
		}

		return errors.ToError()
	}
}

func DocumentProjectDeletions(db *sqlx.DB, c *config.Config, activityBus ActivityBus) Command {
	return func() error {
		tx, err := db.Beginx()
		if err != nil {
			return err
		}
		allProjects := stores.NewDbProjectStore(tx)
		projects, err := allProjects.FindAllArchived()
		if err != nil {
			return err
		}

		errors := NewErrors()
		allActivities := stores.NewDbActivityStore(tx)
		for _, project := range projects {
			command := NewDocumentDeletion(project, allActivities)
			activity, err := command.Activity(activities.ProjectDeleted(project))
			if err != nil {
				errors = append(errors, err)
			}
			if activity == nil {
				continue
			}
			log.Info().Msgf("%s uuid=%s", activity.Name, project.Uuid)
			if err := activityBus.Publish(activity); err != nil {
				errors = append(errors, err)
			}
		}

		return errors.ToError()
	}
}
