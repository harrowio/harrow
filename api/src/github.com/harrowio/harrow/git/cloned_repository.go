package git

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
)

type System interface {
	TempDir() (string, error)
	PersistentDir(key string) (string, error)
	CreateFile(filename string) (io.WriteCloser, error)
	DeleteFile(filename string) error
	SetPermissions(filename string, mode int) error
	Run(cmd *SystemCommand) ([]byte, error)
}

type ClonedRepository struct {
	os       System
	cloneURL *url.URL

	credential   Credential
	tempDir      string
	clonedInto   string
	isPersistent bool
}

func NewClonedRepository(os System, cloneURL *url.URL) *ClonedRepository {
	return &ClonedRepository{
		os:       os,
		cloneURL: cloneURL,
	}
}

func (self *ClonedRepository) UsesSSH() bool {
	return strings.Contains(strings.ToLower(self.cloneURL.Scheme), "ssh")
}

func (self *ClonedRepository) UsesHTTP() bool {
	return strings.HasPrefix(strings.ToLower(self.cloneURL.Scheme), "http")
}

func (self *ClonedRepository) Remove() error {
	if self.tempDir != "" {
		return self.os.DeleteFile(self.tempDir)
	}

	return nil
}

func (self *ClonedRepository) MakePersistent() {
	self.isPersistent = true
}

func (self *ClonedRepository) Pull() error {
	if err := self.Clone(); err != nil {
		return err
	}

	gitReset := NewSystemCommand("git", "reset", "--hard").
		WorkingDirectory(self.clonedInto).
		SetEnv("GIT_PAGER", "/usr/bin/cat").
		SetEnv("GIT_CONFIG_NOSYSTEM", "true").
		SetEnv("GIT_ASKPASS", "/bin/echo").
		SetEnv("GIT_SSH", filepath.Join(self.tempDir, "git-ssh"))

	if output, err := self.os.Run(gitReset); err != nil {
		return fmt.Errorf("Reset: %s\n%s\n%s\n", gitReset, output, err)
	}

	gitPull := NewSystemCommand("git", "pull").
		WorkingDirectory(self.clonedInto).
		SetEnv("GIT_PAGER", "/usr/bin/cat").
		SetEnv("GIT_CONFIG_NOSYSTEM", "true").
		SetEnv("GIT_ASKPASS", "/bin/echo").
		SetEnv("GIT_SSH", filepath.Join(self.tempDir, "git-ssh"))

	if output, err := self.os.Run(gitPull); err != nil {
		return fmt.Errorf("Pull: %s\n%s\n%s\n", gitPull, output, err)
	}

	return nil
}

func (self *ClonedRepository) FetchMetadata() (*RepositoryMetadata, error) {
	refs, err := self.References()
	if err != nil {
		return nil, err
	}

	result := NewRepositoryMetadata()
	for _, ref := range refs {
		result.AddRef(ref)
	}

	contributors, err := self.Contributors()
	if err != nil {
		return nil, err
	}

	for _, contributor := range contributors {
		result.AddContributor(contributor)
	}

	return result, nil
}

func (self *ClonedRepository) Contributors() ([]*Person, error) {
	if err := self.Clone(); err != nil {
		return nil, err
	}

	listContributors := NewSystemCommand("git", "log", "--format=%aN\x1e%aE").
		WorkingDirectory(self.clonedInto)

	output, err := self.os.Run(listContributors)
	if err != nil {
		return nil, fmt.Errorf("Contributors: %s\n%s\n%s\n", listContributors, output, err)
	}

	result := []*Person{}
	seen := map[string]bool{}
	lines := bufio.NewScanner(bytes.NewBuffer(output))
	for lines.Scan() {
		fields := strings.Split(lines.Text(), "\x1e")
		name := fields[0]
		email := fields[1]

		if seen[email] {
			continue
		}
		seen[email] = true

		result = append(result, &Person{
			Name:  name,
			Email: email,
		})
	}

	return result, nil
}

func (self *ClonedRepository) IsAccessible() (bool, error) {
	if err := self.ensureTempDir(); err != nil {
		return false, err
	}

	lsRemote := NewSystemCommand("git", "ls-remote", self.cloneURL.String()).
		WorkingDirectory(self.tempDir).
		SetEnv("GIT_PAGER", "/usr/bin/cat").
		SetEnv("GIT_CONFIG_NOSYSTEM", "true").
		SetEnv("GIT_ASKPASS", "/bin/echo").
		SetEnv("GIT_SSH", filepath.Join(self.tempDir, "git-ssh"))

	if _, err := self.os.Run(lsRemote); err != nil {
		return false, err
	}

	return true, nil
}

