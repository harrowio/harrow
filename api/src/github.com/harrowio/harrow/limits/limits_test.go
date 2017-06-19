package limits

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/uuidhelper"
)

func TestLimits_ignores_other_organizations(t *testing.T) {
	organizationUuid := "474e42bf-3d88-4242-96b2-7b950fa3bfe2"
	otherOrganizationUuid := "88e7349f-20e3-49d8-80d0-d5d6ca2c2e36"

	organization := &domain.Organization{
		Uuid: organizationUuid,
	}
	otherOrganization := &domain.Organization{
		Uuid: otherOrganizationUuid,
	}

	now := time.Date(2016, 3, 3, 11, 10, 8, 0, time.UTC)

	limits := NewLimits(organizationUuid, now)

	organizationCreated := activities.OrganizationCreated(organization, domain.FreePlan)
	organizationCreated.OccurredOn = now.Add(-10 * 24 * time.Hour)

	otherOrganizationCreated := activities.OrganizationCreated(otherOrganization, domain.FreePlan)
	otherOrganizationCreated.OccurredOn = now

	if err := limits.HandleActivity(organizationCreated); err != nil {
		t.Fatal(err)
	}

	if err := limits.HandleActivity(otherOrganizationCreated); err != nil {
		t.Fatal(err)
	}

	if got, want := limits.NumberOfTrialDaysLeft(), 4; got != want {
		t.Errorf(`limits.NumberOfTrialDaysLeft() = %v; want %v`, got, want)
	}

}

func TestLimits_returns_fourteen_days_if_organization_was_created_today(t *testing.T) {
	organizationUuid := "474e42bf-3d88-4242-96b2-7b950fa3bfe2"
	organization := &domain.Organization{
		Uuid: organizationUuid,
	}
	now := time.Date(2016, 3, 3, 11, 10, 8, 0, time.UTC)
	limits := NewLimits(organizationUuid, now)
	organizationCreated := activities.OrganizationCreated(organization, domain.FreePlan)
	organizationCreated.OccurredOn = now

	if err := limits.HandleActivity(organizationCreated); err != nil {
		t.Fatal(err)
	}

	if got, want := limits.NumberOfTrialDaysLeft(), 14; got != want {
		t.Errorf(`limits.NumberOfTrialDaysLeft() = %v; want %v`, got, want)
	}
}

func TestLimits_counts_remaining_trial_days_from_organization_creation_date(t *testing.T) {
	organizationUuid := "474e42bf-3d88-4242-96b2-7b950fa3bfe2"
	organization := &domain.Organization{
		Uuid: organizationUuid,
	}
	now := time.Date(2016, 3, 3, 11, 10, 8, 0, time.UTC)
	limits := NewLimits(organizationUuid, now)
	organizationCreated := activities.OrganizationCreated(organization, domain.FreePlan)
	organizationCreated.OccurredOn = now.Add(-3 * 24 * time.Hour)

	if err := limits.HandleActivity(organizationCreated); err != nil {
		t.Fatal(err)
	}

	if got, want := limits.NumberOfTrialDaysLeft(), 11; got != want {
		t.Errorf(`limits.NumberOfTrialDaysLeft() = %v; want %v`, got, want)
	}
}

func TestLimits_returns_zero_days_if_the_trial_period_has_ended(t *testing.T) {
	organizationUuid := "474e42bf-3d88-4242-96b2-7b950fa3bfe2"
	organization := &domain.Organization{
		Uuid: organizationUuid,
	}
	now := time.Date(2016, 3, 3, 11, 10, 8, 0, time.UTC)
	limits := NewLimits(organizationUuid, now)
	organizationCreated := activities.OrganizationCreated(organization, domain.FreePlan)
	organizationCreated.OccurredOn = now.Add(-20 * 24 * time.Hour)

	if err := limits.HandleActivity(organizationCreated); err != nil {
		t.Fatal(err)
	}

	if got, want := limits.NumberOfTrialDaysLeft(), 0; got != want {
		t.Errorf(`limits.NumberOfTrialDaysLeft() = %v; want %v`, got, want)
	}
}

