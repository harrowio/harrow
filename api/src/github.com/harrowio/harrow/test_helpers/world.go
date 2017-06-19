package test_helpers

import (
	"fmt"
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
)

var (
	TestPlan = &domain.BillingPlan{
		Uuid:             "c9569087-ff00-446b-b002-d29e7c3d38c0",
		PricePerMonth:    domain.Money{Amount: 10, Currency: domain.EUR},
		UsersIncluded:    1,
		ProjectsIncluded: 2,
	}
)

type World struct {
	users                   map[string]*domain.User
	organizations           map[string]*domain.Organization
	organizationMemberships map[string]*domain.OrganizationMembership
	projects                map[string]*domain.Project
	projectMemberships      map[string]*domain.ProjectMembership
	tasks                   map[string]*domain.Task
	environments            map[string]*domain.Environment
	secrets                 map[string]*domain.Secret
	jobs                    map[string]*domain.Job
	schedules               map[string]*domain.Schedule
	workspaceBaseImages     map[string]*domain.WorkspaceBaseImage
	repositories            map[string]*domain.Repository
	repositoryCredentials   map[string]*domain.RepositoryCredential
}

// Use this if you don't care for the stores.KeyValueStore
func MustNewWorld(tx *sqlx.Tx, t *testing.T) *World {
	ss := NewMockSecretKeyValueStore()

	world := MustNewWorldUsingSecretKeyValueStore(tx, ss, t)

	return world
}

func MustNewWorldUsingSecretKeyValueStore(tx *sqlx.Tx, ss stores.SecretKeyValueStore, t *testing.T) *World {
	world := &World{
		users:                   map[string]*domain.User{},
		organizations:           map[string]*domain.Organization{},
		organizationMemberships: map[string]*domain.OrganizationMembership{},
		projects:                map[string]*domain.Project{},
		projectMemberships:      map[string]*domain.ProjectMembership{},
		tasks:                   map[string]*domain.Task{},
		environments:            map[string]*domain.Environment{},
		secrets:                 map[string]*domain.Secret{},
		jobs:                    map[string]*domain.Job{},
		schedules:               map[string]*domain.Schedule{},
		workspaceBaseImages:     map[string]*domain.WorkspaceBaseImage{},
		repositories:            map[string]*domain.Repository{},
		repositoryCredentials:   map[string]*domain.RepositoryCredential{},
	}
	world.MustLoadDefaults(tx, ss, t)

	return world
}

