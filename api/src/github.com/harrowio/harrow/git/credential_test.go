package git

import (
	"strings"
	"testing"
)

// TODO: Many of these tests belong in the general git/repository_test.go file

var (
	httpUsername                   = "harrowio"
	httpPassword                   = "AqO0Kg43uqZldqMXkM9Lf4thxZrTqObL"
	unsupportedGitProtocolUrl      = "git://localhost"
	repoSSHCloneUrlWithUsername    = "git@bitbucket.org:harrowio/mdpf-intergration-test.git"
	repoSSHCloneURLWithoutUsername = "ssh://bitbucket.org:harrowio/mdpf-intergration-test.git"
	repoHttpCloneURLWithoutCreds   = "https://github.com/harrowio/mdpf-intergration-test.git"
	repoHttpCloneURLWithCreds      = "https://max:mustermann@github.com/harrowio/mdpf-intergration-test.git"
	repoUrlWithIllegalCharacters   = "https://über dangerous url ⚠"
	sshPublicKeyFingerprint        = "SHA256:xdWsSgaQ8aVAgaEc+sWmgasWNa30C/MaoJYH4PI7M0A="
)

var sshPrivateKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEApX2FBBzzJaWoq4dv/1Khk1yeGSPzHskyXdwoTQB4lICwCT51
84To2gfrggDqvRSAE/K4jhfY4SBIb85smzAtdLxpKP8cfB/l0e4Hva8kPBBFrkMS
WPAJs2JaDt1lvI+dMONA+SPmFvIsOhi+h4wmXAwH3XuZDQ6YdvIxz333ogMDFokM
eBHcqYlj89w+qjEvMX2U9why9O3JXWJ7TD1DItyiBwSIOOE6dLjs7ZWFCNSwCC39
GSyPFQYKiHRaLzn4vwv02xXVcOGw9JmYifRCibQVf0zQ14sKOAaxPhEIaAwMApiV
mOkMlBAaPTnLnCMHVq2tO4CCiDAjvAfWcJQEwQIDAQABAoIBABYrtsJSTpDgnLQ+
NNbz7wmbAuNDWbLqKYFBmXSXd5ANnYffglXZnIh5Pyfvj4M9V9tUTT1cHIYsmQfB
k/NGhRB6nWwMoXhPna5+QTM8X5Jca7lo6vBXWDVcG8yaBKM6aki+aVn3YT/5ucse
vYfTUuBKDFOz11FUf0CQOfQeYCqoHX8mb3UMMNlKP2RtIE+x7l69llc7W+6NMFn2
Z56J7n/WBRZcrRJofS24ckgH3oFNDSr7OW1pLZvkyAO80Pzo992txl0DlzPzXVKp
eTs3C84sEzdxAPbeFmIOAUaE5S+12P7sKKdE1AcJ4WPjlbgfGA8I2aTM1wZo58gC
cA9qPQECgYEAznuS80BRp/YJciX9m4IvnI971DobUgXGDq10HBGWzIBNTMdQr2DH
SKCJcnTciTDqhzuu6U4vRTMJgOTopoMtXDkMJB5Yp/YZHTr68FJd1r0BYPas4ThR
+NhJXCZIlkeWhTGJyVD1O6VsjrA55nnq/9ktPy4IXYSPgccPM/tNpTkCgYEAzS1V
EysMbP1aIw9mXyNGjof/+NAA+tuHZZKxj/PPuD7TxGKF6W6fnPmxST2HdZHVcbDY
DZfUf0R30R37S1sNSd+FtDLTrkBHFQHZF1gecFAGhVArh1wSiz8wWQsB1fF8YxsV
V2+QHn6AxvrV3KDamrrwQc1eJDg6oqFIo826o8kCgYBfQ/p3yrwh701KYibRQc5v
wG+Uaj7CqDFKAlMoxCC8N5Hyk58xW0h2xMLFkQ9TKMN8I1g/AjijB7ohwvtoH+uk
uhlU7L9gtxW9O8IdcRMkiU2CjC0VOGPxmPC32F3zIBJdX46/2F9c5qTgbIQ6RxPa
eTv8A2QOqaOAb/QeupqHWQKBgGV8Jrh0cpD2P79XvqsQJ7YYTuQi/lkWfMIg7PLn
Bbd8XAKnONVdglWCq84uQPJGT+0MK9GNZ+4LT7h/u+xp/QitJtUaztlBsecSIu8J
BwVGj/Mg1Gb/g6ycdK2WZDIOYBglLUkyRXbP26KQL3gRmA8wp+XkTsxbg6UtYWCk
Qc0BAoGAXKYW2hEfm4hBrYHpM3BL3ifkz/oTvGdckT95SYxSw3aNqC66mTCSV9Zb
PomXRNwaHEfXDnb0X4rcJCyJzBxaMdN+SmHGFOWXXwp5k2MtH0YulLXohj0hwNRO
DjGUcn59iy/VoVx5jCo5+Fikgw7dyep1ao0Ayly9XOAK+HbDk70=
-----END RSA PRIVATE KEY-----`

func Test_Credential_RepoWithUsernameNotOverwrittenWhenSettingCredentialWithoutUsername(t *testing.T) {
	repo, err := NewRepository(repoHttpCloneURLWithCreds)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}
	cred, _ := NewHTTPCredential("", "")
	err = repo.SetCredential(cred)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}
	if repo.Username() != "max" {
		t.Fatalf("expected %q, got %q", "max", repo.Username())
	}
}

func Test_Credential_RepoWithHttpURLWithoutPasswordGetsPasswordFromCredential(t *testing.T) {
	repo, err := NewRepository(repoHttpCloneURLWithoutCreds)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}
	cred, _ := NewHTTPCredential("", "sentinel")
	err = repo.SetCredential(cred)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}
	if !strings.Contains(repo.URL.String(), "sentinel") {
		t.Fatalf("expected %q to contain the string %q", repo.URL.String(), "sentinel")
	}
}

func Test_Credential_RepoWithoutUsernameOverwrittenWhenSettingCredentialWithUsername(t *testing.T) {
	repo, err := NewRepository(repoHttpCloneURLWithoutCreds)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}
	cred, _ := NewHTTPCredential("user", "")
	err = repo.SetCredential(cred)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}
	if repo.Username() != "user" {
		t.Fatalf("expected %q, got %q", "user", repo.Username())
	}
}

func Test_Credential_NewSshCredential_ErrUnparsablePrivateKey(t *testing.T) {
	_, err := NewSshCredential("", []byte("nyan nyan"))
	if err != ErrUnparsablePrivateKey {
		t.Fatalf("expected %q, got %q", ErrUnparsablePrivateKey, err)
	}
}

func Test_Credential_sshCredentialStringIncludesPublicKeyFingerprint(t *testing.T) {
	credential, err := NewSshCredential("", []byte(sshPrivateKeyPEM))
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}
	if !strings.Contains(credential.String(), sshPublicKeyFingerprint) {
		t.Fatalf("expected %q in %q", sshPublicKeyFingerprint, credential.String())
	}
}

func Test_Credential_NoErrorOnCompatibleCredentialType(t *testing.T) {
	repo, err := NewRepository(repoHttpCloneURLWithoutCreds)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}
	cred, _ := NewHTTPCredential("user", "pass")
	err = repo.SetCredential(cred)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}
}

func Test_Credential_ErrorOnIncompatibleCredentialType(t *testing.T) {
	repo, err := NewRepository(repoSSHCloneUrlWithUsername)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}
	cred, _ := NewHTTPCredential("user", "pass")
	err = repo.SetCredential(cred)
	if err != ErrIncompatibleCredentialType {
		t.Fatalf("expected %q, got %q", ErrIncompatibleCredentialType, err)
	}
}

func Test_Credential_NoDefaultUsernameHttpUrl(t *testing.T) {
	repo, err := NewRepository(repoHttpCloneURLWithoutCreds)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}
	if len(repo.Username()) != 0 {
		t.Fatalf("expected %q, got %q", 0, len(repo.Username()))
	}
}

func Test_Credential_InferringCorrectUsernameSSHUrl(t *testing.T) {
	repo, err := NewRepository(repoSSHCloneURLWithoutUsername)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}
	if repo.Username() != "git" {
		t.Fatalf("expected %q, got %q", "git", repo.Username())
	}
}

func Test_Credential_AccessingPrivateRepositorySuccessfullyOverSSH(t *testing.T) {

	t.Skip("Integration tests disabled")

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	repo, err := NewRepository(repoSSHCloneUrlWithUsername)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}

	credential, err := NewSshCredential(repo.Username(), []byte(sshPrivateKeyPEM))
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}

	repo.SetCredential(credential)

	accessible := repo.IsAccessible()
	if !accessible {
		t.Fatalf("expected %q, got %q", true, accessible)
	}

	_, err = repo.References()
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}
}

func Test_Credential_AccessingPrivateRepositorySuccessfullyOverHttp(t *testing.T) {

	t.Skip("Integration tests disabled")

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	repo, err := NewRepository(repoHttpCloneURLWithoutCreds)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}

	credential, err := NewHTTPCredential(httpUsername, httpPassword)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}

	repo.SetCredential(credential)

	accessible := repo.IsAccessible()
	if !accessible {
		t.Fatalf("expected %q, got %q", true, accessible)
	}

	_, err = repo.References()
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}
}

func Test_Credential_FailingToAccessPrivateRepoSuccessfullyOverSSHWithoutCredentials(t *testing.T) {

	t.Skip("Integration tests disabled")

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	repo, err := NewRepository(repoHttpCloneURLWithoutCreds)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}

	accessible := repo.IsAccessible()
	if accessible {
		t.Fatalf("expected %q, got %q", false, accessible)
	}
}

func Test_Credential_FailingToAccessPrivateRepoSuccessfullyOverHttpWithoutCredentials(t *testing.T) {

	t.Skip("Integration tests disabled")

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	repo, err := NewRepository(repoSSHCloneUrlWithUsername)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}

	accessible := repo.IsAccessible()
	if accessible {
		t.Fatalf("expected %q, got %q", false, accessible)
	}
}
