package stores

import (
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"

	"database/sql"

	"github.com/jmoiron/sqlx"
)

func NewDbScheduleStore(tx *sqlx.Tx) *DbScheduleStore {
	return &DbScheduleStore{tx: tx}
}

type DbScheduleStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func (store DbScheduleStore) FindByUuid(uuid string) (*domain.Schedule, error) {

	var schedule *domain.Schedule = &domain.Schedule{Uuid: uuid}

	var q string = `SELECT * FROM schedules WHERE uuid = $1 AND archived_at IS NULL`
	err := store.tx.Get(schedule, q, schedule.Uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return schedule, nil

}

func (store DbScheduleStore) FindInclArchivedByUuid(uuid string) (*domain.Schedule, error) {

	var schedule *domain.Schedule = &domain.Schedule{Uuid: uuid}

	var q string = `SELECT * FROM schedules WHERE uuid = $1`
	err := store.tx.Get(schedule, q, schedule.Uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return schedule, nil
}

func (store DbScheduleStore) Create(schedule *domain.Schedule) (string, error) {

	if err := schedule.Validate(); err != nil {
		return "", err
	}

	if len(schedule.Uuid) == 0 {
		schedule.Uuid = uuidhelper.MustNewV4()
	}

	var q string = `INSERT INTO schedules (uuid, user_uuid, job_uuid, run_once_at, cronspec, timespec, description, parameters)
	VALUES (:uuid, :user_uuid, :job_uuid, :run_once_at, :cronspec, :timespec, :description, :parameters) RETURNING uuid;`
	rows, err := store.tx.NamedQuery(q, schedule)

	if err != nil {
		return "", resolveErrType(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", resolveErrType(rows.Err())
	}

	if err = rows.Scan(&schedule.Uuid); err != nil {
		return "", resolveErrType(err)
	}

	return schedule.Uuid, nil

}

func (store DbScheduleStore) Update(schedule *domain.Schedule) error {

	q := `
UPDATE schedules
SET    timespec = :timespec,
       cronspec = :cronspec,
       description = :description
WHERE  uuid = :uuid
`
	if _, err := store.tx.NamedExec(q, schedule); err != nil {
		return resolveErrType(err)
	}

	return nil
}

func (store DbScheduleStore) ArchiveByUuid(uuid string) error {

	q := `UPDATE schedules SET archived_at = NOW() AT TIME ZONE 'UTC' WHERE uuid = $1`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return &domain.NotFoundError{}
	}

	return nil

}

func (store DbScheduleStore) DeleteByUuid(uuid string) error {

	var q string = `DELETE FROM schedules WHERE uuid = $1;`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil

}

func (store DbScheduleStore) FindAllFutureSchedulesByProjectUuid(projectUuid string) ([]*domain.Schedule, error) {

	result := []*domain.Schedule{}
	q := `
SELECT
  schedules.*
FROM
  schedules,
  jobs_projects
WHERE
  schedules.job_uuid = jobs_projects.uuid
AND
  jobs_projects.project_uuid = $1
AND
  schedules.disabled is null
AND
  schedules.archived_at is null
AND
  (
   schedules.timespec is null
  OR
   schedules.timespec != 'now'
  OR
   schedules.cronspec is not null
  )
`

	if err := store.tx.Select(&result, q, projectUuid); err != nil {
		if err != sql.ErrNoRows {
			return nil, resolveErrType(err)
		}
	}

	return result, nil
}

func (store DbScheduleStore) FindAllByJobUuid(jobUuid string) ([]*domain.Schedule, error) {

	var schedules []*domain.Schedule = []*domain.Schedule{}

	var q string = `SELECT * FROM schedules WHERE job_uuid = $1 AND archived_at IS NULL`

	err := store.tx.Select(&schedules, q, jobUuid)
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

func (store DbScheduleStore) FindAllRecurringByJobUuid(jobUuid string) ([]*domain.Schedule, error) {

	var schedules []*domain.Schedule = []*domain.Schedule{}

	var q string = `SELECT * FROM schedules WHERE job_uuid = $1 AND disabled IS NULL AND (cronspec IS NOT NULL OR timespec != 'now') AND archived_at IS NULL`

	err := store.tx.Select(&schedules, q, jobUuid)
	if err != nil {
		return nil, err
	}

	return schedules, nil
}

func (store DbScheduleStore) DisableSchedule(scheduleUuid, disabled string, disabledBecause *string) error {

	r, err := store.tx.Exec(`UPDATE schedules SET disabled = $1, disabled_because = $2 WHERE uuid = $3`, disabled, disabledBecause, scheduleUuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil
}
