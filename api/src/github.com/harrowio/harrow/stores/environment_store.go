package stores

import (
	"fmt"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"

	"database/sql"

	"github.com/jmoiron/sqlx"
)

func NewDbEnvironmentStore(tx *sqlx.Tx) *DbEnvironmentStore {
	return &DbEnvironmentStore{tx: tx}
}

type DbEnvironmentStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func (self *DbEnvironmentStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *DbEnvironmentStore) SetLogger(l logger.Logger) {
	self.log = l
}

func (store DbEnvironmentStore) FindByUuid(uuid string) (*domain.Environment, error) {

	var env *domain.Environment = &domain.Environment{Uuid: uuid}

	var q string = `SELECT * FROM environments WHERE uuid = $1 AND archived_at IS NULL`
	err := store.tx.Get(env, q, env.Uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	env.PruneVariables()

	return env, nil

}

func (store DbEnvironmentStore) FindByName(name string) (*domain.Environment, error) {

	var env *domain.Environment = &domain.Environment{}

	var q string = `SELECT * FROM environments WHERE name = $1 AND archived_at IS NULL LIMIT 1`
	err := store.tx.Get(env, q, name)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	env.PruneVariables()

	return env, nil

}

func (store DbEnvironmentStore) Create(env *domain.Environment) (string, error) {

	if len(env.Uuid) == 0 {
		env.Uuid = uuidhelper.MustNewV4()
	}

	env.PruneVariables()

	var q string = `INSERT INTO environments (uuid, name, project_uuid, variables, is_default) VALUES (:uuid, :name, :project_uuid, :variables, :is_default) RETURNING uuid`
	rows, err := store.tx.NamedQuery(q, env)

	if err != nil {
		return "", resolveErrType(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", resolveErrType(rows.Err())
	}

	if err = rows.Scan(&env.Uuid); err != nil {
		return "", resolveErrType(err)
	}

	return env.Uuid, nil
}

func (store DbEnvironmentStore) Update(env *domain.Environment) error {

	if len(env.Uuid) == 0 {
		return new(domain.NotFoundError)
	}

	env.PruneVariables()
	var q string = `UPDATE environments SET (name, project_uuid, variables, is_default) = (:name, :project_uuid, :variables, :is_default) WHERE uuid = :uuid AND archived_at IS NULL;`
	result, err := store.tx.NamedExec(q, env)

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

func (store DbEnvironmentStore) ArchiveByUuid(uuid string) error {

	var q string = `UPDATE environments SET archived_at = NOW() AT TIME ZONE 'UTC' WHERE uuid = $1;`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil

}

func (store DbEnvironmentStore) FindAll() ([]*domain.Environment, error) {

	q := `SELECT * FROM environments`
	environments := []*domain.Environment{}
	if err := store.tx.Select(&environments, q); err != nil {
		return nil, err
	}
	return environments, nil
}

func (store DbEnvironmentStore) FindAllByProjectUuid(projUuid string) ([]*domain.Environment, error) {

	var envs []*domain.Environment = []*domain.Environment{}

	var q string = `SELECT * FROM environments WHERE project_uuid = $1 AND archived_at IS NULL`

	err := store.tx.Select(&envs, q, projUuid)
	if err != nil {
		return nil, err
	}

	return envs, nil

}

func (store DbEnvironmentStore) FindByJobUuid(jobUuid string) (*domain.Environment, error) {

	env := domain.Environment{}

	err := store.tx.Get(
		&env,
		`SELECT  e.*
		   FROM  environments e
		   JOIN  jobs j ON j.environment_uuid = e.uuid
		   WHERE j.archived_at IS NULL
		     AND j.uuid = $1`,
		jobUuid,
	)
	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return &env, nil
}
