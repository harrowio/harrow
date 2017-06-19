package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(EnvironmentAdded(&domain.Environment{}))
	registerPayload(EnvironmentEdited(&domain.Environment{}))
}

func EnvironmentAdded(environment *domain.Environment) *domain.Activity {
	return &domain.Activity{
		Name:       "environment.added",
		Payload:    environment,
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
	}
}

func EnvironmentEdited(environment *domain.Environment) *domain.Activity {
	return &domain.Activity{
		Name:       "environment.edited",
		Payload:    environment,
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
	}
}