func (w *World) MustLoadDefaults(tx *sqlx.Tx, ss stores.SecretKeyValueStore, t *testing.T) {
	w.users["default"] = MustCreateUser(t, tx, &domain.User{
		Email:    "vagrant+test@localhost",
		Name:     "Test User",
		Password: "password-is-long-enough",
		UrlHost:  "localhost.localdomain",
	})

	w.users["project-member"] = MustCreateUser(t, tx, &domain.User{
		Email:    "vagrant+test-project-member@localhost",
		Name:     "Project Member",
		Password: "password-is-long-enough",
		UrlHost:  "localhost.localdomain",
	})

	w.users["project-owner"] = MustCreateUser(t, tx, &domain.User{
		Email:    "vagrant+test-project-owner@localhost",
		Name:     "Project Owner",
		Password: "password-is-long-enough",
		UrlHost:  "localhost.localdomain",
	})

	w.users["other"] = MustCreateUser(t, tx, &domain.User{
		Email:    "vagrant+test-other@localhost",
		Name:     "Other User",
		Password: "password-is-long-enough",
		UrlHost:  "localhost.localdomain",
	})

	w.users["without_password"] = MustCreateUser(t, tx, &domain.User{
		Email:           "vagrant+test-without_password@localhost",
		Name:            "Passwordless User",
		UrlHost:         "localhost.localdomain",
		WithoutPassword: true,
	})

	w.users["non-member"] = MustCreateUser(t, tx, &domain.User{
		Email:    "vagrant+test-not-a-member@localhost",
		Name:     "Not A Member",
		Password: "password-is-long-enough",
		UrlHost:  "localhost.localdomain",
	})

	w.organizations["default"] = MustCreateOrganization(t, tx, &domain.Organization{
		Name:   "Acme Inc.",
		Public: false,
	})

	MustSelectBillingPlan(t, tx, w.organizations["default"], TestPlan)

	w.organizationMemberships["owner"] = MustCreateOrganizationMembership(t, tx, &domain.OrganizationMembership{
		UserUuid:         w.users["default"].Uuid,
		OrganizationUuid: w.organizations["default"].Uuid,
		Type:             "owner",
	})

	w.organizationMemberships["member"] = MustCreateOrganizationMembership(t, tx, &domain.OrganizationMembership{
		UserUuid:         w.users["other"].Uuid,
		OrganizationUuid: w.organizations["default"].Uuid,
		Type:             "member",
	})

	w.projects["public"] = MustCreateProject(t, tx, &domain.Project{
		OrganizationUuid: w.organizations["default"].Uuid,
		Name:             "Widgets",
		Public:           true,
	})

	w.projects["private"] = MustCreateProject(t, tx, &domain.Project{
		OrganizationUuid: w.organizations["default"].Uuid,
		Name:             "Private Widgets",
		Public:           false,
	})

	w.projectMemberships["member"] = MustCreateProjectMembership(t, tx, &domain.ProjectMembership{
		ProjectUuid:    w.projects["public"].Uuid,
		UserUuid:       w.users["default"].Uuid,
		MembershipType: domain.MembershipTypeMember,
	})

	w.projectMemberships["project-member-private"] = MustCreateProjectMembership(t, tx, &domain.ProjectMembership{
		ProjectUuid:    w.projects["private"].Uuid,
		UserUuid:       w.users["project-member"].Uuid,
		MembershipType: domain.MembershipTypeMember,
	})

	w.projectMemberships["project-owner-private"] = MustCreateProjectMembership(t, tx, &domain.ProjectMembership{
		ProjectUuid:    w.projects["private"].Uuid,
		UserUuid:       w.users["project-owner"].Uuid,
		MembershipType: domain.MembershipTypeOwner,
	})

	w.repositories["default"] = MustCreateRepository(t, tx, &domain.Repository{
		Name:        "test",
		Url:         "https://example.com/repositories/example.git",
		ProjectUuid: w.projects["public"].Uuid,
	})
	w.repositories["other"] = MustCreateRepository(t, tx, &domain.Repository{
		Name:        "other test repository",
		Url:         "git@example.com:repositories/example.git",
		ProjectUuid: w.projects["public"].Uuid,
	})

	w.environments["default"] = MustCreateEnvironment(t, tx, &domain.Environment{
		Name:        "default",
		ProjectUuid: w.projects["public"].Uuid,
		Variables: domain.EnvironmentVariables{
			M: map[string]string{
				"MY_ENV": "TEST",
			},
		},
	})

	w.environments["astley"] = MustCreateEnvironment(t, tx, &domain.Environment{
		Name:        "astley",
		ProjectUuid: w.projects["public"].Uuid,
		Variables: domain.EnvironmentVariables{
			M: map[string]string{
				"MY_ENV": "ASTLEY",
			},
		},
	})

	w.environments["private"] = MustCreateEnvironment(t, tx, &domain.Environment{
		Name:        "test",
		ProjectUuid: w.projects["private"].Uuid,
		Variables: domain.EnvironmentVariables{
			M: map[string]string{
				"MY_ENV": "TEST",
			},
		},
	})

	secretBytes := []byte(`{"PrivateKey":"private key","PublicKey":"public key"}`)
	w.secrets["default"] = MustCreateSecret(t, tx, ss, &domain.Secret{
		Name:            "test",
		EnvironmentUuid: w.environments["private"].Uuid,
		Type:            domain.SecretSsh,
		Status:          domain.SecretPresent,
		SecretBytes:     secretBytes,
		Key:             []byte("abcd1"),
	})
	envSecretBytes := []byte(`{"Value": "foo"}`)
	w.secrets["env"] = MustCreateSecret(t, tx, ss, &domain.Secret{
		Name:            "test EnvironmentSecret",
		EnvironmentUuid: w.environments["private"].Uuid,
		Type:            domain.SecretEnv,
		Status:          domain.SecretPresent,
		SecretBytes:     envSecretBytes,
		Key:             []byte("abcd2"),
	})

	w.tasks["default"] = MustCreateTask(t, tx, &domain.Task{
		Body:        `#!/bin/sh\nprintf "MY_ENV=%s\n" "$MY_ENV"\n`,
		Name:        "hello-world",
		ProjectUuid: w.projects["public"].Uuid,
		Type:        domain.TaskTypeScript,
	})

	w.tasks["other"] = MustCreateTask(t, tx, &domain.Task{
		Body:        `#!/bin/sh\nprintf "other\n"\n`,
		Name:        "other",
		ProjectUuid: w.projects["public"].Uuid,
		Type:        domain.TaskTypeScript,
	})

	w.jobs["default"] = MustCreateJob(t, tx, &domain.Job{
		EnvironmentUuid: w.environments["default"].Uuid,
		TaskUuid:        w.tasks["default"].Uuid,
		Name:            "test hello-world",
	})
	timespec := "0:00 Jan 01, 2000"
	w.schedules["default"] = MustCreateSchedule(t, tx, &domain.Schedule{
		UserUuid:     w.users["default"].Uuid,
		JobUuid:      w.jobs["default"].Uuid,
		Timespec:     &timespec,
		Description:  "0:00 on January 1st",
		TimezoneName: "UTC",
	})

	w.jobs["other"] = MustCreateJob(t, tx, &domain.Job{
		EnvironmentUuid: w.environments["private"].Uuid,
		TaskUuid:        w.tasks["default"].Uuid,
		Name:            "test other",
	})
	timespec = "0:01 Jan 01, 2000"
	w.schedules["other"] = MustCreateSchedule(t, tx, &domain.Schedule{
		UserUuid:     w.users["default"].Uuid,
		JobUuid:      w.jobs["other"].Uuid,
		Timespec:     &timespec,
		Description:  "0:00 on January 1st",
		TimezoneName: "UTC",
	})

	w.workspaceBaseImages["default"] = MustCreateWorkspaceBaseImage(t, tx, &domain.WorkspaceBaseImage{
		Name:       "harrow-test",
		Repository: "https://github.com/harrowio/container-templates.git",
		Path:       "./harrow-test/",
		Ref:        "master",
		Type:       "container",
	})

	w.repositoryCredentials["default"] = MustCreateRepositoryCredential(t, tx, ss, &domain.RepositoryCredential{
		Name:           fmt.Sprintf("repository-%s", w.repositories["default"].Uuid),
		RepositoryUuid: w.repositories["default"].Uuid,
		Type:           domain.RepositoryCredentialSsh,
		Status:         domain.RepositoryCredentialPending,
		Key:            []byte("abcd2"),
	})

	w.repositoryCredentials["present"] = MustCreateRepositoryCredential(t, tx, ss, &domain.RepositoryCredential{
		Name:           fmt.Sprintf("repository-%s", w.repositories["other"].Uuid),
		RepositoryUuid: w.repositories["other"].Uuid,
		Type:           domain.RepositoryCredentialSsh,
		Status:         domain.RepositoryCredentialPresent,
		SecretBytes:    []byte(`{"PrivateKey":"priv","PublicKey": "public"}`),
		Key:            []byte("abcd2"),
	})

}

