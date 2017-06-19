package stores_test

import (
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
)

type TestEvent struct {
	TestData string
}

func (self *TestEvent) BillingEventName() string { return "test-event" }

func init() {
	domain.RegisterBillingEvent(&TestEvent{})
}

func TestDbBillingEventStore_Create(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbBillingEventStore(tx)
	organizationUuid := "1245440f-7e17-4365-9e3f-705a6505fa9d"
	event := &domain.BillingEvent{
		EventName:        "test-event",
		OrganizationUuid: organizationUuid,
		Data:             &TestEvent{TestData: "test-data"},
	}

	if _, err := store.Create(event); err != nil {
		t.Fatal(err)
	}

	stored, err := store.FindAllByOrganizationUuid(organizationUuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(stored), 1; got != want {
		t.Errorf("len(stored) = %d; want %d", got, want)
	}

	data, ok := stored[0].Data.(*TestEvent)
	if !ok {
		t.Fatalf("stored[0].Data.(type) = %T; want %T", stored[0].Data, data)
	}

	if got, want := data.TestData, "test-data"; got != want {
		t.Errorf("data.TestData = %q; want %q", got, want)
	}
}
