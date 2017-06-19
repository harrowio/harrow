package git

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/harrowio/harrow/posix"
)

type Repository struct {
	URL        *URL
	credential Credential
}

var (
	DefaultSSHUserInfo = url.User("git")
)

// NewRepository returns a new instance of repository.  The URL given must
// be parsable, else an error will be returned.  If the URL includes a
// username and password these will be stored and available to other calls
// which may need such information without requring a call to
// repository.SetCredential()
func NewRepository(u string) (*Repository, error) {

	gu, err := Parse(u)
	if err != nil {
		return nil, err
	}

	repo := &Repository{URL: gu}

	if repo.UsesSSH() && repo.URL.User == nil {
		repo.URL.User = DefaultSSHUserInfo
	}

	return repo, nil
}

// IsAccessible returns true if references can be collected
// without error (self.References())
func (self *Repository) IsAccessible() bool {
	_, err := self.References()
	return err == nil
}

func (self *Repository) SSHHost() string {
	var host string = self.URL.Host
	if h, _, err := net.SplitHostPort(self.URL.Host); err == nil {
		host = h
	}
	return host
}

func (self *Repository) SSHPort() string {
	var port string
	_, port, _ = net.SplitHostPort(self.URL.Host)
	return port
}

// SSHHostAlias returns something usable as an SSH alias in
// ~/.ssh/config for clones over SSH.
func (self *Repository) SSHHostAlias() string {
	urlCpy := self.URL.Copy()
	urlCpy.User = nil
	return strings.Map(posix.SafeStr, urlCpy.String())
}

// References runs the equivalent of a `git ls-remote` returning the
// complete unabridge list of all remote references.
func (self *Repository) References() ([]*Reference, error) {
	return self.lsRemote()
}

// lsRemote parses the URL (string) without credentials to make a safe
// copy for mutation, it then injects the username and password from the
// credential. If the credential has an empty string as the username or
// password the original parsed/inferred username and password are
// retained. According to
// https://git-scm.com/book/en/v2/Git-Commands-Plumbing-Commands
// ls-remote is considered plumbing, not porcelain.  lsRemote takes a
// slice of env variables in the format `KEY=value` consistent with
// os.Environ()
func (self *Repository) lsRemote() ([]*Reference, error) {

	gitExecutable, err := exec.LookPath("git")
	if err != nil {
		return nil, ErrGitExecutableNotFound
	}

	catExecutable, err := exec.LookPath("cat")
	if err != nil {
		return nil, ErrCatExectableNotFound
	}

	sshExecutable, err := exec.LookPath("ssh")
	if err != nil {
		return nil, ErrSshExectableNotFound
	}

	var env []string = []string{
		"GIT_ASKPASS=/bin/echo",
		"GIT_CONFIG_NOSYSTEM=true",
		fmt.Sprintf("GIT_PAGER=%s", catExecutable),
	}

	var lsRemoteCmd *exec.Cmd
	if self.UsesHTTP() {
		// Incase we use https, the hostname will be used to ask the credential
		// helper for the credentials. Use of the credential store mitigates the risk of
		// shell injection commands.
		// [1]: https://git-scm.com/book/en/v2/Git-Tools-Credential-Storage
		// TODO: Fix me not to use the URL with creds embedded (or, check if url.URL.String() embeds creds)
		lsRemoteCmd = exec.Command(gitExecutable, "ls-remote", self.URL.String())
	} else {
		lsRemoteCmd = exec.Command(gitExecutable, "ls-remote", self.URL.String())
	}

	// NOTE: This should move to httpsCredential, however that will require a
	// change to the way the username/password parsing works, never to modify
	// repo.URL but always to create/maintain a repo.credential.(httpsCredential)
	// so that the repo.URL always stays free of credentials, and if given, they
	// are moved into an implicitly created Credential.
	if self.UsesHTTP() && self.URL.User != nil {

		// 1. Setup temp directory, and a tempfile in that directory
		tempDir, err := ioutil.TempDir("", "httpsCredential")
		if err != nil {
			return nil, ErrWritingHttpsCredentialCacheFile
		}
		defer os.RemoveAll(tempDir)

		tempFile, err := ioutil.TempFile(tempDir, "httpsCredentialHelperCache")
		if err != nil {
			return nil, ErrWritingHttpsCredentialCacheFile
		}
		err = os.Mkdir(filepath.Join(tempDir, ".git"), 0700)
		if err != nil {
			return nil, ErrWritingHttpsCredentialCacheFile
		}

		if _, err := tempFile.WriteString(self.URL.String()); err != nil {
			return nil, ErrWritingHttpsCredentialCacheFile
		}

		if err := os.Chmod(tempFile.Name(), 0600); err != nil {
			return nil, ErrWritingHttpsCredentialCacheFile
		}
		if err := tempFile.Close(); err != nil {
			return nil, ErrWritingHttpsCredentialCacheFile
		}

		// 2. Making commands and changing their directories (because `git config')
		storeArg := fmt.Sprintf("store --file %s", tempFile.Name())
		gitConfigCmd := exec.Command(gitExecutable, "config", "credential.helper", storeArg)
		gitConfigCmd.Dir = tempDir
		lsRemoteCmd.Dir = tempDir

		err = gitConfigCmd.Run()
		if err != nil {
			// NOTE: err might be a exec.ExitErr which has a .Stderr
			// property which would be populated, we might use that
			return nil, ErrCallingGitConfig
		}

	}

	if self.UsesSSH() && self.credential != nil {
		if sshCredential, ok := self.credential.(*sshCredential); ok {
			sensitiveFiles, err := sshCredential.writeWrapperScriptToDisk(sshExecutable)
			defer sensitiveFiles.Remove()
			if err != nil {
				// NOTE: is this error handling insane, I don't want to leak
				// errors from a private interface, but I would like to log
				// them here for convenience.
				if err != ErrWritingSshWrapperToDisk {
					return nil, ErrWritingSshPrivateKeyToDisk
				}
				return nil, err
			}
			env = append(env, fmt.Sprintf("GIT_SSH=%s", sensitiveFiles.wrapper))
		}
	}

	// Run the ls-remote with a restricted environment, only keys in the `env'
	// slice are here (so, about 5x GIT_... variables are present.) The paths to
	// all executables are fixed, and Git is able to infer the value for
	// GIT_EXEC_PATH without a fully-populated PATH
	lsRemoteCmd.Env = env

	output, err := lsRemoteCmd.Output()
	if err != nil {
		// NOTE: err might be a exec.ExitErr which has a .Stderr
		// property which would be populated, we might use that
		return nil, ErrGitExecutableNonZeroReturn
	}

	var refs []*Reference
	for _, l := range strings.Split(string(output), "\n") {
		if len(l) == 0 {
			continue
		}
		fields := strings.Fields(l)
		refs = append(refs, &Reference{Hash: fields[0], Name: fields[1]})
	}

	return refs, nil
}

