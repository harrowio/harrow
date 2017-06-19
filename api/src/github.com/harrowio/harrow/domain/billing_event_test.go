package domain

import (
	"encoding/json"
	"fmt"
	"testing"
)

type TestBillingEvent struct {
	TestData string
}

func (self *TestBillingEvent) BillingEventName() string { return "test-billing-event" }

func TestBillingEvent_UnmarshalJSON_deserializes_data_into_the_type_mapped_to_the_event_name(t *testing.T) {
	organizationUuid := "b4c1fe95-848a-4216-a373-9374f66c373e"
	data := fmt.Sprintf(`{"uuid":"9b214984-3751-4ecd-8c5a-3dcb77e90c6d","organizationUuid": %q, "eventName": "extra-limits-granted", "data": {"projects":10}}`, organizationUuid)
	result := new(BillingEvent)
	if err := json.Unmarshal([]byte(data), result); err != nil {
		t.Fatal(err)
	}

	extraLimitsGranted, ok := result.Data.(*BillingExtraLimitsGranted)
	if !ok {
		t.Fatalf("result.Data.(type) = %T; want %T", result.Data, extraLimitsGranted)
	}

	if got, want := extraLimitsGranted.Projects, 10; got != want {
		t.Errorf(`extraLimitsGranted.Projects = %v; want %v`, got, want)
	}
}
