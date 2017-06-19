package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Person struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type RepositoryMetaData struct {
	Contributors map[string]*Person `json:"contributors"`
	Refs         map[string]string  `json:"refs"`
}

func NewRepositoryMetaData() *RepositoryMetaData {
	return &RepositoryMetaData{
		Contributors: map[string]*Person{},
		Refs:         map[string]string{},
	}
}

// WithRef adds a symbolic ref pointing to a given hash to this
// metadata object.
func (self *RepositoryMetaData) WithRef(symbolic, hash string) *RepositoryMetaData {
	self.Refs[symbolic] = hash
	return self
}

// IsEmpty returns true if there are no refs contained in this instance.
func (self *RepositoryMetaData) IsEmpty() bool {
	return len(self.Refs) == 0
}

// Changes returns the changes between this repository metadata and
// the new version provided as an argument.
func (self *RepositoryMetaData) Changes(newVersion *RepositoryMetaData) *RepositoryMetaDataChanges {
	changes := NewRepositoryMetaDataChanges()
	for symbolic, hash := range newVersion.Refs {
		_, refExistsAlready := self.Refs[symbolic]
		if !refExistsAlready {
			changes.Add(NewRepositoryRef(symbolic, hash))
		}
	}

	for symbolic, hash := range self.Refs {
		newHash, refStillExists := newVersion.Refs[symbolic]
		if !refStillExists {
			changes.Remove(NewRepositoryRef(symbolic, hash))
		} else if hash != newHash {
			changes.Change(symbolic, hash, newHash)
		}
	}

	return changes
}

func (self *RepositoryMetaData) Value() (driver.Value, error) {
	return json.Marshal(self)
}

func (self *RepositoryMetaData) Scan(value interface{}) error {
	src, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("RepositoryMetaData: cannot scan from %T", value)
	}

	return json.Unmarshal(src, self)
}
