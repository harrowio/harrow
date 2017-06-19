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

type DbGitTriggerStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func NewDbGitTriggerStore(tx *sqlx.Tx) *DbGitTriggerStore {
	return &DbGitTriggerStore{tx: tx}
}

func (self *DbGitTriggerStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *DbGitTriggerStore) SetLogger(l logger.Logger) {
	self.log = l
}

func (store *DbGitTriggerStore) Create(subject *domain.GitTrigger) (string, error) {

	if subject.Uuid == "" {
		subject.Uuid = uuidhelper.MustNewV4()
	}

	q := `INSERT INTO git_triggers (
	  uuid,
          name,
          project_uuid,
          job_uuid,
          {{if .RepositoryUuid}}repository_uuid,{{end}}
          match_ref,
          creator_uuid,
          change_type
	) VALUES (
	  :uuid,
          :name,
          :project_uuid,
          :job_uuid,
          {{if .RepositoryUuid}}:repository_uuid,{{end}}
          :match_ref,
          :creator_uuid,
          :change_type
	);`

	tmpl := template.Must(template.New("query").Parse(q))
	query := new(bytes.Buffer)
	if err := tmpl.Execute(query, subject); err != nil {
		return "", err
	}

	_, err := store.tx.NamedExec(query.String(), subject)
	if err != nil {
		return "", resolveErrType(err)
	}

	return subject.Uuid, nil
}

func (store *DbGitTriggerStore) Update(subject *domain.GitTrigger) error {

	if !uuidhelper.IsValid(subject.Uuid) {
		return &domain.NotFoundError{}
	}

	q := `UPDATE git_triggers SET
	  name = :name,
          project_uuid = :project_uuid,
          job_uuid = :job_uuid,
          repository_uuid = :repository_uuid,
          match_ref = :match_ref,
          creator_uuid = :creator_uuid,
          change_type = :change_type
	WHERE uuid = :uuid AND archived_at IS NULL`
	tmpl := template.Must(template.New("query").Parse(q))
	query := new(bytes.Buffer)
	if err := tmpl.Execute(query, subject); err != nil {
		return err
	}

	r, err := store.tx.NamedExec(query.String(), subject)
	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return &domain.NotFoundError{}
	}

	return nil
}

func (store *DbGitTriggerStore) FindByUuid(uuid string) (*domain.GitTrigger, error) {

	result := &domain.GitTrigger{}
	q := `SELECT * FROM git_triggers WHERE uuid = $1 AND archived_at IS NULL`
	err := store.tx.Get(result, q, uuid)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}

	return result, err
}

func (store *DbGitTriggerStore) FindByProjectUuid(projectUuid string) ([]*domain.GitTrigger, error) {
	q := `SELECT * FROM git_triggers WHERE project_uuid = $1 AND archived_at IS NULL`

	result := []*domain.GitTrigger{}
	err := store.tx.Select(&result, q, projectUuid)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}

	if err != nil {
		return nil, resolveErrType(err)
	}

	return result, nil
}

func (store *DbGitTriggerStore) FindAllByJobUuid(jobUuid string) ([]*domain.GitTrigger, error) {
	q := `SELECT * FROM git_triggers WHERE job_uuid = $1 AND archived_at IS NULL`

	result := []*domain.GitTrigger{}
	err := store.tx.Select(&result, q, jobUuid)
	if err == sql.ErrNoRows {
		return nil, &domain.NotFoundError{}
	}

	if err != nil {
		return nil, resolveErrType(err)
	}

	return result, nil
}

func (store *DbGitTriggerStore) ArchiveByUuid(uuid string) error {

	q := `UPDATE git_triggers SET archived_at = NOW() AT TIME ZONE 'UTC' WHERE uuid = $1`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return &domain.NotFoundError{}
	}

	return nil
}
