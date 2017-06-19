package domain

import (
	"bytes"
	"encoding/json"
	"testing"
)

func referenceEnvironmentJSON() []byte {
	return []byte(`{"uuid":"7659cc9a-5834-4c52-87c4-9c4e388242ee","name":"Example Environment","projectUuid":"a5fbb519-15ec-422c-9901-6a75760f1663","variables":{"LC_CTYPE":"C"},"archivedAt":null}`)
}

func referenceEnvironment() *Environment {
	return &Environment{
		Uuid:        "7659cc9a-5834-4c52-87c4-9c4e388242ee",
		Name:        "Example Environment",
		ProjectUuid: "a5fbb519-15ec-422c-9901-6a75760f1663",
		Variables: EnvironmentVariables{
			map[string]string{"LC_CTYPE": "C"},
		},
	}
}

func Test_Environment_MarshalJSON(t *testing.T) {
	b, err := json.Marshal(referenceEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, referenceEnvironmentJSON()) {
		t.Fatal("Expected marshalled JSON to have matched", string(referenceEnvironmentJSON()), "got", string(b))
	}
}

func Test_Environment_UnmarshalJSON(t *testing.T) {
	e := &Environment{}
	err := json.Unmarshal(referenceEnvironmentJSON(), e)
	if err != nil {
		t.Fatal(err)
	}
	if !e.Equal(referenceEnvironment()) {
		t.Fatalf("Expected environment to match: %#v != %#v", e, referenceEnvironment())
	}
}

func TestEnvironmentVariables_Diff_marks_additional_variables_as_added(t *testing.T) {
	a := NewEnvironment("78db1ae7-14e9-4c40-a43e-3cc8868c8b78").
		Set("a", "1")
	b := NewEnvironment("e83fb610-337b-4b0b-a6c9-a4b3ef4493d9").
		Set("a", "1").
		Set("b", "2")

	diff := a.Variables.Diff(b.Variables)

	if got, want := len(diff.Added), 1; got != want {
		t.Fatalf(`len(diff.Added) = %v; want %v`, got, want)
	}

	if got, want := diff.Added[0].Name, "b"; got != want {
		t.Errorf(`diff.Added[0].Name = %v; want %v`, got, want)
	}

	if got, want := diff.Added[0].Value, "2"; got != want {
		t.Errorf(`diff.Added[0].Value = %v; want %v`, got, want)
	}
}

func TestEnvironmentVariables_Diff_marks_missing_variables_as_removed(t *testing.T) {
	a := NewEnvironment("78db1ae7-14e9-4c40-a43e-3cc8868c8b78").
		Set("a", "1")
	b := NewEnvironment("e83fb610-337b-4b0b-a6c9-a4b3ef4493d9").
		Set("b", "2")

	diff := a.Variables.Diff(b.Variables)

	if got, want := len(diff.Added), 1; got != want {
		t.Errorf(`len(diff.Added) = %v; want %v`, got, want)
	}

	if got, want := len(diff.Changed), 0; got != want {
		t.Errorf(`len(diff.Changed) = %v; want %v`, got, want)
	}

	if got, want := len(diff.Removed), 1; got != want {
		t.Fatalf(`len(diff.Removed) = %v; want %v`, got, want)
	}

	if got, want := diff.Removed[0].Name, "a"; got != want {
		t.Errorf(`diff.Removed[0].Name = %v; want %v`, got, want)
	}

	if got, want := diff.Removed[0].Value, "1"; got != want {
		t.Errorf(`diff.Removed[0].Value = %v; want %v`, got, want)
	}
}

func TestEnvironmentVariables_Diff_marks_changed_variables_as_changed(t *testing.T) {
	a := NewEnvironment("78db1ae7-14e9-4c40-a43e-3cc8868c8b78").
		Set("a", "1")
	b := NewEnvironment("e83fb610-337b-4b0b-a6c9-a4b3ef4493d9").
		Set("a", "2")

	diff := a.Variables.Diff(b.Variables)

	if got, want := len(diff.Changed), 1; got != want {
		t.Fatalf(`len(diff.Changed) = %v; want %v`, got, want)
	}

	if got, want := diff.Changed[0].Name, "a"; got != want {
		t.Errorf(`diff.Changed[0].Name = %v; want %v`, got, want)
	}

	if got, want := diff.Changed[0].Value, "2"; got != want {
		t.Errorf(`diff.Changed[0].Value = %v; want %v`, got, want)
	}

	if got := diff.Changed[0].OldValue; got == nil {
		t.Fatalf(`diff.Changed[0].OldValue is nil`)
	}

	if got, want := *diff.Changed[0].OldValue, "1"; got != want {
		t.Errorf(`*diff.Changed[0].OldValue = %v; want %v`, got, want)
	}
}

func TestEnvironmentVariables_Diff_returns_no_changes_if_both_variable_sets_are_equal(t *testing.T) {
	a := NewEnvironment("78db1ae7-14e9-4c40-a43e-3cc8868c8b78").
		Set("a", "1")
	b := NewEnvironment("e83fb610-337b-4b0b-a6c9-a4b3ef4493d9").
		Set("a", "1")

	diff := a.Variables.Diff(b.Variables)

	if got, want := len(diff.Changed), 0; got != want {
		t.Errorf(`len(diff.Changed) = %v; want %v`, got, want)
	}

	if got, want := len(diff.Added), 0; got != want {
		t.Errorf(`len(diff.Added) = %v; want %v`, got, want)
	}

	if got, want := len(diff.Removed), 0; got != want {
		t.Errorf(`len(diff.Removed) = %v; want %v`, got, want)
	}
}
