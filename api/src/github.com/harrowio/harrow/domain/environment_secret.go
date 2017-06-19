package domain

import (
	"encoding/json"
	"errors"
)

var ErrNotAnEnvironmentSecret = errors.New("secret type is not env")

// UnprivilegedEnvironmentSecret is meant for transport to unprivileged clients
// (i.e. project members), and doesn't containt the Value
type UnprivilegedEnvironmentSecret struct {
	*Secret
}

// PrivilegedEnvironmentSecret is meant for transport to privileged clients
// (i.e. project owners), and contains the Value
type PrivilegedEnvironmentSecret struct {
	*UnprivilegedEnvironmentSecret
	Value string `json:"value"`
}

type EnvironmentSecret struct {
	*Secret `json:"-"`
	Value   string
}

func AsEnvironmentSecret(s *Secret) (*EnvironmentSecret, error) {
	if s.IsEnvOverride() {
		return &EnvironmentSecret{
			Secret: &Secret{
				Name: s.Name,
				Type: SecretEnv,
			},
			Value: string(s.SecretBytes),
		}, nil
	}
	if !s.IsEnv() {
		return nil, ErrNotAnEnvironmentSecret
	}
	envSecret := &EnvironmentSecret{Secret: s}
	err := json.Unmarshal(s.SecretBytes, envSecret)
	if err != nil {
		return nil, err
	}
	return envSecret, nil
}

func (self *EnvironmentSecret) AsSecret() (*Secret, error) {
	secretBytes, err := json.Marshal(self)
	if err != nil {
		return nil, err
	}
	if self.Secret == nil {
		self.Secret = &Secret{
			Type: SecretEnv,
		}
	}
	self.Secret.SecretBytes = secretBytes
	return self.Secret, nil
}

func (self *EnvironmentSecret) AsPrivileged() (*PrivilegedEnvironmentSecret, error) {
	unpriv, err := self.AsUnprivileged()
	if err != nil {
		return nil, err
	}
	return &PrivilegedEnvironmentSecret{
		UnprivilegedEnvironmentSecret: unpriv,
		Value: self.Value,
	}, nil
}

func (self *EnvironmentSecret) AsUnprivileged() (*UnprivilegedEnvironmentSecret, error) {
	secret, err := self.AsSecret()
	if err != nil {
		return nil, err
	}
	return &UnprivilegedEnvironmentSecret{
		Secret: secret,
	}, nil
}
