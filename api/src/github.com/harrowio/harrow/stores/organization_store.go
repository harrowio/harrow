package stores

import (
	"fmt"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"

	"database/sql"

	"github.com/jmoiron/sqlx"
)

type CachedOrganizationStore struct {
	cache  map[string]*domain.Organization
	onMiss domain.OrganizationStore
	log    logger.Logger
}

func NewCachedOrganizationStore(around domain.OrganizationStore) *CachedOrganizationStore {
	return &CachedOrganizationStore{
		cache:  map[string]*domain.Organization{},
		onMiss: around,
	}
}

func (store *CachedOrganizationStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *CachedOrganizationStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store *CachedOrganizationStore) FindByUuid(organizationUuid string) (*domain.Organization, error) {

	organization, found := store.cache[organizationUuid]
	if found {
		return organization, nil
	}

	organization, err := store.onMiss.FindByUuid(organizationUuid)
	if organization != nil {
		store.cache[organizationUuid] = organization
	}

	return organization, err
}

func (store *CachedOrganizationStore) FindByProjectUuid(projectUuid string) (*domain.Organization, error) {

	organization, found := store.cache[projectUuid]
	if found {
		return organization, nil
	}

	organization, err := store.onMiss.FindByProjectUuid(projectUuid)
	if organization != nil {
		store.cache[organization.Uuid] = organization
		store.cache[projectUuid] = organization
	}

	return organization, err
}

func NewDbOrganizationStore(tx *sqlx.Tx) *DbOrganizationStore {
	return &DbOrganizationStore{tx: tx}
}

type DbOrganizationStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func (store *DbOrganizationStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *DbOrganizationStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store DbOrganizationStore) FindAll() ([]*domain.Organization, error) {

	result := []*domain.Organization{}

	var q string = `SELECT * FROM organizations WHERE archived_at IS NULL`
	err := store.tx.Select(&result, q)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, resolveErrType(err)
	}

	return result, nil

}

func (store DbOrganizationStore) FindAllArchived() ([]*domain.Organization, error) {

	result := []*domain.Organization{}

	var q string = `SELECT * FROM organizations WHERE archived_at IS NOT NULL`
	err := store.tx.Select(&result, q)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, resolveErrType(err)
	}

	return result, nil

}

func (store DbOrganizationStore) FindByUuid(uuid string) (*domain.Organization, error) {

	var org *domain.Organization = &domain.Organization{Uuid: uuid}

	var q string = `SELECT * FROM organizations WHERE uuid = $1 AND archived_at IS NULL`
	err := store.tx.Get(org, q, uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return org, nil

}

func (store DbOrganizationStore) FindByProjectUuid(projectUuid string) (*domain.Organization, error) {

	organization := &domain.Organization{}

	var q string = `
SELECT
  organizations.*
FROM
  organizations,
  projects
WHERE
  organizations.uuid = projects.organization_uuid
AND
  projects.uuid = $1
AND
  organizations.archived_at IS NULL
`
	err := store.tx.Get(organization, q, projectUuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return organization, nil

}

func (store DbOrganizationStore) Create(org *domain.Organization) (string, error) {

	if err := domain.ValidateOrganization(org); err != nil {
		return "", err
	}

	if len(org.Uuid) == 0 {
		org.Uuid = uuidhelper.MustNewV4()
	}

	var q string = `INSERT INTO organizations (uuid, name, public, github_login) VALUES (:uuid, :name, :public, :github_login) RETURNING uuid;`
	rows, err := store.tx.NamedQuery(q, org)

	if err != nil {
		return "", resolveErrType(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", resolveErrType(rows.Err())
	}

	if err = rows.Scan(&org.Uuid); err != nil {
		return "", resolveErrType(err)
	}

	return org.Uuid, nil

}

func (store DbOrganizationStore) FindAllByUserUuidThroughMemberships(userUuid string) ([]*domain.Organization, error) {

	result := []*domain.Organization{}
	q := `
SELECT
   o.*
FROM
   organizations o
WHERE
 (
  SELECT
    COUNT(*)
  FROM
    organization_memberships om
  WHERE
    om.user_uuid = $1
  AND
    om.organization_uuid = o.uuid
  AND
    o.archived_at IS NULL
 ) > 0
OR
 (
  SELECT
    COUNT(*)
  FROM
    project_memberships pm,
    projects p
  WHERE
    pm.user_uuid = $1
  AND
    p.uuid = pm.project_uuid
  AND
    p.organization_uuid = o.uuid
  AND
    pm.archived_at IS NULL
  AND
    p.archived_at IS NULL
 ) > 0
;
`
	if err := store.tx.Select(&result, q, userUuid); err != nil {
		return nil, resolveErrType(err)
	}

	return result, nil
}

func (store DbOrganizationStore) Update(org *domain.Organization) error {

	if len(org.Uuid) == 0 {
		return new(domain.NotFoundError)
	}

	if err := domain.ValidateOrganization(org); err != nil {
		return err
	}

	var q string = `UPDATE organizations SET (name, public, github_login) = (:name, :public, :github_login) WHERE uuid = :uuid AND archived_at IS NULL;`
	result, err := store.tx.NamedExec(q, org)

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

func (store DbOrganizationStore) ArchiveByUuid(uuid string) error {

	var q string = `UPDATE organizations SET archived_at = NOW() AT TIME ZONE 'UTC' WHERE uuid = $1;`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil

}

func (store DbOrganizationStore) DeleteByUuid(uuid string) error {

	var q string = `DELETE FROM organizations WHERE uuid = $1;`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil

}
