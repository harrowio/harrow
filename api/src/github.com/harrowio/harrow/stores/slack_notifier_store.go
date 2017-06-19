package stores

import (
	"database/sql"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"
	"github.com/jmoiron/sqlx"
)

type DbSlackNotifierStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func NewDbSlackNotifierStore(tx *sqlx.Tx) *DbSlackNotifierStore {
	return &DbSlackNotifierStore{tx: tx}
}

func (store *DbSlackNotifierStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *DbSlackNotifierStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store *DbSlackNotifierStore) Create(subject *domain.SlackNotifier) (string, error) {

	if subject.Uuid == "" {
		subject.Uuid = uuidhelper.MustNewV4()
	}

	q := `INSERT INTO slack_notifiers (
	  uuid,
	  webhook_url,
	  url_host,
	  project_uuid,
          name
	) VALUES (
	  :uuid,
	  :webhook_url,
	  :url_host,
	  :project_uuid,
          :name
	);`

	_, err := store.tx.NamedExec(q, subject)
	if err != nil {
		return "", resolveErrType(err)
	}

	return subject.Uuid, nil
}

func (store *DbSlackNotifierStore) Update(subject *domain.SlackNotifier) error {

	if !uuidhelper.IsValid(subject.Uuid) {
		return &domain.NotFoundError{}
	}

	q := `UPDATE slack_notifiers SET
	  name = :name,
	  url_host = :url_host,
	  project_uuid = :project_uuid,
	  webhook_url = :webhook_url
	WHERE uuid = :uuid AND archived_at IS NULL`

	r, err := store.tx.NamedExec(q, subject)
	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return &domain.NotFoundError{}
	}

	return nil
}

func (store *DbSlackNotifierStore) FindByUuid(uuid string) (*domain.SlackNotifier, error) {

	result := &domain.SlackNotifier{}
	q := `SELECT * FROM slack_notifiers WHERE uuid = $1 AND archived_at IS NULL`
	err := store.tx.Get(result, q, uuid)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}

	return result, err
}

func (store *DbSlackNotifierStore) FindByProjectUuid(uuid string) ([]*domain.SlackNotifier, error) {

	result := []*domain.SlackNotifier{}
	q := `SELECT * FROM slack_notifiers WHERE project_uuid = $1 AND archived_at IS NULL`
	err := store.tx.Select(&result, q, uuid)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}

	return result, err
}

func (store *DbSlackNotifierStore) ArchiveByUuid(uuid string) error {

	q := `UPDATE slack_notifiers SET archived_at = NOW() AT TIME ZONE 'UTC' WHERE uuid = $1`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return &domain.NotFoundError{}
	}

	return nil
}
