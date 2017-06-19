package git

import (
	"fmt"
	"testing"
)

func Test_Repository_SSHHostAlias(t *testing.T) {
	want := "ssh___github.com_foo_bar.git"
	repo, _ := NewRepository("git@github.com:foo/bar.git")
	sshHostAlias := repo.SSHHostAlias()
	if want != sshHostAlias {
		t.Fatalf("expected %q got %q", want, sshHostAlias)
	}
}

func Test_Repository_ProtocolSchemelessURLAssumedSsh(t *testing.T) {
	repo, err := NewRepository(repoSSHCloneURLWithoutUsername)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}
	if repo.UsesSSH() != true {
		t.Fatalf("expected %q, got %q", true, repo.UsesSSH())
	}
}

func Test_Repository_DefaultUsernameToGitSSHUrl(t *testing.T) {
	repo, err := NewRepository(repoSSHCloneURLWithoutUsername)
	if err != nil {
		t.Fatalf("expected %q, got %q", nil, err)
	}
	if "git" != repo.Username() {
		t.Fatalf("expected %q, got %q", "git", repo.Username())
	}
}

func Test_Repository_SSHHost(t *testing.T) {
	want := "github.com"
	repo, _ := NewRepository("git@github.com:foo/bar.git")
	sshHost := repo.SSHHost()
	if want != sshHost {
		t.Fatalf("expected %q got %q", want, sshHost)
	}
}

func Test_Repository_SSHPort(t *testing.T) {
	want := "1234"
	repo, _ := NewRepository("ssh://git@github.com:1234/foo/bar.git")
	sshPort := repo.SSHPort()
	if want != sshPort {
		t.Fatalf("expected %q got %q", want, sshPort)
	}
}

func Test_Repository_Username(t *testing.T) {
	want := "git"
	repo, _ := NewRepository("git@github.com:foo/bar.git")
	username := repo.Username()
	if want != username {
		t.Fatalf("expected %q got %q", want, username)
	}
}

func Test_Repository_CloneURLHTTPS(t *testing.T) {
	rawurl := "https://user:pass@github.com/login/repo.git"
	repo, _ := NewRepository(rawurl)
	if repo.CloneURL() != rawurl {
		t.Fatalf("expected %q got %q", rawurl, repo.CloneURL())
	}
}

func Test_Repository_CloneURLSSHMatchesSSHHostAlias(t *testing.T) {
	rawurl := "ssh://user:pass@github.com/login/repo.git"
	repo, _ := NewRepository(rawurl)
	want := fmt.Sprintf("%s:%s", repo.SSHHostAlias(), repo.URL.Path)
	if repo.CloneURL() != want {
		t.Fatalf("expected %q got %q", want, repo.CloneURL())
	}
}

func Test_Repository_FetchMetadata_returns_refs(t *testing.T) {
	t.Skip("Integration tests disabled")
	rawurl := "https://github.com/harrowio/repository-without-master-branch.git"
	repo, err := NewRepository(rawurl)
	if err != nil {
		t.Fatal(err)
	}

	metadata, err := repo.FetchMetadata()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(metadata.Refs), 0; got == want {
		t.Fatalf(`len(metadata.Refs) = %v; want not %v`, got, want)
	}
}

func Test_Repository_FetchMetadata_returns_contributors(t *testing.T) {
	t.Skip("Integration tests disabled")
	rawurl := "https://github.com/harrowio/repository-without-master-branch.git"
	repo, err := NewRepository(rawurl)
	if err != nil {
		t.Fatal(err)
	}

	metadata, err := repo.FetchMetadata()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(metadata.Contributors), 0; got == want {
		t.Fatalf(`len(metadata.Contributors) = %v; want not %v`, got, want)
	}

	for email, contributor := range metadata.Contributors {
		if got, want := email, ""; got == want {
			t.Errorf(`email = %v; want not %v`, got, want)
			continue
		}

		if got, want := contributor.Name, ""; got == want {
			t.Errorf(`contributor[%s].Name = %v; want not %v`, email, got, want)
		}

		if got, want := contributor.Email, ""; got == want {
			t.Errorf(`contributor[%s].Email = %v; want not %v`, email, got, want)
		}
	}

}
