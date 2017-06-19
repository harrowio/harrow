package stores

import (
	"database/sql"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"
	"github.com/jmoiron/sqlx"
)

type DbEmailNotifierStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func NewDbEmailNotifierStore(tx *sqlx.Tx) *DbEmailNotifierStore {
	return &DbEmailNotifierStore{tx: tx}
}

func (store *DbEmailNotifierStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *DbEmailNotifierStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store *DbEmailNotifierStore) Create(subject *domain.EmailNotifier) (string, error) {

	if subject.Uuid == "" {
		subject.Uuid = uuidhelper.MustNewV4()
	}

	q := `INSERT INTO email_notifiers (
	  uuid,
          recipient,
          project_uuid,
          url_host
	) VALUES (
	  :uuid,
          :recipient,
          :project_uuid,
          :url_host
	);`

	_, err := store.tx.NamedExec(q, subject)
	if err != nil {
		return "", resolveErrType(err)
	}

	return subject.Uuid, nil
}

func (store *DbEmailNotifierStore) Update(subject *domain.EmailNotifier) error {

	if !uuidhelper.IsValid(subject.Uuid) {
		return &domain.NotFoundError{}
	}

	q := `UPDATE email_notifiers SET
	  uuid = :uuid,
          recipient = :recipient,
          project_uuid = :project_uuid,
          url_host = :url_host
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

func (store *DbEmailNotifierStore) FindByUuid(uuid string) (*domain.EmailNotifier, error) {

	result := &domain.EmailNotifier{}
	q := `SELECT * FROM email_notifiers WHERE uuid = $1 AND archived_at IS NULL`
	err := store.tx.Get(result, q, uuid)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}

	return result, err
}

func (store *DbEmailNotifierStore) FindByRecipient(uuid string) (*domain.EmailNotifier, error) {

	result := &domain.EmailNotifier{}
	q := `SELECT * FROM email_notifiers WHERE recipient = $1 AND archived_at IS NULL`
	err := store.tx.Get(result, q, uuid)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}

	return result, err
}

func (store *DbEmailNotifierStore) FindAllByProjectUuid(projectUuid string) ([]*domain.EmailNotifier, error) {

	result := []*domain.EmailNotifier{}
	q := `SELECT * FROM email_notifiers WHERE project_uuid = $1 AND archived_at IS NULL;`
	err := store.tx.Select(&result, q, projectUuid)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}

	return result, err
}

func (store *DbEmailNotifierStore) ArchiveByUuid(uuid string) error {

	q := `UPDATE email_notifiers SET archived_at = NOW() AT TIME ZONE 'UTC' WHERE uuid = $1`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return &domain.NotFoundError{}
	}

	return nil
}
