// Package stores provides storage objects, they all have a trivial interface
// for finding by UUID, updateing and deleting
package stores_test

import (
	"fmt"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"

	"testing"
	"time"

	helpers "github.com/harrowio/harrow/test_helpers"
	"github.com/jmoiron/sqlx"
)

func setupSessionStoreTest(t *testing.T) *sqlx.Tx {
	tx := helpers.GetDbTx(t)
	helpers.MustCreateUser(t, tx, &domain.User{
		Uuid:     "11111111-1111-4111-a111-111111111111",
		Name:     "Anne Mustermann",
		Email:    "anne@musterma.nn",
		Password: "changeme123",
	})
	return tx
}

func TestSessionStore_FindByUuid_InvalidatesLoggedOutSessions(t *testing.T) {

	tx := setupSessionStoreTest(t)
	defer tx.Rollback()

	sessionStore := stores.NewDbSessionStore(tx)
	s := &domain.Session{
		UserUuid:      "11111111-1111-4111-a111-111111111111",
		ValidatedAt:   time.Time{},
		UserAgent:     "TestCase",
		ClientAddress: "::1",
	}
	uuid, err := sessionStore.Create(s)
	if err != nil {
		t.Fatal(err)
	}
	s.LogOut()
	sessionStore.MarkAsLoggedOut(s)

	found, err := sessionStore.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}
	if found.Valid {
		t.Fatal("Expected logged out session to be invalid.")
	}

}

func Test_SessionStore_SuccessfullyCreatingANewSession(t *testing.T) {

	tx := setupSessionStoreTest(t)
	defer tx.Rollback()

	sessionStore := stores.NewDbSessionStore(tx)

	s := &domain.Session{
		UserUuid:      "11111111-1111-4111-a111-111111111111",
		ValidatedAt:   time.Time{},
		UserAgent:     "TestCase",
		ClientAddress: "::1",
	}

	_, err := sessionStore.Create(s)
	if err != nil {
		t.Fatal(err)
	}

}

func Test_SessionStore_FailingToCreateSessionForNonExistentUser(t *testing.T) {

	tx := setupSessionStoreTest(t)
	defer tx.Rollback()

	sessionStore := stores.NewDbSessionStore(tx)

	s := &domain.Session{
		UserUuid:      "22222222-2222-4222-a222-222222222222",
		ValidatedAt:   time.Time{},
		UserAgent:     "TestCase",
		ClientAddress: "::1",
	}

	_, err := sessionStore.Create(s)
	if ve, ok := err.(*domain.ValidationError); ok {
		if ve.Errors["user_uuid"][0] != "foreign_key_violation" {
			t.Fatalf("Expected ve.Errors[\"user_uuid\"] != \"foreign_key_violation\" got ve.Errors[\"user_uuid\"] == %v", ve.Errors["user_uuid"])
		}
	} else {
		t.Fatalf("Expected ve.(type) == domain.ValidationError, got %v", err)
	}
}

