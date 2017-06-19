package domain

import (
	"reflect"
	"strings"
	"testing"
)

type mockSubscriptionStore struct {
	subscriptions map[string]*Subscription
	uuidToId      map[string]string
}

func newMockSubscriptionStore() *mockSubscriptionStore {
	return &mockSubscriptionStore{
		subscriptions: map[string]*Subscription{},
		uuidToId:      map[string]string{},
	}
}

func (store *mockSubscriptionStore) Find(watchableUuid, event, userUuid string) (*Subscription, error) {
	id := watchableUuid + ":" + userUuid + ":" + event
	subscription := store.subscriptions[id]
	return subscription, nil
}

func (store *mockSubscriptionStore) FindEventsForUser(watchableId, userUuid string) ([]string, error) {
	prefix := watchableId + ":" + userUuid + ":"
	result := []string{}
	for key, _ := range store.subscriptions {
		if strings.HasPrefix(key, prefix) {
			result = append(result, key[len(prefix):])
		}
	}

	return result, nil
}

func (store *mockSubscriptionStore) Create(subscription *Subscription) (string, error) {
	id := subscription.WatchableUuid + ":" + subscription.UserUuid + ":" + subscription.EventName
	store.subscriptions[id] = subscription
	store.uuidToId[subscription.Uuid] = id
	return id, nil
}

func (store *mockSubscriptionStore) Delete(subscriptionUuid string) error {
	id := store.uuidToId[subscriptionUuid]
	if id != "" {
		delete(store.subscriptions, id)
		delete(store.uuidToId, subscriptionUuid)
	}
	return nil
}

func Test_User_SubscribeTo_SubscribesUserToEvent(t *testing.T) {
	user := &User{
		Uuid: "2560599d-1e46-4869-b2a8-09bb18cb9bc9",
	}

	watchable := &Job{
		Uuid: "1a478524-e7d9-4072-9cc0-27c3ca89e93e",
	}

	subscriptions := newMockSubscriptionStore()

	user.SubscribeTo(watchable, EventOperationStarted, subscriptions)

	if ok, err := user.IsSubscribedTo(watchable, EventOperationStarted, subscriptions); !ok {
		t.Fatalf("Expected to be subscribed to %q", EventOperationStarted)
	} else if err != nil {
		t.Fatal(err)
	}
}

func Test_User_SubscribeTo_DoesNotSuscribeUserToOtherEvents(t *testing.T) {
	user := &User{
		Uuid: "2560599d-1e46-4869-b2a8-09bb18cb9bc9",
	}

	watchable := &Job{
		Uuid: "1a478524-e7d9-4072-9cc0-27c3ca89e93e",
	}

	subscriptions := newMockSubscriptionStore()

	user.SubscribeTo(watchable, EventOperationStarted, subscriptions)

	if ok, err := user.IsSubscribedTo(watchable, EventOperationSucceeded, subscriptions); ok {
		t.Fatalf("Expected not to be subscribed to %q", EventOperationSucceeded)
	} else if err != nil {
		t.Fatal(err)
	}
}

func Test_User_UnsubscribeFrom_UnsubscribesUserFromEvent(t *testing.T) {
	user := &User{
		Uuid: "2560599d-1e46-4869-b2a8-09bb18cb9bc9",
	}

	watchable := &Job{
		Uuid: "1a478524-e7d9-4072-9cc0-27c3ca89e93e",
	}

	subscriptions := newMockSubscriptionStore()

	user.SubscribeTo(watchable, EventOperationStarted, subscriptions)
	user.UnsubscribeFrom(watchable, EventOperationStarted, subscriptions)

	if ok, _ := user.IsSubscribedTo(watchable, EventOperationStarted, subscriptions); ok {
		t.Fatalf("Expected not to be subscribed to %q", EventOperationStarted)
	}
}

func Test_User_Watch_subscribesUserToAllEvents(t *testing.T) {
	user := &User{
		Uuid: "2560599d-1e46-4869-b2a8-09bb18cb9bc9",
	}

	watchable := &Job{
		Uuid: "1a478524-e7d9-4072-9cc0-27c3ca89e93e",
	}

	subscriptions := newMockSubscriptionStore()

	if err := user.Watch(watchable, subscriptions); err != nil {
		t.Fatal(err)
	}

	for _, event := range []string{"operations.failed", "operations.scheduled", "operations.started", "operations.succeeded", "operations.timed_out"} {
		if ok, _ := user.IsSubscribedTo(watchable, event, subscriptions); !ok {
			t.Errorf("Expected to be subscribed to %q", event)
		}
	}
}

func Test_User_Watch_failsIfUserIsAlreadyWatching(t *testing.T) {
	user := &User{
		Uuid: "2560599d-1e46-4869-b2a8-09bb18cb9bc9",
	}

	watchable := &Job{
		Uuid: "1a478524-e7d9-4072-9cc0-27c3ca89e93e",
	}

	subscriptions := newMockSubscriptionStore()

	if err := user.Watch(watchable, subscriptions); err != nil {
		t.Fatal(err)
	}

	err := user.Watch(watchable, subscriptions)
	if err == nil {
		t.Fatal("Expected an error")
	}

	if _, ok := err.(*ValidationError); !ok {
		t.Fatalf("Expected *ValidationError, got %#v", err)
	}
}

func Test_User_Unwatch_unsubscribesUserFromAllEvents(t *testing.T) {
	user := &User{
		Uuid: "2560599d-1e46-4869-b2a8-09bb18cb9bc9",
	}

	watchable := &Job{
		Uuid: "1a478524-e7d9-4072-9cc0-27c3ca89e93e",
	}

	subscriptions := newMockSubscriptionStore()

	if err := user.Watch(watchable, subscriptions); err != nil {
		t.Fatal(err)
	}
	if err := user.Unwatch(watchable, subscriptions); err != nil {
		t.Fatal(err)
	}

	for _, event := range []string{"operations.failed", "operations.scheduled", "operations.started", "operations.succeeded", "operations.timed_out"} {
		if ok, _ := user.IsSubscribedTo(watchable, event, subscriptions); ok {
			t.Errorf("Expected not to be subscribed to %q", event)
		}
	}
}

