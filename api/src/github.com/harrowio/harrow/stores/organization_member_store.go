package stores

import (
	"database/sql"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"

	"github.com/jmoiron/sqlx"
)

type DbOrganizationMemberStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func NewDbOrganizationMemberStore(tx *sqlx.Tx) *DbOrganizationMemberStore {
	return &DbOrganizationMemberStore{tx: tx}
}

func (store *DbOrganizationMemberStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *DbOrganizationMemberStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store *DbOrganizationMemberStore) FindAllByOrganizationUuid(organizationUuid string) ([]*domain.OrganizationMember, error) {

	result := []*domain.OrganizationMember{}

	q := `
	SELECT u.*, om.type as membership_type, om.organization_uuid
	FROM   users u
	JOIN   organization_memberships om  ON u.uuid = om.user_uuid
	WHERE  om.organization_uuid = $1::uuid
	ORDER BY u.name ASC
	`

	err := store.tx.Select(&result, q, organizationUuid)
	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}
