package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(EmailNotifierCreated(&domain.EmailNotifier{}))
	registerPayload(EmailNotifierEdited(&domain.EmailNotifier{}))
	registerPayload(EmailNotifierDeleted(&domain.EmailNotifier{}))
}

func EmailNotifierCreated(payload *domain.EmailNotifier) *domain.Activity {
	return &domain.Activity{
		Name:       "email-notifiers.created",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}

func EmailNotifierEdited(payload *domain.EmailNotifier) *domain.Activity {
	return &domain.Activity{
		Name:       "email-notifiers.edited",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}

func EmailNotifierDeleted(payload *domain.EmailNotifier) *domain.Activity {
	return &domain.Activity{
		Name:       "email-notifiers.deleted",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}
