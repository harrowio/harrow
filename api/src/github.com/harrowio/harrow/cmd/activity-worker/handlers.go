package activityWorker

import (
	"os"

	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
)

var (
	hostname = "localhost"
)

func init() {
	host, err := os.Hostname()
	if err == nil {
		hostname = host
	} else {
		log.Error().Msgf("Failed to determine hostname: %s", err)
		log.Info().Msgf("Using localhost as hostname")
	}
}

type BelongsToProject interface {
	FindProject(store domain.ProjectStore) (*domain.Project, error)
}

type BelongsToJob interface {
	FindJob(store domain.JobStore) (*domain.Job, error)
}

// listProjectMembersTx fetches all project members and uses them as
// the activity's audience if the payload satisfies the
// BelongsToProject interface.
//
// This method exists in parallel to ListProjectMembers because in the
// tests the various stores need to run within the same transaction
// that set up the world state.
func listProjectMembersTx(tx *sqlx.Tx) func(msg activity.Message) {
	return func(msg activity.Message) {
		projectStore := stores.NewDbProjectStore(tx)
		projectMemberStore := stores.NewDbProjectMemberStore(tx)

		activity := msg.Activity()

		belongsToProject, ok := activity.Payload.(BelongsToProject)
		if !ok {
			log.Info().Msgf("listProjectMembersTx.belongsToProject: %T does not implement BelongsToProject", activity.Payload)
			return
		}

		project, err := belongsToProject.FindProject(projectStore)
		if err != nil {
			log.Info().Msgf("listProjectMembersTx.belongsToProject.FindProject: %s", err)
			return
		}

		projectMembers, err := projectMemberStore.FindAllByProjectUuid(project.Uuid)
		if err != nil {
			log.Info().Msgf("listProjectMembersTx.projectMemberStore.FindAllByProjectUuid: %s", err)
			return
		}

		audience := []string{}
		for _, member := range projectMembers {
			audience = append(audience, member.Uuid)

		}
		activity.SetAudience(audience)
	}
}

// ListProjectMembers fetches all project members and uses them as the
// activity's audience if the payload satisfies the BelongsToProject
// interface.
func ListProjectMembers(db *sqlx.DB) func(msg activity.Message) {
	return func(msg activity.Message) {
		tx := db.MustBegin()
		defer tx.Rollback()
		listProjectMembersTx(tx)(msg)
	}
}

// markProjectUuidTx adds the uuid of the project this activity
// belongs to the extra fields of the activity.
func markProjectUuidTx(tx *sqlx.Tx) func(msg activity.Message) {
	return func(msg activity.Message) {
		projectStore := stores.NewDbProjectStore(tx)

		activity := msg.Activity()

		belongsToProject, ok := activity.Payload.(BelongsToProject)
		if !ok {
			log.Info().Msgf("markProjectUuidTx.belongsToProject: %T does not implement BelongsToProject", activity.Payload)
			return
		}

		project, err := belongsToProject.FindProject(projectStore)
		if err != nil {
			log.Info().Msgf("markProjectUuidTx.belongsToProject.FindProject: %s", err)
			return
		}

		activity.SetProjectUuid(project.Uuid)
	}
}

func MarkProjectUuid(db *sqlx.DB) func(msg activity.Message) {
	return func(msg activity.Message) {
		tx := db.MustBegin()
		defer tx.Rollback()
		markProjectUuidTx(tx)(msg)
	}
}

// markJobUuidTx adds the uuid of the job this activity
// belongs to the extra fields of the activity.
func markJobUuidTx(tx *sqlx.Tx) func(msg activity.Message) {
	return func(msg activity.Message) {
		jobStore := stores.NewDbJobStore(tx)

		activity := msg.Activity()

		belongsToJob, ok := activity.Payload.(BelongsToJob)
		if !ok {
			log.Info().Msgf("markJobUuidTx.belongsToJob: %T does not implement BelongsToJob", activity.Payload)
			return
		}

		job, err := belongsToJob.FindJob(jobStore)
		if err != nil {
			log.Info().Msgf("markJobUuidTx.belongsToJob.FindJob: %s", err)
			return
		}

		activity.SetJobUuid(job.Uuid)
	}
}

func MarkJobUuid(db *sqlx.DB) func(msg activity.Message) {
	return func(msg activity.Message) {
		tx := db.MustBegin()
		defer tx.Rollback()
		markJobUuidTx(tx)(msg)
	}
}
