package domain

import (
	"fmt"
	"strings"
	"testing"

	"github.com/harrowio/harrow/git"
)

func Test_Repository_NewMetadataUpdateOperation_returnsOperationForUpdatingMetadata(t *testing.T) {
	repo := &Repository{
		Uuid: "7ce4efb9-5600-490d-9ec6-3c2737083d01",
	}

	operation := repo.NewMetadataUpdateOperation()
	if got := operation; got == nil {
		t.Fatalf("operation is nil")
	}

	if got, want := operation.Type, OperationTypeGitEnumerationBranches; got != want {
		t.Errorf(`operation.Type = %v; want %v`, got, want)
	}

	if got, want := *operation.RepositoryUuid, repo.Uuid; got != want {
		t.Errorf(`*operation.RepositoryUuid = %v; want %v`, got, want)
	}

	if got, want := operation.WorkspaceBaseImageUuid, "31b0127a-6d63-4d22-b32b-e1cfc04f4007"; got != want {
		t.Errorf(`operation.WorkspaceBaseImageUuid = %v; want %v`, got, want)
	}
}

// Test_Repository_ValidationOfKnownBadUrls all examples
// taken from production data set.
func Test_Repository_ValidationOfKnownBadUrls(t *testing.T) {
	var cases = []struct {
		rawurl string
		err    error
	}{
		// {err: git.ErrURLHostMissing, rawurl: "scm:git:git@github.com:hjuergens/date-parser.git"}, # remarkably this parses as a url containing a username and password!
		{err: git.ErrURLPathMissing, rawurl: "pawnshop.git@deploy.myservice.kg"},
	}
	for _, tc := range cases {
		err := ValidateRepository(&Repository{Url: tc.rawurl, ProjectUuid: "31b0127a-6d63-4d22-b32b-e1cfc04f4007"})
		if err == nil {
			t.Fatalf("Got no validation error, but one was expected for URL %s\n", tc.rawurl)
		}
		validationErr, ok := err.(*ValidationError)
		if !ok {
			t.Log("Something fucky with the type assertion")
			t.Fatalf("Got error that wasn't a validation error: %s\n", err)
		}
		if strings.Compare(validationErr.Get("url"), tc.err.Error()) != 0 {
			t.Fatalf("got: %s want %s (case: %s)", validationErr.Get("url"), tc.err, tc.rawurl)
		}
	}
}

func Test_Repository_ValidationOfKnownGoodUrls(t *testing.T) {
	var urls = []string{
		"git@github.com:capistrano/capistrano.git",
		"ssh://git@github.com:capistrano/capistrano.git",
		"ssh://git@69.73.176.35:7978/pedro/bridgeloannetwork.com.git",
		"https://github.com/tools/godep.git",
	}
	for _, url := range urls {
		repo := &Repository{Url: url, ProjectUuid: "31b0127a-6d63-4d22-b32b-e1cfc04f4007"}
		if err := ValidateRepository(repo); err != nil {
			t.Errorf("Expected no error parsing repo url, got", err)
		}
	}
}
func Test_Repository_CloneURLHTTPS(t *testing.T) {
	rawurl := "https://user:pass@github.com/login/repo.git"
	repo, _ := git.NewRepository(rawurl)
	if repo.CloneURL() != rawurl {
		t.Fatalf("expected %q got %q", rawurl, repo.CloneURL())
	}
}

func Test_Repository_CloneURLSSHMatchesSSHHostAlias(t *testing.T) {
	rawurl := "ssh://user:pass@github.com/login/repo.git"
	repo, _ := git.NewRepository(rawurl)
	want := fmt.Sprintf("%s:%s", repo.SSHHostAlias(), repo.URL.Path)
	if repo.CloneURL() != want {
		t.Fatalf("expected %q got %q", want, repo.CloneURL())
	}
}
