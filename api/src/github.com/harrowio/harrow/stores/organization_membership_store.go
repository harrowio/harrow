package stores

import (
	"time"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"

	"database/sql"

	"github.com/jmoiron/sqlx"
)

func NewDbOrganizationMembershipStore(tx *sqlx.Tx) *DbOrganizationMembershipStore {
	return &DbOrganizationMembershipStore{tx: tx}
}

type DbOrganizationMembershipStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func (store *DbOrganizationMembershipStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *DbOrganizationMembershipStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store DbOrganizationMembershipStore) Create(om *domain.OrganizationMembership) error {

	var q string = `INSERT INTO organization_memberships (organization_uuid, user_uuid, type) VALUES (:organization_uuid, :user_uuid, :type)`
	rows, err := store.tx.NamedQuery(q, om)

	if err != nil {
		return resolveErrType(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return resolveErrType(rows.Err())
	}

	return nil

}

func (store DbOrganizationMembershipStore) FindAllByOrganizationAndUserUuidGreaterThan(orgUuid string, userUuid string, minMembershipType string) ([]*domain.OrganizationMembership, error) {

	var orgMemberships []*domain.OrganizationMembership = []*domain.OrganizationMembership{}

	var q string = `
		SELECT organization_memberships.* FROM organization_memberships
		INNER JOIN organizations ON organization_uuid = organziations.uuid
		WHERE organization_uuid = $1
		AND user_uuid = $2
		AND type > $3
		AND organizations.archived_at IS NULL
	`
	err := store.tx.Select(&orgMemberships, q, orgUuid, userUuid, minMembershipType)

	if err != nil {
		return nil, err
	}

	return orgMemberships, nil

}

func (store DbOrganizationMembershipStore) FindByOrganizationAndUserUuids(orgUuid string, userUuid string) (*domain.OrganizationMembership, error) {

	var orgMembership *domain.OrganizationMembership = &domain.OrganizationMembership{OrganizationUuid: orgUuid, UserUuid: userUuid}

	var q string = `
		SELECT organization_memberships.* FROM organization_memberships
		INNER JOIN organizations ON organization_uuid = organizations.uuid
		WHERE organization_uuid = $1
		AND user_uuid = $2
		AND organizations.archived_at IS NULL
	`
	err := store.tx.Get(orgMembership, q, orgMembership.OrganizationUuid, orgMembership.UserUuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return orgMembership, nil

}

func (store DbOrganizationMembershipStore) FindAllByUserUuid(userUuid string) ([]*domain.OrganizationMembership, error) {

	var orgMemberships []*domain.OrganizationMembership = []*domain.OrganizationMembership{}

	var q string = `
		SELECT organization_memberships.* FROM organization_memberships
		INNER JOIN organizations on organization_uuid = organizations.uuid
		WHERE user_uuid = $1
		AND organizations.archived_at IS NULL
	`
	err := store.tx.Select(&orgMemberships, q, userUuid)

	if err != nil {
		return nil, err
	}

	return orgMemberships, nil

}

func (store DbOrganizationMembershipStore) FindAllByOrganizationUuid(userUuid string) ([]*domain.OrganizationMembership, error) {

	var orgMemberships []*domain.OrganizationMembership = []*domain.OrganizationMembership{}

	var q string = `
		SELECT organization_memberships.* FROM organization_memberships
		INNER JOIN organizations ON organization_uuid = organizations.uuid
		WHERE organization_uuid = $1
		AND organizations.archived_at IS NULL
	`
	err := store.tx.Select(&orgMemberships, q, userUuid)

	if err != nil {
		return nil, err
	}

	return orgMemberships, nil

}

func (store DbOrganizationMembershipStore) FindThroughProjectMemberships(organizationUuid, userUuid string) (*domain.OrganizationMembership, error) {

	q := `
        SELECT COUNT(*)
        FROM   project_memberships AS pm
        JOIN   projects AS p
          ON   p.uuid = pm.project_uuid
        WHERE  p.organization_uuid = $1
          AND  pm.user_uuid = $2
          AND  p.archived_at IS NULL
        `

	count := 0
	err := store.tx.Get(&count, q, organizationUuid, userUuid)

	if count == 0 {
		return nil, &domain.NotFoundError{}
	}

	if err != nil {
		return nil, resolveErrType(err)
	}

	return &domain.OrganizationMembership{
		Type:             domain.MembershipTypeGuest,
		OrganizationUuid: organizationUuid,
		UserUuid:         userUuid,
		CreatedAt:        time.Now(),
	}, nil
}