func (w *World) User(tag string) *domain.User {
	return w.users[tag]
}

func (w *World) Organization(tag string) *domain.Organization {
	return w.organizations[tag]
}

func (w *World) OrganizationMembership(tag string) *domain.OrganizationMembership {
	return w.organizationMemberships[tag]
}

func (w *World) Project(tag string) *domain.Project {
	return w.projects[tag]
}

func (w *World) ProjectMembership(tag string) *domain.ProjectMembership {
	return w.projectMemberships[tag]
}

func (w *World) Job(tag string) *domain.Job {
	return w.jobs[tag]
}

func (w *World) Task(tag string) *domain.Task {
	return w.tasks[tag]
}

func (w *World) Environment(tag string) *domain.Environment {
	return w.environments[tag]
}

func (w *World) WorkspaceBaseImage(tag string) *domain.WorkspaceBaseImage {
	return w.workspaceBaseImages[tag]
}

func (w *World) Repository(tag string) *domain.Repository {
	return w.repositories[tag]
}

func (w *World) Schedule(tag string) *domain.Schedule {
	return w.schedules[tag]
}

func (w *World) Secret(tag string) *domain.Secret {
	return w.secrets[tag]
}

func (w *World) RepositoryCredential(tag string) *domain.RepositoryCredential {
	return w.repositoryCredentials[tag]
}
