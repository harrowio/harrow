package domain

import "encoding/json"

type BasicRepositoryCredential struct {
	*RepositoryCredential `json:"-"`
	Username              string `json:"username"`
	Password              string `json:"password"`
}

func AsBasicRepositoryCredential(rc *RepositoryCredential) (*BasicRepositoryCredential, error) {
	if !rc.IsBasic() {
		return nil, NewRepositoryCredentialTypeError(rc.Type, RepositoryCredentialBasic)
	}
	basicRc := &BasicRepositoryCredential{RepositoryCredential: rc}
	err := json.Unmarshal(rc.SecretBytes, basicRc)
	if err != nil {
		return nil, err
	}
	return basicRc, nil
}

func (self *BasicRepositoryCredential) AsRepositoryCredential() (*RepositoryCredential, error) {
	secretBytes, err := json.Marshal(self)
	if err != nil {
		return nil, err
	}
	if self.RepositoryCredential == nil {
		self.RepositoryCredential = &RepositoryCredential{
			Type: RepositoryCredentialBasic,
		}
	}
	self.RepositoryCredential.SecretBytes = secretBytes
	return self.RepositoryCredential, nil
}
