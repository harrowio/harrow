package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(RepositoryMetaDataRefChanged(&domain.ChangedRepositoryRef{}))
	registerPayload(RepositoryMetaDataRefAdded(&domain.RepositoryRef{}))
	registerPayload(RepositoryMetaDataRefRemoved(&domain.RepositoryRef{}))
}

func FromRepositoryMetaDataChanges(changes *domain.RepositoryMetaDataChanges, repositoryUuid string) []*domain.Activity {
	activities := []*domain.Activity{}

	for _, changedRef := range changes.Changed() {
		ref := *changedRef
		ref.RepositoryUuid = repositoryUuid
		activities = append(activities, RepositoryMetaDataRefChanged(&ref))
	}

	for _, addedRef := range changes.Added() {
		ref := *addedRef
		ref.RepositoryUuid = repositoryUuid
		activities = append(activities, RepositoryMetaDataRefAdded(&ref))
	}

	for _, removedRef := range changes.Removed() {
		ref := *removedRef
		ref.RepositoryUuid = repositoryUuid
		activities = append(activities, RepositoryMetaDataRefRemoved(&ref))
	}

	return activities
}

func RepositoryMetaDataRefChanged(changedRef *domain.ChangedRepositoryRef) *domain.Activity {
	return &domain.Activity{
		Name:       "repository-metadata.ref-changed",
		Payload:    changedRef,
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
	}
}

func RepositoryMetaDataRefAdded(changedRef *domain.RepositoryRef) *domain.Activity {
	return &domain.Activity{
		Name:       "repository-metadata.ref-added",
		Payload:    changedRef,
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
	}
}

func RepositoryMetaDataRefRemoved(changedRef *domain.RepositoryRef) *domain.Activity {
	return &domain.Activity{
		Name:       "repository-metadata.ref-removed",
		Payload:    changedRef,
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
	}
}
