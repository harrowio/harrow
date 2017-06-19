package stores

import (
	"io/ioutil"
	"time"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
)

type CachedUserBlockStore struct {
	cache map[string]bool
	next  *DbUserBlockStore
	log   logger.Logger
}

func NewCachedUserBlockStore(next *DbUserBlockStore) *CachedUserBlockStore {
	return &CachedUserBlockStore{
		cache: map[string]bool{},
		next:  next,
	}
}

func (self *CachedUserBlockStore) Log() logger.Logger {
	if self.log == nil {
		self.log = zerolog.New(ioutil.Discard)
	}
	return self.log
}

func (self *CachedUserBlockStore) SetLogger(l logger.Logger) {
	self.log = l
}

func (self *CachedUserBlockStore) UserIsBlocked(userUuid string) (bool, error) {

	if blocked, found := self.cache[userUuid]; found {
		return blocked, nil
	}

	blocks, err := self.next.FindAllByUserUuid(userUuid)
	if err != nil {
		return false, err
	}

	blocked := len(blocks) > 0
	self.cache[userUuid] = blocked

	return blocked, nil
}

type DbUserBlockStore struct {
	tx  *sqlx.Tx
	log logger.Logger
}

func NewDbUserBlockStore(tx *sqlx.Tx) *DbUserBlockStore {
	return &DbUserBlockStore{tx: tx}
}

func (self *DbUserBlockStore) Log() logger.Logger {
	if self.log == nil {
		self.log = zerolog.New(ioutil.Discard)
	}
	return self.log
}

func (self *DbUserBlockStore) SetLogger(l logger.Logger) {
	self.log = l
}

func (store *DbUserBlockStore) FindAllByUserUuid(userUuid string) ([]*domain.UserBlock, error) {

	result := []*domain.UserBlock{}
	q := `SELECT uuid, user_uuid, reason, lower(valid) as valid_from, upper(valid) as valid_to FROM user_blocks WHERE user_uuid = $1 AND valid @> NOW()`
	if err := store.tx.Select(&result, q, userUuid); err != nil {
		return nil, resolveErrType(err)
	}

	return result, nil
}

func (store *DbUserBlockStore) FindAllByUserUuidValidFrom(userUuid string, validFrom time.Time) ([]*domain.UserBlock, error) {

	result := []*domain.UserBlock{}
	q := `SELECT uuid, user_uuid, reason, lower(valid) as valid_from, upper(valid) as valid_to FROM user_blocks WHERE user_uuid = $1 AND $2 >= lower(valid)`
	if err := store.tx.Select(&result, q, userUuid, validFrom); err != nil {
		return nil, resolveErrType(err)
	}

	return result, nil

}

func (store *DbUserBlockStore) Create(block *domain.UserBlock) error {

	q := `INSERT INTO user_blocks (
                uuid,
                user_uuid,
                reason,
                valid
              ) VALUES (
                :uuid,
                :user_uuid,
                :reason,
                tstzrange(:valid_from,:valid_to, '[]')
              );
       `

	_, err := store.tx.NamedExec(q, block)
	if err != nil {
		return resolveErrType(err)
	}

	return nil
}

func (store *DbUserBlockStore) DeleteByUserUuidAndReason(userUuid, reason string) error {

	q := `DELETE FROM user_blocks WHERE user_uuid = $1 AND reason = $2`
	_, err := store.tx.Exec(q, userUuid, reason)
	if err != nil {
		return resolveErrType(err)
	}

	return nil
}
