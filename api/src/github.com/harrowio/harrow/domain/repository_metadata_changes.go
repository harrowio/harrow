package domain

type RepositoryMetaDataChanges struct {
	AddedRefs   []*RepositoryRef `json:"addedRefs"`
	RemovedRefs []*RepositoryRef `json:"removedRefs"`

	ChangedRefs []*ChangedRepositoryRef `json:"changedRefs"`
}

func NewRepositoryMetaDataChanges() *RepositoryMetaDataChanges {
	return &RepositoryMetaDataChanges{
		AddedRefs:   []*RepositoryRef{},
		RemovedRefs: []*RepositoryRef{},
	}
}

type ChangedRepositoryRef struct {
	RepositoryUuid string `json:"repositoryUuid"`
	Symbolic       string `json:"symbolic"`
	OldHash        string `json:"oldHash"`
	NewHash        string `json:"newHash"`
}

type RepositoryRef struct {
	RepositoryUuid string `json:"repositoryUuid"`
	Symbolic       string `json:"symbolic"`
	Hash           string `json:"hash"`
}

func NewRepositoryRef(symbolic, hash string) *RepositoryRef {
	return &RepositoryRef{
		Symbolic: symbolic,
		Hash:     hash,
	}
}

// Added returns the list of refs that have been added with this
// change.  It never returns nil.
func (self *RepositoryMetaDataChanges) Added() []*RepositoryRef {
	if self.AddedRefs == nil {
		return []*RepositoryRef{}
	}

	return self.AddedRefs
}

// Removed returns the list of refs that have been removed with this
// change.  It never returns nil.
func (self *RepositoryMetaDataChanges) Removed() []*RepositoryRef {
	if self.RemovedRefs == nil {
		return []*RepositoryRef{}
	}

	return self.RemovedRefs
}

// Changed returns the list of refs that are pointing to a different
// hash with this set of changes.  It never returns nil.
func (self *RepositoryMetaDataChanges) Changed() []*ChangedRepositoryRef {
	if self.ChangedRefs == nil {
		return []*ChangedRepositoryRef{}
	}

	return self.ChangedRefs
}

// Add marks ref as a ref that has been introduced with this set of
// changes.
func (self *RepositoryMetaDataChanges) Add(ref *RepositoryRef) *RepositoryMetaDataChanges {
	self.AddedRefs = append(self.AddedRefs, ref)
	return self
}

// Remove marks ref as a ref that has been removed with this set of
// changes.  The ref's hash points to the last known hash.
func (self *RepositoryMetaDataChanges) Remove(ref *RepositoryRef) *RepositoryMetaDataChanges {
	self.RemovedRefs = append(self.RemovedRefs, ref)
	return self
}

// Change marks that ref's old hash has been changed to new with this
// set of changes.
func (self *RepositoryMetaDataChanges) Change(symbolic, oldHash, newHash string) *RepositoryMetaDataChanges {
	self.ChangedRefs = append(self.ChangedRefs, &ChangedRepositoryRef{
		Symbolic: symbolic,
		OldHash:  oldHash,
		NewHash:  newHash,
	})

	return self
}

func (self *RepositoryRef) FindProject(projects ProjectStore) (*Project, error) {
	return projects.FindByRepositoryUuid(self.RepositoryUuid)
}

func (self *ChangedRepositoryRef) FindProject(projects ProjectStore) (*Project, error) {
	return projects.FindByRepositoryUuid(self.RepositoryUuid)
}
