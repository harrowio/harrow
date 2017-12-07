package diff

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

var (
	TextA = []byte(`a
b
c
`)
	TextB = []byte(`a
d
c
`)
	TextC = []byte(`a() {
	date
}

b() {
	date
}
`)
	TextD = []byte(`A=1
a() {
	date
}

C=1

b() {
	date
}

B=1
`)
)

func TestChanges_returns_a_list_of_changed_lines(t *testing.T) {
	changes, err := Changes(TextA, TextB)
	if err != nil {
		t.Fatal(err)
	}

	expected := []*Change{
		{Context, 1, "a"},
		{Removal, 2, "b"},
		{Addition, 2, "d"},
		{Context, 3, "c"},
	}

	if got, want := changes, expected; !reflect.DeepEqual(got, want) {
		t.Errorf(`changes = %v; want %v`, got, want)
	}
}

func TestChanges_increases_line_numbers_for_every_addition(t *testing.T) {
	changes, err := Changes(TextC, TextD)
	if err != nil {
		t.Fatal(err)
	}

	expected := []*Change{
		{Addition, 1, "A=1"},
		{Context, 2, "a() {"},
		{Context, 3, "	date"},
		{Context, 4, "}"},
		{Context, 5, ""},
		{Addition, 6, "C=1"},
		{Addition, 7, ""},
		{Context, 8, "b() {"},
		{Context, 9, "	date"},
		{Context, 10, "}"},
		{Addition, 11, ""},
		{Addition, 12, "B=1"},
	}

	if got, want := changes, expected; !reflect.DeepEqual(got, want) {
		t.Errorf(`changes = %v; want %v`, got, want)
	}
}

func TestChanges_extracts_line_number_from_hunk_header(t *testing.T) {
	longTextA := bytes.NewBufferString("")
	longTextB := bytes.NewBufferString("")

	for i := 1; i <= 100; i++ {
		fmt.Fprintf(longTextA, "%d\n", i)
		if i != 68 {
			fmt.Fprintf(longTextB, "%d\n", i)
		}
	}

	changes, err := Changes(longTextA.Bytes(), longTextB.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	expected := []*Change{
		{Context, 65, "65"},
		{Context, 66, "66"},
		{Context, 67, "67"},
		{Removal, 68, "68"},
		{Context, 68, "69"},
		{Context, 69, "70"},
		{Context, 70, "71"},
	}

	if got, want := changes, expected; !reflect.DeepEqual(got, want) {
		t.Errorf(`changes = %v; want %v`, got, want)
	}
}

func TestChanges_extracts_processes_multiple_hunk_headers(t *testing.T) {
	longTextA := bytes.NewBufferString("")
	longTextB := bytes.NewBufferString("")

	for i := 1; i <= 100; i++ {
		fmt.Fprintf(longTextA, "%d\n", i)
		if (i != 68) && (i != 98) {
			fmt.Fprintf(longTextB, "%d\n", i)
		}
	}

	changes, err := Changes(longTextA.Bytes(), longTextB.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	expected := []*Change{
		{Context, 65, "65"},
		{Context, 66, "66"},
		{Context, 67, "67"},
		{Removal, 68, "68"},
		{Context, 68, "69"},
		{Context, 69, "70"},
		{Context, 70, "71"},
		{Context, 95, "95"},
		{Context, 96, "96"},
		{Context, 97, "97"},
		{Removal, 98, "98"},
		{Context, 98, "99"},
		{Context, 99, "100"},
	}

	if got, want := changes, expected; !reflect.DeepEqual(got, want) {
		t.Errorf(`changes = %v; want %v`, got, want)
	}
}

func TestChanges_returns_an_empty_list_if_diff_outputs_nothing(t *testing.T) {
	changes, err := Changes(TextA, TextA)
	if err != nil {
		t.Fatal(err)
	}

	expected := []*Change{}

	if got, want := changes, expected; !reflect.DeepEqual(got, want) {
		t.Errorf(`changes, expected = %v; want %v`, got, want)
	}
}
