package stores_test

import (
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
	"github.com/jmoiron/sqlx"
)

type mockWatchable struct {
	id     string
	events []string
	kind   string
}

func (m *mockWatchable) Id() string                { return m.id }
func (m *mockWatchable) WatchableEvents() []string { return m.events }
func (m *mockWatchable) WatchableType() string     { return m.kind }

func Test_SubscriptionStore_Create_GeneratesUuidForSubscription(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	store := stores.NewDbSubscriptionStore(tx)
	watchable := &mockWatchable{id: "0aa60a94-29e0-4f99-95e8-57b1398e962c", kind: "mock", events: []string{"foo"}}

	subscription := domain.NewSubscription(watchable, "foo", user.Uuid)
	subscription.Uuid = ""
	uuid, err := store.Create(subscription)
	if err != nil {
		t.Fatal(err)
	}

	if uuid == "" {
		t.Fatal("No uuid returned")
	}

	if uuid != subscription.Uuid {
		t.Fatalf("Expected uuid to be %q, got %q", subscription.Uuid, uuid)
	}
}

func Test_SubscriptionStore_Create_SavesEntryInDatabase(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	store := stores.NewDbSubscriptionStore(tx)
	watchable := &mockWatchable{id: "0aa60a94-29e0-4f99-95e8-57b1398e962c", kind: "mock", events: []string{"foo"}}
	subscription := domain.NewSubscription(watchable, "foo", user.Uuid)

	uuid, err := store.Create(subscription)
	if err != nil {
		t.Fatal(err)
	}

	result := domain.Subscription{}
	if err := sqlx.Get(tx, &result, `SELECT * FROM subscriptions WHERE uuid = $1`, uuid); err != nil {
		t.Fatal(err)
	}
}

func Test_SubscriptionStore_Delete_RemovesSubscriptionFromDatabase(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	store := stores.NewDbSubscriptionStore(tx)
	watchable := &mockWatchable{id: "0aa60a94-29e0-4f99-95e8-57b1398e962c", kind: "mock", events: []string{"foo"}}
	subscription := domain.NewSubscription(watchable, "foo", user.Uuid)

	uuid, err := store.Create(subscription)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.Delete(subscription.Uuid); err != nil {
		t.Fatal(err)
	}

	result := domain.Subscription{}
	if err := sqlx.Get(tx, &result, `SELECT * FROM subscriptions WHERE uuid = $1`, uuid); err == nil {
		t.Fatal("Expected an error.")
	}
}

func Test_SubscriptionStore_Delete_ReturnsNotFoundError(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbSubscriptionStore(tx)

	if err := store.Delete("e7b0185e-781b-48a5-8a8a-e946581b505c"); err == nil {
		t.Fatal("Expected an error")
	} else if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatalf("Expected a *domain.NotFoundError, got %T", err)
	}
}

func Test_SubscriptionStore_Delete_DoesNotDeleteOtherSubscriptions(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	store := stores.NewDbSubscriptionStore(tx)
	watchable := &mockWatchable{id: "0aa60a94-29e0-4f99-95e8-57b1398e962c", kind: "mock", events: []string{"foo", "bar"}}

	subscriptions := []*domain.Subscription{}
	for _, event := range []string{"foo", "bar"} {
		subscription := domain.NewSubscription(watchable, event, user.Uuid)

		_, err := store.Create(subscription)
		if err != nil {
			t.Fatal(err)
		}

		subscriptions = append(subscriptions, subscription)
	}

	if err := store.Delete(subscriptions[0].Uuid); err != nil {
		t.Fatal(err)
	}

	result := domain.Subscription{}
	if err := sqlx.Get(tx, &result, `SELECT * FROM subscriptions WHERE uuid = $1`, subscriptions[1].Uuid); err != nil {
		t.Fatal(err)
	}
}

func Test_SubscriptionStore_Find_ReturnsMatchingSubscription(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	store := stores.NewDbSubscriptionStore(tx)
	watchable := &mockWatchable{id: "0aa60a94-29e0-4f99-95e8-57b1398e962c", kind: "mock", events: []string{"foo", "bar"}}

	subscriptions := []*domain.Subscription{}
	for _, event := range []string{"foo", "bar"} {
		subscription := domain.NewSubscription(watchable, event, user.Uuid)

		_, err := store.Create(subscription)
		if err != nil {
			t.Fatal(err)
		}

		subscriptions = append(subscriptions, subscription)
	}

	result, err := store.Find(watchable.Id(), "foo", user.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if result.Uuid != subscriptions[0].Uuid {
		t.Fatalf("Expected to find %q, got %q", subscriptions[0].Uuid, result.Uuid)
	}
}

func Test_SubscriptionStore_Find_ReturnsNotFoundError_IfNoMatchIsFound(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	store := stores.NewDbSubscriptionStore(tx)
	watchable := &mockWatchable{id: "0aa60a94-29e0-4f99-95e8-57b1398e962c", kind: "mock", events: []string{"foo", "bar"}}

	subscription := domain.NewSubscription(watchable, "foo", user.Uuid)

	_, err := store.Create(subscription)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Find(watchable.Id(), "does-not-exist", user.Uuid)
	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatalf("Expected *domain.NotFoundError, got %T", err)
	}
}