func Test_SessionStore_CanFindASessionByUuid(t *testing.T) {

	tx := setupSessionStoreTest(t)
	defer tx.Rollback()

	sessionStore := stores.NewDbSessionStore(tx)

	s := helpers.MustCreateSession(t, tx, &domain.Session{
		UserUuid:      "11111111-1111-4111-a111-111111111111",
		ValidatedAt:   time.Time{},
		UserAgent:     "TestCase",
		ClientAddress: "::1",
	})

	s, err := sessionStore.FindByUuid(s.Uuid)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_SessionStore_FindByUuid_sets_loaded_at_to_the_current_time(t *testing.T) {

	tx := setupSessionStoreTest(t)
	defer tx.Rollback()

	sessionStore := stores.NewDbSessionStore(tx)

	s := helpers.MustCreateSession(t, tx, &domain.Session{
		UserUuid:      "11111111-1111-4111-a111-111111111111",
		ValidatedAt:   time.Time{},
		UserAgent:     "TestCase",
		ClientAddress: "::1",
	})

	s, err := sessionStore.FindByUuid(s.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("now", time.Now())
	fmt.Println("loadedAd", s.LoadedAt)
	loadedAtDelta := time.Now().Sub(s.LoadedAt)
	if got, want := loadedAtDelta, 5*time.Second; got >= want {
		t.Errorf(`loadedAtDelta = %v; want < %v`, got, want)
	}
}

func Test_SessionStore_FindAllByUserUuid_sets_loaded_at_to_the_current_time(t *testing.T) {

	tx := setupSessionStoreTest(t)
	defer tx.Rollback()

	sessionStore := stores.NewDbSessionStore(tx)

	s := helpers.MustCreateSession(t, tx, &domain.Session{
		UserUuid:      "11111111-1111-4111-a111-111111111111",
		ValidatedAt:   time.Time{},
		UserAgent:     "TestCase",
		ClientAddress: "::1",
	})

	userSessions, err := sessionStore.FindAllByUserUuid(s.UserUuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(userSessions), 1; got != want {
		t.Fatalf(`len(userSessions) = %v; want %v`, got, want)
	}

	loadedAtDelta := time.Now().Sub(userSessions[0].LoadedAt)
	if got, want := loadedAtDelta, 5*time.Second; got >= want {
		t.Errorf(`loadedAtDelta = %v; want < %v`, got, want)
	}
}

func Test_SessionStore_SucessfullyDeleteSession(t *testing.T) {

	tx := setupSessionStoreTest(t)
	defer tx.Rollback()

	sessionStore := stores.NewDbSessionStore(tx)

	s := &domain.Session{
		UserUuid:      "11111111-1111-4111-a111-111111111111",
		ValidatedAt:   time.Time{},
		UserAgent:     "TestCase",
		ClientAddress: "::1",
	}

	uuid, err := sessionStore.Create(s)
	if err != nil {
		t.Fatal(err)
	}

	err = sessionStore.DeleteByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}

	_, err = sessionStore.FindByUuid(uuid)
	if err == nil {
		t.Fatal("Expected to get a *domain.NotFoundError, got nil")
	}

}

func Test_SessionStore_FailToDeleteNonExistentSession(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	sessionStore := stores.NewDbSessionStore(tx)

	err := sessionStore.DeleteByUuid("11111111-1111-4111-a111-111111111111")
	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatal("Expected to get a *domain.NotFoundError, got:", err)
	}

}

func Test_SessionStore_ValidAtShouldMatchCreatedAtWhenUserHasNoTOTPPassword(t *testing.T) {

	tx := setupSessionStoreTest(t)
	defer tx.Rollback()

	sessionStore := stores.NewDbSessionStore(tx)

	s := &domain.Session{
		UserUuid:      "11111111-1111-4111-a111-111111111111",
		ValidatedAt:   time.Time{},
		UserAgent:     "TestCase",
		ClientAddress: "::1",
	}

	uuid, err := sessionStore.Create(s)
	if err != nil {
		t.Fatal(err)
	}

	s, err = sessionStore.FindByUuid(uuid)
	if s == nil {
		t.Fatal("FindByUuid Failed", err)
	}

	if s.ValidatedAt.IsZero() {
		t.Fatal("ValidatedAt should not be a zero value")
	}

}

func Test_SessionStore_ValidatedAtShouldBeEmptyWhenUserHasATOTPPassword(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	helpers.MustCreateUser(t, tx, &domain.User{
		Uuid:       "22222222-2222-4222-a222-222222222222",
		Name:       "Anne Mustermann",
		TotpSecret: "anything",
		Email:      "anne@musterma.nn",
		Password:   "changeme123",
	})

	sessionStore := stores.NewDbSessionStore(tx)

	s := &domain.Session{
		UserUuid:      "22222222-2222-4222-a222-222222222222",
		UserAgent:     "TestCase",
		ClientAddress: "::1",
	}

	uuid, err := sessionStore.Create(s)
	if err != nil {
		t.Fatal(err)
	}

	s, err = sessionStore.FindByUuid(uuid)
	if s == nil {
		t.Fatal("FindByUuid Failed", err)
	}

	epoch, err := time.Parse("2006-01-02 15:04:05", "1970-01-01 00:00:00")
	if err != nil {
		panic(err)
	}

	if s.ValidatedAt.IsZero() {
		t.Fatal("ValidatedAt should be ", epoch, " got:", s.ValidatedAt)
	}

}

func Test_SessionStore_FindByUuid_returnsValidSession_ifUserHasTotpSecret_butNoTotpEnabledAt(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	world := helpers.MustNewWorld(tx, t)
	user := world.User("default")

	user.DisableTotp(user.CurrentTotpToken())
	user.GenerateTotpSecret()
	if err := stores.NewDbUserStore(tx, config.GetConfig()).Update(user); err != nil {
		t.Fatal(err)
	}

	session := user.NewSession("Go Test Client", "127.0.0.1")
	sessionStore := stores.NewDbSessionStore(tx)
	uuid, err := sessionStore.Create(session)

	if err != nil {
		t.Fatal(err)
	}

	foundSession, err := sessionStore.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := foundSession.Valid, true; got != want {
		t.Errorf("foundSession.Valid = %v; want %v", got, want)
	}
}
