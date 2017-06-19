package domain

import (
	"testing"
	"time"
)

func Test_Session_NotValidWithoutAValidatedAtTimestamp(t *testing.T) {
	s := &Session{}
	err := s.Validate()
	verr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("Expected a *ValidationError, got: %T", err)
	}

	if got, want := verr.Get("validatedAt"), "never_validated"; got != want {
		t.Errorf("verr.Get(%q) = %q; want %q", "validatedAt", got, want)
	}
}

func Test_Session_NotValidIfValidatedAtBeforeCreatedAt(t *testing.T) {
	noughties, err := time.Parse(time.RFC3339, "1999-01-01T00:00:00+01:00")
	if err != nil {
		t.Fatal(err)
	}
	epoch, err := time.Parse(time.RFC3339, "1970-01-01T00:00:00+01:00")
	if err != nil {
		t.Fatal(err)
	}
	s := &Session{CreatedAt: noughties, ValidatedAt: epoch}
	err = s.Validate()
	verr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("Expected a *ValidationError, got: %T", err)
	}

	if got, want := verr.Get("validatedAt"), "never_validated"; got != want {
		t.Errorf("verr.Get(%q) = %q; want %q", "validatedAt", got, want)
	}
}

func Test_Session_Validate_notValidAnymoreIfLoggedOut(t *testing.T) {
	now := time.Date(2015, 5, 29, 15, 32, 0, 0, time.UTC)
	session := &Session{
		ValidatedAt: now,
		CreatedAt:   now.Add(-2 * time.Hour),
		LoggedOutAt: &now,
	}

	err := session.Validate()
	verr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("err.(type) = %T; want %T", err, verr)
	}

	if got, want := verr.Get("session"), "logged_out"; got != want {
		t.Errorf("verr.Get(%q) = %q; want %q", "session", got, want)
	}
}

func Test_Session_Validate_notValidAnymoreIfExpired(t *testing.T) {
	now := time.Date(2015, 7, 8, 12, 7, 0, 0, time.UTC)
	session := &Session{
		LoadedAt:  now,
		CreatedAt: now.Add(-48 * time.Hour),
		ExpiresAt: now.Add(-24 * time.Hour),
	}

	err := session.Validate()
	verr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("err.(type) = %T; want %T", err, verr)
	}

	if got, want := verr.Get("session"), "expired"; got != want {
		t.Errorf("verr.Get(%q) = %q; want %q", "session", got, want)
	}
}

func Test_Session_OwnedBy_returnsTrueForMatchingUserUuid(t *testing.T) {
	user := &User{
		Uuid: "881b71e9-7843-4fa8-b3ba-4f9338dfcd52",
	}
	session := &Session{
		Uuid:     "363c0be3-2b11-406a-a7d9-66482afb0b0b",
		UserUuid: user.Uuid,
	}

	if !session.OwnedBy(user) {
		t.Fatal("Expected session to be owned by user")
	}
}

func Test_Session_Validate_updatesValid_toFalse_ifSessionInvalid(t *testing.T) {
	session := &Session{Valid: true}
	err := session.Validate()
	if err == nil {
		t.Fatal("Expected an error")
	}

	if got, want := session.Valid, false; got != want {
		t.Errorf("session.Valid = %v; want %v", got, want)
	}
}

func Test_Session_Validate_updatesValid_toTrue_ifSessionValid(t *testing.T) {
	now := time.Date(2015, 5, 29, 15, 32, 0, 0, time.UTC)
	session := &Session{
		ValidatedAt: now,
		CreatedAt:   now.Add(-2 * time.Hour),
		ExpiresAt:   now.Add(24 * time.Hour),
		Valid:       false,
	}

	err := session.Validate()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := session.Valid, true; got != want {
		t.Errorf("session.Valid = %v; want %v", got, want)
	}
}
