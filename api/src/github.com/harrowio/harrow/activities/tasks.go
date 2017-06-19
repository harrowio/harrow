package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(TaskAdded(&domain.Task{}))
	registerPayload(TaskEdited(&domain.Task{}))
}

func TaskAdded(task *domain.Task) *domain.Activity {
	return &domain.Activity{
		Name:       "task.added",
		OccurredOn: Clock.Now(),
		Payload:    task,
		Extra:      map[string]interface{}{},
	}
}

func TaskEdited(task *domain.Task) *domain.Activity {
	return &domain.Activity{
		Name:       "task.edited",
		OccurredOn: Clock.Now(),
		Payload:    task,
		Extra:      map[string]interface{}{},
	}
}
