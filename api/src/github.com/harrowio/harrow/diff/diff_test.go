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
		&Change{Context, 1, "a"},
		&Change{Removal, 2, "b"},
		&Change{Addition, 2, "d"},
		&Change{Context, 3, "c"},
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
		&Change{Addition, 1, "A=1"},
		&Change{Context, 2, "a() {"},
		&Change{Context, 3, "	date"},
		&Change{Context, 4, "}"},
		&Change{Context, 5, ""},
		&Change{Addition, 6, "C=1"},
		&Change{Addition, 7, ""},
		&Change{Context, 8, "b() {"},
		&Change{Context, 9, "	date"},
		&Change{Context, 10, "}"},
		&Change{Addition, 11, ""},
		&Change{Addition, 12, "B=1"},
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
		&Change{Context, 65, "65"},
		&Change{Context, 66, "66"},
		&Change{Context, 67, "67"},
		&Change{Removal, 68, "68"},
		&Change{Context, 68, "69"},
		&Change{Context, 69, "70"},
		&Change{Context, 70, "71"},
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
		&Change{Context, 65, "65"},
		&Change{Context, 66, "66"},
		&Change{Context, 67, "67"},
		&Change{Removal, 68, "68"},
		&Change{Context, 68, "69"},
		&Change{Context, 69, "70"},
		&Change{Context, 70, "71"},
		&Change{Context, 95, "95"},
		&Change{Context, 96, "96"},
		&Change{Context, 97, "97"},
		&Change{Removal, 98, "98"},
		&Change{Context, 98, "99"},
		&Change{Context, 99, "100"},
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
