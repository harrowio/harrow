package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(WebhookCreated(&domain.Webhook{}))
	registerPayload(WebhookEdited(&domain.Webhook{}))
	registerPayload(WebhookDeleted(&domain.Webhook{}))
}

func WebhookCreated(payload *domain.Webhook) *domain.Activity {
	return &domain.Activity{
		Name:       "webhook.created",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}

func WebhookEdited(payload *domain.Webhook) *domain.Activity {
	return &domain.Activity{
		Name:       "webhook.edited",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}

func WebhookDeleted(payload *domain.Webhook) *domain.Activity {
	return &domain.Activity{
		Name:       "webhook.deleted",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}
