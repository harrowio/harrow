package domain

import (
	"bytes"
	"net/http"
	"reflect"
	"testing"
	"time"
)

type MockOperationStore struct {
	byId        map[string]*Operation
	previousFor map[string]*Operation
}

func NewMockOperationStore() *MockOperationStore {
	return &MockOperationStore{
		byId:        map[string]*Operation{},
		previousFor: map[string]*Operation{},
	}
}

func (self *MockOperationStore) MarkExitStatus(operationUuid string, exitStatus int) error {
	return nil
}

func (self *MockOperationStore) FindPreviousOperation(currentOperationUuid string) (*Operation, error) {
	operation, found := self.previousFor[currentOperationUuid]
	if !found {
		return nil, nil
	}
	return operation, nil
}

func (self *MockOperationStore) AddPreviousOperation(previous *Operation, currentUuid string) {
	self.byId[previous.Uuid] = previous
	self.previousFor[currentUuid] = previous
}

type mockEnvStore struct {
	byJobUuid map[string]*Environment
}

func (store *mockEnvStore) FindByJobUuid(jobUuid string) (*Environment, error) {
	return store.byJobUuid[jobUuid], nil
}

type mockRepositoryStore struct {
	byUuid    map[string]*Repository
	byJobUuid map[string][]*Repository
}

func (store *mockRepositoryStore) MarkAsAccessible(repositoryUuid string, accessible bool) error {
	repo := store.byUuid[repositoryUuid]
	if repo == nil {
		return &NotFoundError{}
	}

	repo.Accessible = accessible
	return nil
}

func (store *mockRepositoryStore) FindByUuid(repositoryUuid string) (*Repository, error) {
	return store.byUuid[repositoryUuid], nil
}

func (store *mockRepositoryStore) FindAllByJobUuid(jobUuid string) ([]*Repository, error) {
	return store.byJobUuid[jobUuid], nil
}

type mockRepositoryCredentialsStore struct {
	byRepositoryUuid map[string]*RepositoryCredential
}

func (store *mockRepositoryCredentialsStore) FindByRepositoryUuid(repositoryUuid string) (*RepositoryCredential, error) {
	if store.byRepositoryUuid[repositoryUuid] == nil {
		return nil, new(NotFoundError)
	}
	return store.byRepositoryUuid[repositoryUuid], nil
}

func (store *mockRepositoryCredentialsStore) FindByRepositoryUuidAndType(repositoryUuid string, credentialType RepositoryCredentialType) (*RepositoryCredential, error) {
	if store.byRepositoryUuid[repositoryUuid] == nil {
		return nil, new(NotFoundError)
	}
	cred := store.byRepositoryUuid[repositoryUuid]
	if cred.Type == credentialType {
		return cred, nil
	} else {
		return nil, new(NotFoundError)
	}
}

type mockTaskStore struct {
	byJobUuid map[string]*Task
}

func (store *mockTaskStore) FindByJobUuid(jobUuid string) (*Task, error) {
	return store.byJobUuid[jobUuid], nil
}

type mockSecretStore struct {
	byEnvironmentUuid map[string][]*Secret
}

func (store *mockSecretStore) FindAllByEnvironmentUuid(environmentUuid string) ([]*Secret, error) {
	return store.byEnvironmentUuid[environmentUuid], nil
}

func Test_Operation_Environment_ReturnsNothingWithoutAJob(t *testing.T) {
	operation := &Operation{}
	store := &mockEnvStore{byJobUuid: map[string]*Environment{}}
	if env, err := operation.Environment(store); err != nil {
		t.Fatal(err)
	} else if env != nil {
		t.Fatalf("Unexpected environment: %#v\n", env)
	}
}

func Test_Operation_Environment_ReturnsEnvironmentForAJob(t *testing.T) {
	job := &Job{
		Uuid:            "ce36a377-3250-4a89-8b84-a276084fd511",
		EnvironmentUuid: "4a1e8157-5f9f-4f9d-a1e6-33b9222c272f",
	}

	operation := &Operation{
		Uuid:    "eaaab676-d4e7-4e71-ba48-b13192253587",
		JobUuid: &job.Uuid,
	}

	store := &mockEnvStore{
		byJobUuid: map[string]*Environment{
			job.Uuid: {
				Uuid: job.EnvironmentUuid,
			},
		},
	}

	if env, err := operation.Environment(store); err != nil {
		t.Fatal(err)
	} else if env == nil {
		t.Fatal("Expected an environment to be returned.")
	}
}

