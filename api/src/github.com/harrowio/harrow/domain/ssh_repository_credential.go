package domain

import "encoding/json"

type SshRepositoryCredential struct {
	*RepositoryCredential `json:"-"`
	PrivateKey, PublicKey string
}

func AsSshRepositoryCredential(rc *RepositoryCredential) (*SshRepositoryCredential, error) {
	if !rc.IsSsh() {
		return nil, NewRepositoryCredentialTypeError(rc.Type, RepositoryCredentialSsh)
	}
	sshRc := &SshRepositoryCredential{RepositoryCredential: rc}

	if len(rc.SecretBytes) > 0 {
		if err := json.Unmarshal(rc.SecretBytes, sshRc); err != nil {
			return nil, err
		}
	}

	return sshRc, nil
}

func (self *SshRepositoryCredential) AsRepositoryCredential() (*RepositoryCredential, error) {
	secretBytes, err := json.Marshal(self)
	if err != nil {
		return nil, err
	}
	if self.RepositoryCredential == nil {
		self.RepositoryCredential = &RepositoryCredential{
			Type: RepositoryCredentialSsh,
		}
	}
	self.RepositoryCredential.SecretBytes = secretBytes
	return self.RepositoryCredential, nil
}
