package stores

import (
	"crypto/rand"
	"database/sql"
	"fmt"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"
	"github.com/jmoiron/sqlx"
)

type SecretStore struct {
	ss  SecretKeyValueStore
	tx  *sqlx.Tx
	log logger.Logger
}

func (self *SecretStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *SecretStore) SetLogger(l logger.Logger) {
	self.log = l
}

func NewSecretStore(ss SecretKeyValueStore, tx *sqlx.Tx) *SecretStore {
	return &SecretStore{
		ss: ss,
		tx: tx,
	}
}

func (self *SecretStore) Create(secret *domain.Secret) (string, error) {

	if len(secret.Uuid) == 0 {
		secret.Uuid = uuidhelper.MustNewV4()
	}
	if len(secret.Status) == 0 {
		// if this is a ssh secret, the status is pending until keymaker makes the keys
		if secret.IsSsh() {
			secret.Status = domain.SecretPending
		} else {
			secret.Status = domain.SecretPresent
		}
	}

	// We cannot have the database create the key, because it would only present
	// after reloading the Secret, but we need it right away to write to
	// the SecretKeyValueStore.
	if secret.Key == nil {
		b := make([]byte, 56)
		_, err := rand.Read(b)
		if err != nil {
			return "", err
		}
		secret.Key = b
	}

	var q string = `INSERT INTO secrets (uuid, name, environment_uuid, type, status, key, archived_at) VALUES (:uuid, :name, :environment_uuid, :type, :status, :key, :archived_at) RETURNING uuid;`
	rows, err := self.tx.NamedQuery(q, secret)

	if err != nil {
		return "", resolveErrType(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", resolveErrType(rows.Err())
	}

	var uuid string
	if err = rows.Scan(&uuid); err != nil {
		return "", resolveErrType(err)
	}

	err = self.ss.Set(secret.Uuid, secret.Key, secret.SecretBytes)
	if err != nil {
		return "", err
	}

	return uuid, nil
}

func (self *SecretStore) FindAll() ([]*domain.Secret, error) {

	var secrets []*domain.Secret = []*domain.Secret{}

	var q string = `SELECT * FROM secrets WHERE archived_at IS NULL`

	err := self.tx.Select(&secrets, q)
	if err != nil {
		return nil, err
	}
	for _, secret := range secrets {
		if secret.Status == domain.SecretPresent {
			secretBytes, err := self.ss.Get(secret.Uuid, secret.Key)
			if err != nil {
				self.Log().Error().Msgf("secret not found: %s", secret.Uuid)
			}
			secret.SecretBytes = secretBytes
		}
	}

	return secrets, nil
}

func (self *SecretStore) FindByUuid(uuid string) (*domain.Secret, error) {

	secret := &domain.Secret{}

	var q string = `SELECT * FROM secrets WHERE uuid = $1 AND archived_at IS NULL`
	err := self.tx.Get(secret, q, uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	if secret.Status == domain.SecretPresent {
		secretBytes, err := self.ss.Get(secret.Uuid, secret.Key)
		if err == ErrKeyNotFound {
			return nil, new(domain.NotFoundError)
		}
		if err != nil {
			return nil, err
		}
		secret.SecretBytes = secretBytes
	}

	return secret, nil
}

func (self *SecretStore) FindAllByEnvironmentUuid(environmentUuid string) ([]*domain.Secret, error) {

	var secrets []*domain.Secret = []*domain.Secret{}

	var q string = `SELECT * FROM secrets WHERE environment_uuid = $1 AND archived_at IS NULL`
	err := self.tx.Select(&secrets, q, environmentUuid)

	if err == sql.ErrNoRows {
		return []*domain.Secret{}, nil
	}
	if err != nil {
		return nil, err
	}

	for _, secret := range secrets {
		if secret.Status == domain.SecretPresent {
			secretBytes, err := self.ss.Get(secret.Uuid, secret.Key)
			if err == ErrKeyNotFound {
				return nil, new(domain.NotFoundError)
			}
			if err != nil {
				return nil, err
			}
			secret.SecretBytes = secretBytes
		}
	}

	return secrets, nil
}

func (self *SecretStore) Update(secret *domain.Secret) error {

	if len(secret.Uuid) == 0 {
		return new(domain.NotFoundError)
	}

	var q string = `UPDATE secrets SET (name, environment_uuid, type, status, archived_at) = (:name, :environment_uuid, :type, :status, :archived_at) WHERE uuid = :uuid AND archived_at IS NULL;`
	result, err := self.tx.NamedExec(q, secret)

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

	err = self.ss.Set(secret.Uuid, secret.Key, secret.SecretBytes)
	if err != nil {
		return err
	}
	return nil
}

func (self *SecretStore) ArchiveByUuid(uuid string) error {

	var q string = `UPDATE secrets SET archived_at = NOW() AT TIME ZONE 'UTC' WHERE uuid = $1;`
	r, err := self.tx.Exec(q, uuid)

	if err != nil {
		return resolveErrType(err)
	}

	if n, _ := r.RowsAffected(); n == 0 {
		return new(domain.NotFoundError)
	}

	err = self.ss.Del(uuid)

	return err

}
