package stores

import (
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"

	"database/sql"

	"github.com/jmoiron/sqlx"
)

func NewDbTargetStore(tx *sqlx.Tx) *DbTargetStore {
	return &DbTargetStore{tx: tx}
}

type DbTargetStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func (self DbTargetStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self DbTargetStore) SetLogger(l logger.Logger) {
	self.log = l
}

func (store DbTargetStore) FindByUuid(uuid string) (*domain.Target, error) {

	var target *domain.Target = &domain.Target{Uuid: uuid}

	var q string = `SELECT * FROM targets WHERE uuid = $1 AND archived_at IS NULL`
	err := store.tx.Get(target, q, target.Uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return target, nil

}

func (store DbTargetStore) Create(target *domain.Target) (string, error) {

	if len(target.Uuid) == 0 {
		target.Uuid = uuidhelper.MustNewV4()
	}

	var q string = `INSERT INTO targets (uuid, type, project_uuid, environment_uuid, url, identifier, secret) VALUES (:uuid, :type, :project_uuid, :environment_uuid, :url, :identifier, :secret) RETURNING uuid`
	rows, err := store.tx.NamedQuery(q, target)

	if err != nil {
		return "", resolveErrType(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", resolveErrType(rows.Err())
	}

	if err = rows.Scan(&target.Uuid); err != nil {
		return "", resolveErrType(err)
	}

	return target.Uuid, nil
}

func (store DbTargetStore) FindAllByProjectUuid(projUuid string) ([]*domain.Target, error) {

	var targets []*domain.Target = []*domain.Target{}

	var q string = `SELECT * FROM targets WHERE project_uuid = $1 AND archived_at IS NULL`

	err := store.tx.Select(&targets, q, projUuid)
	if err != nil {
		return nil, err
	}

	return targets, nil

}

func (store DbTargetStore) FindAllByEnvironmentUuid(envUuid string) ([]*domain.Target, error) {

	var targets []*domain.Target = []*domain.Target{}

	var q string = `SELECT * FROM targets WHERE environment_uuid = $1 AND archived_at IS NULL`

	err := store.tx.Select(&targets, q, envUuid)
	if err != nil {
		return nil, err
	}

	return targets, nil

}