func Test_Operation_Secrets(t *testing.T) {

	job := &Job{
		Uuid:            "ce36a377-3250-4a89-8b84-a276084fd511",
		EnvironmentUuid: "4a1e8157-5f9f-4f9d-a1e6-33b9222c272f",
	}
	operation := &Operation{
		Uuid:    "eaaab676-d4e7-4e71-ba48-b13192253587",
		JobUuid: &job.Uuid,
	}

	env := &Environment{
		Uuid: job.EnvironmentUuid,
	}
	envStore := &mockEnvStore{
		byJobUuid: map[string]*Environment{
			job.Uuid: env,
		},
	}

	expSecrets := []*Secret{
		{EnvironmentUuid: env.Uuid},
		{EnvironmentUuid: env.Uuid},
	}
	secretStore := &mockSecretStore{
		byEnvironmentUuid: map[string][]*Secret{
			env.Uuid: expSecrets,
		},
	}

	if secrets, err := operation.Secrets(envStore, secretStore); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(secrets, expSecrets) {
		t.Fatalf("secrets=%#v, want=%#v\n", secrets, expSecrets)
	}
}

func Test_Operation_IsReady(t *testing.T) {

	job := &Job{
		Uuid:            "ce36a377-3250-4a89-8b84-a276084fd511",
		EnvironmentUuid: "4a1e8157-5f9f-4f9d-a1e6-33b9222c272f",
	}
	operation := &Operation{
		Uuid:    "eaaab676-d4e7-4e71-ba48-b13192253587",
		JobUuid: &job.Uuid,
	}

	env := &Environment{
		Uuid: job.EnvironmentUuid,
	}
	envStore := &mockEnvStore{
		byJobUuid: map[string]*Environment{
			job.Uuid: env,
		},
	}

	secrets := []*Secret{
		{EnvironmentUuid: env.Uuid, Status: SecretPending},
		{EnvironmentUuid: env.Uuid, Status: SecretPending},
	}
	secretStore := &mockSecretStore{
		byEnvironmentUuid: map[string][]*Secret{
			env.Uuid: secrets,
		},
	}

	repos := []*Repository{
		{Uuid: "485f5eb0-b1e6-4617-b409-efd7d29ecfac"},
		{Uuid: "3b58159a-39f8-4227-9e07-429dbbf89063"},
	}
	reposStore := &mockRepositoryStore{
		byJobUuid: map[string][]*Repository{
			job.Uuid: repos,
		},
	}
	repoCredentialsStore := &mockRepositoryCredentialsStore{}
	rdy, err := operation.IsReady(reposStore, repoCredentialsStore, envStore, secretStore)
	if err != nil {
		t.Error(err)
	}
	if rdy {
		t.Error("Expected the operation to be NOT ready")
	}

	// Adding only the RepositoryCredentials should not make the operation ready
	repoCredentials := []*RepositoryCredential{
		{RepositoryUuid: repos[0].Uuid},
		{RepositoryUuid: repos[1].Uuid},
	}
	repoCredentialsStore = &mockRepositoryCredentialsStore{
		byRepositoryUuid: map[string]*RepositoryCredential{
			repos[0].Uuid: repoCredentials[0],
			repos[1].Uuid: repoCredentials[1],
		},
	}

	rdy, err = operation.IsReady(reposStore, repoCredentialsStore, envStore, secretStore)
	if err != nil {
		t.Error(err)
	}
	if rdy {
		t.Error("Expected the operation NOT to be ready")
	}

	// Finally, making the Secrets be "present" should make the Operation ready
	secrets[0].Status = SecretPresent
	secrets[1].Status = SecretPresent

	rdy, err = operation.IsReady(reposStore, repoCredentialsStore, envStore, secretStore)
	if err != nil {
		t.Error(err)
	}
	if !rdy {
		t.Error("Expected the operation to be ready")
	}

}

func Test_Operation_Repositories_JobOperation(t *testing.T) {

	job := &Job{
		Uuid:            "ce36a377-3250-4a89-8b84-a276084fd511",
		EnvironmentUuid: "4a1e8157-5f9f-4f9d-a1e6-33b9222c272f",
	}
	operation := &Operation{
		Uuid:    "eaaab676-d4e7-4e71-ba48-b13192253587",
		JobUuid: &job.Uuid,
	}

	repos := []*Repository{
		{Uuid: "485f5eb0-b1e6-4617-b409-efd7d29ecfac"},
		{Uuid: "3b58159a-39f8-4227-9e07-429dbbf89063"},
	}
	reposStore := &mockRepositoryStore{
		byJobUuid: map[string][]*Repository{
			job.Uuid: repos,
		},
	}

	rs, err := operation.Repositories(reposStore)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rs, repos) {
		t.Errorf("rs=%#v, want %#v", rs, repos)
	}
}

