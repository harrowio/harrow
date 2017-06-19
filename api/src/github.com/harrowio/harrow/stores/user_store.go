package stores

import (
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"

	"database/sql"

	"github.com/jameskeane/bcrypt"
	"github.com/jmoiron/sqlx"
)

type CachedUserStore struct {
	cache  map[string]*domain.User
	onMiss domain.UserStore
	log    logger.Logger
}

func NewCachedUserStore(around domain.UserStore) *CachedUserStore {
	return &CachedUserStore{
		cache:  map[string]*domain.User{},
		onMiss: around,
	}
}

func (store *CachedUserStore) Log() logger.Logger {
	if store.log == nil {
		store.log = logger.Discard
	}
	return store.log
}

func (store *CachedUserStore) SetLogger(l logger.Logger) {
	store.log = l
}

func (store *CachedUserStore) FindByUuid(userUuid string) (*domain.User, error) {

	user, found := store.cache[userUuid]
	if found {
		return user, nil
	}

	user, err := store.onMiss.FindByUuid(userUuid)
	if err != nil {
		return nil, err
	}

	store.cache[userUuid] = user
	return user, nil
}

func (store *CachedUserStore) FindAllSubscribers(watchableId, event string) ([]*domain.User, error) {

	return store.onMiss.FindAllSubscribers(watchableId, event)
}

func NewDbUserStore(tx *sqlx.Tx, c *config.Config) *DbUserStore {
	return &DbUserStore{tx: tx, config: c}
}

// getDbUserStore stores users in the database
type DbUserStore struct {
	tx     *sqlx.Tx
	config *config.Config
	log    logger.Logger
}

func (self DbUserStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self DbUserStore) SetLogger(l logger.Logger) {
	self.log = l
}

func (store DbUserStore) mustHashPassword(password string) string {

	var salt string
	if store.config.Environment() == "test" {
		salt, _ = bcrypt.Salt(4)
	} else {
		salt, _ = bcrypt.Salt(12)
	}

	passwordHash, err := bcrypt.Hash(password, salt)
	if err != nil {
		panic(err)
	}

	return passwordHash

}

func (store DbUserStore) FindUuidByEmailAddressAndPassword(email string, password string) (string, error) {

	var r = struct {
		Uuid         string `db:"uuid"`
		PasswordHash string `db:"password_hash"`
	}{}

	q := `
		SELECT
			uuid,
			password_hash
		FROM
			users
		WHERE
			email = $1 AND
			without_password = FALSE
	`

	err := store.tx.Get(&r, q, email)

	if err == sql.ErrNoRows {
		return "", new(domain.NotFoundError)
	}

	if err != nil {
		return "", resolveErrType(err)
	}

	if !bcrypt.Match(password, r.PasswordHash) {
		return "", new(domain.NotFoundError)
	}

	return r.Uuid, nil

}

// Used for password reset. The user is neved as a response. We need it to
// trigger an email to the user that the password reset was requested for.
func (store DbUserStore) FindByEmailAddress(email string) (*domain.User, error) {

	user := new(domain.User)

	var q string = `SELECT * FROM users WHERE email = $1`
	err := store.tx.Get(user, q, email)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (store DbUserStore) FindAllForProject(projectUuid string) ([]*domain.User, error) {

	result := []*domain.User{}

	q := `
SELECT u.*
FROM   users AS u

INNER JOIN organization_memberships AS om
        ON u.uuid = om.user_uuid

INNER JOIN organizations AS o
        ON o.uuid = om.organization_uuid

INNER JOIN projects AS p
        ON o.uuid = p.organization_uuid

WHERE p.uuid = $1
AND   o.archived_at IS NULL
AND   p.archived_at IS NULL
`
	err := store.tx.Select(&result, q, projectUuid)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (store *DbUserStore) FindAllSubscribers(watchableId, event string) ([]*domain.User, error) {

	result := []*domain.User{}

	q := `
SELECT u.*
FROM   users AS u

INNER JOIN subscriptions as s
        ON s.user_uuid = u.uuid

WHERE s.watchable_uuid = $1
  AND s.event_name = $2
`
	err := store.tx.Select(&result, q, watchableId, event)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (store DbUserStore) FindByUuid(uuid string) (*domain.User, error) {

	var user *domain.User = &domain.User{Uuid: uuid}

	var q string = `SELECT * FROM users WHERE uuid = $1`
	err := store.tx.Get(user, q, user.Uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (store DbUserStore) Create(u *domain.User) (string, error) {

	if err := domain.ValidateUser(u); err != nil {
		return "", err
	}

	if !u.WithoutPassword {
		u.PasswordHash = store.mustHashPassword(u.Password)
	}

	if len(u.Uuid) == 0 {
		u.Uuid = uuidhelper.MustNewV4()
	}

	var q string = `
		INSERT INTO users (
			uuid,
			name,
			gh_username,
			email,
			totp_secret,
			without_password,
			password_hash,
			url_host,
			signup_parameters
		) VALUES (
			:uuid,
			:name,
			:gh_username,
			:email,
			:totp_secret,
			:without_password,
			:password_hash,
			:url_host,
			:signup_parameters
		) RETURNING uuid
	`
	rows, err := store.tx.NamedQuery(q, u)

	if err != nil {
		return "", resolveErrType(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", resolveErrType(rows.Err())
	}

	if err = rows.Scan(&u.Uuid); err != nil {
		return "", resolveErrType(err)
	}

	return u.Uuid, nil

}

func (store DbUserStore) Update(u *domain.User) error {

	if len(u.Uuid) == 0 {
		return new(domain.NotFoundError)
	}

	if u.Password != "" {
		u.WithoutPassword = false
	}

	if err := domain.ValidateUser(u); err != nil {
		return err
	}

	if u.Password != "" {
		u.PasswordHash = store.mustHashPassword(u.Password)
	}

	var q string = `
		UPDATE
			users
		SET
                        name = :name,
                        gh_username = :gh_username,
                        email = :email,
                        without_password = :without_password,
                        password_hash = :password_hash,
                        totp_secret = :totp_secret,
                        totp_enabled_at = :totp_enabled_at

		WHERE uuid = :uuid
	`
	r, err := store.tx.NamedExec(q, u)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	return resolveErrType(err)

}
