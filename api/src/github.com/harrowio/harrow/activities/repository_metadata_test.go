package activities

import (
	"testing"

	"github.com/harrowio/harrow/domain"
)

func TestFromRepositoryMetaDataChanges_returnsAnActivityForEachChangedRef(t *testing.T) {
	repositoryUuid := "1918edcc-353e-423c-b5df-515d332abc19"
	changes := domain.NewRepositoryMetaDataChanges()
	changes.
		Change(
			"refs/heads/master",
			"f1d2d2f924e986ac86fdf7b36c94bcdf32beec15",
			"e242ed3bffccdf271b7fbaf34ed72d089537b42f",
		).
		Change(
			"refs/heads/feature-branch",
			"886447c6c37777beedf01287d9da1da069993062",
			"e80321bc4e7a72053d508ecf8e9a02a1c24c47ef",
		)

	activities := FromRepositoryMetaDataChanges(changes, repositoryUuid)

	if got, want := len(activities), 2; got != want {
		t.Fatalf(`len(activities) = %v; want %v`, got, want)
	}

	if got, want := activities[0].Name, "repository-metadata.ref-changed"; got != want {
		t.Errorf(`activities[0].Name = %v; want %v`, got, want)
	}

	if got, want := activities[1].Name, "repository-metadata.ref-changed"; got != want {
		t.Errorf(`activities[1].Name = %v; want %v`, got, want)
	}
}

func TestFromRepositoryMetaDataChanges_associatesARepositoryUuidWithEachChangedRef(t *testing.T) {
	repositoryUuid := "1918edcc-353e-423c-b5df-515d332abc19"
	changes := domain.NewRepositoryMetaDataChanges()
	changes.
		Change(
			"refs/heads/master",
			"f1d2d2f924e986ac86fdf7b36c94bcdf32beec15",
			"e242ed3bffccdf271b7fbaf34ed72d089537b42f",
		)

	activities := FromRepositoryMetaDataChanges(changes, repositoryUuid)

	if got, want := len(activities), 1; got != want {
		t.Fatalf(`len(activities) = %v; want %v`, got, want)
	}

	payload, ok := activities[0].Payload.(*domain.ChangedRepositoryRef)
	if !ok {
		t.Fatalf("activities[0].Payload.(type) = %T; want %T", activities[0].Payload, payload)
	}

	if got, want := payload.RepositoryUuid, repositoryUuid; got != want {
		t.Errorf(`payload.RepositoryUuid = %v; want %v`, got, want)
	}
}

func TestFromRepositoryMetaDataChanges_returnsAnActivityForEachAddedRef(t *testing.T) {
	repositoryUuid := "ad7ab2b3-27fd-43c0-8ca6-37ed1b037fa7"
	changes := domain.NewRepositoryMetaDataChanges()
	changes.
		Add(domain.NewRepositoryRef("refs/heads/feature-branch", "e80321bc4e7a72053d508ecf8e9a02a1c24c47ef"))

	activities := FromRepositoryMetaDataChanges(changes, repositoryUuid)

	if got, want := len(activities), 1; got != want {
		t.Fatalf(`len(activities) = %v; want %v`, got, want)
	}

	if got, want := activities[0].Name, "repository-metadata.ref-added"; got != want {
		t.Errorf(`activities[0].Name = %v; want %v`, got, want)
	}
}

func TestFromRepositoryMetaDataChanges_associatesARepositoryUuidWithEachAddedRef_withoutChangingTheOriginalRef(t *testing.T) {
	repositoryUuid := "ad7ab2b3-27fd-43c0-8ca6-37ed1b037fa7"
	ref := domain.NewRepositoryRef("refs/heads/feature-branch", "e80321bc4e7a72053d508ecf8e9a02a1c24c47ef")
	changes := domain.NewRepositoryMetaDataChanges()
	changes.
		Add(ref)

	activities := FromRepositoryMetaDataChanges(changes, repositoryUuid)

	if got, want := len(activities), 1; got != want {
		t.Fatalf(`len(activities) = %v; want %v`, got, want)
	}

	payload, ok := activities[0].Payload.(*domain.RepositoryRef)
	if !ok {
		t.Fatalf("activities[0].Payload.(type) = %T; want %T", activities[0].Payload, payload)
	}

	if got, want := payload.RepositoryUuid, repositoryUuid; got != want {
		t.Errorf(`payload.RepositoryUuid = %v; want %v`, got, want)
	}

	if got, want := ref.RepositoryUuid, ""; got != want {
		t.Errorf(`ref.RepositoryUuid = %v; want %v`, got, want)
	}
}

func TestFromRepositoryMetaDataChanges_returnsAnActivityForEachRemovedRef(t *testing.T) {
	repositoryUuid := "ad7ab2b3-27fd-43c0-8ca6-37ed1b037fa7"
	ref := domain.NewRepositoryRef("refs/heads/feature-branch", "e80321bc4e7a72053d508ecf8e9a02a1c24c47ef")
	changes := domain.NewRepositoryMetaDataChanges()
	changes.
		Remove(ref)

	activities := FromRepositoryMetaDataChanges(changes, repositoryUuid)

	if got, want := len(activities), 1; got != want {
		t.Fatalf(`len(activities) = %v; want %v`, got, want)
	}

	if got, want := activities[0].Name, "repository-metadata.ref-removed"; got != want {
		t.Errorf(`activities[0].Name = %v; want %v`, got, want)
	}

}

func TestFromRepositoryMetaDataChanges_associatesARepositoryUuidWithEachRemovedRef_withoutChangingTheOriginalRef(t *testing.T) {
	repositoryUuid := "ad7ab2b3-27fd-43c0-8ca6-37ed1b037fa7"
	ref := domain.NewRepositoryRef("refs/heads/feature-branch", "e80321bc4e7a72053d508ecf8e9a02a1c24c47ef")
	changes := domain.NewRepositoryMetaDataChanges()
	changes.
		Remove(ref)

	activities := FromRepositoryMetaDataChanges(changes, repositoryUuid)

	if got, want := len(activities), 1; got != want {
		t.Fatalf(`len(activities) = %v; want %v`, got, want)
	}

	payload, ok := activities[0].Payload.(*domain.RepositoryRef)
	if !ok {
		t.Fatalf("activities[0].Payload.(type) = %T; want %T", activities[0].Payload, payload)
	}

	if got, want := payload.RepositoryUuid, repositoryUuid; got != want {
		t.Errorf(`payload.RepositoryUuid = %v; want %v`, got, want)
	}

	if got, want := ref.RepositoryUuid, ""; got != want {
		t.Errorf(`ref.RepositoryUuid = %v; want %v`, got, want)
	}
}
