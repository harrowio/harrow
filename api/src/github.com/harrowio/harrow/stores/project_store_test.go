// Package stores provides storage objects, they all have a trivial interface
// for finding by UUID, updateing and deleting
package stores_test

import (
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
)

func Test_ProjectStore_FindByWebhookUuid_returnsProject(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	project := world.Project("public")
	user := world.User("default")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if uuid, err := stores.NewDbWebhookStore(tx).Create(webhook); err != nil {
		t.Fatal(err)
	} else {
		webhook.Uuid = uuid
	}

	projectStore := stores.NewDbProjectStore(tx)
	found, err := projectStore.FindByWebhookUuid(webhook.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := found.Uuid, project.Uuid; got != want {
		t.Fatalf("found.Uuid = %q; want %q", got, want)
	}
}

func Test_ProjectStore_FindByWebhookUuid_doesNotReturnArchivedProjects(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	project := world.Project("public")
	user := world.User("default")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if uuid, err := stores.NewDbWebhookStore(tx).Create(webhook); err != nil {
		t.Fatal(err)
	} else {
		webhook.Uuid = uuid
	}

	projectStore := stores.NewDbProjectStore(tx)
	if err := projectStore.ArchiveByUuid(project.Uuid); err != nil {
		t.Fatal(err)
	}
	_, err := projectStore.FindByWebhookUuid(webhook.Uuid)
	if err == nil {
		t.Fatal("Expected an error")
	}

	if derr, ok := err.(*domain.NotFoundError); !ok {
		t.Fatalf("err.(type) = %T; want %T", err, derr)
	}
}

func Test_ProjectStore_Update_returnsDomainNotFoundError_whenUpdatingNonExistingProject(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbProjectStore(tx)
	got := store.Update(&domain.Project{})
	want, ok := got.(*domain.NotFoundError)
	if !ok {
		t.Errorf("err.(type) = %T; want %T", got, want)
	}
}

// func Test_ProjectStore_SuccessfullyCreatingANewProject(t *testing.T) {
//
// 	tx := helpers.GetDbTx(t)
// 	defer tx.Rollback()
//
// 	o := helpers.MustCreateOrganization(t, tx, &domain.Organization{Name: "Test Organization"})
// 	p := &domain.Project{Name: "Test Project", OrganizationUuid: o.Uuid}
//
// 	_, err := stores.NewDbProjectStore(tx).Create(p)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// }
//
// func Test_ProjectStore_SucessfullyDeleteProject(t *testing.T) {
//
// 	var err error
//
// 	tx := helpers.GetDbTx(t)
// 	defer tx.Rollback()
//
// 	ps := stores.NewDbProjectStore(tx)
//
// 	o := helpers.MustCreateOrganization(t, tx, &domain.Organization{Name: "Test Organization"})
//
// 	p := helpers.MustCreateProject(t, tx, &domain.Project{
// 		Name:             "Example Project 1",
// 		OrganizationUuid: o.Uuid,
// 	})
//
// 	err = ps.DeleteByUuid(p.Uuid)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	_, err = ps.FindByUuid(p.Uuid)
// 	if err == nil {
// 		t.Fatal("Expected *domain.NotFoundError")
// 	}
//
// }
//
// func Test_ProjectStore_FailToDeleteNonExistentProject(t *testing.T) {
//
// 	tx := helpers.GetDbTx(t)
// 	defer tx.Rollback()
//
// 	err := stores.NewDbProjectStore(tx).DeleteByUuid("11111111-1111-4111-a111-111111111111")
// 	if _, ok := err.(*domain.NotFoundError); !ok {
// 		t.Fatal("Expected to get a *domain.NotFoundError, got:", err)
// 	}
//
// }
//
// func Test_ProjectStore_FailingToCreateUserDueToShortName(t *testing.T) {
//
// 	tx := helpers.GetDbTx(t)
// 	defer tx.Rollback()
//
// 	store := stores.NewDbProjectStore(tx)
//
// 	_, err := store.Create(&domain.Project{Name: ""})
// 	if ve, ok := err.(*domain.ValidationError); ok {
// 		if ve.Errors["name"][0] != "required" {
// 			t.Fatalf("Expected err.Errors[\"name\"] to be `required' got: %v", ve.Errors["name"])
// 		}
// 	} else {
// 		t.Fatalf("Expected to get a domain.ValidationError, got: %v", err)
// 	}
//
// }
//
// func Test_ProjectStore_LookingUpAllProjectsByOrganizationUuid(t *testing.T) {
//
// 	var err error
//
// 	tx := helpers.GetDbTx(t)
// 	defer tx.Rollback()
//
// 	projectStore := stores.NewDbProjectStore(tx)
//
// 	o := helpers.MustCreateOrganization(t, tx, &domain.Organization{Name: "Example"})
//
// 	p := helpers.MustCreateProject(t, tx, &domain.Project{
// 		Name:             "Example Project 1",
// 		OrganizationUuid: o.Uuid,
// 	})
//
// 	projects, err := projectStore.FindAllByOrganizationUuid(o.Uuid)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	if len(projects) != 1 {
// 		t.Fatalf("Expected len(projects) != 1, got %v", len(projects))
// 	}
//
// 	if projects[0].Uuid != p.Uuid {
// 		t.Fatalf("Expected projects[0].Uuid != p.Uuid, got %v", p.Uuid)
// 	}
//
// }
//
// func Test_ProjectStore_LookingUpProjectForOperationUuid(t *testing.T) {
// 	tx := helpers.GetDbTx(t)
// 	defer tx.Rollback()
//
// 	projectStore := stores.NewDbProjectStore(tx)
//
// 	o := helpers.MustCreateOrganization(t, tx, &domain.Organization{Name: "Example"})
//
// 	p := helpers.MustCreateProject(t, tx, &domain.Project{
// 		Name:             "Example Project 1",
// 		OrganizationUuid: o.Uuid,
// 	})
//
// 	helpers.MustCreateProject(t, tx, &domain.Project{
// 		Name:             "Example Project 2",
// 		OrganizationUuid: o.Uuid,
// 	})
//
// 	task := helpers.MustCreateTask(t, tx, &domain.Task{
// 		Name:        "Example Task 1",
// 		ProjectUuid: p.Uuid,
// 		Type:        domain.TaskTypeScript,
// 	})
//
// 	env := helpers.MustCreateEnvironment(t, tx, &domain.Environment{
// 		Name:        "Example Environment 1",
// 		ProjectUuid: p.Uuid,
// 		Variables: domain.EnvironmentVariables{M: map[string]string{
// 			"VAR": "VALUE",
// 		}},
// 	})
//
// 	job := helpers.MustCreateJob(t, tx, &domain.Job{
// 		Name:            "Example Job 1",
// 		TaskUuid:        task.Uuid,
// 		EnvironmentUuid: env.Uuid,
// 	})
//
// 	op := helpers.MustCreateOperation(t, tx, &domain.Operation{
// 		Type:                   domain.OperationTypeJobScheduled,
// 		JobUuid:                &job.Uuid,
// 		WorkspaceBaseImageUuid: "31b0127a-6d63-4d22-b32b-e1cfc04f4007",
// 	})
//
// 	project, err := projectStore.FindForAction("operations", op.Uuid)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	if project.Uuid != p.Uuid {
// 		t.Fatalf("Expected to find project %s, got %s\n", p.Uuid, project.Uuid)
// 	}
//
// }
