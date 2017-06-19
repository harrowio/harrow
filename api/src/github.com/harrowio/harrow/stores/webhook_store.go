package stores

import (
	"database/sql"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"
	"github.com/jmoiron/sqlx"
)

type DbWebhookStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func NewDbWebhookStore(tx *sqlx.Tx) *DbWebhookStore {
	return &DbWebhookStore{tx: tx}
}

func (store *DbWebhookStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *DbWebhookStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store *DbWebhookStore) Create(webhook *domain.Webhook) (string, error) {

	if webhook.Uuid == "" {
		webhook.Uuid = uuidhelper.MustNewV4()
	}

	q := `INSERT INTO webhooks (
	  uuid,
	  slug,
	  name,
	  creator_uuid,
	  job_uuid,
	  project_uuid
	) VALUES (
	  :uuid,
	  :slug,
	  :name,
	  :creator_uuid,
		:job_uuid,
	  :project_uuid
	);`

	_, err := store.tx.NamedExec(q, webhook)
	if err != nil {
		return "", resolveErrType(err)
	}

	return webhook.Uuid, nil
}

func (store *DbWebhookStore) Update(webhook *domain.Webhook) error {

	if !uuidhelper.IsValid(webhook.Uuid) {
		return &domain.NotFoundError{}
	}

	q := `UPDATE webhooks SET
	  slug = :slug,
	  name = :name,
	  creator_uuid = :creator_uuid,
	  job_uuid = :job_uuid,
	  project_uuid = :project_uuid
	WHERE uuid = :uuid AND archived_at IS NULL`

	r, err := store.tx.NamedExec(q, webhook)
	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return &domain.NotFoundError{}
	}

	return nil
}

func (store *DbWebhookStore) FindByUuid(uuid string) (*domain.Webhook, error) {

	result := &domain.Webhook{}
	q := `SELECT * FROM webhooks WHERE uuid = $1 AND archived_at IS NULL`
	err := store.tx.Get(result, q, uuid)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}

	return result, err
}

func (store *DbWebhookStore) FindAllByJobUuid(uuid string) ([]*domain.Webhook, error) {

	result := []*domain.Webhook{}
	q := `SELECT * FROM webhooks WHERE job_uuid = $1 AND archived_at IS NULL`
	err := store.tx.Select(&result, q, uuid)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}

	return result, err
}

func (store *DbWebhookStore) FindBySlug(slug string) (*domain.Webhook, error) {

	result := &domain.Webhook{}
	q := `SELECT * FROM webhooks WHERE slug = $1 AND archived_at IS NULL`
	err := store.tx.Get(result, q, slug)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}

	return result, err
}

func (store *DbWebhookStore) ArchiveByUuid(uuid string) error {

	q := `UPDATE webhooks SET archived_at = NOW() AT TIME ZONE 'UTC' WHERE uuid = $1`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return &domain.NotFoundError{}
	}

	return nil
}

func (store *DbWebhookStore) FindByProjectUuid(uuid string) ([]*domain.Webhook, error) {

	q := `SELECT * FROM webhooks WHERE project_uuid = $1 AND archived_at IS NULL AND name NOT LIKE 'urn:harrow:job-notifier:%'`

	result := []*domain.Webhook{}
	err := store.tx.Select(&result, q, uuid)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}

	if err != nil {
		return nil, resolveErrType(err)
	}

	return result, nil
}
