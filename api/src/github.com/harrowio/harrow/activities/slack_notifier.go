package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(SlackNotifierCreated(&domain.SlackNotifier{}))
	registerPayload(SlackNotifierEdited(&domain.SlackNotifier{}))
	registerPayload(SlackNotifierDeleted(&domain.SlackNotifier{}))
}

func SlackNotifierCreated(payload *domain.SlackNotifier) *domain.Activity {
	return &domain.Activity{
		Name:       "slack-notifiers.created",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}

func SlackNotifierEdited(payload *domain.SlackNotifier) *domain.Activity {
	return &domain.Activity{
		Name:       "slack-notifiers.edited",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}

func SlackNotifierDeleted(payload *domain.SlackNotifier) *domain.Activity {
	return &domain.Activity{
		Name:       "slack-notifiers.deleted",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}
