package domain

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type SecretType string

var (
	SecretEnv         SecretType = "env"
	SecretEnvOverride SecretType = "env-override"
	SecretSsh         SecretType = "ssh"
)

// driver.Valuer
func (self SecretType) Value() (driver.Value, error) {
	return string(self), nil
}

// driver.Scanner
func (self *SecretType) Scan(value interface{}) error {
	switch t := value.(type) {
	default:
		return fmt.Errorf("unexpected type %T", t)
	case []byte:
		*self = SecretType(t)
		return nil
	}
}

type SecretStatus string

var (
	SecretPending SecretStatus = "pending"
	SecretPresent SecretStatus = "present"
)

// driver.Valuer
func (self SecretStatus) Value() (driver.Value, error) {
	return string(self), nil
}

// driver.Scanner
func (self *SecretStatus) Scan(value interface{}) error {
	switch t := value.(type) {
	default:
		return fmt.Errorf("unexpected type %T", t)
	case []byte:
		*self = SecretStatus(value.([]byte))
		return nil
	}
}

type Secret struct {
	defaultSubject
	Uuid            string       `json:"uuid"`
	Name            string       `json:"name"`
	EnvironmentUuid string       `json:"environmentUuid" db:"environment_uuid"`
	Type            SecretType   `json:"type"`
	Status          SecretStatus `json:"status"`
	ArchivedAt      *time.Time   `json:"archivedAt"      db:"archived_at"`
	Key             []byte       `json:"-"               db:"key"`
	SecretBytes     []byte       `json:"-"               db:"-"`
}

func (self *Secret) IsPending() bool {
	return self.Status == SecretPending
}

func (self *Secret) IsSsh() bool {
	return self.Type == SecretSsh
}

func (self *Secret) IsEnv() bool {
	return self.Type == SecretEnv
}

func (self *Secret) IsEnvOverride() bool {
	return self.Type == SecretEnvOverride
}

// authz.BelongsToProject
func (self *Secret) FindProject(store ProjectStore) (*Project, error) {
	return store.FindByEnvironmentUuid(self.EnvironmentUuid)
}

// domain.Subject
func (self *Secret) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/secrets/%s", requestScheme, requestBaseUri, self.Uuid)
}

// domain.Subject
func (self *Secret) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	response["environment"] = map[string]string{"href": fmt.Sprintf("%s://%s/environments/%s", requestScheme, requestBaseUri, self.EnvironmentUuid)}
	return response
}

// authz.Subject
func (self *Secret) AuthorizationName() string {
	return "secret"
}