func (self *ClonedRepository) SetCredential(credential Credential) error {
	self.credential = credential
	if credential.Protocol() != self.cloneURL.Scheme {
		return fmt.Errorf("cannot use %q credential for %q URL", credential.Protocol(), self.cloneURL.Scheme)
	}

	switch protocol := credential.Protocol(); protocol {
	case "https":
		self.cloneURL.User = url.UserPassword(credential.Username(), credential.Secret())
	case "ssh":
		self.cloneURL.User = url.User(credential.Username())
		if err := self.ensureTempDir(); err != nil {
			return err
		}

		sshCredentialPath, err := self.writeSSHCredential(credential)
		if err != nil {
			return err
		}

		if err := self.writeGitSSHWrapper(sshCredentialPath); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported protocol %q", protocol)
	}

	return nil
}

func (self *ClonedRepository) References() ([]*Reference, error) {
	if err := self.Clone(); err != nil {
		return nil, err
	}

	lsRemote := NewSystemCommand("git", "ls-remote", "repository").
		WorkingDirectory(self.tempDir)
	refs, err := self.os.Run(lsRemote)
	if err != nil {
		return nil, fmt.Errorf("References: %s\n%s\n%s\n", lsRemote, refs, err)
	}

	result := []*Reference{}
	errorLines := []string{}
	lines := bufio.NewScanner(bytes.NewBuffer(refs))
	for lines.Scan() {
		fields := strings.Fields(lines.Text())
		if len(fields) != 2 {
			errorLines = append(errorLines, lines.Text())
		} else {
			result = append(result, &Reference{
				Hash: strings.TrimSpace(fields[0]),
				Name: strings.TrimSpace(fields[1]),
			})
		}
	}

	if len(errorLines) > 0 {
		err = fmt.Errorf("failed to parse:\n%s", strings.Join(errorLines, "\n"))
	} else {
		err = nil
	}

	return result, err
}

func (self *ClonedRepository) Clone() error {
	if self.clonedInto != "" {
		return nil
	}

	if err := self.ensureTempDir(); err != nil {
		return err
	}

	gitRevParse := NewSystemCommand("git", "rev-parse", "--git-dir").
		WorkingDirectory(filepath.Join(self.tempDir, "repository"))
	if output, err := self.os.Run(gitRevParse); err == nil && strings.HasSuffix(string(output), ".git\n") {
		self.clonedInto = filepath.Join(self.tempDir, "repository")
		return nil
	}

	gitClone := NewSystemCommand("git", "clone", self.cloneURL.String(), "repository").
		WorkingDirectory(self.tempDir).
		SetEnv("GIT_PAGER", "/usr/bin/cat").
		SetEnv("GIT_CONFIG_NOSYSTEM", "true").
		SetEnv("GIT_ASKPASS", "/bin/echo").
		SetEnv("GIT_SSH", filepath.Join(self.tempDir, "git-ssh"))

	output, err := self.os.Run(gitClone)
	if err != nil {
		return fmt.Errorf("Clone: %s\n%s\n%s\n", gitClone, output, err)
	}
	self.clonedInto = filepath.Join(self.tempDir, "repository")

	return nil
}

func (self *ClonedRepository) ensureTempDir() error {
	if self.tempDir != "" {
		return nil
	}

	tempDir := ""
	err := (error)(nil)
	if self.isPersistent {
		dirname := fmt.Sprintf("%s%s", self.cloneURL.Host, self.cloneURL.Path)
		if strings.HasSuffix(dirname, ".git") {
			dirname = dirname[0 : len(dirname)-len(".git")]
		}
		dirname = strings.Replace(dirname, "_", "__", -1)
		dirname = strings.Replace(dirname, "-", "--", -1)
		dirname = strings.Replace(dirname, ".", "_", -1)
		dirname = strings.Replace(dirname, "/", "-", -1)
		tempDir, err = self.os.PersistentDir(dirname)
	} else {
		tempDir, err = self.os.TempDir()
	}
	if err != nil {
		return err
	}

	self.tempDir = tempDir
	return nil
}

func (self *ClonedRepository) writeSSHCredential(credential Credential) (string, error) {
	sshCredentialPath := filepath.Join(self.tempDir, "ssh_credential")
	sshCredential, err := self.os.CreateFile(sshCredentialPath)
	if err != nil {
		return "", err
	}

	if _, err := fmt.Fprintf(sshCredential, credential.Secret()); err != nil {
		return "", err
	}
	if err := sshCredential.Close(); err != nil {
		return "", err
	}

	if err := self.os.SetPermissions(sshCredentialPath, 0600); err != nil {
		return "", err
	}

	return sshCredentialPath, nil
}

func (self *ClonedRepository) writeGitSSHWrapper(sshCredentialPath string) error {
	gitSSHWrapperName := filepath.Join(self.tempDir, "git-ssh")
	gitSSHWrapper, err := self.os.CreateFile(gitSSHWrapperName)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(gitSSHWrapper, sshWrapperFormatTpl, "/usr/bin/ssh", sshCredentialPath); err != nil {
		return err
	}
	if err := gitSSHWrapper.Close(); err != nil {
		return err
	}

	if err := self.os.SetPermissions(gitSSHWrapperName, 0755); err != nil {
		return err
	}

	return nil
}
