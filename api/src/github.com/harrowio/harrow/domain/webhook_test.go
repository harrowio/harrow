package domain

import (
	"reflect"
	"testing"
)

func Test_Webhook_OwnUrl(t *testing.T) {
	webhook := &Webhook{
		Uuid: "607982b8-b0f1-4762-8bd1-ac230310ef8e",
	}

	actual := webhook.OwnUrl("http", "www.example.com")
	expected := "http://www.example.com/webhooks/607982b8-b0f1-4762-8bd1-ac230310ef8e"
	if actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}

func Test_Webhook_Links(t *testing.T) {
	webhook := &Webhook{
		Uuid:        "607982b8-b0f1-4762-8bd1-ac230310ef8e",
		ProjectUuid: "07597891-3495-46af-b702-babfaacec89a",
		CreatorUuid: "a175d3c4-8bc9-42f9-afd1-adeaf9d2e306",
		JobUuid:     "6d17d7f0-06aa-47e4-b5a3-6c20b4005f40",
	}

	expected := map[string]map[string]string{
		"self": {
			"href": "http://www.example.com/webhooks/607982b8-b0f1-4762-8bd1-ac230310ef8e",
		},
		"project": {
			"href": "http://www.example.com/projects/07597891-3495-46af-b702-babfaacec89a",
		},
		"creator": {
			"href": "http://www.example.com/users/a175d3c4-8bc9-42f9-afd1-adeaf9d2e306",
		},
		"job": {
			"href": "http://www.example.com/jobs/6d17d7f0-06aa-47e4-b5a3-6c20b4005f40",
		},
	}

	actual := webhook.Links(map[string]map[string]string{}, "http", "www.example.com")

	for resource, links := range expected {
		actualResource, found := actual[resource]
		if !found {
			t.Errorf("Expected to link to %q", resource)
			continue
		}

		if actualResource["href"] != links["href"] {
			t.Errorf("Expected %q, got %q", links["href"], actualResource["href"])
		}
	}

}

func Test_Webhook_AuthorizationNameIsWebhook(t *testing.T) {
	webhook := &Webhook{}
	authname := webhook.AuthorizationName()
	expected := "webhook"
	if authname != expected {
		t.Fatalf("Expected %q, got %q", expected, authname)
	}
}

func Test_Webhook_BelongsToProject(t *testing.T) {
	project := &Project{
		Uuid: "43dd7e4a-7958-488f-900d-9a23de865727",
	}

	webhook := &Webhook{
		ProjectUuid: project.Uuid,
	}

	store := &mockProjectStore{
		byId: map[string]*Project{
			project.Uuid: project,
		},
	}

	found, _ := webhook.FindProject(store)
	if found.Uuid != webhook.ProjectUuid {
		t.Fatalf("Expected ProjectUuid to be %q, got %q", webhook.ProjectUuid, found.Uuid)
	}
}

func Test_Webhook_Validate_requiresSlugToBeSet(t *testing.T) {
	project := &Project{
		Uuid: "43dd7e4a-7958-488f-900d-9a23de865727",
	}

	webhook := &Webhook{
		ProjectUuid: project.Uuid,
	}

	err := webhook.Validate()
	verr := err.(*ValidationError)

	expected := []string{"empty"}
	actual := verr.Errors["slug"]

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Expected %#v to be %#v", actual, expected)
	}

}

func Test_Webhook_Validate_requiresNameToBeSet(t *testing.T) {
	project := &Project{
		Uuid: "43dd7e4a-7958-488f-900d-9a23de865727",
	}

	webhook := &Webhook{
		ProjectUuid: project.Uuid,
	}

	err := webhook.Validate()
	verr := err.(*ValidationError)

	expected := []string{"empty"}
	actual := verr.Errors["name"]

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Expected %#v to be %#v", actual, expected)
	}

}

func Test_Webhook_Validate_requiresJobUuidToBeSet(t *testing.T) {
	project := &Project{
		Uuid: "43dd7e4a-7958-488f-900d-9a23de865727",
	}

	webhook := &Webhook{
		ProjectUuid: project.Uuid,
	}

	err := webhook.Validate()
	verr := err.(*ValidationError)

	expected := []string{"empty"}
	actual := verr.Errors["jobUuid"]

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Expected %#v to be %#v", actual, expected)
	}

}
