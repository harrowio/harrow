package stores

import (
	"bytes"
	"errors"
	"io/ioutil"

	redis "gopkg.in/redis.v2"

	"github.com/harrowio/harrow/logger"

	"golang.org/x/crypto/openpgp"
)

type RedisSecretKeyValueStore struct {
	c   *redis.Client
	log logger.Logger
}

func NewRedisSecretKeyValueStore(c *redis.Client) SecretKeyValueStore {
	return &RedisSecretKeyValueStore{c: c}
}

func (self *RedisSecretKeyValueStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *RedisSecretKeyValueStore) SetLogger(l logger.Logger) {
	self.log = l
}

func (self *RedisSecretKeyValueStore) Get(key string, passphrase []byte) ([]byte, error) {
	r := self.c.Get(key)
	if r.Err() == redis.Nil {
		return nil, ErrKeyNotFound
	}
	if r.Err() != nil {
		return nil, r.Err()
	}
	enc := []byte(r.Val())
	return decrypt(enc, passphrase)
}
func (self *RedisSecretKeyValueStore) Set(key string, passphrase, data []byte) error {
	enc, err := encrypt(data, passphrase)
	if err != nil {
		return err
	}
	r := self.c.Set(key, string(enc))
	return r.Err()
}

func (self *RedisSecretKeyValueStore) Del(key string) error {
	r := self.c.Del(key)
	return r.Err()
}

func encrypt(plaintext, passphrase []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	plainWriter, err := openpgp.SymmetricallyEncrypt(buf, passphrase, nil, nil)
	if err != nil {
		return nil, err
	}
	_, err = plainWriter.Write(plaintext)
	if err != nil {
		return nil, err
	}
	err = plainWriter.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decrypt(ciphertext, passphrase []byte) ([]byte, error) {
	// Provide the key []byte in a openpgp.PromptFunc
	prompt := func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		if !symmetric {
			return nil, errors.New("Can only decrypt symmetrically encrypted messages")
		}
		return passphrase, nil
	}
	md, err := openpgp.ReadMessage(bytes.NewReader(ciphertext), nil, prompt, nil)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(md.UnverifiedBody)
}
