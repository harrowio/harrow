package git

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func cloneURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   "example.com",
		Path:   "/repository.git",
	}
}

func usernameAndPassword() Credential {
	credential, err := NewHTTPCredential("john-doe", "super-secret")
	if err != nil {
		panic(err)
	}
	return credential
}

func sshUserAndSecret() Credential {
	credential, err := NewSshCredential("git", []byte(sshPrivateKeyPEM))
	if err != nil {
		panic(err)
	}

	return credential
}

func TestClonedRepository_UsesHTTP_returns_true_if_URL_scheme_starts_with_http(t *testing.T) {
	repo := NewClonedRepository(NewMockSystem(), &url.URL{
		Scheme: "https",
		Host:   "example.com",
	})

	if got, want := repo.UsesHTTP(), true; got != want {
		t.Errorf(`repo.UsesHTTP() = %v; want %v`, got, want)
	}
}

func TestClonedRepository_UsesSSH_returns_true_if_URL_scheme_starts_with_ssh(t *testing.T) {
	repo := NewClonedRepository(NewMockSystem(), &url.URL{
		Scheme: "ssh",
		Host:   "example.com",
	})

	if got, want := repo.UsesSSH(), true; got != want {
		t.Errorf(`repo.UsesSSH() = %v; want %v`, got, want)
	}
}

func TestClonedRepository_SetCredential_returns_error_if_setting_a_credential_for_a_different_protocol(t *testing.T) {
	mockSystem := NewMockSystem()
	repo := NewClonedRepository(mockSystem, cloneURL())

	if err := repo.SetCredential(sshUserAndSecret()); err == nil {
		t.Fatal("expected an error")
	} else {
		if got, want := err.Error(), `cannot use "ssh" credential for "https" URL`; got != want {
			t.Fatalf(`err.Error() = %v; want %v`, got, want)
		}
	}
}

func TestClonedRepository_Clone_clones_into_temporary_directory(t *testing.T) {
	mockSystem := NewMockSystem()
	repo := NewClonedRepository(mockSystem, cloneURL())

	if err := repo.Clone(); err != nil {
		t.Fatal(err)
	}

	if got, want := len(mockSystem.tempDirs), 1; got != want {
		t.Fatalf(`len(mockSystem.tempDirs) = %v; want %v`, got, want)
	}

	expectedCommand := NewSystemCommand("git", "clone", cloneURL().String(), "repository").
		WorkingDirectory(mockSystem.tempDirs[0])
	mockSystem.setExpectedEnvironment(expectedCommand)
	if !mockSystem.Ran(expectedCommand) {
		t.Fatalf("mockSystem did not run %s", expectedCommand)
	}
}

func TestClonedRepository_Clone_checks_for_existing_repository_on_disk(t *testing.T) {
	mockSystem := NewMockSystem()
	repositoryURL := cloneURL()
	repo := NewClonedRepository(mockSystem, repositoryURL)
	repo.MakePersistent()

	if err := repo.Clone(); err != nil {
		t.Fatal(err)
	}

	expectedCommand := NewSystemCommand("git", "rev-parse", "--git-dir").
		WorkingDirectory(filepath.Join(mockSystem.persistentDirs[0], "repository"))
	if !mockSystem.Ran(expectedCommand) {
		t.Fatalf("mockSystem did not run %s", expectedCommand)
	}
}

func TestClonedRepository_Clone_does_clone_repository_if_it_does_not_exist_already(t *testing.T) {
	mockSystem := NewMockSystem()
	repositoryURL := cloneURL()
	repo := NewClonedRepository(mockSystem, repositoryURL)
	repo.MakePersistent()

	mockSystem.failCommandWith(errors.New("fatal: Not a git repository"))

	repo.Clone()

	if !mockSystem.RanBin("git", "clone") {
		t.Fatal("mockSystem did not run `git clone`")
	}
}

