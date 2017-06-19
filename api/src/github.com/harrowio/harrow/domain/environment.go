package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/harrowio/harrow/hstore"
)

type Environment struct {
	defaultSubject
	Uuid        string               `json:"uuid"`
	Name        string               `json:"name"`
	ProjectUuid string               `json:"projectUuid" db:"project_uuid"`
	IsDefault   bool                 `json:"-"  db:"is_default"`
	Variables   EnvironmentVariables `json:"variables"`
	ArchivedAt  *time.Time           `json:"archivedAt"  db:"archived_at"`
	CreatedAt   time.Time            `json:"-"`
}

type EnvironmentVariables struct {
	M map[string]string
}

func NewEnvironment(uuid string) *Environment {
	return &Environment{
		Uuid: uuid,
		Variables: EnvironmentVariables{
			M: map[string]string{},
		},
	}
}

func (self *Environment) Id() string { return self.Uuid }

func (self *Environment) CreationDate() time.Time {
	return self.CreatedAt
}

func (self *Environment) Get(variable string) string {
	return self.Variables.M[variable]
}

func (self *Environment) Set(variable, value string) *Environment {
	self.Variables.M[variable] = value
	return self
}

func (self *Environment) PruneVariables() {
	delete(self.Variables.M, "")
}

func (self *Environment) Equal(b *Environment) bool {
	if self.Uuid != b.Uuid {
		return false
	}
	if self.Name != b.Name {
		return false
	}
	if self.ProjectUuid != b.ProjectUuid {
		return false
	}
	return true
}

type EnvironmentVariable struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	OldValue *string `json:"oldValue,omitempty"`
}

type EnvironmentVariableChanges struct {
	Added   []*EnvironmentVariable `json:"added"`
	Removed []*EnvironmentVariable `json:"removed"`
	Changed []*EnvironmentVariable `json:"changed"`
}

func NewEnvironmentVariableChanges() *EnvironmentVariableChanges {
	return &EnvironmentVariableChanges{
		Added:   []*EnvironmentVariable{},
		Removed: []*EnvironmentVariable{},
		Changed: []*EnvironmentVariable{},
	}
}

func (self EnvironmentVariables) Diff(other EnvironmentVariables) *EnvironmentVariableChanges {
	result := NewEnvironmentVariableChanges()

	for otherName, otherValue := range other.M {
		if _, found := self.M[otherName]; !found {
			result.Added = append(result.Added, &EnvironmentVariable{
				Name:  otherName,
				Value: otherValue,
			})
		}
	}

	for selfName, selfValue := range self.M {
		if otherValue, found := other.M[selfName]; !found {
			result.Removed = append(result.Removed, &EnvironmentVariable{
				Name:  selfName,
				Value: selfValue,
			})
		} else if otherValue != selfValue {
			result.Changed = append(result.Changed, &EnvironmentVariable{
				Name:     selfName,
				Value:    otherValue,
				OldValue: &selfValue,
			})
		}
	}

	return result
}

func (self *EnvironmentVariables) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.M)
}

func (self *EnvironmentVariables) UnmarshalJSON(data []byte) error {
	vars := make(map[string]string)
	if err := json.Unmarshal(data, &vars); err != nil {
		return err
	}
	if self.M == nil {
		self.M = make(map[string]string)
	}
	for name, value := range vars {
		self.M[name] = value
	}
	return nil
}

func (self *EnvironmentVariables) Scan(value interface{}) error {
	if m, err := hstore.Scan(value); err != nil {
		return err
	} else {
		self.M = m
		return nil
	}
}

func (self EnvironmentVariables) Value() (driver.Value, error) {
	return hstore.Value(self.M)
}

func (self *Environment) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/environments/%s", requestScheme, requestBaseUri, self.Uuid)
}

func (self *Environment) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["project"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s", requestScheme, requestBaseUri, self.ProjectUuid)}
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	response["targets"] = map[string]string{"href": fmt.Sprintf("%s://%s/environments/%s/targets", requestScheme, requestBaseUri, self.Uuid)}
	response["jobs"] = map[string]string{"href": fmt.Sprintf("%s://%s/environments/%s/jobs", requestScheme, requestBaseUri, self.Uuid)}
	response["secrets"] = map[string]string{"href": fmt.Sprintf("%s://%s/environments/%s/secrets", requestScheme, requestBaseUri, self.Uuid)}
	return response
}

// FindProject satisfies authz.BelongsToProject in order to determine
// authorization.
func (self *Environment) FindProject(store ProjectStore) (*Project, error) {
	return store.FindByUuid(self.ProjectUuid)
}

func (self *Environment) AuthorizationName() string {
	if self.IsDefault {
		return "default-environment"
	}

	return "environment"
}

// NewSecret returns a new secret associated with this environment
func (self *Environment) NewSecret(name string, kind SecretType) *Secret {
	return &Secret{
		EnvironmentUuid: self.Uuid,
		Name:            name,
		Type:            kind,
		Status:          SecretPending,
	}
}
