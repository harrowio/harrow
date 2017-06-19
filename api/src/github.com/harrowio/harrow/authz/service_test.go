package authz

import (
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
)

type testSubject struct{}

func (self *testSubject) AuthorizationName() string { return "test-subject" }

func Test_txService_HasCapabilities_returnsTrue_ifArgumentImplementsSubject(t *testing.T) {
	t.Parallel()
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	subject := &testSubject{}
	service := NewService(tx, user, config.GetConfig())

	if !service.HasCapabilities(subject) {
		t.Fatalf("Expected subject to have capabilities")
	}

}

func Test_txService_HasCapabilities_returnsFalse_ifArgumentDoesNotImplementSubject(t *testing.T) {
	t.Parallel()
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	subject := 1
	service := NewService(tx, user, config.GetConfig())

	if service.HasCapabilities(subject) {
		t.Fatalf("Expected subject not to have capabilities")
	}
}

func Test_txService_LoggedInUserCanCreateOrganization(t *testing.T) {
	t.Parallel()
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("non-member")
	subject := &domain.Organization{
		Uuid: "278fba0a-968c-4122-b29d-4ad4cb03dad3",
	}
	service := NewService(tx, user, config.GetConfig())

	allowed, err := service.Can(domain.CapabilityCreate, subject)
	if !allowed {
		t.Fatalf("Expected logged in user to be able to create an organization. Error: %s", err)
	}
}

func Test_txService_LoggedOutUserCanSignUp(t *testing.T) {
	t.Parallel()
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	subject := &domain.User{
		Uuid: "dda5bdf3-be3f-4ce5-94be-1b2eed823a39",
	}
	service := NewService(tx, nil, config.GetConfig())

	allowed, err := service.Can(domain.CapabilitySignUp, subject)
	if !allowed {
		t.Fatalf("Expected logged out user to be able to sign up. Error: %s", err)
	}
}

func Test_txService_loadForUser_doesNotFailIfSubjectHasNoId(t *testing.T) {
	t.Parallel()
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	subject := &domain.User{}
	service := NewService(tx, nil, config.GetConfig())

	_, err := service.Can(domain.CapabilitySignUp, subject)
	if err != nil {
		aerr := err.(*Error)
		t.Fatalf("Unexpected error: %s: %s", err, aerr.Internal())
	}
}

func Test_txService_DomainObjectsHaveAuthorizationNameDefined(t *testing.T) {
	t.Parallel()
	// The following vi(m) command extracts all exported type names
	// from the domain package.  This is useful for maintaining the
	// list of entities to test here.
	//:r!grep -ho '^type [A-Z][^ ]* struct' src/github.com/harrowio/harrow/domain/*.go|awk '{print "&domain." $2 "{},"}'

	subjects := []interface{}{
		// Internal use only, no need to check
		// &domain.Change{},
		// &domain.Changes{},
		// &domain.EnvironmentVariables{},
		// &domain.LogLine{},
		// &domain.LoggableWrapper{},
		// &domain.LogSubscription{},
		// &domain.WrapperScriptVars{},
		// &domain.JobOperationVars{},
		// &domain.RepositoryOperationVars{},
		// &domain.SetupScriptVars{},
		// &domain.KeyPair{},
		// &domain.SessionNotValidError{},
		// &domain.ValidationError{},

		// Accessible to the user
		&domain.Delivery{},
		&domain.Environment{},
		&domain.Invitation{},
		&domain.Job{},
		&domain.Loggable{},
		&domain.OAuthToken{},
		&domain.Operation{},
		&domain.Organization{},
		&domain.OrganizationMember{},
		&domain.OrganizationMembership{},
		&domain.Project{},
		&domain.ProjectMember{},
		&domain.ProjectMembership{},
		&domain.Repository{},
		&domain.ScheduledExecution{},
		&domain.Schedule{},
		&domain.Session{},
		&domain.Subscription{},
		&domain.Subscriptions{},
		&domain.Target{},
		&domain.Task{},
		&domain.User{},
		&domain.Webhook{},
		&domain.WorkspaceBaseImage{},
	}

	service := NewService(nil, nil, config.GetConfig())

	for _, subject := range subjects {
		if !service.HasCapabilities(subject) {
			t.Errorf("%T has no authorization name", subject)
		}
	}
}