func TestClonedRepository_Clone_does_nothing_if_repository_has_already_been_cloned(t *testing.T) {
	mockSystem := NewMockSystem()
	repo := NewClonedRepository(mockSystem, cloneURL())

	if err := repo.Clone(); err != nil {
		t.Fatal(err)
	}

	if err := repo.Clone(); err != nil {
		t.Fatal(err)
	}

	if got, want := len(mockSystem.tempDirs), 1; got != want {
		t.Errorf(`len(mockSystem.tempDirs) = %v; want %v`, got, want)
	}
}

func TestClonedRepository_Clone_injects_provided_credential_into_the_URL_when_cloning_via_HTTPS(t *testing.T) {
	mockSystem := NewMockSystem()

	credential := usernameAndPassword()
	repositoryURL := cloneURL()
	repositoryURL.Scheme = "https"
	repo := NewClonedRepository(mockSystem, repositoryURL)
	if err := repo.SetCredential(credential); err != nil {
		t.Fatal(err)
	}

	if err := repo.Clone(); err != nil {
		t.Fatal(err)
	}

	expectedURL := &url.URL{
		Scheme: "https",
		User:   url.UserPassword(credential.Username(), credential.Secret()),
		Host:   repositoryURL.Host,
		Path:   repositoryURL.Path,
	}

	expectedCommand := NewSystemCommand("git", "clone", expectedURL.String(), "repository").
		WorkingDirectory(mockSystem.tempDirs[0])
	mockSystem.setExpectedEnvironment(expectedCommand)

	if !mockSystem.Ran(expectedCommand) {
		t.Fatalf("mockSystem did not run %s", expectedCommand)
	}
}

func TestClonedRepository_Clone_uses_persistent_directory_based_on_repository_URL_if_persistent_option_is_set(t *testing.T) {
	mockSystem := NewMockSystem()
	repositoryURL := cloneURL()
	repo := NewClonedRepository(mockSystem, repositoryURL)
	repo.MakePersistent()

	if err := repo.Clone(); err != nil {
		t.Fatal(err)
	}

	expectedCommand := NewSystemCommand("git", "clone", repositoryURL.String(), "repository").
		WorkingDirectory("/persistent/git/example_com-repository")
	mockSystem.setExpectedEnvironment(expectedCommand)

	if !mockSystem.Ran(expectedCommand) {
		t.Fatalf("mockSystem did not run %s", expectedCommand)
	}
}

func TestClonedRepository_Clone_does_not_inject_SSH_credential_secret_into_clone_url(t *testing.T) {
	mockSystem := NewMockSystem()

	credential := sshUserAndSecret()
	repositoryURL := cloneURL()
	repositoryURL.Scheme = "ssh"
	repo := NewClonedRepository(mockSystem, repositoryURL)
	if err := repo.SetCredential(credential); err != nil {
		t.Fatal(err)
	}

	if err := repo.Clone(); err != nil {
		t.Fatal(err)
	}

	expectedURL := &url.URL{
		Scheme: "ssh",
		User:   url.User(credential.Username()),
		Host:   repositoryURL.Host,
		Path:   repositoryURL.Path,
	}

	expectedCommand := NewSystemCommand("git", "clone", expectedURL.String(), "repository").
		WorkingDirectory(mockSystem.tempDirs[0])
	mockSystem.setExpectedEnvironment(expectedCommand)
	if !mockSystem.Ran(expectedCommand) {
		t.Fatalf("mockSystem did not run %s", expectedCommand)
	}
}

func TestClonedRepository_SetCredential_writes_SSH_credential_secret_into_file_named_ssh_credential(t *testing.T) {
	mockSystem := NewMockSystem()

	credential := sshUserAndSecret()
	repositoryURL := cloneURL()
	repositoryURL.Scheme = "ssh"
	repo := NewClonedRepository(mockSystem, repositoryURL)
	if err := repo.SetCredential(credential); err != nil {
		t.Fatal(err)
	}

	expectedFilename := filepath.Join(mockSystem.tempDirs[0], "ssh_credential")

	if hasFile, contents := mockSystem.HasFile(expectedFilename, []byte(credential.Secret())); !hasFile {
		t.Fatalf("Expected file %q to have contents:\n%s\nGot:\n%s\n", expectedFilename, []byte(credential.Secret()), contents)
	}
}

