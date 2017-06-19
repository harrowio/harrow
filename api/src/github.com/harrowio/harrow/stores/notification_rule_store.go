package stores

import (
	"database/sql"
	"text/template"

	"bytes"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"
	"github.com/jmoiron/sqlx"
)

type DbNotificationRuleStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func NewDbNotificationRuleStore(tx *sqlx.Tx) *DbNotificationRuleStore {
	return &DbNotificationRuleStore{tx: tx}
}

func (store *DbNotificationRuleStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *DbNotificationRuleStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store *DbNotificationRuleStore) FindByProjectUuid(projectUuid string) ([]*domain.NotificationRule, error) {
	q := `SELECT * FROM notification_rules WHERE archived_at IS NULL AND project_uuid = $1`
	results := []*domain.NotificationRule{}
	if err := store.tx.Select(&results, q, projectUuid); err != nil {
		return nil, resolveErrType(err)
	}

	return results, nil
}

func (store *DbNotificationRuleStore) FindAllByJobUuid(jobUuid string) ([]*domain.NotificationRule, error) {
	q := `SELECT * FROM notification_rules WHERE archived_at IS NULL AND job_uuid = $1`
	results := []*domain.NotificationRule{}
	if err := store.tx.Select(&results, q, jobUuid); err != nil {
		return nil, resolveErrType(err)
	}

	return results, nil
}

func (store *DbNotificationRuleStore) FindByNotifierAndJobUuidAndType(notifierUuid, jobUuid, notifierType string) (*domain.NotificationRule, error) {
	q := `SELECT * FROM notification_rules WHERE archived_at IS NULL AND notifier_uuid = $1 AND job_uuid = $2 AND notifier_type = $3`
	result := domain.NotificationRule{}
	if err := store.tx.Get(&result, q, notifierUuid, jobUuid, notifierType); err != nil {
		if err == sql.ErrNoRows {
			return nil, new(domain.NotFoundError)
		}
		return nil, resolveErrType(err)
	}

	return &result, nil
}

func (store *DbNotificationRuleStore) FindAllByNotifierUuid(notifierUuid string) ([]*domain.NotificationRule, error) {
	q := `SELECT * FROM notification_rules WHERE archived_at IS NULL AND notifier_uuid = $1`
	result := []*domain.NotificationRule{}
	if err := store.tx.Select(&result, q, notifierUuid); err != nil {
		if err == sql.ErrNoRows {
			return nil, new(domain.NotFoundError)
		}
		return nil, resolveErrType(err)
	}

	return result, nil
}

func (store *DbNotificationRuleStore) Create(subject *domain.NotificationRule) (string, error) {

	if subject.Uuid == "" {
		subject.Uuid = uuidhelper.MustNewV4()
	}

	if err := subject.Validate(); err != nil {
		return "", err
	}

	qTemplate := `INSERT INTO notification_rules (
	  uuid,
          project_uuid,
{{if .JobUuid}}job_uuid,{{end}}
          notifier_uuid,
          notifier_type,
          match_activity,
          creator_uuid
	) VALUES (
	  :uuid,
          :project_uuid,
{{if .JobUuid}}:job_uuid,{{end}}
          :notifier_uuid,
          :notifier_type,
          :match_activity,
          :creator_uuid
	);`

	t := template.Must(template.New("main").Parse(qTemplate))
	q := new(bytes.Buffer)
	if err := t.Execute(q, subject); err != nil {
		return "", err
	}

	_, err := store.tx.NamedExec(q.String(), subject)
	if err != nil {
		return "", resolveErrType(err)
	}

	return subject.Uuid, nil
}

func (store *DbNotificationRuleStore) Update(subject *domain.NotificationRule) error {

	if !uuidhelper.IsValid(subject.Uuid) {
		return &domain.NotFoundError{}
	}

	qTemplate := `UPDATE notification_rules SET
	  project_uuid = :project_uuid,
{{if .JobUuid}} job_uuid = :job_uuid,{{end}}
	  match_activity = :match_activity,
	  notifier_uuid = :notifier_uuid,
	  notifier_type = :notifier_type,
	  creator_uuid = :creator_uuid
	WHERE uuid = :uuid AND archived_at IS NULL`

	t := template.Must(template.New("main").Parse(qTemplate))
	q := new(bytes.Buffer)
	if err := t.Execute(q, subject); err != nil {
		return err
	}

	r, err := store.tx.NamedExec(q.String(), subject)
	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return &domain.NotFoundError{}
	}

	return nil
}

func (store *DbNotificationRuleStore) FindByUuid(uuid string) (*domain.NotificationRule, error) {

	result := &domain.NotificationRule{}
	q := `SELECT * FROM notification_rules WHERE uuid = $1 AND archived_at IS NULL`
	err := store.tx.Get(result, q, uuid)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}

	return result, err
}

func (store *DbNotificationRuleStore) ArchiveByUuid(uuid string) error {

	q := `UPDATE notification_rules SET archived_at = NOW() AT TIME ZONE 'UTC' WHERE uuid = $1`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return &domain.NotFoundError{}
	}

	return nil
}