func Test_txService_CapabilitiesBySubject_returnsSetOfAllSubjectsWithCapabilities(t *testing.T) {
	t.Parallel()
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	service := NewService(tx, user, config.GetConfig())

	expected := map[string][]string{
		"public":       []string{"read"},
		"session":      []string{"create", "validate"},
		"user":         []string{"signup"},
		"organization": []string{"create"},
	}

	capabilities := service.CapabilitiesBySubject()
	for subject, got := range capabilities {
		want, found := expected[subject]
		if !found {
			t.Errorf("subject %q not found", subject)
			continue
		}

		if !reflect.DeepEqual(want, got) {
			t.Errorf("verbs = %v; want %v", want, got)
			continue
		}
	}
}

func Test_txService_Can_returnsTrue_ifThingOwnedByCurrentUser(t *testing.T) {
	t.Parallel()
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	subject := &domain.Session{
		Uuid:     "c4e45f56-6350-46ca-8cff-f3cbdeeb5b0e",
		UserUuid: user.Uuid,
	}
	service := NewService(tx, user, config.GetConfig())

	allowed, err := service.CanRead(subject)
	if err != nil {
		t.Fatal(err)
	}

	if !allowed {
		t.Fatalf("Expected %#v to be able to read %#v", user, subject)
	}
}

func Test_txService_Can_returnsError_ifCurrentUserIsBlocked(t *testing.T) {
	t.Parallel()
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	block, err := user.NewBlock("testing")
	if err != nil {
		t.Fatal(err)
	}
	block.BlockForever(time.Now().Add(-24 * time.Hour))
	if err := stores.NewDbUserBlockStore(tx).Create(block); err != nil {
		t.Fatal(err)
	}
	service := NewService(tx, user, config.GetConfig())
	subject := user
	allowed, err := service.CanUpdate(subject)
	if err == nil {
		t.Fatalf("expected an error")
	}

	if got, want := allowed, false; got != want {
		t.Errorf("allowed = %v; want %v", got, want)
	}

	aerr, ok := err.(*Error)
	if !ok {
		t.Fatalf("err.(type) = %T; want %T", err, aerr)
	}
}

func Test_txService_CanRead_returns_false_for_projects_you_are_not_a_member_of(t *testing.T) {
	t.Parallel()
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")

	organization := test_helpers.MustCreateOrganization(t, tx, &domain.Organization{
		Uuid: "42a87d88-fb49-46ad-8e05-3c391a94a5c4",
		Name: "Test Organization",
	})

	projectWithAccess := test_helpers.MustCreateProject(t, tx, &domain.Project{
		Uuid:             "8e6347a4-c9fc-45af-a46c-80a15ffc5c5c",
		OrganizationUuid: organization.Uuid,
		Name:             "Test A",
	})

	projectWithoutAccess := test_helpers.MustCreateProject(t, tx, &domain.Project{
		Uuid:             "b0470a40-777b-4232-9880-87c6b8bb5e76",
		OrganizationUuid: organization.Uuid,
		Name:             "Test B",
	})

	test_helpers.MustCreateProjectMembership(t, tx, &domain.ProjectMembership{
		UserUuid:       user.Uuid,
		ProjectUuid:    projectWithAccess.Uuid,
		MembershipType: domain.MembershipTypeMember,
	})

	service := NewService(tx, user, config.GetConfig())

	canRead, _ := service.CanRead(projectWithAccess)
	if got, want := canRead, true; got != want {
		t.Errorf(`canRead = %v; want %v`, got, want)
	}

	cannotRead, _ := service.CanRead(projectWithoutAccess)
	if got, want := cannotRead, false; got != want {
		t.Errorf(`cannotRead = %v; want %v`, got, want)
	}
}

func Test_txService_CanRead_returns_true_for_projects_whose_organization_you_are_a_member_of(t *testing.T) {
	t.Parallel()
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")

	organization := test_helpers.MustCreateOrganization(t, tx, &domain.Organization{
		Uuid: "42a87d88-fb49-46ad-8e05-3c391a94a5c4",
		Name: "Test Organization",
	})

	projectA := test_helpers.MustCreateProject(t, tx, &domain.Project{
		Uuid:             "8e6347a4-c9fc-45af-a46c-80a15ffc5c5c",
		OrganizationUuid: organization.Uuid,
		Name:             "Test A",
	})

	projectB := test_helpers.MustCreateProject(t, tx, &domain.Project{
		Uuid:             "b0470a40-777b-4232-9880-87c6b8bb5e76",
		OrganizationUuid: organization.Uuid,
		Name:             "Test B",
	})

	test_helpers.MustCreateOrganizationMembership(t, tx, &domain.OrganizationMembership{
		UserUuid:         user.Uuid,
		OrganizationUuid: organization.Uuid,
		Type:             domain.MembershipTypeMember,
	})

	service := NewService(tx, user, config.GetConfig())

	canReadA, _ := service.CanRead(projectA)
	canReadB, _ := service.CanRead(projectB)

	if got, want := canReadA, true; got != want {
		t.Errorf(`canReadA = %v; want %v`, got, want)
	}

	if got, want := canReadB, true; got != want {
		t.Errorf(`canReadB = %v; want %v`, got, want)
	}
}