func TestClonedRepository_SetCredential_writes_git_SSH_wrapper_to_disk_when_setting_SSH_credential(t *testing.T) {
	mockSystem := NewMockSystem()

	credential := sshUserAndSecret()
	repositoryURL := cloneURL()
	repositoryURL.Scheme = "ssh"
	repo := NewClonedRepository(mockSystem, repositoryURL)
	if err := repo.SetCredential(credential); err != nil {
		t.Fatal(err)
	}

	credentialFilename := filepath.Join(mockSystem.tempDirs[0], "ssh_credential")
	expectedFilename := filepath.Join(mockSystem.tempDirs[0], "git-ssh")
	expectedContents := fmt.Sprintf("/usr/bin/ssh -o IdentitiesOnly=yes -o PasswordAuthentication=no -o StrictHostKeyChecking=no -i %s \"$@\"\n", credentialFilename)
	if hasFile, contents := mockSystem.HasFile(expectedFilename, []byte(expectedContents)); !hasFile {
		t.Fatalf("Expected file %q to have contents:\n%s\nGot:\n%s\n", expectedFilename, expectedContents, contents)
	}

}

func TestClonedRepository_SetCredential_marks_git_SSH_wrapper_as_executable(t *testing.T) {
	mockSystem := NewMockSystem()

	credential := sshUserAndSecret()
	repositoryURL := cloneURL()
	repositoryURL.Scheme = "ssh"
	repo := NewClonedRepository(mockSystem, repositoryURL)
	if err := repo.SetCredential(credential); err != nil {
		t.Fatal(err)
	}

	if got, want := mockSystem.Permissions(filepath.Join(mockSystem.tempDirs[0], "git-ssh")), 0755; got != want {
		t.Errorf(`mockSystem.Permissions(filepath.Join(mockSystem.tempDirs[0], "git-ssh")) = %v; want %v`, got, want)
	}
}

func TestClonedRepository_SetCredential_returns_an_error_if_protocol_is_not_https_or_ssh(t *testing.T) {
	mockSystem := NewMockSystem()

	repositoryURL := cloneURL()
	repositoryURL.Scheme = "http"
	repo := NewClonedRepository(mockSystem, repositoryURL)
	if err := repo.SetCredential(NewMockCredential("http")); err == nil {
		t.Fatal("Expected an error")
	} else {
		if got, want := err.Error(), `unsupported protocol "http"`; got != want {
			t.Errorf(`err.Error() = %v; want %v`, got, want)
		}
	}
}

func TestClonedRepository_SetCredential_sets_permissions_of_ssh_credential_to_0600(t *testing.T) {
	mockSystem := NewMockSystem()

	credential := sshUserAndSecret()
	repositoryURL := cloneURL()
	repositoryURL.Scheme = "ssh"
	repo := NewClonedRepository(mockSystem, repositoryURL)
	if err := repo.SetCredential(credential); err != nil {
		t.Fatal(err)
	}

	if got, want := mockSystem.Permissions(filepath.Join(mockSystem.tempDirs[0], "ssh_credential")), 0600; got != want {
		t.Errorf(`mockSystem.Permissions(filepath.Join(mockSystem.tempDirs[0], "ssh_credential")) = %v; want %v`, got, want)
	}
}

func TestClonedRepository_Remove_removes_temporary_directory(t *testing.T) {
	mockSystem := NewMockSystem()

	credential := sshUserAndSecret()
	repositoryURL := cloneURL()
	repositoryURL.Scheme = "ssh"
	repo := NewClonedRepository(mockSystem, repositoryURL)
	if err := repo.SetCredential(credential); err != nil {
		t.Fatal(err)
	}

	if err := repo.Remove(); err != nil {
		t.Fatal(err)
	}

	if got, want := mockSystem.DeletedFile(mockSystem.tempDirs[0]), true; got != want {
		t.Errorf(`mockSystem.DeletedFile(mockSystem.tempDirs[0]) = %v; want %v`, got, want)
	}

}

