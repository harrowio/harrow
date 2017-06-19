package domain

import (
	"bytes"
	"net/http"
	"testing"
)

func Test_Delivery_OwnUrl(t *testing.T) {
	delivery := &Delivery{
		Uuid: "607982b8-b0f1-4762-8bd1-ac230310ef8e",
	}

	actual := delivery.OwnUrl("http", "www.example.com")
	expected := "http://www.example.com/deliveries/607982b8-b0f1-4762-8bd1-ac230310ef8e"
	if actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}

func Test_Delivery_Links(t *testing.T) {
	scheduleUuid := "13fdf67b-b672-4713-914f-d02b066e3173"
	delivery := &Delivery{
		Uuid:         "607982b8-b0f1-4762-8bd1-ac230310ef8e",
		WebhookUuid:  "54d4fc6c-a5e3-4fd0-b707-7a295c3da6d7",
		ScheduleUuid: &scheduleUuid,
	}

	expected := map[string]map[string]string{
		"self": map[string]string{
			"href": "http://www.example.com/deliveries/607982b8-b0f1-4762-8bd1-ac230310ef8e",
		},
		"webhook": map[string]string{
			"href": "http://www.example.com/webhooks/54d4fc6c-a5e3-4fd0-b707-7a295c3da6d7",
		},
		"schedule": map[string]string{
			"href": "http://www.example.com/schedules/13fdf67b-b672-4713-914f-d02b066e3173",
		},
	}

	actual := delivery.Links(map[string]map[string]string{}, "http", "www.example.com")

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

func Test_Delivery_AuthorizationNameIsDelivery(t *testing.T) {
	delivery := &Delivery{}
	authname := delivery.AuthorizationName()
	expected := "delivery"
	if authname != expected {
		t.Fatalf("Expected %q, got %q", expected, authname)
	}
}

func Test_Delivery_BelongsToProject(t *testing.T) {
	project := &Project{
		Uuid: "43dd7e4a-7958-488f-900d-9a23de865727",
	}

	webhook := &Webhook{
		Uuid:        "162be101-8ed8-4952-ac4e-93e80b1597fb",
		ProjectUuid: project.Uuid,
	}

	delivery := webhook.NewDelivery(nil)

	store := &mockProjectStore{
		byWebhookUuid: map[string]*Project{
			webhook.Uuid: project,
		},
	}

	found, _ := delivery.FindProject(store)
	if found.Uuid != webhook.ProjectUuid {
		t.Fatalf("Expected ProjectUuid to be %q, got %q", webhook.ProjectUuid, found.Uuid)
	}
}

func TestDelivery_GitRef_extractsRefFromGitHubWebhookFormat(t *testing.T) {
	src := bytes.NewBufferString(`{"ref":"feature-branch"}`)
	req, err := http.NewRequest("POST", "http://www.example.com/wh/slug", src)
	if err != nil {
		t.Fatal(err)
	}

	payload := DeliveredRequest{Request: req}

	delivery := &Delivery{
		Uuid:    "607982b8-b0f1-4762-8bd1-ac230310ef8e",
		Request: payload,
	}

	if got, want := delivery.GitRef(), "feature-branch"; got != want {
		t.Errorf(`delivery.GitRef() = %v; want %v`, got, want)
	}
}

func TestDelivery_RepositoryName_extractsRepoNameFromGitHubWebhookFormat(t *testing.T) {
	src := bytes.NewBufferString(`{"ref":"feature-branch", "repository":{"full_name": "repository/name"}}`)
	req, err := http.NewRequest("POST", "http://www.example.com/wh/slug", src)
	if err != nil {
		t.Fatal(err)
	}

	payload := DeliveredRequest{Request: req}

	delivery := &Delivery{
		Uuid:    "607982b8-b0f1-4762-8bd1-ac230310ef8e",
		Request: payload,
	}

	if got, want := delivery.RepositoryName(), "repository/name"; got != want {
		t.Errorf(`delivery.RepositoryName() = %v; want %v`, got, want)
	}
}
func TestDelivery_RepositoryName_extractsRepoNameFromBitBucketWebhookFormat(t *testing.T) {
	src := bytes.NewBufferString(`{"repository":{"full_name": "repository/name"}, "push": {"changes": [{"new": {"name": "feature-branch"}}]}}`)
	req, err := http.NewRequest("POST", "http://www.example.com/wh/slug", src)
	if err != nil {
		t.Fatal(err)
	}

	payload := DeliveredRequest{Request: req}

	delivery := &Delivery{
		Uuid:    "607982b8-b0f1-4762-8bd1-ac230310ef8e",
		Request: payload,
	}

	if got, want := delivery.RepositoryName(), "repository/name"; got != want {
		t.Errorf(`delivery.RepositoryName() = %v; want %v`, got, want)
	}
}

func TestDelivery_GitRef_returnsFirstNewRef_fromBitBucketWebhookFormat(t *testing.T) {
	src := bytes.NewBufferString(`{"push": {"changes": [{"new": {"name": "feature-branch"}}]}}`)
	req, err := http.NewRequest("POST", "http://www.example.com/wh/slug", src)
	if err != nil {
		t.Fatal(err)
	}

	payload := DeliveredRequest{Request: req}

	delivery := &Delivery{
		Uuid:    "607982b8-b0f1-4762-8bd1-ac230310ef8e",
		Request: payload,
	}

	if got, want := delivery.GitRef(), "feature-branch"; got != want {
		t.Errorf(`delivery.GitRef() = %v; want %v`, got, want)
	}
}

func TestDelivery_GitRef_doesNotFailWhenCalledMultipleTimes(t *testing.T) {
	src := bytes.NewBufferString(`{"ref":"feature-branch"}`)
	req, err := http.NewRequest("POST", "http://www.example.com/wh/slug", src)
	if err != nil {
		t.Fatal(err)
	}

	payload := DeliveredRequest{Request: req}

	delivery := &Delivery{
		Uuid:    "607982b8-b0f1-4762-8bd1-ac230310ef8e",
		Request: payload,
	}

	delivery.GitRef() // drain delivered request body

	if got, want := delivery.GitRef(), "feature-branch"; got != want {
		t.Errorf(`delivery.GitRef() = %v; want %v`, got, want)
	}
}

type repositoriesByName []*Repository

func (self repositoriesByName) FindAllByProjectUuidAndRepositoryName(projectUuid, repositoryName string) ([]*Repository, error) {
	return []*Repository(self), nil
}

func TestDelivery_OperationParameters_returnsRepositoriesToCheckOut(t *testing.T) {
	src := bytes.NewBufferString(`{"ref":"feature-branch", "repository":{"full_name": "repository/name"}}`)
	req, err := http.NewRequest("POST", "http://www.example.com/wh/slug", src)
	if err != nil {
		t.Fatal(err)
	}

	payload := DeliveredRequest{Request: req}

	delivery := &Delivery{
		Uuid:    "607982b8-b0f1-4762-8bd1-ac230310ef8e",
		Request: payload,
	}

	repository := &Repository{
		Uuid: "23a1d67d-6d17-403e-9092-479805c9cbc8",
		Name: "the repository",
	}

	projectUuid := "492dddc0-0212-4f49-bae8-65f57f133902"
	params := delivery.OperationParameters(projectUuid, repositoriesByName{repository})

	if got, want := len(params.Checkout), 1; got != want {
		t.Fatalf(`len(params.Checkout) = %v; want %v`, got, want)
	}

	if got, want := params.Checkout[repository.Uuid], "feature-branch"; got != want {
		t.Errorf(`params.Checkout[repository.Uuid] = %v; want %v`, got, want)
	}
}

func TestDelivery_OperationParameters_returnsEmptyParametersIfRepositoryDoesNotMatch(t *testing.T) {
	src := bytes.NewBufferString(`{"ref":"feature-branch", "repository":{"full_name": "repository/name"}}`)
	req, err := http.NewRequest("POST", "http://www.example.com/wh/slug", src)
	if err != nil {
		t.Fatal(err)
	}

	payload := DeliveredRequest{Request: req}

	delivery := &Delivery{
		Uuid:    "607982b8-b0f1-4762-8bd1-ac230310ef8e",
		Request: payload,
	}

	projectUuid := "492dddc0-0212-4f49-bae8-65f57f133902"
	params := delivery.OperationParameters(projectUuid, repositoriesByName{})

	if got, want := len(params.Checkout), 0; got != want {
		t.Fatalf(`len(params.Checkout) = %v; want %v`, got, want)
	}
}

func TestDelivery_OperationParameters_associatesThisDeliveryWithOperation(t *testing.T) {
	src := bytes.NewBufferString(`{"ref":"feature-branch", "repository":{"full_name": "repository/name"}}`)
	req, err := http.NewRequest("POST", "http://www.example.com/wh/slug", src)
	if err != nil {
		t.Fatal(err)
	}

	payload := DeliveredRequest{Request: req}

	delivery := &Delivery{
		Uuid:    "607982b8-b0f1-4762-8bd1-ac230310ef8e",
		Request: payload,
	}

	projectUuid := "492dddc0-0212-4f49-bae8-65f57f133902"
	params := delivery.OperationParameters(projectUuid, repositoriesByName{})
	if got, want := params.TriggeredByDelivery, delivery.Uuid; got != want {
		t.Errorf(`params.TriggeredByDelivery = %v; want %v`, got, want)
	}
}

func TestDelivery_OperationParameters_setsTriggerReasonToWebhook(t *testing.T) {
	src := bytes.NewBufferString(`{"ref":"feature-branch", "repository":{"full_name": "repository/name"}}`)
	req, err := http.NewRequest("POST", "http://www.example.com/wh/slug", src)
	if err != nil {
		t.Fatal(err)
	}

	payload := DeliveredRequest{Request: req}

	delivery := &Delivery{
		Uuid:    "607982b8-b0f1-4762-8bd1-ac230310ef8e",
		Request: payload,
	}

	projectUuid := "492dddc0-0212-4f49-bae8-65f57f133902"
	params := delivery.OperationParameters(projectUuid, repositoriesByName{})
	if got, want := params.Reason, OperationTriggeredByWebhook; got != want {
		t.Errorf(`params.Reason = %v; want %v`, got, want)
	}

}
