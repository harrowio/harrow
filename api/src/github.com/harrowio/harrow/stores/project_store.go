package stores

import (
	"fmt"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"

	"database/sql"

	"github.com/jmoiron/sqlx"
)

type CachedProjectStore struct {
	cache  map[string]*domain.Project
	onMiss domain.ProjectStore
	log    logger.Logger
}

func NewCachedProjectStore(around domain.ProjectStore) *CachedProjectStore {
	return &CachedProjectStore{
		cache:  map[string]*domain.Project{},
		onMiss: around,
	}
}

func (store *CachedProjectStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *CachedProjectStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store *CachedProjectStore) cached(uuid string, onMiss func(string) (*domain.Project, error)) (*domain.Project, error) {
	project, found := store.cache[uuid]
	if found {
		return project, nil
	}

	project, err := onMiss(uuid)
	if err != nil {
		return nil, err
	}

	store.cache[uuid] = project
	return project, nil
}

func (store *CachedProjectStore) FindByUuid(projectUuid string) (*domain.Project, error) {

	return store.cached(projectUuid, store.onMiss.FindByUuid)
}

func (store *CachedProjectStore) FindByJobUuid(jobUuid string) (*domain.Project, error) {

	return store.cached(jobUuid, store.onMiss.FindByJobUuid)
}

func (store *CachedProjectStore) FindByTaskUuid(taskUuid string) (*domain.Project, error) {

	return store.cached(taskUuid, store.onMiss.FindByTaskUuid)
}

func (store *CachedProjectStore) FindByRepositoryUuid(repositoryUuid string) (*domain.Project, error) {

	return store.cached(repositoryUuid, store.onMiss.FindByRepositoryUuid)
}

func (store *CachedProjectStore) FindByOrganizationUuid(organizationUuid string) (*domain.Project, error) {

	return store.cached(organizationUuid, store.onMiss.FindByOrganizationUuid)
}

func (store *CachedProjectStore) FindByNotifierUuid(notifierUuid, notifierType string) (*domain.Project, error) {

	project, found := store.cache[notifierUuid]
	if found {
		return project, nil
	}

	project, err := store.onMiss.FindByNotifierUuid(notifierUuid, notifierType)
	if err != nil {
		return nil, err
	}

	store.cache[notifierUuid] = project
	return project, nil
}

func (store *CachedProjectStore) FindByNotificationRule(notifierUuid, notifierType string) (*domain.Project, error) {

	project, found := store.cache[notifierUuid]
	if found {
		return project, nil
	}

	project, err := store.onMiss.FindByNotificationRule(notifierUuid, notifierType)
	if err != nil {
		return nil, err
	}

	store.cache[notifierUuid] = project
	return project, nil
}

func (store *CachedProjectStore) FindByMemberUuid(memberUuid string) (*domain.Project, error) {

	return store.cached(memberUuid, store.onMiss.FindByMemberUuid)
}

func (store *CachedProjectStore) FindByEnvironmentUuid(environmentUuid string) (*domain.Project, error) {

	return store.cached(environmentUuid, store.onMiss.FindByEnvironmentUuid)
}

func (store *CachedProjectStore) FindByWebhookUuid(webhookUuid string) (*domain.Project, error) {

	return store.cached(webhookUuid, store.onMiss.FindByWebhookUuid)
}

func NewDbProjectStore(tx *sqlx.Tx) *DbProjectStore {
	return &DbProjectStore{tx: tx}
}

type DbProjectStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func (store DbProjectStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store DbProjectStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store DbProjectStore) FindByUuid(uuid string) (*domain.Project, error) {

	var project *domain.Project = &domain.Project{Uuid: uuid}

	var q string = `SELECT * FROM projects WHERE uuid = $1 and archived_at IS NULL`
	err := store.tx.Get(project, q, project.Uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return project, nil

}

func (store DbProjectStore) FindByUuidWithDeleted(uuid string) (*domain.Project, error) {

	var project *domain.Project = &domain.Project{Uuid: uuid}

	var q string = `SELECT * FROM projects WHERE uuid = $1`
	err := store.tx.Get(project, q, project.Uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return project, nil

}

func (store DbProjectStore) FindByMemberUuid(uuid string) (*domain.Project, error) {

	project := &domain.Project{}

	var q string = `SELECT *
	FROM projects p
	WHERE p.archived_at IS NULL
	AND   (
	  EXISTS (SELECT true FROM project_memberships pm
		     WHERE pm.user_uuid = $1
		     AND   pm.project_uuid = p.uuid)
	OR
	  EXISTS (SELECT true FROM organization_memberships om
		     WHERE om.user_uuid = $1
		     AND   p.organization_uuid = om.organization_uuid)
	)`

	err := store.tx.Get(project, q, uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return project, nil

}

func (store DbProjectStore) FindByOrganizationUuid(uuid string) (*domain.Project, error) {

	project := &domain.Project{}

	var q string = `SELECT * FROM projects WHERE organization_uuid = $1 and archived_at IS NULL`
	err := store.tx.Get(project, q, uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return project, nil
}

func (store DbProjectStore) FindByRepositoryUuid(uuid string) (*domain.Project, error) {

	project := domain.Project{}

	q := `
SELECT     p.*
FROM       projects   AS p

INNER JOIN repositories AS r
        ON p.uuid = r.project_uuid

WHERE      r.uuid = $1
AND        p.archived_at IS NULL;
`
	err := store.tx.Get(&project, q, uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (store DbProjectStore) FindByEnvironmentUuid(uuid string) (*domain.Project, error) {

	project := domain.Project{}

	q := `
SELECT     p.*
FROM       projects   AS p

INNER JOIN environments AS e
        ON p.uuid = e.project_uuid

WHERE      e.uuid = $1
AND        p.archived_at IS NULL;
`
	err := store.tx.Get(&project, q, uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (store DbProjectStore) FindByTaskUuid(uuid string) (*domain.Project, error) {

	project := domain.Project{}

	q := `
SELECT     p.*
FROM       projects   AS p

INNER JOIN tasks      AS t
        ON p.uuid = t.project_uuid

WHERE      t.uuid = $1
AND        p.archived_at IS NULL;
`
	err := store.tx.Get(&project, q, uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (store DbProjectStore) FindByWebhookUuid(uuid string) (*domain.Project, error) {

	project := domain.Project{}
	q := `
SELECT      p.*
FROM        projects AS p
INNER JOIN  webhooks AS wh
        ON wh.project_uuid = p.uuid
     WHERE wh.uuid = $1
       AND p.archived_at IS NULL
`

	err := store.tx.Get(&project, q, uuid)
	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, resolveErrType(err)
	}

	return &project, nil
}

func (store DbProjectStore) FindByJobUuid(uuid string) (*domain.Project, error) {

	project := domain.Project{}

	q := `
SELECT     p.*
FROM       projects   AS p

INNER JOIN tasks      AS t
        ON p.uuid = t.project_uuid

INNER JOIN jobs       AS j
        ON t.uuid = j.task_uuid

WHERE      j.uuid = $1
AND        p.archived_at IS NULL;
`
	err := store.tx.Get(&project, q, uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (store DbProjectStore) FindByNotifierUuid(uuid, notifierType string) (*domain.Project, error) {

	project := domain.Project{}

	q := fmt.Sprintf(`
SELECT     p.*
FROM       projects   AS p

INNER JOIN notification_rules as nr
        ON p.uuid = nr.project_uuid

INNER JOIN %q  AS n
        ON n.uuid = nr.notifier_uuid

WHERE      n.uuid = $1
AND        p.archived_at IS NULL;
`, notifierType)
	err := store.tx.Get(&project, q, uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (store DbProjectStore) FindForAction(table, uuid string) (*domain.Project, error) {

	project := domain.Project{}

	q := fmt.Sprintf(`
SELECT     p.*
FROM       projects   AS p

INNER JOIN tasks      AS t
        ON p.uuid = t.project_uuid

INNER JOIN jobs       AS j
        ON t.uuid = j.task_uuid

INNER JOIN %s AS x
        ON j.uuid = x.job_uuid

WHERE      x.uuid = $1
AND        p.archived_at IS NULL;
`, table)
	err := store.tx.Get(&project, q, uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (store DbProjectStore) Create(project *domain.Project) (string, error) {

	if err := domain.ValidateProject(project); err != nil {
		return "", err
	}

	if len(project.Uuid) == 0 {
		project.Uuid = uuidhelper.MustNewV4()
	}

	var q string = `INSERT INTO projects (uuid, organization_uuid, public, name) VALUES (:uuid, :organization_uuid, :public, :name) RETURNING uuid;`
	rows, err := store.tx.NamedQuery(q, project)

	if err != nil {
		return "", resolveErrType(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", resolveErrType(rows.Err())
	}

	if err = rows.Scan(&project.Uuid); err != nil {
		return "", resolveErrType(err)
	}

	return project.Uuid, nil

}

func (store DbProjectStore) Update(project *domain.Project) error {

	if len(project.Uuid) == 0 {
		return new(domain.NotFoundError)
	}

	var q string = `UPDATE projects SET (name, organization_uuid, public) = (:name, :organization_uuid, :public) WHERE uuid = :uuid AND archived_at IS NULL;`
	result, err := store.tx.NamedExec(q, project)

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

func (store DbProjectStore) DeleteByUuid(uuid string) error {

	var q string = `DELETE FROM projects WHERE uuid = $1;`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil

}

func (store DbProjectStore) ArchiveByUuid(uuid string) error {

	var q string = `UPDATE projects SET archived_at = NOW() AT TIME ZONE 'UTC' WHERE uuid = $1;`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil

}

func (store DbProjectStore) FindAllIncludingArchived() ([]*domain.Project, error) {

	var projects []*domain.Project = []*domain.Project{}

	var q string = `SELECT * FROM projects`

	err := store.tx.Select(&projects, q)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (store DbProjectStore) FindAllArchived() ([]*domain.Project, error) {

	var projects []*domain.Project = []*domain.Project{}

	var q string = `SELECT * FROM projects WHERE archived_at is not NULL`

	err := store.tx.Select(&projects, q)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (store DbProjectStore) FindAll() ([]*domain.Project, error) {

	var projects []*domain.Project = []*domain.Project{}

	var q string = `SELECT * FROM projects WHERE archived_at IS NULL`

	err := store.tx.Select(&projects, q)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (store DbProjectStore) FindAllByUserUuid(userUuid string) ([]*domain.Project, error) {

	var projects []*domain.Project = []*domain.Project{}

	var q string = `
	WITH orgs (uuid) AS (
	  SELECT organization_uuid
	  FROM   organization_memberships
	  WHERE  user_uuid = $1
	),
        by_membership (uuid) AS (
          SELECT project_uuid
          FROM   project_memberships
          WHERE  user_uuid = $1
        )
        SELECT p.*
        FROM   projects AS p
        JOIN   orgs AS o
          ON   p.organization_uuid = o.uuid
        WHERE  p.archived_at is NULL
        UNION
        SELECT p.*
        FROM   projects AS p
        JOIN   by_membership AS m
          ON   p.uuid = m.uuid
        WHERE  p.archived_at is NULL
`
	err := store.tx.Select(&projects, q, userUuid)
	if err != nil {
		return nil, err
	}

	return projects, nil

}

func (store DbProjectStore) FindAllByUserUuidOnlyThroughProjectMemberships(userUuid string) ([]*domain.Project, error) {

	var projects []*domain.Project = []*domain.Project{}

	var q string = `
        SELECT p.*
        FROM   projects AS p
        JOIN   project_memberships AS pm
          ON   p.uuid = pm.project_uuid
        WHERE  p.archived_at is NULL
          AND  pm.archived_at IS NULL
          AND  pm.user_uuid = $1
`
	err := store.tx.Select(&projects, q, userUuid)
	if err != nil {
		return nil, err
	}

	return projects, nil

}

func (store DbProjectStore) FindAllByOrganizationUuid(orgUuid string) ([]*domain.Project, error) {

	var projects []*domain.Project = []*domain.Project{}

	var q string = `SELECT * FROM projects WHERE organization_uuid = $1 and archived_at IS NULL`

	err := store.tx.Select(&projects, q, orgUuid)
	if err != nil {
		return nil, err
	}

	return projects, nil

}

func (store DbProjectStore) FindByNotificationRule(notifierType string, notifierUuid string) (*domain.Project, error) {

	project := &domain.Project{}
	q := `SELECT * FROM projects WHERE uuid = (SELECT project_uuid FROM notification_rules WHERE notifier_type = $1 AND notifier_uuid = $2 LIMIT 1) AND archived_at IS NULL`
	if err := store.tx.Get(project, q, notifierType, notifierUuid); err != nil {
		return nil, resolveErrType(err)
	}

	return project, nil
}
