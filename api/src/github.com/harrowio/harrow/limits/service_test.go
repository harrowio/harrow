package limits

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/uuidhelper"
)

type ThingBelongingToOrganization struct {
	ProjectUuid string
}

func NewThingBelongingToOrganization(projectUuid string) *ThingBelongingToOrganization {
	return &ThingBelongingToOrganization{
		ProjectUuid: projectUuid,
	}
}

func (self *ThingBelongingToOrganization) FindProject(projects domain.ProjectStore) (*domain.Project, error) {
	return projects.FindByUuid(self.ProjectUuid)
}

func TestService_Exceeded_finds_organization_of_subject_through_project_uuid(t *testing.T) {
	project := &domain.Project{
		Uuid: "21d28ae8-8f02-4a13-b994-911adb53b9f5",
	}
	projects := NewDummyProjectStore().Add(project)
	organization := &domain.Organization{
		Uuid: "8dc275ca-899d-47de-9b41-4962724e3850",
	}
	now := time.Date(2016, 3, 4, 14, 14, 7, 0, time.UTC)
	organizations := NewDummyOrganizationStore().Add(project.Uuid, organization)
	billingPlans := NewDummyBillingPlanStore().Add(domain.FreePlan)
	billingHistory := NewDummyBillingHistory().Add(organization.Uuid, domain.FreePlan.Uuid)
	limits := NewDummyLimitsStore().Add(organization.Uuid, NewLimits(organization.Uuid, now))
	thing := NewThingBelongingToOrganization(project.Uuid)

	service := NewService(organizations, projects, billingPlans, billingHistory, limits)

	_, err := service.Exceeded(thing)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := service.Organization(), organization; !reflect.DeepEqual(got, want) {
		t.Errorf(`service.Organization() = %v; want %v`, got, want)
	}
}

func TestService_Exceeded_fetches_billing_plan_for_organization(t *testing.T) {
	project := &domain.Project{
		Uuid: "21d28ae8-8f02-4a13-b994-911adb53b9f5",
	}
	projects := NewDummyProjectStore().Add(project)
	organization := &domain.Organization{
		Uuid: "8dc275ca-899d-47de-9b41-4962724e3850",
	}
	now := time.Date(2016, 3, 4, 14, 14, 21, 0, time.UTC)
	organizations := NewDummyOrganizationStore().Add(project.Uuid, organization)
	billingPlans := NewDummyBillingPlanStore().Add(domain.FreePlan)
	billingHistory := NewDummyBillingHistory().Add(organization.Uuid, domain.FreePlan.Uuid)
	limits := NewDummyLimitsStore().Add(organization.Uuid, NewLimits(organization.Uuid, now))
	thing := NewThingBelongingToOrganization(project.Uuid)

	service := NewService(organizations, projects, billingPlans, billingHistory, limits)

	service.Exceeded(thing)

	if got, want := service.BillingPlan(), domain.FreePlan; !reflect.DeepEqual(got, want) {
		t.Errorf(`service.BillingPlan() = %v; want %v`, got, want)
	}
}

func TestService_Exceeded_respects_extra_projects_and_users_from_billing_history(t *testing.T) {
	project := &domain.Project{
		Uuid: "21d28ae8-8f02-4a13-b994-911adb53b9f5",
	}
	projects := NewDummyProjectStore().Add(project)
	organization := &domain.Organization{
		Uuid: "8dc275ca-899d-47de-9b41-4962724e3850",
	}
	now := time.Date(2016, 3, 4, 14, 14, 21, 0, time.UTC)
	organizations := NewDummyOrganizationStore().Add(project.Uuid, organization)
	billingPlans := NewDummyBillingPlanStore().Add(domain.FreePlan)
	billingHistory := NewDummyBillingHistory().
		Add(organization.Uuid, domain.FreePlan.Uuid).
		SetExtraProjects(organization.Uuid, 100).
		SetExtraUsers(organization.Uuid, 100)
	actualLimits := NewLimits(organization.Uuid, now)
	addProjectsToLimits(organization.Uuid, actualLimits, 100)
	addUsersToLimits(organization.Uuid, actualLimits, 100)
	limits := NewDummyLimitsStore().Add(organization.Uuid, actualLimits)
	thing := NewThingBelongingToOrganization(project.Uuid)

	service := NewService(organizations, projects, billingPlans, billingHistory, limits)

	exceeded, err := service.Exceeded(thing)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := exceeded, false; got != want {
		t.Errorf(`exceeded = %v; want %v`, got, want)
	}

}

func addProjectsToLimits(organizationUuid string, limits *Limits, count int) {
	for i := 0; i < count; i++ {
		project := &domain.Project{
			Uuid:             uuidhelper.MustNewV4(),
			OrganizationUuid: organizationUuid,
			Name:             fmt.Sprintf("Project %d", count),
		}

		limits.HandleActivity(activities.ProjectCreated(project))
	}
}

func addUsersToLimits(organizationUuid string, limits *Limits, count int) {
	project := &domain.Project{
		Uuid:             uuidhelper.MustNewV4(),
		OrganizationUuid: organizationUuid,
		Name:             fmt.Sprintf("Project %d", count),
	}

	limits.HandleActivity(activities.ProjectCreated(project))

	for i := 0; i < count; i++ {
		user := &domain.User{
			Uuid: uuidhelper.MustNewV4(),
			Name: fmt.Sprintf("User %d", count),
		}

		limits.HandleActivity(activities.UserJoinedProject(user, project))
	}

}