func TestLimits_counts_users_joining_a_project_as_members(t *testing.T) {
	organizationUuid := "8baa5c90-6365-4d11-af8c-aa23475ffd15"
	member := &domain.User{Uuid: "ae979a7f-9f85-48fb-8d8f-380cd4972eef"}
	project := &domain.Project{
		Uuid:             "79334b6b-5ab0-4a4e-8ba9-d29862985fae",
		OrganizationUuid: organizationUuid,
	}
	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 31, 23, 0, time.UTC))

	activity := activities.UserJoinedProject(member, project)

	if err := limits.HandleActivity(activity); err != nil {
		t.Fatal(err)
	}

	if got, want := limits.NumberOfMembers(), 1; got != want {
		t.Errorf(`limits.NumberOfMembers() = %v; want %v`, got, want)
	}
}

func TestLimits_does_not_count_the_user_joining_a_project_belonging_to_a_different_organization(t *testing.T) {
	organizationUuid := "8baa5c90-6365-4d11-af8c-aa23475ffd15"
	anotherOrganizationUuid := "37646bc6-3dfd-401d-bba9-3bb91b354f26"
	member := &domain.User{
		Uuid: "4a5db290-3475-44d9-921c-cea99d6114cd",
	}
	foreignProject := &domain.Project{
		Uuid:             "9d45f704-57ec-48ed-8cf3-59edd6315ba6",
		OrganizationUuid: anotherOrganizationUuid,
	}

	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 31, 27, 0, time.UTC))
	history := []*domain.Activity{
		activities.UserJoinedProject(member, foreignProject),
	}

	for _, activity := range history {
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	if got, want := limits.NumberOfMembers(), 0; got != want {
		t.Errorf(`limits.NumberOfMembers() = %v; want %v`, got, want)
	}
}

func TestLimits_does_not_count_the_same_user_twice_when_a_user_joins_a_project(t *testing.T) {
	organizationUuid := "8baa5c90-6365-4d11-af8c-aa23475ffd15"
	member := &domain.User{
		Uuid: "28ded9d4-2c55-4533-b394-e52ce371f2f8",
	}
	projects := []*domain.Project{
		{
			Uuid:             "d3ac68a7-fb99-4899-a395-ed247322910e",
			OrganizationUuid: organizationUuid,
		},
		{
			Uuid:             "4bae5f35-4888-4de1-bd8d-0b7cb10da5b3",
			OrganizationUuid: organizationUuid,
		},
	}
	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 31, 53, 0, time.UTC))
	history := []*domain.Activity{
		activities.ProjectCreated(projects[0]),
		activities.ProjectCreated(projects[1]),
		activities.UserJoinedProject(member, projects[0]),
		activities.UserJoinedProject(member, projects[1]),
	}

	for _, activity := range history {
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	if got, want := limits.NumberOfMembers(), 1; got != want {
		t.Errorf(`limits.NumberOfMembers() = %v; want %v`, got, want)
	}

}

func TestLimits_counts_the_number_of_projects_for_this_organization(t *testing.T) {
	organizationUuid := "986459b1-5f29-4d3d-87e3-e157fc208c2e"
	anotherOrganizationUuid := "9cf8b005-5b17-4abc-85e0-3e5e5dabdb17"

	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 31, 56, 0, time.UTC))

	ownProject := &domain.Project{
		OrganizationUuid: organizationUuid,
	}
	foreignProject := &domain.Project{
		OrganizationUuid: anotherOrganizationUuid,
	}

	activities := []*domain.Activity{
		activities.ProjectCreated(ownProject),
		activities.ProjectCreated(foreignProject),
	}

	for _, activity := range activities {
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	if got, want := limits.NumberOfProjects(), 1; got != want {
		t.Errorf(`limits.NumberOfProjects() = %v; want %v`, got, want)
	}
}