func Test_Operation_Repositories_RepoOperation(t *testing.T) {

	repo := &Repository{Uuid: "3b58159a-39f8-4227-9e07-429dbbf89063"}

	reposStore := &mockRepositoryStore{
		byUuid: map[string]*Repository{
			repo.Uuid: repo,
		},
	}
	operation := &Operation{
		Uuid:           "eaaab676-d4e7-4e71-ba48-b13192253587",
		RepositoryUuid: &repo.Uuid,
	}

	rs, err := operation.Repositories(reposStore)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(rs, []*Repository{repo}) {
		t.Errorf("rs=%#v, want %#v", rs, []*Repository{repo})
	}
}

func Test_keyPair_Filename_producesPortableFilename(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Deployment Key", "deployment_key-e354020"},
		{"deployment key", "deployment_key-e28f86d"},
		{"  trimmed ", "trimmed-d7e0bd8"},
		{"нет", "___-ced07fd"},
		{"ναι", "___-fab58b9"},
		{"../../../../../etc/passwd", "passwd-648ce0e"},
	}

	subject := &keyPair{}
	for _, test := range tests {
		subject.Name = test.in
		if got := subject.Filename(); got != test.want {
			t.Errorf("keyPair{Name: %q}.Filename() = %q; want %q", test.in, got, test.want)
		}
	}
}

func TestOperationParameters_Scan_restoresStateFromValue(t *testing.T) {
	uuid := "4dc6b75d-1878-47ee-8f96-3fb4c1a5b157"
	operation := &Operation{
		Parameters: &OperationParameters{
			Checkout: map[string]string{
				uuid: "feature",
			},
		},
	}

	value, err := operation.Parameters.Value()
	if err != nil {
		t.Fatal(err)
	}

	scanned := &OperationParameters{}

	if err := scanned.Scan(value); err != nil {
		t.Fatal(err)
	}

	if got, want := scanned.Checkout[uuid], "feature"; got != want {
		t.Errorf(`scanned.Checkout[uuid] = %v; want %v`, got, want)
	}
}

type deliveriesByUuid map[string]*Delivery

func (self deliveriesByUuid) FindByUuid(uuid string) (*Delivery, error) {
	delivery, found := self[uuid]
	if !found {
		return nil, new(NotFoundError)
	}
	return delivery, nil
}

func TestNewOperationSetupScriptCtxt_loadsPreviousOperationFromStore(t *testing.T) {
	jobUuid := "7f316f70-e2d0-46d9-b2ca-44093245156c"
	operations := NewMockOperationStore()
	previousOperationId := "620f928c-add8-47fb-a7fc-3831786571ea"
	currentOperationId := "5008a687-d76f-477e-af89-322052fd2821"
	operations.AddPreviousOperation(&Operation{
		Uuid:    previousOperationId,
		Type:    OperationTypeJobScheduled,
		JobUuid: &jobUuid,
	}, currentOperationId)
	operation := &Operation{
		Uuid:    currentOperationId,
		Type:    OperationTypeJobScheduled,
		JobUuid: &jobUuid,
	}

	ctxt, err := operation.NewSetupScriptCtxt(&WorkspaceBaseImage{}, &Project{})
	if err != nil {
		t.Fatal(err)
	}

	if err := ctxt.LoadPreviousOperation(operations); err != nil {
		t.Fatal(err)
	}

	if got := ctxt.PreviousOperation; got == nil {
		t.Fatalf("ctxt.PreviousOperation is nil")
	}

	if got, want := ctxt.PreviousOperation.Uuid, previousOperationId; got != want {
		t.Errorf(`ctxt.PreviousOperation.Uuid = %v; want %v`, got, want)
	}

}

