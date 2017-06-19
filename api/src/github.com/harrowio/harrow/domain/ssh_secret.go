package domain

import (
	"encoding/json"
	"errors"
)

var ErrNotASshSecret = errors.New("secret type is not ssh")

type SshSecret struct {
	*Secret               `json:"-"`
	PrivateKey, PublicKey string
}

// UnprivilegedSshSecret is meant for transport to unprivileged clients
// (i.e. project members), and only decodes the public key
type UnprivilegedSshSecret struct {
	*Secret
	PublicKey string `json:"publicKey"`
}

// PrivilegedSshSecret is meant for transport to privileged clients
// (i.e. project owners), and decodes both public and private keys
type PrivilegedSshSecret struct {
	*UnprivilegedSshSecret
	PrivateKey string `json:"privateKey"`
}

func AsSshSecret(s *Secret) (*SshSecret, error) {
	if !s.IsSsh() {
		return nil, ErrNotASshSecret
	}
	sshSecret := &SshSecret{Secret: s}
	err := json.Unmarshal(s.SecretBytes, sshSecret)
	if err != nil {
		return nil, err
	}
	return sshSecret, nil
}

func (self *SshSecret) AsSecret() (*Secret, error) {
	secretBytes, err := json.Marshal(self)
	if err != nil {
		return nil, err
	}
	if self.Secret == nil {
		self.Secret = &Secret{
			Type: SecretSsh,
		}
	}
	self.Secret.SecretBytes = secretBytes
	return self.Secret, nil
}

func (self *SshSecret) AsPrivileged() (*PrivilegedSshSecret, error) {
	unpriv, err := self.AsUnprivileged()
	if err != nil {
		return nil, err
	}
	return &PrivilegedSshSecret{
		UnprivilegedSshSecret: unpriv,
		PrivateKey:            self.PrivateKey,
	}, nil
}

func (self *SshSecret) AsUnprivileged() (*UnprivilegedSshSecret, error) {
	secret, err := self.AsSecret()

	if err != nil {
		return nil, err
	}
	return &UnprivilegedSshSecret{
		Secret:    secret,
		PublicKey: self.PublicKey,
	}, nil
}
