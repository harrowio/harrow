package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(RepositoryAdded(&domain.Repository{}))
	registerPayload(RepositoryEdited(&domain.Repository{}))
	registerPayload(RepositoryDetectedAsPrivate(&domain.Repository{}))
	registerPayload(RepositoryDetectedAsPublic(&domain.Repository{}))
	registerPayload(RepositoryConnectedSuccessfully(&domain.Repository{}))
}

func RepositoryAdded(repository *domain.Repository) *domain.Activity {
	return &domain.Activity{
		Name:       "repository.added",
		Payload:    repository,
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
	}
}

func RepositoryDetectedAsPrivate(repository *domain.Repository) *domain.Activity {
	return &domain.Activity{
		Name:       "repository.detected-as-private",
		Payload:    repository,
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
	}
}

func RepositoryDetectedAsPublic(repository *domain.Repository) *domain.Activity {
	return &domain.Activity{
		Name:       "repository.detected-as-public",
		Payload:    repository,
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
	}
}

func RepositoryConnectedSuccessfully(repository *domain.Repository) *domain.Activity {
	return &domain.Activity{
		Name:       "repository.connected-successfully",
		Payload:    repository,
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
	}
}

func RepositoryEdited(repository *domain.Repository) *domain.Activity {
	return &domain.Activity{
		Name:       "repository.edited",
		Payload:    repository,
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
	}
}
