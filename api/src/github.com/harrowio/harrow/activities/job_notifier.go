package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(JobNotifierCreated(&domain.JobNotifier{}))
	registerPayload(JobNotifierEdited(&domain.JobNotifier{}))
	registerPayload(JobNotifierDeleted(&domain.JobNotifier{}))
}

func JobNotifierCreated(payload *domain.JobNotifier) *domain.Activity {
	return &domain.Activity{
		Name:       "job-notifiers.created",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}

func JobNotifierEdited(payload *domain.JobNotifier) *domain.Activity {
	return &domain.Activity{
		Name:       "job-notifiers.edited",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}

func JobNotifierDeleted(payload *domain.JobNotifier) *domain.Activity {
	return &domain.Activity{
		Name:       "job-notifiers.deleted",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}
