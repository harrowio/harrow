package stores

import (
	"database/sql"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"
	"github.com/jmoiron/sqlx"
)

type DbProjectMembershipStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func NewDbProjectMembershipStore(tx *sqlx.Tx) *DbProjectMembershipStore {
	return &DbProjectMembershipStore{tx: tx}
}

func (store *DbProjectMembershipStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *DbProjectMembershipStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store *DbProjectMembershipStore) FindByUuid(uuid string) (*domain.ProjectMembership, error) {

	membership := &domain.ProjectMembership{}
	q := `SELECT * FROM project_memberships WHERE uuid = $1 AND archived_at IS NULL`
	err := store.tx.Get(membership, q, uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, resolveErrType(err)
	}

	return membership, nil
}

func (store *DbProjectMembershipStore) ArchiveByUuid(uuid string) error {

	var q string = `UPDATE project_memberships SET archived_at = NOW() AT TIME ZONE 'UTC' WHERE uuid = $1;`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil
}

func (store *DbProjectMembershipStore) FindByUserAndProjectUuid(userUuid, projectUuid string) (*domain.ProjectMembership, error) {

	membership := &domain.ProjectMembership{}
	q := `SELECT * FROM project_memberships WHERE user_uuid = $1 and project_uuid = $2 AND archived_at IS NULL LIMIT 1`
	err := store.tx.Get(membership, q, userUuid, projectUuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, resolveErrType(err)
	}

	return membership, nil
}

func (store *DbProjectMembershipStore) Create(membership *domain.ProjectMembership) (string, error) {

	if !uuidhelper.IsValid(membership.Uuid) {
		membership.Uuid = uuidhelper.MustNewV4()
	}

	q := `INSERT INTO project_memberships
		(uuid,
		 project_uuid,
		 user_uuid,
		 membership_type)
		VALUES (:uuid, :project_uuid, :user_uuid, :membership_type);
	`

	rows, err := store.tx.NamedQuery(q, membership)
	if err != nil {
		return "", resolveErrType(err)
	}
	rows.Close()

	return membership.Uuid, nil
}

func (store *DbProjectMembershipStore) Update(membership *domain.ProjectMembership) error {

	if len(membership.Uuid) == 0 {
		return new(domain.NotFoundError)
	}

	q := `UPDATE project_memberships SET membership_type = :membership_type WHERE uuid = :uuid`

	r, err := store.tx.NamedExec(q, membership)
	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil
}

func (store *DbProjectMembershipStore) FindAllByProjectUuid(projectUuid string) ([]*domain.ProjectMembership, error) {

	q := `SELECT *
		FROM  project_memberships
		WHERE project_uuid = $1
		 AND  archived_at IS NULL
	`

	memberships := []*domain.ProjectMembership{}
	err := store.tx.Select(&memberships, q, projectUuid)
	if err != nil {
		return nil, err
	}

	return memberships, nil
}
