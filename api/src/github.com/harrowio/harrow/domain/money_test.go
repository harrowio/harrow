package domain

import (
	"encoding/json"
	"testing"
)

func TestMoney_String_treatsAmountAsNumberOfCents(t *testing.T) {
	price := &Money{1023, EUR}
	if got, want := price.String(), "10.23 EUR"; got != want {
		t.Errorf("price.String() = %q; want %q", got, want)
	}
}

func TestMoney_String_onlyShowsSignInFrontOfTotal(t *testing.T) {
	price := &Money{-10, EUR}
	if got, want := price.String(), "-0.10 EUR"; got != want {
		t.Errorf("price.String() = %q; want %q", got, want)
	}
}

func TestMoney_Scan_readsSameFormatAsStringRepresentation(t *testing.T) {
	price := &Money{}
	text := "10.23 EUR"

	if err := price.Scan(text); err != nil {
		t.Fatal(err)
	}

	if got, want := price.Amount, 1023; got != want {
		t.Errorf("price.Amount = %d; want %d", got, want)
	}

	if got, want := price.Currency, EUR; got != want {
		t.Errorf("price.Currency = %q; want %q", got, want)
	}
}

func TestMoney_Value_returnsStringThatCanBeScanned(t *testing.T) {
	price := &Money{-10, EUR}
	valued, ok := price.Value().(string)
	if !ok {
		t.Fatalf("price.Value().(type) = %T; want %T", price.Value(), valued)
	}

	scanned := &Money{}
	if err := scanned.Scan(valued); err != nil {
		t.Fatal(err)
	}

	if got, want := scanned.String(), price.String(); got != want {
		t.Errorf("scanned.String() = %q; want %q", got, want)
	}
}

func TestMoney_MarshalsToJSONAsAString(t *testing.T) {
	price := &Money{-10, EUR}
	marshalled, err := json.Marshal(price)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(marshalled), `"-0.10 EUR"`; got != want {
		t.Errorf("marshalled = %q; want %q", got, want)
	}
}

func TestMoney_UnmarshalsFromJSONAsAString(t *testing.T) {
	price := &Money{-10, EUR}
	marshalled, err := json.Marshal(price)
	if err != nil {
		t.Fatal(err)
	}
	unmarshalled := &Money{}
	if err := json.Unmarshal(marshalled, &unmarshalled); err != nil {
		t.Fatal(err)
	}

	if !price.Equal(unmarshalled) {
		t.Errorf("expected %q to equal %q", price, unmarshalled)
	}
}
