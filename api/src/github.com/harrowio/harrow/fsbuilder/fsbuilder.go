package fsbuilder

import (
	"io"

	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
)

func NewConfig(secrets stores.SecretKeyValueStore, tx *sqlx.Tx) *Config {
	return &Config{secrets: secrets, tx: tx}
}

type Config struct {
	secrets stores.SecretKeyValueStore
	tx      *sqlx.Tx
}

func (self *Config) Secrets() stores.SecretKeyValueStore {
	return self.secrets
}

func (self *Config) Tx() *sqlx.Tx {
	return self.tx
}

type FsBuilder interface {
	Build(operationUuid string) (io.Reader, error)
}