func TestClonedRepository_IsAccessible_runs_git_ls_remote_to_determine_accessibility(t *testing.T) {
	mockSystem := NewMockSystem()
	repo := NewClonedRepository(mockSystem, cloneURL())

	if _, err := repo.IsAccessible(); err != nil {
		t.Fatal(err)
	}

	expectedCommand := NewSystemCommand("git", "ls-remote", cloneURL().String()).
		WorkingDirectory(mockSystem.tempDirs[0])

	mockSystem.setExpectedEnvironment(expectedCommand)

	if !mockSystem.Ran(expectedCommand) {
		t.Fatalf("mockSystem did not run %s", expectedCommand)
	}
}

func TestClonedRepository_IsAccessible_returns_false_and_an_error_if_git_ls_remote_fails(t *testing.T) {
	mockSystem := NewMockSystem()
	repo := NewClonedRepository(mockSystem, cloneURL())

	expectedError := errors.New("something went wrong")
	mockSystem.failCommandWith(expectedError)

	accessible, err := repo.IsAccessible()
	if got, want := accessible, false; got != want {
		t.Errorf(`accessible = %v; want %v`, got, want)
	}

	if got := err; got == nil {
		t.Fatalf(`err is nil`)
	}

	if got, want := err.Error(), expectedError.Error(); got != want {
		t.Errorf(`err.Error() = %v; want %v`, got, want)
	}

}

func TestClonedRepository_IsAccessible_returns_true_if_git_ls_remote_reported_no_error(t *testing.T) {
	mockSystem := NewMockSystem()
	repo := NewClonedRepository(mockSystem, cloneURL())

	accessible, err := repo.IsAccessible()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := accessible, true; got != want {
		t.Errorf(`accessible = %v; want %v`, got, want)
	}
}

func TestClonedRepository_References_attempts_to_clone_repository(t *testing.T) {
	mockSystem := NewMockSystem()
	repo := NewClonedRepository(mockSystem, cloneURL())

	if _, err := repo.References(); err != nil {
		t.Fatal(err)
	}

	if got, want := len(mockSystem.tempDirs), 1; got != want {
		t.Fatalf(`len(mockSystem.tempDirs) = %v; want %v`, got, want)
	}

	expectedCommand := NewSystemCommand("git", "clone", cloneURL().String(), "repository").
		WorkingDirectory(mockSystem.tempDirs[0])
	mockSystem.setExpectedEnvironment(expectedCommand)
	if !mockSystem.Ran(expectedCommand) {
		t.Fatalf("mockSystem did not run %s", expectedCommand)
	}
}

func TestClonedRepository_References_parses_output_of_git_ls_remote(t *testing.T) {
	mockSystem := NewMockSystem().UseTempDir("/tmp/some-tmp-dir")
	repo := NewClonedRepository(mockSystem, cloneURL())

	tmpDir, err := mockSystem.TempDir()
	if err != nil {
		t.Fatal(err)
	}

	expectedCommand := NewSystemCommand("git", "ls-remote", "repository").
		WorkingDirectory(tmpDir)

	mockSystem.setOutputForCommand(expectedCommand, `34cfff4153bf54639ffeee11cbb538bd992adfc5        HEAD
2df126d911291120a98d707da9311e29c04d150f        refs/deploys/a-deploy
`)
	references, err := repo.References()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(references), 2; got != want {
		t.Fatalf(`len(references) = %v; want %v`, got, want)
	}

	if got, want := references[0].Name, "HEAD"; got != want {
		t.Errorf(`references[0].Name = %v; want %v`, got, want)
	}

	if got, want := references[0].Hash, "34cfff4153bf54639ffeee11cbb538bd992adfc5"; got != want {
		t.Errorf(`references[0].Hash = %v; want %v`, got, want)
	}

	if got, want := references[1].Name, "refs/deploys/a-deploy"; got != want {
		t.Errorf(`references[1].Name = %v; want %v`, got, want)
	}

	if got, want := references[1].Hash, "2df126d911291120a98d707da9311e29c04d150f"; got != want {
		t.Errorf(`references[1].Hash = %v; want %v`, got, want)
	}
}

