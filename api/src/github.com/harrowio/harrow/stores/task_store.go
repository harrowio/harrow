package stores

import (
	"fmt"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"

	"database/sql"

	"github.com/jmoiron/sqlx"
)

func NewDbTaskStore(tx *sqlx.Tx) *DbTaskStore {
	return &DbTaskStore{tx: tx}
}

type DbTaskStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func (store DbTaskStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store DbTaskStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store DbTaskStore) FindByUuid(uuid string) (*domain.Task, error) {

	var task *domain.Task = &domain.Task{Uuid: uuid}

	var q string = `SELECT * FROM tasks WHERE uuid = $1 AND archived_at IS NULL`
	err := store.tx.Get(task, q, task.Uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return task, nil

}

func (store DbTaskStore) FindByName(name string) (*domain.Task, error) {

	var task *domain.Task = &domain.Task{}

	var q string = `SELECT * FROM tasks WHERE name = $1 AND archived_at IS NULL LIMIT 1`
	err := store.tx.Get(task, q, name)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return task, nil

}

func (store DbTaskStore) Create(task *domain.Task) (string, error) {

	if len(task.Uuid) == 0 {
		task.Uuid = uuidhelper.MustNewV4()
	}

	var q string = `INSERT INTO tasks (uuid, name, body, project_uuid) VALUES (:uuid, :name, :body, :project_uuid) RETURNING uuid;`
	rows, err := store.tx.NamedQuery(q, task)

	if err != nil {
		return "", resolveErrType(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", resolveErrType(rows.Err())
	}

	if err = rows.Scan(&task.Uuid); err != nil {
		return "", resolveErrType(err)
	}

	return task.Uuid, nil

}

func (store DbTaskStore) Update(task *domain.Task) error {

	if len(task.Uuid) == 0 {
		return new(domain.NotFoundError)
	}

	var q string = `UPDATE tasks SET (name, body, project_uuid) = (:name, :body, :project_uuid) WHERE uuid = :uuid AND archived_at IS NULL;`
	result, err := store.tx.NamedExec(q, task)

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

func (store DbTaskStore) ArchiveByUuid(uuid string) error {

	var q string = `UPDATE tasks SET archived_at = NOW() AT TIME ZONE 'UTC' WHERE uuid = $1;`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil

}

func (store DbTaskStore) FindAll() ([]*domain.Task, error) {

	var tasks []*domain.Task = []*domain.Task{}

	var q string = `SELECT * FROM tasks`

	err := store.tx.Select(&tasks, q)
	if err != nil {
		return nil, err
	}

	return tasks, nil

}

func (store DbTaskStore) FindAllByProjectUuid(projUuid string) ([]*domain.Task, error) {

	var tasks []*domain.Task = []*domain.Task{}

	var q string = `SELECT * FROM tasks WHERE project_uuid = $1 AND archived_at IS NULL`

	err := store.tx.Select(&tasks, q, projUuid)
	if err != nil {
		return nil, err
	}

	return tasks, nil

}

func (store *DbTaskStore) FindByJobUuid(jobUuid string) (*domain.Task, error) {

	task := domain.Task{}

	q := `SELECT t.*
	      FROM   tasks t
	      JOIN   jobs  j ON j.task_uuid = t.uuid
	      WHERE  j.uuid = $1 AND j.archived_at IS NULL;`

	err := store.tx.Get(&task, q, jobUuid)
	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}
	if err != nil {
		return nil, err
	}

	return &task, nil
}
