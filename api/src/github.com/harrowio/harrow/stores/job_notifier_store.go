package stores

import (
	"database/sql"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"
	"github.com/jmoiron/sqlx"
)

type DbJobNotifierStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func NewDbJobNotifierStore(tx *sqlx.Tx) *DbJobNotifierStore {
	return &DbJobNotifierStore{tx: tx}
}

func (store *DbJobNotifierStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *DbJobNotifierStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store *DbJobNotifierStore) Create(subject *domain.JobNotifier) (string, error) {

	if subject.Uuid == "" {
		subject.Uuid = uuidhelper.MustNewV4()
	}

	q := `INSERT INTO job_notifiers (
	  uuid,
          webhook_url,
          job_uuid,
          project_uuid
	) VALUES (
	  :uuid,
          :webhook_url,
          :job_uuid,
          :project_uuid
	);`

	_, err := store.tx.NamedExec(q, subject)
	if err != nil {
		return "", resolveErrType(err)
	}

	return subject.Uuid, nil
}

func (store *DbJobNotifierStore) Update(subject *domain.JobNotifier) error {

	if !uuidhelper.IsValid(subject.Uuid) {
		return &domain.NotFoundError{}
	}

	q := `UPDATE job_notifiers SET
	  webhook_url = :webhook_url,
	  job_uuid = :job_uuid,
          project_uuid = :project_uuid
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

func (store *DbJobNotifierStore) FindByUuid(uuid string) (*domain.JobNotifier, error) {

	result := &domain.JobNotifier{}
	q := `
SELECT
 job_notifiers.*,
 (
  SELECT
   name
  FROM
   jobs
  WHERE
   jobs.uuid = job_notifiers.job_uuid
 ) AS job_name
FROM
 job_notifiers
WHERE
 uuid = $1
AND
 archived_at IS NULL
;
`
	err := store.tx.Get(result, q, uuid)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}

	return result, err
}

func (store *DbJobNotifierStore) ArchiveByUuid(uuid string) error {

	q := `UPDATE job_notifiers SET archived_at = NOW() AT TIME ZONE 'UTC' WHERE uuid = $1`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return &domain.NotFoundError{}
	}

	return nil
}

func (store *DbJobNotifierStore) FindAllByTriggeredJobUuid(jobUuid string) ([]*domain.JobNotifier, error) {

	q := `SELECT * FROM job_notifiers WHERE job_uuid = $1 AND archived_at IS NULL`
	result := []*domain.JobNotifier{}
	if err := store.tx.Select(&result, q, jobUuid); err != nil {
		return nil, resolveErrType(err)
	}

	return result, nil
}

func (store *DbJobNotifierStore) FindAllByJobUuid(jobUuid string) ([]*domain.JobNotifier, error) {

	q := `
SELECT DISTINCT ON (job_notifiers.uuid)
  job_notifiers.*,
  (SELECT name FROM jobs WHERE jobs.uuid = job_notifiers.job_uuid) as job_name
FROM
  job_notifiers,
  notification_rules
WHERE
  job_notifiers.archived_at IS NULL
AND
  notification_rules.archived_at IS NULL
AND
  notification_rules.job_uuid = $1
AND
  notification_rules.notifier_type = 'job_notifiers'
;
`
	result := []*domain.JobNotifier{}
	if err := store.tx.Select(&result, q, jobUuid); err != nil {
		return nil, resolveErrType(err)
	}

	return result, nil
}

func (store *DbJobNotifierStore) FindAllByProjectUuid(projectUuid string) ([]*domain.JobNotifier, error) {

	q := `
SELECT DISTINCT ON (job_notifiers.uuid)
  job_notifiers.*,
  (SELECT name FROM jobs WHERE jobs.uuid = job_notifiers.job_uuid) as job_name
FROM
  job_notifiers
WHERE
  job_notifiers.archived_at IS NULL
AND
  job_notifiers.project_uuid = $1
;
`
	result := []*domain.JobNotifier{}
	if err := store.tx.Select(&result, q, projectUuid); err != nil {
		return nil, resolveErrType(err)
	}

	return result, nil
}