func TestLimits_decreases_the_number_of_projects_when_a_project_is_deleted(t *testing.T) {
	organizationUuid := "986459b1-5f29-4d3d-87e3-e157fc208c2e"
	anotherOrganizationUuid := "9cf8b005-5b17-4abc-85e0-3e5e5dabdb17"

	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 31, 59, 0, time.UTC))

	ownProject := &domain.Project{
		OrganizationUuid: organizationUuid,
	}
	foreignProject := &domain.Project{
		OrganizationUuid: anotherOrganizationUuid,
	}

	activities := []*domain.Activity{
		activities.ProjectCreated(ownProject),
		activities.ProjectCreated(foreignProject),
		activities.ProjectDeleted(foreignProject),
		activities.ProjectDeleted(ownProject),
	}

	for _, activity := range activities {
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	if got, want := limits.NumberOfProjects(), 0; got != want {
		t.Errorf(`limits.NumberOfProjects() = %v; want %v`, got, want)
	}
}

func TestLimits_decreases_the_number_of_members_when_a_user_is_removed_from_all_projects_of_the_organization(t *testing.T) {
	organizationUuid := "8baa5c90-6365-4d11-af8c-aa23475ffd15"
	member := &domain.User{
		Uuid: "28ded9d4-2c55-4533-b394-e52ce371f2f8",
	}
	projects := []*domain.Project{
		{
			Uuid:             "d3ac68a7-fb99-4899-a395-ed247322910e",
			OrganizationUuid: organizationUuid,
		},
		{
			Uuid:             "4bae5f35-4888-4de1-bd8d-0b7cb10da5b3",
			OrganizationUuid: organizationUuid,
		},
	}
	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 32, 3, 0, time.UTC))
	history := []*domain.Activity{
		activities.ProjectCreated(projects[0]),
		activities.ProjectCreated(projects[1]),
		activities.UserJoinedProject(member, projects[0]),
		activities.UserJoinedProject(member, projects[1]),
		activities.UserLeftProject(member, projects[0]),
	}

	for _, activity := range history {
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	if got, want := limits.NumberOfMembers(), 1; got != want {
		t.Errorf(`limits.NumberOfMembers() = %v; want %v`, got, want)
	}

	if err := limits.HandleActivity(activities.UserRemovedFromProject(member, projects[1])); err != nil {
		t.Fatal(err)
	}

	if got, want := limits.NumberOfMembers(), 0; got != want {
		t.Errorf(`limits.NumberOfMembers() = %v; want %v`, got, want)
	}
}

func TestLimits_counts_number_of_successful_operations_for_this_organization_by_year_and_month(t *testing.T) {
	organizationUuid := "55fb1519-d4bd-4e9f-a3ff-2d3838d3fff3"
	project := &domain.Project{
		Uuid:             "a44435cb-bc77-4e59-bff7-9330bf7efd67",
		OrganizationUuid: organizationUuid,
	}

	occurredOn := func(date time.Time, activity *domain.Activity) *domain.Activity {
		activity.OccurredOn = date
		activity.SetProjectUuid(project.Uuid)
		return activity
	}

	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 32, 6, 0, time.UTC))
	history := []*domain.Activity{
		occurredOn(
			time.Date(2016, 1, 1, 11, 0, 0, 0, time.UTC),
			activities.ProjectCreated(project),
		),
		occurredOn(
			time.Date(2016, 1, 1, 12, 0, 0, 0, time.UTC),
			activities.OperationSucceeded(nil),
		),
		occurredOn(
			time.Date(2016, 2, 1, 12, 0, 0, 0, time.UTC),
			activities.OperationSucceeded(nil),
		),
	}

	for _, activity := range history {
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	if got, want := limits.NumberOfSuccessfulOperationsIn(2016, 1), 1; got != want {
		t.Errorf(`limits.SuccessfulOperationIn(2016, 1) = %v; want %v`, got, want)
	}

	if got, want := limits.NumberOfSuccessfulOperationsIn(2016, 2), 1; got != want {
		t.Errorf(`limits.SuccessfulOperationsIn(2016 = %v; want %v`, got, want)
	}

	if got, want := limits.NumberOfSuccessfulOperationsIn(2016, 3), 0; got != want {
		t.Errorf(`limits.SuccessfulOperationsIn(2016, 3) = %v; want %v`, got, want)
	}

}

