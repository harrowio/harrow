package stores

import (
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/jmoiron/sqlx"
)

type DbProjectMemberStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func NewDbProjectMemberStore(tx *sqlx.Tx) *DbProjectMemberStore {
	return &DbProjectMemberStore{tx: tx}
}

func (store *DbProjectMemberStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *DbProjectMemberStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store *DbProjectMemberStore) FindAllByProjectUuid(projectUuid string) ([]*domain.ProjectMember, error) {

	members := []*domain.ProjectMember{}
	// TODO(dh): find a better way to do this.
	q := `
        WITH
        project_organization_uuid AS (
          SELECT organization_uuid
            FROM projects
           WHERE uuid = $1::uuid
        ),
        members_of_project AS (
          SELECT u.*, pm.project_uuid, pm.membership_type, pm.uuid as membership_uuid
          FROM users u
          JOIN project_memberships pm ON u.uuid = pm.user_uuid
          WHERE pm.project_uuid = $1::uuid
            AND pm.archived_at IS NULL
        ),
        members_of_org AS (
          SELECT u.*, $1::uuid as project_uuid, om.type as membership_type, null::uuid as membership_uuid
          FROM users u
          JOIN organization_memberships om ON u.uuid = om.user_uuid
          WHERE om.organization_uuid = (SELECT * FROM project_organization_uuid)
        ),
        members AS (
          SELECT uuid, project_uuid, membership_type, membership_uuid FROM members_of_project
          UNION
          SELECT uuid, project_uuid, membership_type, membership_uuid FROM members_of_org
        )
        SELECT u.*, m.project_uuid, m.membership_type, m.membership_uuid FROM users u, members m WHERE u.uuid = m.uuid
        ORDER BY name ASC
        `

	err := store.tx.Select(&members, q, projectUuid)
	if err != nil {
		return nil, err
	}

	return members, err
}
