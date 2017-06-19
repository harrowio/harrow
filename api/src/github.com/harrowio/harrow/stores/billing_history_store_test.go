package stores_test

import (
	"fmt"
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	helpers "github.com/harrowio/harrow/test_helpers"
)

var (
	organizationUuids = []string{
		"2cfb0b9c-d3ab-4f0d-8f10-a488dea32d3f",
		"ca8e651a-0704-487c-b4de-ebaa7fd6d593",
		"cc1724cc-c48d-4241-be8c-d01de0ff4bb7",
	}
	userUuids = []string{
		"df4b5d49-7585-4b5a-996c-effd6105063b",
		"a4c9d964-1968-4c59-bb39-8affc32a3214",
		"487a6d81-f60e-434e-a8f7-35991adc527f",
	}
	planUuid = "40b16493-7b66-41aa-ac60-fbcbe93db435"
)

func generatePlanSelectedEvent(seed int) *domain.BillingEvent {
	return &domain.BillingEvent{
		OrganizationUuid: organizationUuids[seed%len(organizationUuids)],
		EventName:        "plan-selected",
		Data: &domain.BillingPlanSelected{
			UserUuid:       userUuids[seed%len(userUuids)],
			PlanUuid:       planUuid,
			SubscriptionId: fmt.Sprintf("subscription-%d", seed),
		},
	}
}

func TestDbBillingHistoryStore_Load_replaysAllHistory(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	events := []*domain.BillingEvent{
		generatePlanSelectedEvent(1),
		generatePlanSelectedEvent(2),
	}

	billingEvents := stores.NewDbBillingEventStore(tx)
	for _, event := range events {
		if _, err := billingEvents.Create(event); err != nil {
			t.Fatal(err)
		}
	}

	history, err := stores.NewDbBillingHistoryStore(tx, helpers.NewMockKeyValueStore()).Load()
	if err != nil {
		t.Fatal(err)
	}

	if got := history; got == nil {
		t.Fatalf("history is nil")
	}

	if got := history.Subscription("subscription-1"); got == nil {
		t.Fatalf(`history.Subscription("subscription-1") is nil`)
	}

}

func TestDbBillingHistoryStore_CachesUsingAKeyValueStore(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	events := []*domain.BillingEvent{
		generatePlanSelectedEvent(1),
		generatePlanSelectedEvent(2),
	}

	billingEvents := stores.NewDbBillingEventStore(tx)
	for _, event := range events {
		if _, err := billingEvents.Create(event); err != nil {
			t.Fatal(err)
		}
	}

	kvStore := helpers.NewMockKeyValueStore()
	_, err := stores.NewDbBillingHistoryStore(tx, kvStore).Load()
	if err != nil {
		t.Fatal(err)
	}

	history, err := kvStore.Get("billing-history")
	if err != nil {
		t.Fatal(err)
	}

	if got := history; got == nil {
		t.Fatalf("history is nil")
	}
}

func TestDbBillingHistoryStore_returnsCachedVersionWithNewerEventsApplied(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	events := []*domain.BillingEvent{
		generatePlanSelectedEvent(1),
		generatePlanSelectedEvent(2),
	}

	cachedHistory := domain.NewBillingHistory()
	cachedHistory.HandleEvent(generatePlanSelectedEvent(100))
	billingEvents := stores.NewDbBillingEventStore(tx)
	for _, event := range events {
		if _, err := billingEvents.Create(event); err != nil {
			t.Fatal(err)
		}
	}

	kvStore := helpers.NewMockKeyValueStore()
	store := stores.NewDbBillingHistoryStore(tx, kvStore)
	store.CacheHistory(cachedHistory)
	history, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if got := history.Subscription("subscription-100"); got == nil {
		t.Errorf(`history.Subscription("subscription-100") is nil`)
	}

	if got := history.Subscription("subscription-2"); got == nil {
		t.Errorf(`history.Subscription("subscription-2") is nil`)
	}

}