func TestLimits_does_not_count_successful_operations_of_other_organizations(t *testing.T) {
	organizationUuid := "55fb1519-d4bd-4e9f-a3ff-2d3838d3fff3"
	project := &domain.Project{
		Uuid:             "a44435cb-bc77-4e59-bff7-9330bf7efd67",
		OrganizationUuid: organizationUuid,
	}
	anotherProjectUuid := "97f0c931-655b-4f6b-9aba-cadca26378eb"
	foreignOperation := activities.OperationSucceeded(nil)
	foreignOperation.OccurredOn = time.Date(2016, 1, 1, 12, 0, 0, 0, time.UTC)
	foreignOperation.SetProjectUuid(anotherProjectUuid)

	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 32, 8, 0, time.UTC))
	history := []*domain.Activity{
		activities.ProjectCreated(project),
		foreignOperation,
	}

	for _, activity := range history {
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	if got, want := limits.NumberOfSuccessfulOperationsIn(2016, 1), 0; got != want {
		t.Errorf(`limits.SuccessfulOperationIn(2016, 1) = %v; want %v`, got, want)
	}
}

func TestLimits_counts_private_and_public_repositories(t *testing.T) {

	organizationUuid := "55fb1519-d4bd-4e9f-a3ff-2d3838d3fff3"
	project := &domain.Project{
		Uuid:             "a44435cb-bc77-4e59-bff7-9330bf7efd67",
		OrganizationUuid: organizationUuid,
	}
	privateRepository := &domain.Repository{
		Uuid:        "b4d125f7-dd40-4665-8707-f63bc518d2df",
		ProjectUuid: project.Uuid,
	}
	publicRepository := &domain.Repository{
		Uuid:        "6778a012-7599-43e5-b5f6-3c772f76f513",
		ProjectUuid: project.Uuid,
	}

	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 32, 11, 0, time.UTC))
	history := []*domain.Activity{
		activities.ProjectCreated(project),
		activities.RepositoryAdded(publicRepository),
		activities.RepositoryAdded(privateRepository),
		activities.RepositoryDetectedAsPrivate(privateRepository),
	}

	for _, activity := range history {
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	if got, want := limits.NumberOfPrivateRepositories(), 1; got != want {
		t.Errorf(`limits.NumberOfPrivateRepositories() = %v; want %v`, got, want)
	}

	if got, want := limits.NumberOfPublicRepositories(), 1; got != want {
		t.Errorf(`limits.NumberOfPublicRepositories() = %v; want %v`, got, want)
	}

}

func TestLimits_marks_repository_as_public_if_it_is_detected_as_public(t *testing.T) {

	organizationUuid := "55fb1519-d4bd-4e9f-a3ff-2d3838d3fff3"
	project := &domain.Project{
		Uuid:             "a44435cb-bc77-4e59-bff7-9330bf7efd67",
		OrganizationUuid: organizationUuid,
	}
	repository := &domain.Repository{
		Uuid:        "b4d125f7-dd40-4665-8707-f63bc518d2df",
		ProjectUuid: project.Uuid,
	}

	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 32, 15, 0, time.UTC))
	history := []*domain.Activity{
		activities.ProjectCreated(project),
		activities.RepositoryAdded(repository),
		activities.RepositoryDetectedAsPrivate(repository),
		activities.RepositoryDetectedAsPublic(repository),
	}

	for _, activity := range history {
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	if got, want := limits.NumberOfPrivateRepositories(), 0; got != want {
		t.Errorf(`limits.NumberOfPrivateRepositories() = %v; want %v`, got, want)
	}

	if got, want := limits.NumberOfPublicRepositories(), 1; got != want {
		t.Errorf(`limits.NumberOfPublicRepositories() = %v; want %v`, got, want)
	}

}

