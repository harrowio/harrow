package domain

import "testing"

func TestOrganization_NewBillingEvent_setsOrganizationUuidOnEvent(t *testing.T) {
	organization := &Organization{
		Uuid: "6be49a96-cc66-4083-a284-f836befed34a",
	}

	event := organization.NewBillingEvent(&TestBillingEvent{
		TestData: "foo",
	})

	if got, want := event.OrganizationUuid, organization.Uuid; got != want {
		t.Errorf("event.OrganizationUuid = %q; want %q", got, want)
	}

}
