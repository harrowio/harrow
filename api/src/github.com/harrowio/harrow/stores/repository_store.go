package stores

import (
	"fmt"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"

	"database/sql"

	"github.com/jmoiron/sqlx"
)

func NewDbRepositoryStore(tx *sqlx.Tx) *DbRepositoryStore {
	return &DbRepositoryStore{tx: tx}
}

type DbRepositoryStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func (self DbRepositoryStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self DbRepositoryStore) SetLogger(l logger.Logger) {
	self.log = l
}

func (store DbRepositoryStore) FindAll() ([]*domain.Repository, error) {

	var repositories []*domain.Repository = []*domain.Repository{}

	var q string = `SELECT * FROM repositories WHERE archived_at IS NULL`

	err := store.tx.Select(&repositories, q)
	if err != nil {
		return nil, err
	}

	return repositories, nil
}

func (store DbRepositoryStore) FindAllWithTriggers() ([]*domain.Repository, error) {

	var repositories []*domain.Repository = []*domain.Repository{}

	var q string = `
SELECT
  r.*
FROM
 repositories AS r
WHERE
 archived_at IS NULL
AND
 (
  SELECT
    count(*)
  FROM
    git_triggers       AS gt
  WHERE
    (
     gt.repository_uuid = r.uuid
     OR
     (
      gt.repository_uuid IS NULL
      AND
      gt.project_uuid = r.project_uuid
     )
    )
  AND
    gt.archived_at IS NULL
 ) > 0
`

	err := store.tx.Select(&repositories, q)
	if err != nil {
		return nil, err
	}

	return repositories, nil
}

func (store DbRepositoryStore) FindByUuid(uuid string) (*domain.Repository, error) {

	var repo *domain.Repository = &domain.Repository{Uuid: uuid}

	var q string = `
SELECT
  *,
  (
   SELECT
     COUNT(*)
   FROM
     activities
   WHERE
     name = 'repository.connected-successfully'
   AND
     (payload->>'uuid')::uuid = repositories.uuid
  ) > 0 AS connected_successfully
FROM
 repositories
WHERE
 uuid = $1
AND
 archived_at IS NULL
;`
	err := store.tx.Get(repo, q, repo.Uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return repo, nil

}

func (store DbRepositoryStore) Create(repo *domain.Repository) (string, error) {

	if len(repo.Uuid) == 0 {
		repo.Uuid = uuidhelper.MustNewV4()
	}

	var q string = `INSERT INTO repositories (uuid, project_uuid, url, name, github_imported, github_login, github_repo) VALUES (:uuid, :project_uuid, :url, :name, :github_imported, :github_login, :github_repo) RETURNING uuid;`
	rows, err := store.tx.NamedQuery(q, repo)

	if err != nil {
		return "", resolveErrType(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", resolveErrType(rows.Err())
	}

	if err = rows.Scan(&repo.Uuid); err != nil {
		return "", resolveErrType(err)
	}

	return repo.Uuid, nil

}

func (store DbRepositoryStore) Update(repo *domain.Repository) error {

	if len(repo.Uuid) == 0 {
		return new(domain.NotFoundError)
	}

	var q string = `UPDATE repositories SET (url, name, accessible, github_imported, github_login, github_repo, project_uuid, metadata) = (:url, :name, :accessible, :github_imported, :github_login, :github_repo, :project_uuid, :metadata) WHERE uuid = :uuid AND archived_at IS NULL;`
	result, err := store.tx.NamedExec(q, repo)

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

func (store DbRepositoryStore) ArchiveByUuid(uuid string) error {

	var q string = `UPDATE repositories SET archived_at = NOW() AT TIME ZONE 'UTC' WHERE uuid = $1;`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil

}

func (store *DbRepositoryStore) MarkAsAccessible(uuid string, accessible bool) error {

	var q string = `UPDATE repositories SET accessible = $2 WHERE uuid = $1;`
	r, err := store.tx.Exec(q, uuid, accessible)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil
}

func (store DbRepositoryStore) FindAllByProjectUuid(projUuid string) ([]*domain.Repository, error) {

	var repos []*domain.Repository = []*domain.Repository{}

	var q string = `SELECT * FROM repositories WHERE project_uuid = $1 AND archived_at IS NULL`

	err := store.tx.Select(&repos, q, projUuid)
	if err != nil {
		return nil, err
	}

	return repos, nil

}

func (store DbRepositoryStore) UpdateMetadata(repositoryUuid string, metadata *domain.RepositoryMetaData) error {

	q := `UPDATE repositories SET metadata = $2, metadata_updated_at = NOW() WHERE uuid = $1`
	result, err := store.tx.Exec(q, repositoryUuid, metadata)

	if err != nil {
		return resolveErrType(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	switch rowsAffected {
	case 0:
		return new(domain.NotFoundError)
	case 1:
		return nil
	default:
		return fmt.Errorf("Wanted to update 1 row but updated %d rows", rowsAffected)
	}
}

func (store DbRepositoryStore) FindAllByProjectUuidAndRepositoryName(projUuid, repositoryName string) ([]*domain.Repository, error) {

	var repos []*domain.Repository = []*domain.Repository{}

	var q string = `SELECT * FROM repositories WHERE project_uuid = $1 AND archived_at IS NULL AND url LIKE '%' || $2 || '.git'`

	err := store.tx.Select(&repos, q, projUuid, repositoryName)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func (store *DbRepositoryStore) FindAllByJobUuid(jobUuid string) ([]*domain.Repository, error) {

	var repos []*domain.Repository = []*domain.Repository{}

	q := `SELECT r.*
	      FROM repositories r
	      JOIN tasks t ON t.project_uuid = r.project_uuid
	      JOIN jobs j ON j.task_uuid = t.uuid
	      WHERE j.uuid = $1 AND r.archived_at IS NULL`

	err := store.tx.Select(&repos, q, jobUuid)
	if err != nil {
		return nil, err
	}

	return repos, nil

}

func (self *DbRepositoryStore) FindByOldestMetadata() (*domain.Repository, error) {
	q := `SELECT * FROM repositories WHERE accessible AND archived_at IS NULL ORDER BY metadata_updated_at ASC LIMIT 1`
	repo := &domain.Repository{}
	err := self.tx.Get(repo, q)
	if err != nil {
		return nil, resolveErrType(err)
	}

	return repo, nil
}