func TestNewOperationSetupScriptCtxt_ensuresPreviousOperationHasRepositoryCheckouts(t *testing.T) {
	jobUuid := "7f316f70-e2d0-46d9-b2ca-44093245156c"
	operations := NewMockOperationStore()
	previousOperationId := "620f928c-add8-47fb-a7fc-3831786571ea"
	currentOperationId := "5008a687-d76f-477e-af89-322052fd2821"
	operations.AddPreviousOperation(&Operation{
		Uuid:    previousOperationId,
		Type:    OperationTypeJobScheduled,
		JobUuid: &jobUuid,
	}, currentOperationId)
	operation := &Operation{
		Uuid:    currentOperationId,
		Type:    OperationTypeJobScheduled,
		JobUuid: &jobUuid,
	}

	ctxt, err := operation.NewSetupScriptCtxt(&WorkspaceBaseImage{}, &Project{})
	if err != nil {
		t.Fatal(err)
	}

	if err := ctxt.LoadPreviousOperation(operations); err != nil {
		t.Fatal(err)
	}

	if got := ctxt.PreviousOperation; got == nil {
		t.Fatalf("ctxt.PreviousOperation is nil")
	}

	if got := ctxt.PreviousOperation.RepositoryCheckouts; got == nil {
		t.Fatalf("ctxt.PreviousOperation.RepositoryCheckouts is nil")
	}

}

func TestNewOperationSetupScriptCtxt_LoadPreviousOperation_doesNotFailIfThereIsNoPreviousOperation(t *testing.T) {
	jobUuid := "7f316f70-e2d0-46d9-b2ca-44093245156c"
	operations := NewMockOperationStore()
	currentOperationId := "5008a687-d76f-477e-af89-322052fd2821"
	operation := &Operation{
		Uuid:    currentOperationId,
		Type:    OperationTypeJobScheduled,
		JobUuid: &jobUuid,
	}

	ctxt, err := operation.NewSetupScriptCtxt(&WorkspaceBaseImage{}, &Project{})
	if err != nil {
		t.Fatal(err)
	}

	if err := ctxt.LoadPreviousOperation(operations); err != nil {
		t.Fatal(err)
	}
}

func TestOperationSetupScriptCtxt_LoadWebhookBody_loadsBodyFromDelivery(t *testing.T) {

	src := `{"ref":"master"}`
	req, err := http.NewRequest("POST", "https://example.com/wh/slug", bytes.NewBufferString(src))
	delivery := &Delivery{
		Uuid:    "2f555379-7573-49f1-8637-57b9c3d83403",
		Request: DeliveredRequest{req},
	}
	deliveries := deliveriesByUuid{
		delivery.Uuid: delivery,
	}

	jobUuid := "7f316f70-e2d0-46d9-b2ca-44093245156c"
	operation := &Operation{
		Type:    OperationTypeJobScheduled,
		JobUuid: &jobUuid,
		Parameters: &OperationParameters{
			TriggeredByDelivery: delivery.Uuid,
		},
	}

	ctxt, err := operation.NewSetupScriptCtxt(&WorkspaceBaseImage{}, &Project{})
	if err != nil {
		t.Fatal(err)
	}
	if err := ctxt.LoadWebhookBody(deliveries); err != nil {
		t.Fatal(err)
	}

	if got, want := string(ctxt.WebhookBody), src; got != want {
		t.Errorf(`string(ctxt.WebhookBody) = %v; want %v`, got, want)
	}
}

func Test_OperationCtxt_AddSshConfig_AddsOneNewEntry_ForSshUrls(t *testing.T) {

	operation := &Operation{
		Type: OperationTypeJobScheduled,
	}
	repository := &Repository{
		Uuid: "faeecdc3-a14a-402e-bdaa-a4f21b4c4259",
		Url:  "git@github.com:foo/bar.git",
	}
	repositoryCredential := &RepositoryCredential{
		Name: "repository-faeecdc3-a14a-402e-bdaa-a4f21b4c4259@harrow.io",
	}
	operationCtxt, err := operation.NewSetupScriptCtxt(&WorkspaceBaseImage{}, &Project{})
	if err != nil {
		t.Fatal(err)
	}

	operationCtxt.AddSshConfig(repository, repositoryCredential)

	if have, want := len(operationCtxt.SshConfigs), 1; have != want {
		t.Fatalf("len(operationCtxt.SshConfigs), have=%d, want %d", have, want)
	}
	if have, want := operationCtxt.SshConfigs[0].Host, "github.com"; have != want {
		t.Errorf("operationCtxt.SshConfigs[0].Host, have=%s, want %s", have, want)
	}
	expectedAlias := "ssh___github.com_foo_bar.git"
	if have, want := operationCtxt.SshConfigs[0].SSHHostAlias, expectedAlias; have != want {
		t.Errorf("operationCtxt.SshConfigs[0].HostAlias, have=%s, want %s", have, want)
	}
	if have, want := operationCtxt.SshConfigs[0].User, "git"; have != want {
		t.Errorf("operationCtxt.SshConfigs[0].User, have=%s, want %s", have, want)
	}
	if have, want := operationCtxt.SshConfigs[0].RepoUuid, repository.Uuid; have != want {
		t.Errorf("operationCtxt.SshConfigs[0].RepoUuid, have=%s, want %s", have, want)
	}
	key := keyPair{
		Name: repositoryCredential.Name,
	}
	keyFileName := key.Filename()
	if have, want := operationCtxt.SshConfigs[0].KeyFileName, keyFileName; have != want {
		t.Errorf("operationCtxt.SshConfigs[0].KeyFileName, have=%s, want %s", have, want)
	}
}