func TestLimits_does_not_count_repositories_beloning_to_another_organization(t *testing.T) {

	organizationUuid := "55fb1519-d4bd-4e9f-a3ff-2d3838d3fff3"
	anotherOrganizationUuid := "c2f01a69-e83f-40b0-8935-05820d230dee"
	project := &domain.Project{
		Uuid:             "a44435cb-bc77-4e59-bff7-9330bf7efd67",
		OrganizationUuid: anotherOrganizationUuid,
	}
	privateRepository := &domain.Repository{
		Uuid:        "b4d125f7-dd40-4665-8707-f63bc518d2df",
		ProjectUuid: project.Uuid,
	}
	publicRepository := &domain.Repository{
		Uuid:        "6778a012-7599-43e5-b5f6-3c772f76f513",
		ProjectUuid: project.Uuid,
	}

	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 32, 17, 0, time.UTC))
	history := []*domain.Activity{
		activities.ProjectCreated(project),
		activities.RepositoryAdded(publicRepository),
		activities.RepositoryAdded(privateRepository),
		activities.RepositoryDetectedAsPrivate(privateRepository),
	}

	for _, activity := range history {
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	if got, want := limits.NumberOfPrivateRepositories(), 0; got != want {
		t.Errorf(`limits.NumberOfPrivateRepositories() = %v; want %v`, got, want)
	}

	if got, want := limits.NumberOfPublicRepositories(), 0; got != want {
		t.Errorf(`limits.NumberOfPublicRepositories() = %v; want %v`, got, want)
	}

}

func TestLimits_does_not_handle_activities_older_than_its_current_state(t *testing.T) {
	organizationUuid := "83c85452-456a-4333-ae3c-6dae0be16216"
	projects := []*domain.Project{
		{
			Uuid:             "4952fd60-d255-4264-a05e-afb1acaee3ed",
			OrganizationUuid: organizationUuid,
		},
		{
			Uuid:             "42eba2f6-53f5-4a72-9644-e41b2cf5d951",
			OrganizationUuid: organizationUuid,
		},
	}
	history := []*domain.Activity{
		activities.ProjectCreated(projects[0]),
		activities.ProjectCreated(projects[1]),
	}
	history[0].OccurredOn = time.Now()
	history[1].OccurredOn = history[0].OccurredOn.Add(-1 * time.Hour)

	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 32, 19, 0, time.UTC))

	for _, activity := range history {
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	if got, want := limits.NumberOfProjects(), 1; got != want {
		t.Errorf(`limits.NumberOfProjects() = %v; want %v`, got, want)
	}
}

func TestLimits_records_the_occurrence_date_of_the_last_activity_handled_as_version(t *testing.T) {
	organizationUuid := "83c85452-456a-4333-ae3c-6dae0be16216"
	projects := []*domain.Project{
		{
			Uuid:             "4952fd60-d255-4264-a05e-afb1acaee3ed",
			OrganizationUuid: organizationUuid,
		},
		{
			Uuid:             "42eba2f6-53f5-4a72-9644-e41b2cf5d951",
			OrganizationUuid: organizationUuid,
		},
	}
	history := []*domain.Activity{
		activities.ProjectCreated(projects[0]),
		activities.ProjectCreated(projects[1]),
	}
	history[0].OccurredOn = time.Now()
	history[1].OccurredOn = history[0].OccurredOn.Add(1 * time.Hour)

	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 32, 21, 0, time.UTC))

	for _, activity := range history {
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	if got, want := limits.Version(), history[1].OccurredOn; got != want {
		t.Errorf(`limits.Version() = %v; want %v`, got, want)
	}
}

