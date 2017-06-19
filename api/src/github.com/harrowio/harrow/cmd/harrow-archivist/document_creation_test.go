package harrowArchivist

import (
	"reflect"
	"testing"
	"time"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
)

func TestDocumentCreation_Activity_returns_an_activity_if_none_exists_yet(t *testing.T) {
	subject := &domain.Organization{
		Uuid: "789f1e9d-ad17-4b63-9471-9f5a6fc333a9",
	}
	activitiesInMemory := NewActivitiesInMemory()
	command := NewDocumentCreation(subject, activitiesInMemory)
	activity, err := command.Activity(activities.OrganizationCreated(nil, nil))
	if err != nil {
		t.Fatal(err)
	}

	if got := activity; got == nil {
		t.Fatalf(`activity is nil`)
	}

}

func TestDocumentCreation_Activity_uses_the_name_of_the_provided_activity(t *testing.T) {
	testCases := []struct {
		subject   CreationSubject
		prototype *domain.Activity
	}{
		{
			subject:   &domain.Organization{},
			prototype: activities.OrganizationCreated(nil, nil),
		},
		{
			subject:   &domain.Project{},
			prototype: activities.ProjectCreated(nil),
		},
		{
			subject:   &domain.Task{},
			prototype: activities.TaskAdded(nil),
		},
	}

	for _, testCase := range testCases {
		activitiesInMemory := NewActivitiesInMemory()
		command := NewDocumentCreation(testCase.subject, activitiesInMemory)
		activity, err := command.Activity(testCase.prototype)
		if err != nil {
			t.Fatal(err)
		}

		if got, want := activity.Name, testCase.prototype.Name; got != want {
			t.Errorf(`activity.Name = %v; want %v`, got, want)
		}
	}
}

func TestDocumentCreation_Activity_uses_the_provided_subject_as_the_payload(t *testing.T) {
	subject := &domain.Project{
		Uuid: "789f1e9d-ad17-4b63-9471-9f5a6fc333a9",
	}
	activitiesInMemory := NewActivitiesInMemory()
	command := NewDocumentCreation(subject, activitiesInMemory)
	activity, err := command.Activity(activities.ProjectCreated(nil))
	if err != nil {
		t.Fatal(err)
	}

	if got, want := activity.Payload, subject; !reflect.DeepEqual(got, want) {
		t.Errorf(`activity.Payload = %v; want %v`, got, want)
	}
}

func TestDocumentCreation_Activity_marks_organizations_as_being_created_with_the_free_billing_plan(t *testing.T) {
	subject := &domain.Organization{
		Uuid: "789f1e9d-ad17-4b63-9471-9f5a6fc333a9",
	}
	activitiesInMemory := NewActivitiesInMemory()
	command := NewDocumentCreation(subject, activitiesInMemory)
	activity, err := command.Activity(activities.OrganizationCreated(nil, nil))
	if err != nil {
		t.Fatal(err)
	}

	payload := activities.OrganizationCreated(subject, domain.FreePlan).Payload
	if got, want := activity.Payload, payload; !reflect.DeepEqual(got, want) {
		t.Errorf(`activity.Payload = %v; want %v`, got, want)
	}
}

func TestDocumentCreation_Activity_dates_the_activity_to_the_creation_date_of_the_subject(t *testing.T) {
	createdAt := time.Date(2016, 3, 1, 14, 40, 26, 0, time.UTC)
	subject := &domain.Organization{
		Uuid:      "789f1e9d-ad17-4b63-9471-9f5a6fc333a9",
		CreatedAt: createdAt,
	}
	activitiesInMemory := NewActivitiesInMemory()
	command := NewDocumentCreation(subject, activitiesInMemory)
	activity, err := command.Activity(activities.OrganizationCreated(nil, nil))
	if err != nil {
		t.Fatal(err)
	}

	if got, want := activity.OccurredOn, createdAt; got != want {
		t.Errorf(`activity.OccurredOn = %v; want %v`, got, want)
	}
}

func TestDocumentCreation_Activity_initializes_the_Extra_field(t *testing.T) {
	createdAt := time.Date(2016, 3, 1, 14, 40, 26, 0, time.UTC)
	subject := &domain.Organization{
		Uuid:      "789f1e9d-ad17-4b63-9471-9f5a6fc333a9",
		CreatedAt: createdAt,
	}
	activitiesInMemory := NewActivitiesInMemory()
	command := NewDocumentCreation(subject, activitiesInMemory)
	activity, err := command.Activity(activities.OrganizationCreated(nil, nil))
	if err != nil {
		t.Fatal(err)
	}

	if got := activity.Extra; got == nil {
		t.Fatalf(`activity.Extra is nil`)
	}

}

func TestDocumentCreation_Activity_returns_nil_if_no_activity_needs_to_be_created(t *testing.T) {
	subject := &domain.Project{
		Uuid: "789f1e9d-ad17-4b63-9471-9f5a6fc333a9",
	}
	activitiesInMemory := NewActivitiesInMemory()
	activitiesInMemory.Add(activities.ProjectCreated(subject))

	command := NewDocumentCreation(subject, activitiesInMemory)
	activity, err := command.Activity(activities.ProjectCreated(nil))
	if err != nil {
		t.Fatal(err)
	}

	if got, want := activity, (*domain.Activity)(nil); got != want {
		t.Errorf(`activity = %v; want %v`, got, want)
	}
}