func Test_OperationCtxt_AddSshConfig_AddsNoEntry_ForHttpsUrls(t *testing.T) {

	operation := &Operation{
		Type: OperationTypeJobScheduled,
	}
	repository := &Repository{
		Uuid: "faeecdc3-a14a-402e-bdaa-a4f21b4c4259",
		Url:  "https://github.com/foo/bar.git",
	}
	repositoryCredential := &RepositoryCredential{
		Name: "repository-faeecdc3-a14a-402e-bdaa-a4f21b4c4259@harrow.io",
	}
	operationCtxt, err := operation.NewSetupScriptCtxt(&WorkspaceBaseImage{}, &Project{})
	if err != nil {
		t.Fatal(err)
	}

	operationCtxt.AddSshConfig(repository, repositoryCredential)

	if have, want := len(operationCtxt.SshConfigs), 0; have != want {
		t.Fatalf("len(operationCtxt.SshConfigs), have=%d, want %d", have, want)
	}
}

func TestOperation_IsGitAccessCheck_returnsFalseForEnumeratingBranches(t *testing.T) {
	operation := &Operation{
		Type: OperationTypeGitEnumerationBranches,
	}

	if got, want := operation.IsGitAccessCheck(), false; got != want {
		t.Errorf(`operation.IsGitAccessCheck() = %v; want %v`, got, want)
	}
}

func TestOperation_IsGitMetadataCollect_returnsTrueForOperationTypeGitEnumerationBranches(t *testing.T) {
	operation := &Operation{
		Type: OperationTypeGitEnumerationBranches,
	}

	if got, want := operation.IsGitMetadataCollect(), true; got != want {
		t.Errorf(`operation.IsGitMetadataCollect() = %v; want %v`, got, want)
	}
}

func TestOperation_Status_returns_failure_if_operation_has_failed_at_timestamp(t *testing.T) {
	now := time.Date(2016, 4, 5, 17, 12, 21, 0, time.UTC)
	operation := &Operation{
		FailedAt: &now,
	}

	if got, want := operation.Status(), "failure"; got != want {
		t.Errorf(`operation.Status() = %v; want %v`, got, want)
	}
}

func TestOperation_Status_returns_success_if_operation_has_finished_at_set_and_was_successful(t *testing.T) {
	now := time.Date(2016, 4, 5, 17, 12, 21, 0, time.UTC)
	operation := &Operation{
		FinishedAt: &now,
		ExitStatus: 0,
	}

	if got, want := operation.Status(), "success"; got != want {
		t.Errorf(`operation.Status() = %v; want %v`, got, want)
	}
}

func TestOperation_Status_returns_active_if_operation_has_not_finished_yet(t *testing.T) {
	operation := &Operation{}
	if got, want := operation.Status(), "active"; got != want {
		t.Errorf(`operation.Status() = %v; want %v`, got, want)
	}
}

func TestOperation_Status_returns_timeout_if_operation_has_timed_out(t *testing.T) {
	now := time.Date(2016, 4, 5, 18, 36, 33, 0, time.UTC)
	operation := &Operation{
		TimedOutAt: &now,
	}
	if got, want := operation.Status(), "timeout"; got != want {
		t.Errorf(`operation.Status() = %v; want %v`, got, want)
	}
}

func TestOperation_Status_returns_fatal_if_operation_has_failed_fatally(t *testing.T) {
	fatalError := "fatal error"
	operation := &Operation{
		FatalError: &fatalError,
	}
	if got, want := operation.Status(), "fatal"; got != want {
		t.Errorf(`operation.Status() = %v; want %v`, got, want)
	}
}

func TestOperation_Status_returns_canceled_if_operation_has_canceled_at_set(t *testing.T) {
	now := time.Date(2016, 5, 6, 11, 42, 9, 0, time.UTC)
	operation := &Operation{
		CanceledAt: &now,
	}

	if got, want := operation.Status(), "canceled"; got != want {
		t.Errorf(`operation.Status() = %v; want %v`, got, want)
	}
}