func Test_txService_CanCreate_returns_true_for_projects_whose_organization_you_are_a_member_of(t *testing.T) {
	t.Parallel()
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")

	organization := test_helpers.MustCreateOrganization(t, tx, &domain.Organization{
		Uuid: "42a87d88-fb49-46ad-8e05-3c391a94a5c4",
		Name: "Test Organization",
	})

	project := &domain.Project{
		Uuid:             "b0470a40-777b-4232-9880-87c6b8bb5e76",
		OrganizationUuid: organization.Uuid,
		Name:             "Test",
	}

	test_helpers.MustCreateOrganizationMembership(t, tx, &domain.OrganizationMembership{
		UserUuid:         user.Uuid,
		OrganizationUuid: organization.Uuid,
		Type:             domain.MembershipTypeMember,
	})

	service := NewService(tx, user, config.GetConfig())

	canCreate, _ := service.CanCreate(project)

	if got, want := canCreate, true; got != want {
		t.Errorf(`canCreate = %v; want %v`, got, want)
	}
}

func Test_txService_CanRead_returns_true_for_limits_for_an_organization_whose_projects_you_are_a_member_of(t *testing.T) {
	t.Parallel()
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")

	organization := test_helpers.MustCreateOrganization(t, tx, &domain.Organization{
		Uuid: "42a87d88-fb49-46ad-8e05-3c391a94a5c4",
		Name: "Test Organization",
	})

	project := test_helpers.MustCreateProject(t, tx, &domain.Project{
		Uuid:             "b0470a40-777b-4232-9880-87c6b8bb5e76",
		OrganizationUuid: organization.Uuid,
		Name:             "Test",
	})

	test_helpers.MustCreateProjectMembership(t, tx, &domain.ProjectMembership{
		UserUuid:       user.Uuid,
		ProjectUuid:    project.Uuid,
		MembershipType: domain.MembershipTypeMember,
	})

	theLimits := &domain.Limits{
		OrganizationUuid: organization.Uuid,
	}

	service := NewService(tx, user, config.GetConfig())

	canRead, _ := service.CanRead(theLimits)

	if got, want := canRead, true; got != want {
		t.Errorf(`canRead = %v; want %v`, got, want)
	}
}

var trackTime func(time.Time, string) = func(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func Test_txService_organization_membership_type_trumps_project_membership_type(t *testing.T) {
	t.Parallel()
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)

	user := world.User("default")

	organization := test_helpers.MustCreateOrganization(t, tx, &domain.Organization{
		Uuid: "42a87d88-fb49-46ad-8e05-3c391a94a5c4",
		Name: "Test Organization",
	})

	project := test_helpers.MustCreateProject(t, tx, &domain.Project{
		Uuid:             "8e6347a4-c9fc-45af-a46c-80a15ffc5c5c",
		OrganizationUuid: organization.Uuid,
		Name:             "Test Project",
	})

	environment := test_helpers.MustCreateEnvironment(t, tx, &domain.Environment{
		Uuid:        "bb5d495b-23d7-4ba1-9732-2e3df818a54f",
		ProjectUuid: project.Uuid,
		Name:        "Test Environment",
		Variables: domain.EnvironmentVariables{
			M: map[string]string{},
		},
	})

	test_helpers.MustCreateOrganizationMembership(t, tx, &domain.OrganizationMembership{
		UserUuid:         user.Uuid,
		OrganizationUuid: organization.Uuid,
		Type:             domain.MembershipTypeOwner,
	})

	test_helpers.MustCreateProjectMembership(t, tx, &domain.ProjectMembership{
		UserUuid:       user.Uuid,
		ProjectUuid:    project.Uuid,
		MembershipType: domain.MembershipTypeMember,
	})

	service := NewService(tx, user, config.GetConfig())

	canArchive, _ := service.CanUpdate(environment)

	if got, want := canArchive, true; got != want {
		t.Errorf(`canArchive = %v; want %v`, got, want)
	}
}
