package stores

import "fmt"

type KeyValueStore interface {
	Get(key string) ([]byte, error)
	Exists(key string) (bool, error)
	Set(key string, data []byte) error
	Del(key string) error

	LRange(key string, start, stop int64) ([]string, error)
	LPush(key string, data string) error
	RPush(key string, data string) error

	SMembers(key string) ([]string, error)
	SIsMember(key, member string) (bool, error)
	SAdd(key, member string) (int64, error)
	SRem(key, member string) (int64, error)
}

type SecretKeyValueStore interface {
	Get(key string, passphrase []byte) ([]byte, error)
	Set(key string, passphrase, data []byte) error
	Del(key string) error
}

var ErrKeyNotFound = fmt.Errorf("KeyValueStore: nil")
