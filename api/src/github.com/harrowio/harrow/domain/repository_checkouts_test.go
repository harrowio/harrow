package domain

import (
	"reflect"
	"testing"
)

func TestRepositoryCheckouts_on_Checkout_recordsSymbolicRef_andHash(t *testing.T) {
	uuid := "7e511287-0353-4bbc-b59e-4ffb752547bd"
	hash := "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15"
	event := MapPayload{
		"event":      "checkout",
		"ref":        "feature-branch",
		"repository": uuid,
		"hash":       hash,
	}

	subject := NewRepositoryCheckouts()
	subject.HandleEvent(event)

	if got, want := subject.Ref(uuid), "feature-branch"; got != want {
		t.Errorf(`subject.Ref(uuid) = %v; want %v`, got, want)
	}

	if got, want := subject.Hash(uuid), hash; got != want {
		t.Errorf(`subject.Hash(uuid) = %v; want %v`, got, want)
	}

}

func TestRepositoryCheckouts_on_Checkout_recordsNothing_ifRepositoryOrRefOrHashAreEmpty(t *testing.T) {
	uuid := "7e511287-0353-4bbc-b59e-4ffb752547bd"
	hash := "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15"
	events := []MapPayload{
		{
			"event":      "checkout",
			"ref":        "",
			"repository": uuid,
			"hash":       hash,
		},
		{
			"event":      "checkout",
			"ref":        "master",
			"repository": uuid,
			"hash":       "",
		},
		{
			"event":      "checkout",
			"ref":        "master",
			"repository": "",
			"hash":       hash,
		},
	}

	subject := NewRepositoryCheckouts()
	for _, event := range events {
		subject.HandleEvent(event)
	}

	if got, want := subject.Ref(uuid), ""; got != want {
		t.Errorf(`subject.Ref(uuid) = %v; want %v`, got, want)
	}

	if got, want := subject.Hash(uuid), ""; got != want {
		t.Errorf(`subject.Hash(uuid) = %v; want %v`, got, want)
	}

	if got, want := subject.Ref(""), ""; got != want {
		t.Errorf(`subject.Ref("") = %v; want %v`, got, want)
	}

	if got, want := subject.Hash(""), ""; got != want {
		t.Errorf(`subject.Hash("") = %v; want %v`, got, want)
	}
}

func TestRepositoryCheckouts_Ref_returnsRef_ofLatestCheckout(t *testing.T) {
	uuid := "7e511287-0353-4bbc-b59e-4ffb752547bd"
	hash := "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15"
	event := MapPayload{
		"event":      "checkout",
		"ref":        "feature-branch",
		"repository": uuid,
		"hash":       hash,
	}

	subject := NewRepositoryCheckouts()
	subject.HandleEvent(event)
	subject.HandleEvent(MapPayload{
		"event":      "checkout",
		"ref":        "other-feature",
		"repository": uuid,
		"hash":       hash,
	})
	if got, want := subject.Ref(uuid), "other-feature"; got != want {
		t.Errorf(`subject.Ref(uuid) = %v; want %v`, got, want)
	}
}

func TestRepositoryCheckouts_Hash_returnsHash_ofLatestCheckout(t *testing.T) {
	uuid := "7e511287-0353-4bbc-b59e-4ffb752547bd"
	hash := "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15"
	event := MapPayload{
		"event":      "checkout",
		"ref":        "feature-branch",
		"repository": uuid,
		"hash":       "e242ed3bffccdf271b7fbaf34ed72d089537b42f",
	}

	subject := NewRepositoryCheckouts()
	subject.HandleEvent(event)
	subject.HandleEvent(MapPayload{
		"event":      "checkout",
		"ref":        "other-feature",
		"repository": uuid,
		"hash":       hash,
	})
	if got, want := subject.Hash(uuid), hash; got != want {
		t.Errorf(`subject.Hash(uuid) = %v; want %v`, got, want)
	}
}

func TestRepositoryCheckouts_Scan_and_Value_are_inverse_operations(t *testing.T) {
	uuid := "7e511287-0353-4bbc-b59e-4ffb752547bd"
	hash := "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15"
	event := MapPayload{
		"event":      "checkout",
		"ref":        "feature-branch",
		"repository": uuid,
		"hash":       hash,
	}

	subject := NewRepositoryCheckouts()
	subject.HandleEvent(event)

	valued, err := subject.Value()
	if err != nil {
		t.Fatal(err)
	}

	scanned := NewRepositoryCheckouts()
	if err := scanned.Scan(valued); err != nil {
		t.Fatal(err)
	}

	if got, want := scanned, subject; !reflect.DeepEqual(got, want) {
		t.Errorf(`scanned = %v; want %v`, got, want)
	}
}
