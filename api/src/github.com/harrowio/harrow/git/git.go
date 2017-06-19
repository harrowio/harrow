// Package git provides access to git repository metadata such as refs,
// authors, branches and commits.
//
// The package goes to some length to intelligently handle URLs, and be
// wise about protocol schemes.
//
// Since URLs can embed credentials (username, password) the package
// attempts to parse them and handle them correctly. Incase of SSH-like
// protocol schemes (handled by giturl package) an intelligent guess of
// `git` as a username is made.
//
// Calls to SetCredential mutate the given URL if the URL protocol scheme
// is http(s) and the given credential is compatible. Hence calls to
// NewRepository() with a URL sans credentials, followed by a call to
// SetCredential() with an httpCredential instance is equivilent to having
// called NewRepository() with the credentials embedded in the URL
package git

import "errors"

var (
	ErrSshExectableNotFound            = errors.New("git: ssh executable not found")
	ErrCatExectableNotFound            = errors.New("git: ssh executable not found")
	ErrGitExecutableNonZeroReturn      = errors.New("git: git executable returned non-zero")
	ErrGitExecutableNotFound           = errors.New("git: git executable not found")
	ErrIllegalCharactersInURL          = errors.New("git: illegal characters in url")
	ErrIncompatibleCredentialType      = errors.New("git: credential type is not compatible with the url protocol scheme")
	ErrNotImplemented                  = errors.New("git: not implemented")
	ErrUnparsablePrivateKey            = errors.New("git: unparsable private key")
	ErrUnparsableURL                   = errors.New("git: unparsable url")
	ErrWritingSshWrapperToDisk         = errors.New("git: error writing ssh wrapper to disk")
	ErrWritingSshPrivateKeyToDisk      = errors.New("git: error writing ssh private key to disk")
	ErrRemovingSensitveFiles           = errors.New("git: error removing sensitive files")
	ErrWritingHttpsCredentialCacheFile = errors.New("git: error writing https credential cache helper file to disk")
	ErrCallingGitConfig                = errors.New("git: error calling `git config'")
)