func (self *Repository) SetCredential(c Credential) error {
	if _, ok := c.(*httpCredential); !ok && self.UsesHTTP() {
		return ErrIncompatibleCredentialType
	}
	if _, ok := c.(*sshCredential); !ok && self.UsesSSH() {
		return ErrIncompatibleCredentialType
	}
	if len(self.Username()) == 0 && len(c.Username()) > 0 {
		self.URL.User = url.User(c.Username())
	}
	if httpCredential, ok := c.(*httpCredential); ok {
		self.URL.User = url.UserPassword(self.Username(), httpCredential.Password())
	}
	self.credential = c
	return nil
}

func (self *Repository) UsesSSH() bool {
	return strings.Contains(strings.ToLower(self.URL.Scheme), "ssh")
}

func (self *Repository) UsesHTTP() bool {
	return strings.HasPrefix(strings.ToLower(self.URL.Scheme), "http")
}

func (self *Repository) Username() string {
	if self.URL.User != nil {
		return self.URL.User.Username()
	} else {
		return ""
	}
}

// Contributors returns the list of contributors to this repository.
// In order to do its work, it requires a clone of the repository.  If
// this repository has not been cloned yet, this method attempts to
// clone the repository first before extracting information from it.
func (self *Repository) Contributors() ([]*Person, error) {
	return nil, errors.New("not implemented")
}

// CloneURL returns something we can pass to `ls-remote` or similar it is
// mostly used in the older code (rootfs builder) where the newer `Credential`
// and `git.Repository` stuff is more modern.  In the SSH case it returns the
// host alias with the clone path, assuming that a corresponding entry is made
// somewhere else in a `man (1) ssh_config`.
func (self *Repository) CloneURL() string {
	if self.UsesSSH() {
		return fmt.Sprintf("%s:%s", self.SSHHostAlias(), self.URL.Path)
	} else {
		return self.URL.String()
	}
}

// FetchMetadata gathers metadata about the repository.  Metadata
// includes all refs and what they point to as well as contributors
// (identified by their email address) to the repository.
func (self *Repository) FetchMetadata() (*RepositoryMetadata, error) {
	result := NewRepositoryMetadata()
	refs, err := self.References()
	if err != nil {
		return nil, err
	}
	for _, ref := range refs {
		result.AddRef(ref)
	}

	contributors, err := self.Contributors()
	for _, contributor := range contributors {
		result.AddContributor(contributor)
	}

	return result, nil
}
