package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(NotificationRuleCreated(&domain.NotificationRule{}))
	registerPayload(NotificationRuleEdited(&domain.NotificationRule{}))
	registerPayload(NotificationRuleDeleted(&domain.NotificationRule{}))
}

func NotificationRuleCreated(payload *domain.NotificationRule) *domain.Activity {
	return &domain.Activity{
		Name:       "notification-rules.created",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}

func NotificationRuleEdited(payload *domain.NotificationRule) *domain.Activity {
	return &domain.Activity{
		Name:       "notification-rules.edited",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}

func NotificationRuleDeleted(payload *domain.NotificationRule) *domain.Activity {
	return &domain.Activity{
		Name:       "notification-rules.deleted",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}
