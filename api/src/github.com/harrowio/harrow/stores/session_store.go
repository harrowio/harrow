package stores

import (
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"

	"database/sql"

	"github.com/jmoiron/sqlx"
)

func NewDbSessionStore(tx *sqlx.Tx) *DbSessionStore {
	return &DbSessionStore{tx: tx}
}

type DbSessionStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func (store DbSessionStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store DbSessionStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store DbSessionStore) FindByUuid(uuid string) (*domain.Session, error) {

	var session *domain.Session = &domain.Session{Uuid: uuid}

	var q string = `SELECT *, now() as loaded_at FROM sessions WHERE uuid = $1`
	err := store.tx.Get(session, q, session.Uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	session.Validate()

	return session, nil
}

func (store DbSessionStore) MarkAsLoggedOut(session *domain.Session) error {

	q := `
	UPDATE sessions SET logged_out_at = $2 WHERE uuid = $1;
	`

	rows, err := store.tx.Query(q, session.Uuid, session.LoggedOutAt)
	if err != nil {
		return resolveErrType(err)
	}
	rows.Close()

	return nil
}

func (store DbSessionStore) Create(session *domain.Session) (string, error) {

	if len(session.Uuid) == 0 {
		session.Uuid = uuidhelper.MustNewV4()
	}

	var q string = `
		INSERT INTO sessions (
			uuid,
			validated_at,
			user_uuid,
			user_agent,
			client_address
		) VALUES (
			$1,
			CASE WHEN (SELECT COUNT(uuid) FROM users WHERE uuid = $2 AND totp_enabled_at IS NULL) = 0 THEN timestamp 'epoch' ELSE NOW() END,
			$2,
			$3,
			$4
		) RETURNING uuid;
	`
	rows, err := store.tx.Query(q, session.Uuid, session.UserUuid, session.UserAgent, session.ClientAddress)

	if err != nil {
		return "", resolveErrType(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", resolveErrType(rows.Err())
	}

	if err = rows.Scan(&session.Uuid); err != nil {
		return "", resolveErrType(err)
	}

	return session.Uuid, nil

}

func (store DbSessionStore) ExpireAllSessionsForUserUuid(userUuid string) error {
	q := `UPDATE sessions SET expires_at = now() - '1s'::interval WHERE user_uuid = $1`
	r, err := store.tx.Exec(q, userUuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil
}

func (store DbSessionStore) MarkAsValid(session *domain.Session) error {

	var q string = `
	  UPDATE sessions
	  SET    validated_at = NOW()
	  WHERE  uuid = :uuid;
	`
	r, err := store.tx.NamedExec(q, session)
	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return resolveErrType(err)
}

func (store DbSessionStore) FindAllByUserUuid(userUuid string) ([]*domain.Session, error) {

	var sessions []*domain.Session = []*domain.Session{}

	var q string = `SELECT *, now() as loaded_at FROM sessions WHERE user_uuid = $1`
	err := store.tx.Select(&sessions, q, userUuid)

	if err != nil {
		return nil, resolveErrType(err)
	}

	for _, s := range sessions {
		s.Validate()
	}

	return sessions, nil

}

func (store DbSessionStore) InvalidateAllButMostRecentSessionForUser(userUuid string) (int64, error) {

	q := `WITH sessions_to_invalidate (uuid) AS (
  SELECT uuid
    FROM sessions
   WHERE user_uuid = $1
     AND logged_out_at IS NULL
     AND invalidated_at IS NULL
     AND now() < expires_at
ORDER BY created_at DESC
  OFFSET 1
)
UPDATE sessions
   SET invalidated_at = NOW()
 WHERE uuid IN (SELECT uuid FROM sessions_to_invalidate)
;
`
	r, err := store.tx.Exec(q, userUuid)
	if err != nil {
		return 0, resolveErrType(err)
	}

	return r.RowsAffected()
}

func (store DbSessionStore) DeleteByUuid(uuid string) error {

	var q string = `DELETE FROM sessions WHERE uuid = $1;`
	r, err := store.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return nil

}
