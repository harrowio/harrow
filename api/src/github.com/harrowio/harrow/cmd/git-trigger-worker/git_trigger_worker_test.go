package gitTriggerWorker

import (
	"testing"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
)

type InMemoryTriggerIndex struct {
	triggers []*domain.GitTrigger
}

func NewInMemoryTriggerIndex() *InMemoryTriggerIndex {
	return &InMemoryTriggerIndex{
		triggers: []*domain.GitTrigger{},
	}
}

// Add adds trigger to this index and includes trigger in the search
// for a trigger matcching an activity.
func (self *InMemoryTriggerIndex) Add(trigger *domain.GitTrigger) *InMemoryTriggerIndex {
	self.triggers = append(self.triggers, trigger)
	return self
}

func (self *InMemoryTriggerIndex) FindTriggersForActivity(activity *domain.Activity) ([]*domain.GitTrigger, error) {
	return self.triggers, nil
}

type InMemoryScheduler struct {
	scheduled  map[string]bool
	parameters map[string]*domain.OperationParameters
}

func NewInMemoryScheduler() *InMemoryScheduler {
	return &InMemoryScheduler{
		scheduled:  map[string]bool{},
		parameters: map[string]*domain.OperationParameters{},
	}
}

func (self *InMemoryScheduler) ScheduleJob(forTrigger *domain.GitTrigger, params *domain.OperationParameters) error {
	self.scheduled[forTrigger.JobUuid] = true
	self.parameters[forTrigger.Uuid] = params
	return nil
}

func (self *InMemoryScheduler) Parameters(triggerUuid string) *domain.OperationParameters {
	return self.parameters[triggerUuid]
}

func (self *InMemoryScheduler) HasScheduled(jobUuid string) bool {
	return self.scheduled[jobUuid]
}

func (self *InMemoryScheduler) ScheduledJobsCount() int {
	return len(self.scheduled)
}

func TestGitTriggerWorker_HandleActivity_schedulesAJob_ifAnyTriggerMatchesActivity(t *testing.T) {
	creatorUuid := "e2eb693c-def3-414e-b4dd-ef17d2460b4f"
	repositoryUuid := "f7364b78-011a-4fe7-b012-dfa3cf166a8c"
	jobUuid := "894826d5-6734-4e01-a850-5d93b77937c6"
	matchesAnyRef := domain.NewGitTrigger("test", creatorUuid).
		ForJob(jobUuid).
		ForChangeType("add").
		MatchingRef(".")

	triggerIndex := NewInMemoryTriggerIndex().
		Add(matchesAnyRef)

	scheduler := NewInMemoryScheduler()
	activity := activities.RepositoryMetaDataRefAdded(&domain.RepositoryRef{
		RepositoryUuid: repositoryUuid,
		Symbolic:       "refs/heads/master",
		Hash:           "e242ed3bffccdf271b7fbaf34ed72d089537b42f",
	})

	worker := NewGitTriggerWorker(triggerIndex, scheduler)
	if err := worker.HandleActivity(activity); err != nil {
		t.Fatal(err)
	}

	if got, want := scheduler.HasScheduled(jobUuid), true; got != want {
		t.Errorf(`scheduler.HasScheduled(jobUuid) = %v; want %v`, got, want)
	}
}

func TestGitTriggerWorker_HandleActivity_doesNotScheduleAJob_ifNoTriggerMatchesActivity(t *testing.T) {
	repositoryUuid := "f7364b78-011a-4fe7-b012-dfa3cf166a8c"

	triggerIndex := NewInMemoryTriggerIndex()

	scheduler := NewInMemoryScheduler()
	activity := activities.RepositoryMetaDataRefAdded(&domain.RepositoryRef{
		RepositoryUuid: repositoryUuid,
		Symbolic:       "refs/heads/master",
		Hash:           "e242ed3bffccdf271b7fbaf34ed72d089537b42f",
	})

	worker := NewGitTriggerWorker(triggerIndex, scheduler)
	if err := worker.HandleActivity(activity); err != nil {
		t.Fatal(err)
	}

	if got, want := scheduler.ScheduledJobsCount(), 0; got != want {
		t.Errorf(`scheduler.ScheduledJobsCount() = %v; want %v`, got, want)
	}
}

func TestOperationParametersForActivity_setsCheckoutToRefFromActivity(t *testing.T) {
	repositoryUuid := "78f71eec-3c21-48b9-82cc-5fd5e6b27b32"
	ref := domain.NewRepositoryRef("refs/heads/feature-branch", "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15")
	ref.RepositoryUuid = repositoryUuid
	changedRef := &domain.ChangedRepositoryRef{
		RepositoryUuid: repositoryUuid,
		Symbolic:       "refs/heads/feature-branch",
		OldHash:        "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15",
		NewHash:        "e242ed3bffccdf271b7fbaf34ed72d089537b42f",
	}

	testcases := []struct {
		Activity *domain.Activity
		Checkout string
	}{
		{activities.RepositoryMetaDataRefAdded(ref), "feature-branch"},
		{activities.RepositoryMetaDataRefRemoved(ref), "feature-branch"},
		{activities.RepositoryMetaDataRefChanged(changedRef), "feature-branch"},
	}

	for _, testcase := range testcases {
		params := OperationParametersForActivity(testcase.Activity)
		if got, want := params.Checkout[repositoryUuid], testcase.Checkout; got != want {
			t.Errorf(`params.Checkout[repositoryUuid] = %v; want %v`, got, want)
		}
	}
}