func TestClonedRepository_FetchMetadata_returns_the_references_in_a_repository(t *testing.T) {
	mockSystem := NewMockSystem().UseTempDir("/tmp/some-tmp-dir")
	repo := NewClonedRepository(mockSystem, cloneURL())

	tmpDir, err := mockSystem.TempDir()
	if err != nil {
		t.Fatal(err)
	}

	expectedCommand := NewSystemCommand("git", "ls-remote", "repository").
		WorkingDirectory(tmpDir)

	mockSystem.setOutputForCommand(expectedCommand, `34cfff4153bf54639ffeee11cbb538bd992adfc5        HEAD
2df126d911291120a98d707da9311e29c04d150f        refs/deploys/a-deploy
`)
	metadata, err := repo.FetchMetadata()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := metadata.Refs["HEAD"], "34cfff4153bf54639ffeee11cbb538bd992adfc5"; got != want {
		t.Errorf(`metadata.Refs["HEAD"] = %v; want %v`, got, want)
	}

	if got, want := metadata.Refs["refs/deploys/a-deploy"], "2df126d911291120a98d707da9311e29c04d150f"; got != want {
		t.Errorf(`metadata.Refs["refs/deploys/a-deploy"] = %v; want %v`, got, want)
	}

}

func TestClonedRepository_FetchMetadata_returns_the_contributors_in_a_repository(t *testing.T) {
	mockSystem := NewMockSystem().UseTempDir("/tmp/some-tmp-dir")
	repo := NewClonedRepository(mockSystem, cloneURL())

	tmpDir, err := mockSystem.TempDir()
	if err != nil {
		t.Fatal(err)
	}

	expectedCommand := NewSystemCommand("git", "log", "--format=%aN\x1e%aE").
		WorkingDirectory(filepath.Join(tmpDir, "repository"))

	mockSystem.setOutputForCommand(expectedCommand,
		"John Doe\x1ejohn.doe@example.com\n"+
			"Alice Baker\x1ealice.baker@example.com\n",
	)
	metadata, err := repo.FetchMetadata()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := metadata.Contributors["john.doe@example.com"], (&Person{Name: "John Doe", Email: "john.doe@example.com"}); !reflect.DeepEqual(got, want) {
		t.Errorf(`metadata.Contributors["john.doe@example.com"] = %v; want %v`, got, want)
	}

	if got, want := metadata.Contributors["alice.baker@example.com"], (&Person{Name: "Alice Baker", Email: "alice.baker@example.com"}); !reflect.DeepEqual(got, want) {
		t.Errorf(`metadata.Contributors["alice.baker@example.com"] = %v; want %v`, got, want)
	}
}

func TestClonedRepository_Contributors_returns_list_of_commit_authors(t *testing.T) {
	mockSystem := NewMockSystem().UseTempDir("/tmp/some-tmp-dir")
	repo := NewClonedRepository(mockSystem, cloneURL())

	tmpDir, err := mockSystem.TempDir()
	if err != nil {
		t.Fatal(err)
	}

	expectedCommand := NewSystemCommand("git", "log", "--format=%aN\x1e%aE").
		WorkingDirectory(filepath.Join(tmpDir, "repository"))

	mockSystem.setOutputForCommand(expectedCommand,
		"John Doe\x1ejohn.doe@example.com\n"+
			"Alice Baker\x1ealice.baker@example.com\n",
	)

	contributors, err := repo.Contributors()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(contributors), 2; got != want {
		t.Fatalf(`len(contributors) = %v; want %v`, got, want)
	}

	if got, want := contributors[0].Name, "John Doe"; got != want {
		t.Errorf(`contributors[0].Name = %v; want %v`, got, want)
	}

	if got, want := contributors[1].Name, "Alice Baker"; got != want {
		t.Errorf(`contributors[1].Name = %v; want %v`, got, want)
	}
}