func TestLimits_can_be_serialized_and_deserialized_as_JSON_without_losing_state(t *testing.T) {

	organizationUuid := "55fb1519-d4bd-4e9f-a3ff-2d3838d3fff3"
	project := &domain.Project{
		Uuid:             "a44435cb-bc77-4e59-bff7-9330bf7efd67",
		OrganizationUuid: organizationUuid,
	}
	privateRepository := &domain.Repository{
		Uuid:        "b4d125f7-dd40-4665-8707-f63bc518d2df",
		ProjectUuid: project.Uuid,
	}
	publicRepository := &domain.Repository{
		Uuid:        "6778a012-7599-43e5-b5f6-3c772f76f513",
		ProjectUuid: project.Uuid,
	}

	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 32, 24, 0, time.UTC))
	beforeSerialization := []*domain.Activity{
		activities.ProjectCreated(project),
		activities.RepositoryAdded(publicRepository),
		activities.RepositoryAdded(privateRepository),
	}
	afterDeserialization := []*domain.Activity{
		activities.RepositoryDetectedAsPrivate(privateRepository),
	}

	for _, activity := range beforeSerialization {
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	data, err := json.Marshal(limits)
	if err != nil {
		t.Fatal(err)
	}

	deserialized := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 32, 26, 0, time.UTC))
	if err := json.Unmarshal(data, deserialized); err != nil {
		t.Fatal(err)
	}

	for _, activity := range afterDeserialization {
		if err := deserialized.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	if got, want := deserialized.NumberOfPublicRepositories(), 1; got != want {
		t.Errorf(`deserialized.NumberOfPublicRepositories() = %v; want %v`, got, want)
	}

	if got, want := deserialized.NumberOfPrivateRepositories(), 1; got != want {
		t.Errorf(`deserialized.NumberOfPrivateRepositories() = %v; want %v`, got, want)
	}
}

func TestLimits_Report_returns_a_domain_subject_representing_this_instance(t *testing.T) {

	organizationUuid := "55fb1519-d4bd-4e9f-a3ff-2d3838d3fff3"
	project := &domain.Project{
		Uuid:             "a44435cb-bc77-4e59-bff7-9330bf7efd67",
		OrganizationUuid: organizationUuid,
	}
	privateRepository := &domain.Repository{
		Uuid:        "b4d125f7-dd40-4665-8707-f63bc518d2df",
		ProjectUuid: project.Uuid,
	}
	publicRepository := &domain.Repository{
		Uuid:        "6778a012-7599-43e5-b5f6-3c772f76f513",
		ProjectUuid: project.Uuid,
	}

	latestVersion := time.Now()
	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 32, 29, 0, time.UTC))
	history := []*domain.Activity{
		activities.ProjectCreated(project),
		activities.RepositoryAdded(publicRepository),
		activities.RepositoryAdded(privateRepository),
		activities.RepositoryDetectedAsPrivate(privateRepository),
	}

	for i, activity := range history {
		activity.OccurredOn = latestVersion.Add(
			(time.Duration)(-(len(history) - i - 1)) * time.Hour,
		)
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	subject := limits.Report(nil, 0, 0)

	if got, want := subject.Projects, 1; got != want {
		t.Errorf(`subject.Projects = %v; want %v`, got, want)
	}

	if got, want := subject.PublicRepositories, 1; got != want {
		t.Errorf(`subject.PublicRepositories = %v; want %v`, got, want)
	}

	if got, want := subject.PrivateRepositories, 1; got != want {
		t.Errorf(`subject.PrivateRepositories = %v; want %v`, got, want)
	}

	if got, want := subject.OrganizationUuid, organizationUuid; got != want {
		t.Errorf(`subject.OrganizationUuid = %v; want %v`, got, want)
	}

	if got, want := subject.Version, latestVersion; got != want {
		t.Errorf(`subject.Version = %v; want %v`, got, want)
	}
}

