package stores

import (
	"fmt"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"

	"database/sql"

	"github.com/jmoiron/sqlx"
)

func NewDbOperationStore(tx *sqlx.Tx) *DbOperationStore {
	return &DbOperationStore{tx: tx}
}

type DbOperationStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func (store *DbOperationStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *DbOperationStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store DbOperationStore) FindByUuid(uuid string) (*domain.Operation, error) {

	var op *domain.Operation = &domain.Operation{Uuid: uuid}

	var q string = `SELECT * FROM operations WHERE uuid = $1`

	err := store.tx.Get(op, q, op.Uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return op, nil

}

func (store DbOperationStore) Create(op *domain.Operation) (string, error) {

	if len(op.Uuid) == 0 {
		op.Uuid = uuidhelper.MustNewV4()
	}

	var q string = `
		INSERT INTO operations (
                        uuid,
                        type,
                        workspace_base_image_uuid,
                        repository_uuid,
                        time_limit,
                        job_uuid,
                        notifier_uuid,
                        notifier_type,
                        repository_refs,
                        git_logs,
                        parameters
		) VALUES (
                        :uuid,
                        :type,
                        :workspace_base_image_uuid,
                        :repository_uuid,
                        :time_limit,
                        :job_uuid,
                        :notifier_uuid,
                        :notifier_type,
                        :repository_refs,
                        :git_logs,
                        :parameters
		) RETURNING uuid
	`
	rows, err := store.tx.NamedQuery(q, op)

	if err != nil {
		return "", resolveErrType(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", resolveErrType(rows.Err())
	}

	if err = rows.Scan(&op.Uuid); err != nil {
		return "", resolveErrType(err)
	}

	return op.Uuid, nil

}

func (store DbOperationStore) FindAllByScheduleUuid(scheduleUuid string) ([]*domain.Operation, error) {

	result := []*domain.Operation{}
	q := `
SELECT
  operations.*
FROM
  operations
WHERE
  parameters->>'scheduleUuid' = $1::text
;`
	if err := store.tx.Select(&result, q, scheduleUuid); err != nil {
		return nil, resolveErrType(err)
	}

	return result, nil
}

func (store DbOperationStore) FindAll() ([]*domain.Operation, error) {

	var operations []*domain.Operation = []*domain.Operation{}

	var q string = `
		SELECT
			operations.*
		FROM operations
	`

	err := store.tx.Select(&operations, q)
	if err != nil {
		return nil, err
	}

	return operations, nil

}

func (store DbOperationStore) FindAllByRepositoryUuid(repoUuid string) ([]*domain.Operation, error) {

	var operations []*domain.Operation = []*domain.Operation{}

	var q string = `
		SELECT
			operations.*
		FROM operations
		LEFT OUTER JOIN repositories
			ON repository_uuid = repositories.uuid
		WHERE
			repository_uuid = $1
			AND repositories.archived_at IS NULL
	`

	err := store.tx.Select(&operations, q, repoUuid)
	if err != nil {
		return nil, err
	}

	return operations, nil

}

const findAllOperations = -1

func (store DbOperationStore) FindAllByProjectUuid(projectUuid string) ([]*domain.Operation, error) {

	return store.findAllByProjectUuidWithLimit(projectUuid, findAllOperations, "")
}

func (store DbOperationStore) FindRecentByProjectUuid(projectUuid string) ([]*domain.Operation, error) {

	return store.findAllByProjectUuidWithLimit(projectUuid, 20, "AND jp.name NOT LIKE 'urn:harrow%'")
}

func (store DbOperationStore) FindMostRecentUserOperationByProjectUuid(projectUuid string) (*domain.Operation, error) {

	recent, err := store.findAllByProjectUuidWithLimit(projectUuid, 1, "AND jp.name NOT LIKE 'urn:harrow%' AND o.type::text like 'job.%'")
	if err != nil {
		return nil, err
	}
	if len(recent) == 0 {
		return nil, sql.ErrNoRows
	}
	return recent[0], nil
}

func (store DbOperationStore) findAllByProjectUuidWithLimit(projectUuid string, limit int, where string) ([]*domain.Operation, error) {

	var operations []*domain.Operation = []*domain.Operation{}
	if where == "" {
		where = "AND true"
	}

	q := `
		SELECT
			o.*
		FROM operations AS o
		LEFT OUTER JOIN jobs_projects AS jp
			ON jp.uuid = o.job_uuid
		WHERE
		jp.project_uuid = $1
		AND jp.archived_at IS NULL
                %s
    ORDER BY o.created_at DESC
                %s
;
	`
	if limit == findAllOperations {
		q = fmt.Sprintf(q, where, "")
	} else {
		q = fmt.Sprintf(q, where, fmt.Sprintf("LIMIT %d", limit))
	}

	err := store.tx.Select(&operations, q, projectUuid)
	if err != nil {
		return nil, err
	}

	return operations, nil

}

func (store DbOperationStore) FindAllByJobUuid(jobUuid string) ([]*domain.Operation, error) {

	var operations []*domain.Operation = []*domain.Operation{}

	var q string = `
		SELECT
			operations.*
		FROM operations
		LEFT OUTER JOIN jobs_projects AS jp
			ON jp.uuid = operations.job_uuid
		WHERE
			job_uuid = $1
			AND jp.archived_at is NULL
	`

	err := store.tx.Select(&operations, q, jobUuid)
	if err != nil {
		return nil, err
	}

	return operations, nil

}

func (store DbOperationStore) FindRecentByJobUuid(n int, jobUuid string) ([]*domain.Operation, error) {

	var operations []*domain.Operation = []*domain.Operation{}

	var q string = `
		SELECT
			operations.*
		FROM operations
		LEFT OUTER JOIN jobs_projects AS jp
			ON jp.uuid = operations.job_uuid
		WHERE
			job_uuid = $1
                AND     jp.archived_at is NULL
                ORDER BY created_at DESC
                LIMIT $2
	`

	err := store.tx.Select(&operations, q, jobUuid, n)
	if err != nil {
		return nil, err
	}

	return operations, nil

}

func (store *DbOperationStore) updateTimestamp(operationUuid, columnName string) error {
	q := fmt.Sprintf(`
	      UPDATE operations
	      SET %s = NOW() at time zone 'utc'
	      WHERE uuid = $1::uuid
	`, columnName)
	fmt.Println("Running", q, operationUuid)
	r, err := store.tx.Exec(q, operationUuid)
	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil
}

func (store *DbOperationStore) updateColumn(operationUuid, columnName string, value interface{}) error {
	q := fmt.Sprintf(`
	      UPDATE operations
	      SET %s = $2
	      WHERE uuid = $1::uuid
	`, columnName)

	r, err := store.tx.Exec(q, operationUuid, value)
	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil
}

func (store *DbOperationStore) MarkAsCanceled(operationUuid string) error {

	return store.updateTimestamp(operationUuid, "canceled_at")
}

func (store *DbOperationStore) MarkAsStarted(operationUuid string) error {

	return store.updateTimestamp(operationUuid, "started_at")
}

func (store *DbOperationStore) MarkAsFinished(operationUuid string) error {

	return store.updateTimestamp(operationUuid, "finished_at")
}

func (store *DbOperationStore) MarkAsTimedOut(operationUuid string) error {

	return store.updateTimestamp(operationUuid, "timed_out_at")
}

func (store *DbOperationStore) MarkAsFailed(operationUuid string) error {

	return store.updateTimestamp(operationUuid, "failed_at")
}

func (store *DbOperationStore) MarkFatalError(operationUuid, fatalErrorMessage string) error {

	return store.updateColumn(operationUuid, "fatal_error", fatalErrorMessage)
}

func (store *DbOperationStore) MarkExitStatus(operationUuid string, exitStatus int) error {

	return store.updateColumn(operationUuid, "exit_status", exitStatus)
}

func (store *DbOperationStore) MarkRepositoryCheckouts(operationUuid string, checkouts *domain.RepositoryCheckouts) error {

	return store.updateColumn(operationUuid, "repository_refs", checkouts)
}

func (store *DbOperationStore) MarkGitLogs(operationUuid string, logs *domain.GitLogs) error {

	return store.updateColumn(operationUuid, "git_logs", logs)
}

func (store *DbOperationStore) MarkStatusLogs(operationUuid string, logs *domain.StatusLogs) error {

	return store.updateColumn(operationUuid, "status_logs", logs)
}

func (store *DbOperationStore) FindPreviousOperation(currentOperationUuid string) (*domain.Operation, error) {

	q := `SELECT * FROM operations
        WHERE finished_at IS NOT NULL
        AND   archived_at IS NULL
        AND   job_uuid = (select job_uuid from operations where uuid = $1)
        AND   uuid <> $1
        ORDER BY finished_at DESC
        LIMIT 1;`

	result := &domain.Operation{}
	err := store.tx.Get(result, q, currentOperationUuid)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return result, err
}