func Test_User_SubscriptionsFor_returnsSubscribedEvents(t *testing.T) {
	user := &User{
		Uuid: "2560599d-1e46-4869-b2a8-09bb18cb9bc9",
	}

	watchable := &Job{
		Uuid: "1a478524-e7d9-4072-9cc0-27c3ca89e93e",
	}

	subscriptions := newMockSubscriptionStore()

	if err := user.Watch(watchable, subscriptions); err != nil {
		t.Fatal(err)
	}
	if err := user.UnsubscribeFrom(watchable, EventOperationStarted, subscriptions); err != nil {
		t.Fatal(err)
	}

	userSubscriptions, err := user.SubscriptionsFor(watchable, subscriptions)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]bool{
		EventOperationStarted:   false,
		EventOperationFailed:    true,
		EventOperationSucceeded: true,
		EventOperationTimedOut:  true,
		EventOperationScheduled: true,
	}

	actual := userSubscriptions.Subscribed

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Expected userSubscriptions.Subscribed to be %#v, got %#v\n", expected, actual)
	}
}

func Test_User_FindProject_returnsNotFoundErrorForUserWithoutId(t *testing.T) {
	user := &User{}
	_, err := user.FindProject(&mockProjectStore{})
	if err == nil {
		t.Fatalf("Expected error")
	}

	if _, ok := err.(*NotFoundError); !ok {
		t.Fatalf("Expected error to be of type %T, got %T", &NotFoundError{}, err)
	}
}

func Test_User_FindUser_doesNotAccessTheProvidedStore(t *testing.T) {
	user := &User{}
	found, err := user.FindUser(nil)
	if err != nil {
		t.Fatal(err)
	}

	if found != user {
		t.Fatalf("Expected %#v to be the same object as %#v", found, user)
	}
}

func Test_User_EnableTotp_RequiresCurrentTotpToken(t *testing.T) {
	user := &User{}
	user.GenerateTotpSecret()
	user.EnableTotp(user.CurrentTotpToken())

	if got, want := user.TotpEnabled(), true; got != want {
		t.Fatalf("User(%q) = %v; want %v", "TotpEnabled", got, want)
	}
}

func Test_User_EnableTotp_FailsIfInvalidTokenIsProvided(t *testing.T) {
	user := &User{}
	user.GenerateTotpSecret()
	err := user.EnableTotp(0)

	if err != ErrTotpTokenNotValid {
		t.Errorf("user.EnableTotp(0) = %q; want %q", err, ErrTotpTokenNotValid)
	}

	if got, want := user.TotpEnabled(), false; got != want {
		t.Fatalf("User(%q) = %v; want %v", "TotpEnabled", got, want)
	}

}

func Test_User_DisableTotp_RequiresCurrentToken(t *testing.T) {
	user := &User{}

	user.GenerateTotpSecret()
	user.EnableTotp(user.CurrentTotpToken())
	err := user.DisableTotp(0)

	if got, want := err, ErrTotpTokenNotValid; got != want {
		t.Fatalf("User(%q) = %v; want %v", "DisableTotp", got, want)
	}
}

func Test_User_DisableTotp_TotpNotEnabledAnymore(t *testing.T) {
	user := &User{}

	user.GenerateTotpSecret()
	user.EnableTotp(user.CurrentTotpToken())
	user.DisableTotp(user.CurrentTotpToken())

	if got, want := user.TotpEnabled(), false; got != want {
		t.Fatalf("User(%q) = %v; want %v", "TotpEnabled", got, want)
	}
}

func Test_User_NewBlock_setsUserUuid(t *testing.T) {
	user := &User{Uuid: "ee7db5a0-73b3-4a84-908b-f2474f468c9e"}
	block, err := user.NewBlock("for testing")
	if err != nil {
		t.Fatal(err)
	}

	if got, want := block.UserUuid, user.Uuid; got != want {
		t.Errorf("block.UserUuid = %q; want %q", got, want)
	}
}

func Test_User_NewBlock_tracksTheReasonForTheBlock(t *testing.T) {
	user := &User{Uuid: "ee7db5a0-73b3-4a84-908b-f2474f468c9e"}
	block, err := user.NewBlock("for testing")
	if err != nil {
		t.Fatal(err)
	}

	if got, want := block.Reason, "for testing"; got != want {
		t.Errorf("block.Reason = %q; want %q", got, want)
	}
}

func Test_User_NewBlock_doesNotAllowAnEmptyReason(t *testing.T) {
	user := &User{Uuid: "ee7db5a0-73b3-4a84-908b-f2474f468c9e"}
	_, err := user.NewBlock("   ")
	if err == nil {
		t.Fatalf("expected an error")
	}

	verr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("err.(type) = %T; want %T", err, verr)
	}

	if got, want := verr.Get("reason"), "empty"; got != want {
		t.Fatalf("verr.Get(%q) = %q; want %q", "reason", got, want)
	}
}

func Test_User_Scrub_nulls_out_the_users_password(t *testing.T) {
	user := &User{
		Uuid:     "4948aa51-584b-45a7-a3ee-d8c949fa4495",
		Password: "password",
	}

	scrubbed := user.Scrub()

	if got, want := scrubbed.Password, ""; got != want {
		t.Errorf(`scrubbed.Password = %v; want %v`, got, want)
	}
}