func TestClonedRepository_Contributors_does_not_return_two_entries_for_the_same_email_address(t *testing.T) {
	mockSystem := NewMockSystem().UseTempDir("/tmp/some-tmp-dir")
	repo := NewClonedRepository(mockSystem, cloneURL())

	tmpDir, err := mockSystem.TempDir()
	if err != nil {
		t.Fatal(err)
	}

	expectedCommand := NewSystemCommand("git", "log", "--format=%aN\x1e%aE").
		WorkingDirectory(filepath.Join(tmpDir, "repository"))

	mockSystem.setOutputForCommand(expectedCommand,
		"John Doe\x1ejohn.doe@example.com\n"+
			"Alice Baker\x1ealice.baker@example.com\n"+
			"John Doe\x1ejohn.doe@example.com\n",
	)

	contributors, err := repo.Contributors()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(contributors), 2; got != want {
		t.Fatalf(`len(contributors) = %v; want %v`, got, want)
	}

	if got, want := contributors[0].Name, "John Doe"; got != want {
		t.Errorf(`contributors[0].Name = %v; want %v`, got, want)
	}

	if got, want := contributors[1].Name, "Alice Baker"; got != want {
		t.Errorf(`contributors[1].Name = %v; want %v`, got, want)
	}
}

func TestClonedRepository_Pull_clones_the_repository_before_pulling_it(t *testing.T) {
	mockSystem := NewMockSystem()
	repo := NewClonedRepository(mockSystem, cloneURL())

	if err := repo.Pull(); err != nil {
		t.Fatal(err)
	}

	if !mockSystem.RanBin("git", "clone") {
		t.Fatal("Expected `git clone` to have run")
	}
}

func TestClonedRepository_Pull_runs_git_pull_to_update_the_repository(t *testing.T) {
	mockSystem := NewMockSystem()
	repo := NewClonedRepository(mockSystem, cloneURL())

	if err := repo.Pull(); err != nil {
		t.Fatal(err)
	}

	expectedCommand := NewSystemCommand("git", "pull").
		WorkingDirectory(filepath.Join(mockSystem.tempDirs[0], "repository"))
	mockSystem.setExpectedEnvironment(expectedCommand)

	if !mockSystem.Ran(expectedCommand) {
		t.Fatalf("mockSystem did not run %q", expectedCommand)
	}
}

func TestClonedRepository_IsAccessible_successfully_reports_SSH_repository_as_accessible(t *testing.T) {
	skipIntegrationTests(t)
	system := NewOperatingSystem("/tmp")
	repositoryURL := &url.URL{
		Scheme: "ssh",
		Host:   "bitbucket.org",
		Path:   "/harrowio/mdpf-intergration-test.git",
	}
	repo := NewClonedRepository(system, repositoryURL)
	defer func() {
		if err := repo.Remove(); err != nil {
			t.Fatal(err)
		}
	}()
	if err := repo.SetCredential(sshUserAndSecret()); err != nil {
		t.Fatal(err)
	}

	accessible, err := repo.IsAccessible()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := accessible, true; got != want {
		t.Errorf(`accessible = %v; want %v`, got, want)
	}
}

func TestClonedRepository_IsAccessible_reports_HTTPS_repository_with_wrong_credentials_as_inaccessible(t *testing.T) {
	skipIntegrationTests(t)
	system := NewOperatingSystem("/tmp")
	repositoryURL := &url.URL{
		Scheme: "https",
		Host:   "bitbucket.org",
		Path:   "/harrowio/mdpf-intergration-test.git",
	}
	repo := NewClonedRepository(system, repositoryURL)
	defer func() {
		if err := repo.Remove(); err != nil {
			t.Fatal(err)
		}
	}()

	httpUsernameAndPassword, err := NewHTTPCredential("john-doe", "not_a_real_password")
	if err != nil {
		t.Fatal(err)
	}

	if err := repo.SetCredential(httpUsernameAndPassword); err != nil {
		t.Fatal(err)
	}

	accessible, err := repo.IsAccessible()
	if got := err; got == nil {
		t.Fatalf(`err is nil`)
	}

	if got, want := accessible, false; got != want {
		t.Errorf(`accessible = %v; want %v`, got, want)
	}
}

