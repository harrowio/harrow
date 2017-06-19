package stores

import (
	"fmt"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"

	"database/sql"

	"github.com/jmoiron/sqlx"
)

func NewDbJobStore(tx *sqlx.Tx) *DbJobStore {
	return &DbJobStore{tx: tx}
}

type DbJobStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func (self DbJobStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self DbJobStore) SetLogger(l logger.Logger) {
	self.log = l
}

func (store DbJobStore) FindByUuid(uuid string) (*domain.Job, error) {

	job := &domain.Job{Uuid: uuid}
	q := `SELECT * FROM jobs_projects WHERE uuid = $1 AND archived_at IS NULL`
	err := store.tx.Get(job, q, job.Uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return job, nil

}

func (store DbJobStore) FindByName(name string) (*domain.Job, error) {

	job := &domain.Job{}
	q := `SELECT * FROM jobs_projects WHERE name = $1 AND archived_at IS NULL`
	err := store.tx.Get(job, q, name)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return job, nil

}

func (store DbJobStore) FindAllByUuids(uuids []string) ([]*domain.Job, error) {

	jobs := make([]*domain.Job, 0)
	if len(uuids) == 0 {
		return jobs, nil
	}

	query, args, err := sqlx.In(`SELECT * FROM jobs_projects WHERE uuid IN(?) AND archived_at IS NULL;`, uuids)
	if err != nil {
		return nil, err
	}
	query = store.tx.Rebind(query)

	err = store.tx.Select(&jobs, query, args...)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return jobs, nil

}

func (store DbJobStore) FindAll() ([]*domain.Job, error) {

	jobs := []*domain.Job{}

	q := `SELECT * FROM jobs_projects`

	err := store.tx.Select(&jobs, q)
	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return jobs, nil

}

func (store DbJobStore) Create(job *domain.Job) (string, error) {

	if len(job.Uuid) == 0 {
		job.Uuid = uuidhelper.MustNewV4()
	}

	q := `INSERT INTO jobs (uuid, name, description, task_uuid, environment_uuid) VALUES (:uuid, :name, :description, :task_uuid, :environment_uuid) RETURNING uuid;`
	rows, err := store.tx.NamedQuery(q, job)

	if err != nil {
		return "", resolveErrType(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", resolveErrType(rows.Err())
	}

	if err = rows.Scan(&job.Uuid); err != nil {
		return "", resolveErrType(err)
	}

	return job.Uuid, nil

}

func (store DbJobStore) Update(job *domain.Job) error {

	if len(job.Uuid) == 0 {
		return new(domain.NotFoundError)
	}

	var q string = `UPDATE jobs SET (name, description, task_uuid, environment_uuid) = (:name, :description, :task_uuid, :environment_uuid) WHERE uuid = :uuid AND archived_at IS NULL;`
	result, err := store.tx.NamedExec(q, job)

	if err != nil {
		return resolveErrType(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return fmt.Errorf("Wanted to update 1 row but updated %d rows", rowsAffected)
	}
	return nil
}

func (store DbJobStore) ArchiveByUuid(uuid string) error {

	var q string = `UPDATE jobs SET archived_at = NOW() AT TIME ZONE 'UTC' WHERE uuid = $1;`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil

}

func (store DbJobStore) FindAllByEnvironmentUuid(envUuid string) ([]*domain.Job, error) {

	jobs := []*domain.Job{}

	q := `SELECT * FROM jobs_projects WHERE environment_uuid = $1 AND archived_at IS NULL`

	err := store.tx.Select(&jobs, q, envUuid)
	if err != nil {
		return nil, err
	}

	return jobs, nil

}

func (store DbJobStore) FindAllByTaskUuid(taskUuid string) ([]*domain.Job, error) {

	var jobs []*domain.Job = []*domain.Job{}

	q := `SELECT * FROM jobs_projects WHERE task_uuid = $1 AND archived_at IS NULL`

	err := store.tx.Select(&jobs, q, taskUuid)
	if err != nil {
		return nil, err
	}

	return jobs, nil

}

func (store DbJobStore) FindAllByProjectUuid(projUuid string) ([]*domain.Job, error) {

	jobs := []*domain.Job{}

	q := `SELECT * FROM jobs_projects WHERE project_uuid = $1 AND archived_at IS NULL`

	err := store.tx.Select(&jobs, q, projUuid)
	if err != nil {
		return nil, err
	}

	return jobs, nil

}

func (store DbJobStore) FindAllByProjectUuids(projectUuids []string) ([]*domain.Job, error) {

	jobs := []*domain.Job{}
	uuids, err := pqConcatUuids(projectUuids)
	if err != nil {
		return nil, err
	}

	if len(projectUuids) == 0 {
		return jobs, nil
	}

	q := `SELECT * FROM jobs_projects WHERE project_uuid IN(%s) AND archived_at IS NULL`

	if err := store.tx.Select(&jobs, fmt.Sprintf(q, uuids)); err != nil {
		return nil, err
	}

	return jobs, nil
}
