package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(GitTriggerCreated(&domain.GitTrigger{}))
	registerPayload(GitTriggerEdited(&domain.GitTrigger{}))
	registerPayload(GitTriggerDeleted(&domain.GitTrigger{}))
}

func GitTriggerCreated(payload *domain.GitTrigger) *domain.Activity {
	return &domain.Activity{
		Name:       "git-triggers.created",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}

func GitTriggerEdited(payload *domain.GitTrigger) *domain.Activity {
	return &domain.Activity{
		Name:       "git-triggers.edited",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}

func GitTriggerDeleted(payload *domain.GitTrigger) *domain.Activity {
	return &domain.Activity{
		Name:       "git-triggers.deleted",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}
