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

type RepositoryCredentialStore struct {
	ss  SecretKeyValueStore
	tx  *sqlx.Tx
	log logger.Logger
}

func NewRepositoryCredentialStore(ss SecretKeyValueStore, tx *sqlx.Tx) *RepositoryCredentialStore {
	return &RepositoryCredentialStore{ss: ss, tx: tx}
}

func (self *RepositoryCredentialStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *RepositoryCredentialStore) SetLogger(l logger.Logger) {
	self.log = l
}

func (self *RepositoryCredentialStore) Create(repositoryCredential *domain.RepositoryCredential) (string, error) {

	if len(repositoryCredential.Uuid) == 0 {
		repositoryCredential.Uuid = uuidhelper.MustNewV4()
	}
	if len(repositoryCredential.Status) == 0 {
		repositoryCredential.Status = domain.RepositoryCredentialPending
	}

	// We cannot have the database create the key, because it would only present
	// after reloading the Secret, but we need it right away to write to
	// the SecretKeyValueStore.
	if repositoryCredential.Key == nil {
		b := make([]byte, 56)
		_, err := rand.Read(b)
		if err != nil {
			return "", err
		}
		repositoryCredential.Key = b
	}

	var q string = `INSERT INTO repository_credentials (uuid, name, repository_uuid, type, status, key, archived_at) VALUES (:uuid, :name, :repository_uuid, :type, :status, :key, :archived_at) RETURNING uuid;`
	rows, err := self.tx.NamedQuery(q, repositoryCredential)

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

	err = self.ss.Set(repositoryCredential.Uuid, repositoryCredential.Key, repositoryCredential.SecretBytes)
	if err != nil {
		return "", err
	}

	return uuid, nil
}

func (self *RepositoryCredentialStore) FindByUuid(uuid string) (*domain.RepositoryCredential, error) {

	repositoryCredential := &domain.RepositoryCredential{}

	var q string = `SELECT * FROM repository_credentials WHERE uuid = $1 AND archived_at IS NULL`
	err := self.tx.Get(repositoryCredential, q, uuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	if repositoryCredential.Status == domain.RepositoryCredentialPresent {
		secretBytes, err := self.ss.Get(repositoryCredential.Uuid, repositoryCredential.Key)
		if err == ErrKeyNotFound {
			return nil, new(domain.NotFoundError)
		}
		if err != nil {
			return nil, err
		}
		repositoryCredential.SecretBytes = secretBytes
	}

	return repositoryCredential, nil
}

func (self *RepositoryCredentialStore) FindByRepositoryUuidAndType(uuid string, credentialType domain.RepositoryCredentialType) (*domain.RepositoryCredential, error) {

	repositoryCredential := &domain.RepositoryCredential{}

	q := `SELECT * FROM repository_credentials WHERE repository_uuid = $1 AND type = $2 AND archived_at IS NULL`
	err := self.tx.Get(repositoryCredential, q, uuid, credentialType)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	if repositoryCredential.Status == domain.RepositoryCredentialPresent {
		secretBytes, err := self.ss.Get(repositoryCredential.Uuid, repositoryCredential.Key)
		if err == ErrKeyNotFound {
			return nil, new(domain.NotFoundError)
		}
		if err != nil {
			return nil, err
		}
		repositoryCredential.SecretBytes = secretBytes
	}

	return repositoryCredential, nil
}

func (self *RepositoryCredentialStore) FindByRepositoryUuidAndTypeNoLoad(uuid string, credentialType domain.RepositoryCredentialType) (*domain.RepositoryCredential, error) {

	repositoryCredential := &domain.RepositoryCredential{}

	q := `SELECT * FROM repository_credentials WHERE repository_uuid = $1 AND type = $2 AND archived_at IS NULL`
	err := self.tx.Get(repositoryCredential, q, uuid, credentialType)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	return repositoryCredential, nil
}

func (self *RepositoryCredentialStore) FindByRepositoryUuid(repositoryUuid string) (*domain.RepositoryCredential, error) {

	repositoryCredential := &domain.RepositoryCredential{}

	var q string = `SELECT * FROM repository_credentials WHERE repository_uuid = $1 AND archived_at IS NULL`
	err := self.tx.Get(repositoryCredential, q, repositoryUuid)

	if err == sql.ErrNoRows {
		return nil, new(domain.NotFoundError)
	}

	if err != nil {
		return nil, err
	}

	if repositoryCredential.Status == domain.RepositoryCredentialPresent {
		secret, err := self.ss.Get(repositoryCredential.Uuid, repositoryCredential.Key)
		if err == ErrKeyNotFound {
			fmt.Println("about to raise not found err because not present")
			return nil, new(domain.NotFoundError)
		}
		if err != nil {
			return nil, err
		}
		repositoryCredential.SecretBytes = secret
	}

	return repositoryCredential, nil
}

func (self *RepositoryCredentialStore) Update(rc *domain.RepositoryCredential) error {

	if len(rc.Uuid) == 0 {
		return new(domain.NotFoundError)
	}

	var q string = `UPDATE repository_credentials SET (name, repository_uuid, type, status, archived_at) = (:name, :repository_uuid, :type, :status, :archived_at) WHERE uuid = :uuid AND archived_at IS NULL;`
	result, err := self.tx.NamedExec(q, rc)

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

	err = self.ss.Set(rc.Uuid, rc.Key, rc.SecretBytes)
	if err != nil {
		return err
	}
	return nil
}