func TestClonedRepository_IsAccessible_successfully_reports_HTTPS_repository_as_accessible(t *testing.T) {
	skipIntegrationTests(t)
	system := NewOperatingSystem("/tmp")
	repositoryURL := &url.URL{
		Scheme: "https",
		Host:   "bitbucket.org",
		Path:   "/harrowio/mdpf-intergration-test.git",
	}
	repo := NewClonedRepository(system, repositoryURL)
	defer func() {
		if err := repo.Remove(); err != nil {
			t.Fatal(err)
		}
	}()

	httpUsernameAndPassword, err := NewHTTPCredential(httpUsername, httpPassword)
	if err != nil {
		t.Fatal(err)
	}

	if err := repo.SetCredential(httpUsernameAndPassword); err != nil {
		t.Fatal(err)
	}

	accessible, err := repo.IsAccessible()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := accessible, true; got != want {
		t.Errorf(`accessible = %v; want %v`, got, want)
	}
}

func TestClonedRepository_Clone_successfully_clones_repository_via_SSH(t *testing.T) {
	skipIntegrationTests(t)
	system := NewOperatingSystem("/tmp")
	repositoryURL := &url.URL{
		Scheme: "ssh",
		Host:   "bitbucket.org",
		Path:   "/harrowio/mdpf-intergration-test.git",
	}
	repo := NewClonedRepository(system, repositoryURL)
	defer func() {
		if err := repo.Remove(); err != nil {
			t.Fatal(err)
		}
	}()
	if err := repo.SetCredential(sshUserAndSecret()); err != nil {
		t.Fatal(err)
	}

	refs, err := repo.References()
	if err != nil {
		t.Fatal(err)
	}

	foundHEAD := false
	for _, ref := range refs {
		if ref.Name == "HEAD" {
			foundHEAD = true
			break
		}
	}

	if got, want := foundHEAD, true; got != want {
		t.Errorf(`foundHEAD = %v; want %v`, got, want)
	}
}

func TestClonedRepository_Contributors_works_with_another_instance_of_persistent_clone(t *testing.T) {
	skipIntegrationTests(t)
	system := NewOperatingSystem("/tmp")
	repositoryURL := &url.URL{
		Scheme: "ssh",
		Host:   "bitbucket.org",
		Path:   "/harrowio/mdpf-intergration-test.git",
	}
	repo := NewClonedRepository(system, repositoryURL)
	repo.MakePersistent()
	defer func() {
		if err := repo.Remove(); err != nil {
			t.Fatal(err)
		}
	}()
	if err := repo.SetCredential(sshUserAndSecret()); err != nil {
		t.Fatal(err)
	}

	if err := repo.Clone(); err != nil {
		t.Fatal(nil)
	}

	otherRepo := NewClonedRepository(system, repositoryURL)
	otherRepo.MakePersistent()
	contributors, err := otherRepo.Contributors()
	if err != nil {
		t.Fatal(err)
	}

	nContributors := len(contributors)
	if got, want := nContributors, 0; got == want {
		t.Errorf(`nContributors = %v; want not %v`, got, want)
	}
}

func skipIntegrationTests(t *testing.T) {
	if os.Getenv("HARROW_INTEGRATION_TEST") == "" {
		t.Skip("skipping integration tests")
	}
}

type MockSystem struct {
	tempDirs       []string
	persistentDirs []string
	commands       []*SystemCommand
	files          map[string]*bytes.Buffer
	executable     map[string]int
	deletedFiles   map[string]bool
	output         map[string]string
	runError       error
	returnTempDir  string
}

