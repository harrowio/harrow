package domain

import "testing"

func TestRepositoryMetaData_Changes_returnsRefsThatHaveBeenAdded(t *testing.T) {
	current := NewRepositoryMetaData().
		WithRef("refs/heads/master", "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15")
	next := NewRepositoryMetaData().
		WithRef("refs/heads/master", "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15").
		WithRef("refs/heads/feature-branch", "823da6941730ddb8262e618e2465c6b167e26591")

	changes := current.Changes(next)

	if got := changes; got == nil {
		t.Fatalf("changes is nil")
	}

	if got, want := len(changes.Added()), 1; got != want {
		t.Fatalf(`len(changes.Added()) = %v; want %v`, got, want)
	}

	if got, want := changes.Added()[0].Symbolic, "refs/heads/feature-branch"; got != want {
		t.Errorf(`changes.Added()[0] = %v; want %v`, got, want)
	}

}

func TestRepositoryMetaData_Changes_returnsRefsThatHaveBeenRemoved(t *testing.T) {
	current := NewRepositoryMetaData().
		WithRef("refs/heads/master", "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15").
		WithRef("refs/heads/feature-branch", "823da6941730ddb8262e618e2465c6b167e26591")
	next := NewRepositoryMetaData().
		WithRef("refs/heads/master", "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15")

	changes := current.Changes(next)

	if got := changes; got == nil {
		t.Fatalf("changes is nil")
	}

	if got, want := len(changes.Removed()), 1; got != want {
		t.Fatalf(`len(changes.Removed()) = %v; want %v`, got, want)
	}

	if got, want := changes.Removed()[0].Symbolic, "refs/heads/feature-branch"; got != want {
		t.Errorf(`changes.Removed()[0] = %v; want %v`, got, want)
	}

}

func TestRepositoryMetaData_Changes_doesNotListRemovedRefsAsChangedRefs(t *testing.T) {
	current := NewRepositoryMetaData().
		WithRef("refs/heads/master", "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15").
		WithRef("refs/heads/feature-branch", "823da6941730ddb8262e618e2465c6b167e26591")
	next := NewRepositoryMetaData().
		WithRef("refs/heads/master", "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15")

	changes := current.Changes(next)

	if got := changes; got == nil {
		t.Fatalf("changes is nil")
	}

	if got, want := len(changes.Changed()), 0; got != want {
		t.Errorf(`len(changes.Changed()) = %v; want %v`, got, want)
	}
}

func TestRepositoryMetaData_Changes_returnsRefsThatHaveChanged(t *testing.T) {
	current := NewRepositoryMetaData().
		WithRef("refs/heads/master", "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15")
	next := NewRepositoryMetaData().
		WithRef("refs/heads/master", "e242ed3bffccdf271b7fbaf34ed72d089537b42f")

	changes := current.Changes(next)

	if got := changes; got == nil {
		t.Fatalf("changes is nil")
	}

	if got, want := len(changes.Changed()), 1; got != want {
		t.Fatalf(`len(changes.Changed()) = %v; want %v`, got, want)
	}

	changed := changes.Changed()[0]

	if got, want := changed.Symbolic, "refs/heads/master"; got != want {
		t.Errorf(`changed.Symbolic = %v; want %v`, got, want)
	}

	if got, want := changed.OldHash, "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15"; got != want {
		t.Errorf(`changed.OldHash = %v; want %v`, got, want)
	}

	if got, want := changed.NewHash, "e242ed3bffccdf271b7fbaf34ed72d089537b42f"; got != want {
		t.Errorf(`changed.NewHash = %v; want %v`, got, want)
	}
}
