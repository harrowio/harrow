package controllerLXD

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/clock"
	"github.com/harrowio/harrow/domain"
)

type ActivitySinkInMemory struct {
	EmittedActivities []*domain.Activity
}

func NewActivitySinkInMemory() *ActivitySinkInMemory {
	return &ActivitySinkInMemory{
		EmittedActivities: []*domain.Activity{},
	}
}

// EmitActivity emits activity by recording it in an in-memory array.
func (self *ActivitySinkInMemory) EmitActivity(activity *domain.Activity) {
	self.EmittedActivities = append(self.EmittedActivities, activity)
}

type RepositoryMetaDataInMemory struct {
	MetaDataFor map[string]*domain.RepositoryMetaData
	Error       error
}

func NewRepositoryMetaDataInMemory() *RepositoryMetaDataInMemory {
	return &RepositoryMetaDataInMemory{
		MetaDataFor: map[string]*domain.RepositoryMetaData{},
		Error:       nil,
	}
}

func (self *RepositoryMetaDataInMemory) UpdateMetadata(repositoryUuid string, metadata *domain.RepositoryMetaData) error {
	self.MetaDataFor[repositoryUuid] = metadata
	return self.Error
}

func TestUpdateRepositoryMetaData_emitsAllActivitiesProducedByChanges(t *testing.T) {
	now := time.Now()
	defer func(clock clock.Interface) { activities.Clock = clock }(activities.Clock)
	activities.Clock = clock.At(now)

	repositoryUuid := "ee503829-640e-42bf-a64a-6d2e0d9bf239"
	sink := NewActivitySinkInMemory()
	action := NewUpdateRepositoryMetaData(sink, NewRepositoryMetaDataInMemory())
	old := domain.NewRepositoryMetaData().
		WithRef("refs/heads/master", "e242ed3bffccdf271b7fbaf34ed72d089537b42f")

	new := domain.NewRepositoryMetaData().
		WithRef("refs/heads/master", "e242ed3bffccdf271b7fbaf34ed72d089537b42f").
		WithRef("refs/heads/feature-branch", "6eadeac2dade6347e87c0d24fd455feffa7069f0")

	if err := action.Update(repositoryUuid, old, new); err != nil {
		t.Fatal(err)
	}

	expectedActivities := activities.FromRepositoryMetaDataChanges(
		old.Changes(new),
		repositoryUuid,
	)

	{
		got, _ := json.MarshalIndent(sink.EmittedActivities, "", "  ")
		want, _ := json.MarshalIndent(expectedActivities, "", "  ")
		if !bytes.Equal(got, want) {
			t.Errorf(`sink.EmittedActivities = %s; want %s`, got, want)
		}
	}
}

func TestUpdateRepositoryMetaData_persistsNewMetaData(t *testing.T) {
	now := time.Now()
	defer func(clock clock.Interface) { activities.Clock = clock }(activities.Clock)
	activities.Clock = clock.At(now)
	repositoryUuid := "ee503829-640e-42bf-a64a-6d2e0d9bf239"
	store := NewRepositoryMetaDataInMemory()
	sink := NewActivitySinkInMemory()
	action := NewUpdateRepositoryMetaData(sink, store)
	old := domain.NewRepositoryMetaData().
		WithRef("refs/heads/master", "e242ed3bffccdf271b7fbaf34ed72d089537b42f")

	new := domain.NewRepositoryMetaData().
		WithRef("refs/heads/master", "e242ed3bffccdf271b7fbaf34ed72d089537b42f").
		WithRef("refs/heads/feature-branch", "6eadeac2dade6347e87c0d24fd455feffa7069f0")

	if err := action.Update(repositoryUuid, old, new); err != nil {
		t.Fatal(err)
	}

	{
		got, _ := json.MarshalIndent(store.MetaDataFor[repositoryUuid], "", "  ")
		want, _ := json.MarshalIndent(new, "", "  ")
		if !bytes.Equal(got, want) {
			t.Errorf(`store.MetaDataFor[repositoryUuid] = %s; want %s`, got, want)
		}
	}
}

func TestUpdateRepositoryMetaData_doesNotEmitActivitiesIfOldMetaDataIsNil(t *testing.T) {
	repositoryUuid := "ee503829-640e-42bf-a64a-6d2e0d9bf239"
	store := NewRepositoryMetaDataInMemory()
	sink := NewActivitySinkInMemory()
	action := NewUpdateRepositoryMetaData(sink, store)
	new := domain.NewRepositoryMetaData().
		WithRef("refs/heads/master", "e242ed3bffccdf271b7fbaf34ed72d089537b42f").
		WithRef("refs/heads/feature-branch", "6eadeac2dade6347e87c0d24fd455feffa7069f0")

	if err := action.Update(repositoryUuid, nil, new); err != nil {
		t.Fatal(err)
	}

	if got, want := len(sink.EmittedActivities), 0; got != want {
		t.Errorf(`len(sink.EmittedActivities) = %v; want %v`, got, want)
	}
}

func TestUpdateRepositoryMetaData_doesNotEmitActivitiesIfOldMetaDataHasNoRefs(t *testing.T) {
	repositoryUuid := "ee503829-640e-42bf-a64a-6d2e0d9bf239"
	store := NewRepositoryMetaDataInMemory()
	sink := NewActivitySinkInMemory()
	action := NewUpdateRepositoryMetaData(sink, store)
	old := domain.NewRepositoryMetaData()
	new := domain.NewRepositoryMetaData().
		WithRef("refs/heads/master", "e242ed3bffccdf271b7fbaf34ed72d089537b42f").
		WithRef("refs/heads/feature-branch", "6eadeac2dade6347e87c0d24fd455feffa7069f0")

	if err := action.Update(repositoryUuid, old, new); err != nil {
		t.Fatal(err)
	}

	if got, want := len(sink.EmittedActivities), 0; got != want {
		t.Errorf(`len(sink.EmittedActivities) = %v; want %v`, got, want)
	}
}