func NewMockSystem() *MockSystem {
	return &MockSystem{
		tempDirs:       []string{},
		persistentDirs: []string{},
		files:          map[string]*bytes.Buffer{},
		deletedFiles:   map[string]bool{},
		executable:     map[string]int{},
		output:         map[string]string{},
	}
}

type NopCloser struct{ io.Writer }

func (self NopCloser) Close() error { return nil }

func (self *MockSystem) CreateFile(filename string) (io.WriteCloser, error) {
	self.files[filename] = new(bytes.Buffer)
	return NopCloser{self.files[filename]}, nil
}

func (self *MockSystem) DeleteFile(filename string) error {
	self.deletedFiles[filename] = true
	return nil
}

func (self *MockSystem) SetPermissions(filename string, mode int) error {
	self.executable[filename] = mode
	return nil
}

func (self *MockSystem) UseTempDir(dir string) *MockSystem {
	self.returnTempDir = dir
	return self
}

func (self *MockSystem) TempDir() (string, error) {
	name := fmt.Sprintf("/tmp/tmp-%02d", len(self.tempDirs))

	if self.returnTempDir != "" {
		name = self.returnTempDir
	}

	self.tempDirs = append(self.tempDirs, name)
	return name, nil
}

func (self *MockSystem) Run(cmd *SystemCommand) ([]byte, error) {
	self.commands = append(self.commands, cmd)
	output, found := self.output[cmd.String()]
	if found {
		return []byte(output), self.runError
	}

	return []byte{}, self.runError
}

func (self *MockSystem) Ran(cmd *SystemCommand) bool {
	expected := cmd.String()
	for _, command := range self.commands {
		if command.String() == expected {
			return true
		}
	}

	return false
}

func (self *MockSystem) RanBin(bin string, args ...string) bool {
	for _, command := range self.commands {
		if command.Exec == bin && strings.HasPrefix(
			strings.Join(command.Args, " "),
			strings.Join(args, " "),
		) {
			return true
		}
	}

	return false
}

func (self *MockSystem) PersistentDir(key string) (string, error) {
	dir := fmt.Sprintf("/persistent/git/%s", key)
	self.persistentDirs = append(self.persistentDirs, dir)
	return dir, nil
}

func (self *MockSystem) DeletedFile(filename string) bool {
	return self.deletedFiles[filename]
}

func (self *MockSystem) HasFile(filename string, contents []byte) (bool, []byte) {
	buffer, found := self.files[filename]
	if !found {
		return false, []byte("<file not found>")
	}
	return bytes.Equal(buffer.Bytes(), contents), buffer.Bytes()
}

func (self *MockSystem) Permissions(filename string) int {
	return self.executable[filename]
}

func (self *MockSystem) setOutputForCommand(cmd *SystemCommand, output string) {
	self.output[cmd.String()] = output
}

func (self *MockSystem) failCommandWith(err error) {
	self.runError = err
}

func (self *MockSystem) setExpectedEnvironment(cmd *SystemCommand) {
	dir := ""
	if len(self.tempDirs) == 0 {
		dir = self.persistentDirs[0]
	} else {
		dir = self.tempDirs[0]
	}

	cmd.
		SetEnv("GIT_SSH", filepath.Join(dir, "git-ssh")).
		SetEnv("GIT_PAGER", "/usr/bin/cat").
		SetEnv("GIT_CONFIG_NOSYSTEM", "true").
		SetEnv("GIT_ASKPASS", "/bin/echo")
}

type MockCredential struct {
	protocol string
}

func NewMockCredential(protocol string) *MockCredential {
	return &MockCredential{
		protocol: protocol,
	}
}

func (self *MockCredential) String() string {
	return fmt.Sprintf("%s-credential", self.protocol)
}

func (self *MockCredential) Username() string {
	return "john-doe"
}

func (self *MockCredential) Secret() string {
	return "super-secret"
}

func (self *MockCredential) Protocol() string {
	return self.protocol
}
