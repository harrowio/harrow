package domain

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type RepositoryCredentialType string

type RepositoryCredentialTypeError struct {
	ExpectedType RepositoryCredentialType `json:"expectedType"`
	ActualType   RepositoryCredentialType `json:"actualType"`
}

func NewRepositoryCredentialTypeError(actual, expected RepositoryCredentialType) *RepositoryCredentialTypeError {
	return &RepositoryCredentialTypeError{
		ExpectedType: expected,
		ActualType:   actual,
	}
}

func (self *RepositoryCredentialTypeError) Error() string {
	return fmt.Sprintf("Repository credential of type %q, expected %q", self.ActualType, self.ExpectedType)
}

var (
	RepositoryCredentialSsh   RepositoryCredentialType = "ssh"
	RepositoryCredentialBasic RepositoryCredentialType = "basic"
)

// driver.Valuer
func (self RepositoryCredentialType) Value() (driver.Value, error) {
	return string(self), nil
}

// driver.Scanner
func (self *RepositoryCredentialType) Scan(value interface{}) error {
	switch t := value.(type) {
	default:
		return fmt.Errorf("unexpected type %T", t)
	case []byte:
		*self = RepositoryCredentialType(t)
		return nil
	}
}

type RepositoryCredentialStatus string

var (
	RepositoryCredentialPending RepositoryCredentialStatus = "pending"
	RepositoryCredentialPresent RepositoryCredentialStatus = "present"
)

// driver.Valuer
func (self RepositoryCredentialStatus) Value() (driver.Value, error) {
	return string(self), nil
}

// driver.Scanner
func (self *RepositoryCredentialStatus) Scan(value interface{}) error {
	switch t := value.(type) {
	default:
		return fmt.Errorf("unexpected type %T", t)
	case []byte:
		*self = RepositoryCredentialStatus(value.([]byte))
		return nil
	}
}

type RepositoryCredential struct {
	defaultSubject
	Uuid           string                     `json:"uuid"`
	Name           string                     `json:"name"`
	RepositoryUuid string                     `json:"repositoryUuid" db:"repository_uuid"`
	Type           RepositoryCredentialType   `json:"type"`
	Status         RepositoryCredentialStatus `json:"status"`
	ArchivedAt     *time.Time                 `json:"archivedAt"     db:"archived_at"`
	SecretBytes    []byte                     `json:"-"              db:"-"`
	Key            []byte                     `json:"-"               db:"key"`
	PublicKey      string                     `json:"publicKey,omitempty"`
}

func (self *RepositoryCredential) IsPending() bool {
	return self.Status == RepositoryCredentialPending
}

func (self *RepositoryCredential) IsSsh() bool {
	return self.Type == RepositoryCredentialSsh
}

func (self *RepositoryCredential) IsBasic() bool {
	return self.Type == RepositoryCredentialBasic
}

// authz.BelongsToProject
func (self *RepositoryCredential) FindProject(store ProjectStore) (*Project, error) {
	return store.FindByRepositoryUuid(self.RepositoryUuid)
}

// domain.Subject
func (self *RepositoryCredential) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/repositories/%s/credential", requestScheme, requestBaseUri, self.RepositoryUuid)
}

// domain.Subject
func (self *RepositoryCredential) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["repository"] = map[string]string{"href": fmt.Sprintf("%s://%s/repositories/%s", requestScheme, requestBaseUri, self.RepositoryUuid)}
	return response
}

// authz.Subject
func (self *RepositoryCredential) AuthorizationName() string {
	return "repository-credential"
}
