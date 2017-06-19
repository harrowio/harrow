package stores

import (
	"database/sql"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"
	"github.com/jmoiron/sqlx"
)

type CachedInvitationStore struct {
	cache  map[string]*domain.Invitation
	onMiss domain.InvitationStore
	log    logger.Logger
}

func NewCachedInvitationStore(onMiss domain.InvitationStore) *CachedInvitationStore {
	return &CachedInvitationStore{
		cache:  map[string]*domain.Invitation{},
		onMiss: onMiss,
	}
}

func (store *CachedInvitationStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *CachedInvitationStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store CachedInvitationStore) FindByUserAndProjectUuid(userId, projectId string) (*domain.Invitation, error) {

	key := userId + ":" + projectId
	invitation, found := store.cache[key]
	if found {
		return invitation, nil
	}

	invitation, err := store.onMiss.FindByUserAndProjectUuid(userId, projectId)
	if err != nil {
		return nil, err
	}

	store.cache[key] = invitation

	return invitation, nil
}

type DbInvitationStore struct {
	tx *sqlx.Tx
}

func NewDbInvitationStore(tx *sqlx.Tx) *DbInvitationStore {
	return &DbInvitationStore{tx: tx}
}

func (store *DbInvitationStore) Create(invitation *domain.Invitation) (string, error) {

	if len(invitation.Uuid) == 0 {
		invitation.Uuid = uuidhelper.MustNewV4()
	}

	if len(invitation.InviteeUuid) == 0 {
		// Generate a new id.  This will be used as the user uuid
		// when a new user signs up due to receiving an invitation.
		invitation.InviteeUuid = uuidhelper.MustNewV4()
	}

	q := `INSERT INTO invitations
               (uuid,
                recipient_name,
                email,
                organization_uuid,
                project_uuid,
                membership_type,
                creator_uuid,
                invitee_uuid,
                message)
        VALUES (:uuid,
                :recipient_name,
                :email,
                :organization_uuid,
                :project_uuid,
                :membership_type,
                :creator_uuid,
                :invitee_uuid,
                :message)
        RETURNING uuid;
	`
	rows, err := store.tx.NamedQuery(q, invitation)

	if err != nil {
		return "", resolveErrType(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", resolveErrType(rows.Err())
	}

	if err = rows.Scan(&invitation.Uuid); err != nil {
		return "", resolveErrType(err)
	}

	return invitation.Uuid, nil

}

func (store *DbInvitationStore) FindByUuid(uuid string) (*domain.Invitation, error) {

	inv := &domain.Invitation{Uuid: uuid}
	q := `SELECT * FROM invitations WHERE uuid = $1 AND archived_at IS NULL;`
	err := store.tx.Get(inv, q, inv.Uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return inv, nil
}

func (store *DbInvitationStore) FindByUserAndProjectUuid(userId, projectId string) (*domain.Invitation, error) {

	inv := &domain.Invitation{}
	q := `SELECT * FROM invitations WHERE invitee_uuid = $1 AND project_uuid = $2 AND archived_at IS NULL;`
	err := store.tx.Get(inv, q, userId, projectId)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return inv, nil
}

func (store *DbInvitationStore) Update(inv *domain.Invitation) error {

	if len(inv.Uuid) == 0 {
		return new(domain.NotFoundError)
	}

	q := `
        UPDATE invitations
        SET
                accepted_at       = :accepted_at,
                creator_uuid      = :creator_uuid,
                email             = :email,
                invitee_uuid      = :invitee_uuid,
                membership_type   = :membership_type,
                message           = :message,
                organization_uuid = :organization_uuid,
                project_uuid      = :project_uuid,
                recipient_name    = :recipient_name,
                refused_at        = :refused_at
        WHERE   uuid = :uuid
          AND   archived_at IS NULL;
	`
	rows, err := store.tx.NamedQuery(q, inv)
	if err == sql.ErrNoRows {
		return new(domain.NotFoundError)
	}

	if err != nil {
		return err
	}
	rows.Close()

	return nil
}
