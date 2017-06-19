package domain

import (
	"fmt"
	"net/url"
	"regexp"
	"time"

	"github.com/harrowio/harrow/git"
)

type Repository struct {
	defaultSubject
	Uuid           string     `json:"uuid"`
	Url            string     `json:"url"`
	Name           string     `json:"name"`
	Accessible     bool       `json:"accessible"`
	VisibleTo      string     `json:"-"              db:"visible_to"`
	ProjectUuid    string     `json:"projectUuid"    db:"project_uuid"`
	GithubImported bool       `json:"githubImported" db:"github_imported"`
	GithubLogin    string     `json:"githubLogin"    db:"github_login"`
	GithubRepo     string     `json:"githubRepo"     db:"github_repo"`
	CreatedAt      time.Time  `json:"createdAt"      db:"created_at"`
	ArchivedAt     *time.Time `json:"archivedAt"     db:"archived_at"`

	MetadataUpdatedAt *time.Time          `json:"metadataUpdatedAt" db:"metadata_updated_at"`
	Metadata          *RepositoryMetaData `json:"metadata" db:"metadata"`

	ConnectedSuccessfully *bool `json:"connectedSuccessfully,omitempty" db:"connected_successfully"`
}

func ValidateRepository(u *Repository) error {

	if len(u.Url) == 0 {
		return NewValidationError("url", "too_short")
	}

	if _, err := u.Git(); err != nil {
		return NewValidationError("url", err.Error())
	}

	if len(u.ProjectUuid) == 0 {
		return NewValidationError("projectUuid", "required")
	}

	return nil
}

func (self *Repository) Git() (*git.Repository, error) {
	return git.NewRepository(self.Url)
}

// PublicURL returns the URL used for cloning this repository with all
// sensitive information removed.
func (self *Repository) PublicURL() string {
	gitRepo, err := self.Git()
	if err != nil {
		return self.Url
	}

	return gitRepo.URL.String()
}

// CloneURL returns the URL to use for cloning the repository.  This
// url can include sensitive informaiton (usernames, passwords).
func (self *Repository) CloneURL() string {
	if g, err := self.Git(); err == nil {
		return g.CloneURL()
	} else {
		return ""
	}
}

func (self *Repository) ClonedGit(OS git.System, credentials RepositoryCredentialStore) (*git.ClonedRepository, error) {
	repositoryURL, err := git.Parse(self.Url)
	if err != nil {
		return nil, fmt.Errorf("git.Parse(%q): %s", self.Url, err)
	}
	clonedRepository := git.NewClonedRepository(OS, &repositoryURL.URL)
	clonedRepository.MakePersistent()
	credentialType := RepositoryCredentialBasic
	if clonedRepository.UsesSSH() {
		credentialType = RepositoryCredentialSsh
	}
	repositoryCredential, err := credentials.FindByRepositoryUuidAndType(self.Uuid, credentialType)
	if err != nil && !IsNotFound(err) {
		return nil, fmt.Errorf("credentials.FindByRepositoryUuidAndType(%q, %q): %s", self.Uuid, credentialType, err)
	}

	if repositoryCredential != nil {
		if err := self.setRepositoryCredential(clonedRepository, repositoryCredential, &repositoryURL.URL); err != nil {
			return nil, err
		}
	}

	return clonedRepository, nil
}

func (self *Repository) setRepositoryCredential(clonedRepository *git.ClonedRepository, repositoryCredential *RepositoryCredential, repositoryURL *url.URL) error {
	gitCredential := (git.Credential)(nil)
	if clonedRepository.UsesHTTP() {
		credential, err := AsBasicRepositoryCredential(repositoryCredential)
		if err != nil {
			return err
		}
		httpCredential, err := git.NewHTTPCredential(credential.Username, credential.Password)
		if err != nil {
			return err
		}
		gitCredential = httpCredential
	} else if clonedRepository.UsesSSH() {
		credential, err := AsSshRepositoryCredential(repositoryCredential)
		if err != nil {
			return err
		}
		user := "git"
		if repositoryURL.User != nil {
			user = repositoryURL.User.Username()
		}

		if len(credential.PrivateKey) > 0 {
			sshCredential, err := git.NewSshCredential(user, []byte(credential.PrivateKey))
			if err != nil {
				return err
			}
			gitCredential = sshCredential
		}
	}

	if gitCredential != nil {
		if err := clonedRepository.SetCredential(gitCredential); err != nil {
			clonedRepository.Remove()
			return err
		}
	}

	return nil
}

func (self *Repository) SetCredential(credential *RepositoryCredential) *Repository {
	gitRepo, err := self.Git()
	if err != nil {
		return self
	}

	if !gitRepo.UsesHTTP() {
		return self
	}

	if credential.Type != RepositoryCredentialBasic {
		return self
	}

	basicCredential, err := AsBasicRepositoryCredential(credential)
	if err != nil {
		return self
	}

	gitCredential, err := git.NewHTTPCredential(basicCredential.Username, basicCredential.Password)
	if err != nil {
		return self
	}

	gitRepo.SetCredential(gitCredential)

	self.Url = gitRepo.CloneURL()

	return self
}

func (self *Repository) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/repositories/%s", requestScheme, requestBaseUri, self.Uuid)
}

func (self *Repository) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["checks"] = map[string]string{"href": fmt.Sprintf("%s://%s/repositories/%s/checks", requestScheme, requestBaseUri, self.Uuid)}
	response["metadata"] = map[string]string{"href": fmt.Sprintf("%s://%s/repositories/%s/metadata", requestScheme, requestBaseUri, self.Uuid)}
	response["operations"] = map[string]string{"href": fmt.Sprintf("%s://%s/repositories/%s/operations", requestScheme, requestBaseUri, self.Uuid)}
	response["project"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s", requestScheme, requestBaseUri, self.ProjectUuid)}
	response["credential"] = map[string]string{"href": fmt.Sprintf("%s://%s/repositories/%s/credential", requestScheme, requestBaseUri, self.Uuid)}
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	response["github-deploy-key"] = map[string]string{"href": fmt.Sprintf("%s://%s/oauth/github/repositories/%s/keys", requestScheme, requestBaseUri, self.Uuid)}
	return response
}

var sshUrlRegexp = regexp.MustCompile("([^@]*)@([^:]*):(.*)")

// FindProject satisfies authz.BelongsToProject in order to determine
// authorization.
func (self *Repository) FindProject(store ProjectStore) (*Project, error) {
	return store.FindByUuid(self.ProjectUuid)
}

func (self *Repository) AuthorizationName() string { return "repository" }

// NewMetadataUpdateOperation returns an operation for updating this
// repository's metadata.
func (self *Repository) NewMetadataUpdateOperation() *Operation {
	uuid := self.Uuid
	return &Operation{
		RepositoryUuid:         &uuid,
		WorkspaceBaseImageUuid: "31b0127a-6d63-4d22-b32b-e1cfc04f4007",
		Type:       OperationTypeGitEnumerationBranches,
		Parameters: NewOperationParameters(),
	}
}