func TestLimits_Report_includes_maximum_values_from_billing_plan(t *testing.T) {

	organizationUuid := "55fb1519-d4bd-4e9f-a3ff-2d3838d3fff3"
	project := &domain.Project{
		Uuid:             "a44435cb-bc77-4e59-bff7-9330bf7efd67",
		OrganizationUuid: organizationUuid,
	}
	users := []*domain.User{
		{
			Uuid: "d2c8c46a-d835-4b2f-b763-5d082aa58bf3",
		},
		{
			Uuid: "7f2c77c3-1f83-40a0-8bfe-b6eb91de71b8",
		},
	}
	privateRepository := &domain.Repository{
		Uuid:        "b4d125f7-dd40-4665-8707-f63bc518d2df",
		ProjectUuid: project.Uuid,
	}
	publicRepository := &domain.Repository{
		Uuid:        "6778a012-7599-43e5-b5f6-3c772f76f513",
		ProjectUuid: project.Uuid,
	}

	latestVersion := time.Now()
	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 32, 31, 0, time.UTC))
	history := []*domain.Activity{
		activities.ProjectCreated(project),
		activities.UserJoinedProject(users[0], project),
		activities.UserJoinedProject(users[1], project),
		activities.RepositoryAdded(publicRepository),
		activities.RepositoryAdded(privateRepository),
		activities.RepositoryDetectedAsPrivate(privateRepository),
	}

	plan := &domain.BillingPlan{
		Name:                 "Free",
		UsersIncluded:        1,
		ProjectsIncluded:     1,
		PrivateCodeAvailable: false,
	}

	for i, activity := range history {
		activity.OccurredOn = latestVersion.Add(
			(time.Duration)(-(len(history) - i - 1)) * time.Hour,
		)
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	subject := limits.Report(plan, 0, 0)

	if got := subject.Plan; got == nil {
		t.Fatalf(`subject.Plan is nil`)
	}

	if got, want := subject.Plan.UsersExceedingLimit, 1; got != want {
		t.Errorf(`subject.Plan.UsersExceedingLimit = %v; want %v`, got, want)
	}

	if got, want := subject.Plan.ProjectsExceedingLimit, 0; got != want {
		t.Errorf(`subject.Plan.ProjectsExceedingLimit = %v; want %v`, got, want)
	}

	if got, want := subject.Plan.RequiresUpgradeForPrivateCode, true; got != want {
		t.Errorf(`subject.Plan.RequiresUpgradeForPrivateCode = %v; want %v`, got, want)
	}

	if got, want := subject.Plan.UsersIncluded, plan.UsersIncluded; got != want {
		t.Errorf(`subject.Plan.UsersIncluded = %v; want %v`, got, want)
	}

	if got, want := subject.Plan.ProjectsIncluded, plan.ProjectsIncluded; got != want {
		t.Errorf(`subject.Plan.ProjectsIncluded = %v; want %v`, got, want)
	}
}

func TestLimits_Report_includes_number_of_trial_days_left(t *testing.T) {

	organizationUuid := "55fb1519-d4bd-4e9f-a3ff-2d3838d3fff3"
	organizationCreated := activities.OrganizationCreated(&domain.Organization{
		Uuid: organizationUuid,
	}, domain.FreePlan)
	now := time.Date(2016, 3, 3, 11, 32, 31, 0, time.UTC)
	organizationCreated.OccurredOn = now

	limits := NewLimits(organizationUuid, now)
	history := []*domain.Activity{
		organizationCreated,
	}

	for _, activity := range history {
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	subject := limits.Report(domain.FreePlan, 0, 0)

	if got := subject.Plan; got == nil {
		t.Fatalf(`subject.Plan is nil`)
	}

	if got, want := subject.TrialDaysLeft, 14; got != want {
		t.Errorf(`subject.TrialDaysLeft = %v; want %v`, got, want)
	}
}

func TestLimits_Report_returns_zero_projects_above_limit_for_Platinum_plan(t *testing.T) {
	organizationUuid := "83c85452-456a-4333-ae3c-6dae0be16216"
	history := []*domain.Activity{}
	for i := 0; i < domain.PlatinumPlan.ProjectsIncluded*2; i++ {
		project := &domain.Project{
			Uuid:             uuidhelper.MustNewV4(),
			OrganizationUuid: organizationUuid,
		}

		history = append(history, activities.ProjectCreated(project))
	}

	limits := NewLimits(organizationUuid, time.Date(2016, 3, 3, 11, 32, 21, 0, time.UTC))

	for _, activity := range history {
		if err := limits.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	plan := limits.Report(domain.PlatinumPlan, 0, 0).Plan

	if got, want := plan.ProjectsExceedingLimit, 0; got != want {
		t.Errorf(`plan.ProjectsExceedingLimit = %v; want %v`, got, want)
	}
}
