package stores

import (
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"

	"database/sql"

	"github.com/jmoiron/sqlx"
)

func NewDbOAuthTokenStore(tx *sqlx.Tx) *DbOAuthTokenStore {
	return &DbOAuthTokenStore{tx: tx}
}

type DbOAuthTokenStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func (store *DbOAuthTokenStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *DbOAuthTokenStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store DbOAuthTokenStore) FindByUuid(uuid string) (*domain.OAuthToken, error) {

	var token *domain.OAuthToken = &domain.OAuthToken{Uuid: uuid}

	var q string = `SELECT * FROM oauth_tokens WHERE uuid = $1`
	err := store.tx.Get(token, q, token.Uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return token, nil

}

func (store DbOAuthTokenStore) FindByAccessToken(accessToken string) (*domain.OAuthToken, error) {

	token := new(domain.OAuthToken)

	var q string = `SELECT * FROM oauth_tokens WHERE access_token = $1`
	err := store.tx.Get(token, q, accessToken)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return token, nil

}

func (store DbOAuthTokenStore) FindByUserUuid(userUuid string) (*domain.OAuthToken, error) {

	var token *domain.OAuthToken = &domain.OAuthToken{UserUuid: userUuid}

	var q string = `SELECT * FROM oauth_tokens WHERE user_uuid = $1`
	err := store.tx.Get(token, q, token.UserUuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return token, nil

}

func (store DbOAuthTokenStore) Create(token *domain.OAuthToken) (string, error) {

	if len(token.Uuid) == 0 {
		token.Uuid = uuidhelper.MustNewV4()
	}

	var q string = `INSERT INTO oauth_tokens (uuid, user_uuid, provider, access_token, token_type, scopes) VALUES (:uuid, :user_uuid, :provider, :access_token, :token_type, :scopes) RETURNING uuid;`
	rows, err := store.tx.NamedQuery(q, token)

	if err != nil {
		return "", resolveErrType(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", resolveErrType(rows.Err())
	}

	if err = rows.Scan(&token.Uuid); err != nil {
		return "", resolveErrType(err)
	}

	return token.Uuid, nil

}

func (store DbOAuthTokenStore) FindByProviderAndUserUuid(provider, userUuid string) (*domain.OAuthToken, error) {

	var token *domain.OAuthToken = &domain.OAuthToken{Provider: provider, UserUuid: userUuid}

	var q string = `SELECT * FROM oauth_tokens WHERE provider = $1 AND user_uuid = $2`
	err := store.tx.Get(token, q, token.Provider, token.UserUuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return token, nil

}

func (store DbOAuthTokenStore) FindAllByUserUuid(userUuid string) ([]*domain.OAuthToken, error) {

	var tokens []*domain.OAuthToken = []*domain.OAuthToken{}

	var q string = `SELECT * FROM oauth_tokens WHERE user_uuid = $1`

	err := store.tx.Select(&tokens, q, userUuid)
	if err != nil {
		return nil, err
	}

	return tokens, nil

}

func (store DbOAuthTokenStore) FindByRepositoryUUID(repositoryUUID string) (*domain.OAuthToken, error) {

	q := `
SELECT
  oauth_tokens.*
FROM
  oauth_tokens,
  (
    SELECT
      t.uuid,
      (SELECT
          COUNT(*)
        FROM
          project_memberships AS pm
        WHERE
          pm.user_uuid = u.uuid
        AND
          pm.project_uuid = p.uuid
        AND
          pm.archived_at is NULL
       ) AS project_memberships,
       (SELECT
          COUNT(*)
        FROM
          organization_memberships AS om
        WHERE
          om.organization_uuid = o.uuid
        AND
          om.user_uuid = u.uuid
       ) AS organization_memberships
    FROM
      oauth_tokens AS t,
      repositories AS r,
      users AS u,
      projects AS p,
      organizations AS o
    WHERE
      r.uuid = $1
    AND
      p.uuid = r.project_uuid
    AND
      o.uuid = p.organization_uuid
    AND
      r.archived_at IS NULL
  ) AS repository_tokens
WHERE
  oauth_tokens.uuid = repository_tokens.uuid
AND
  repository_tokens.project_memberships + repository_tokens.organization_memberships > 0
`

	token := &domain.OAuthToken{}
	if err := store.tx.Get(token, q, repositoryUUID); err != nil {
		if err == sql.ErrNoRows {
			return nil, new(domain.NotFoundError)
		}
		return nil, resolveErrType(err)
	}

	return token, nil
}

func (store DbOAuthTokenStore) DeleteByUuid(uuid string) error {

	var q string = `DELETE FROM oauth_tokens WHERE uuid = $1`

	_, err := store.tx.Exec(q, uuid)
	if err != nil {
		return err
	}

	return nil
}
