package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(SecretAdded(&domain.Secret{}))
	registerPayload(SecretEdited(&domain.Secret{}))
	registerPayload(SecretDeleted(&domain.Secret{}))
}

func SecretAdded(secret *domain.Secret) *domain.Activity {
	return &domain.Activity{
		Name:       "secret.added",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    secret,
	}
}

func SecretEdited(secret *domain.Secret) *domain.Activity {
	return &domain.Activity{
		Name:       "secret.edited",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    secret,
	}
}

func SecretDeleted(secret *domain.Secret) *domain.Activity {
	return &domain.Activity{
		Name:       "secret.deleted",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    secret,
	}
}
