package git

type RepositoryMetadata struct {
	Contributors map[string]*Person `json:"contributors"`
	Refs         map[string]string  `json:"refs"`
}

type Person struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func NewRepositoryMetadata() *RepositoryMetadata {
	return &RepositoryMetadata{
		Contributors: map[string]*Person{},
		Refs:         map[string]string{},
	}
}

func (self *RepositoryMetadata) AddRef(ref *Reference) *RepositoryMetadata {
	self.Refs[ref.Name] = ref.Hash
	return self
}

func (self *RepositoryMetadata) AddContributor(contributor *Person) *RepositoryMetadata {
	self.Contributors[contributor.Email] = contributor
	return self
}
