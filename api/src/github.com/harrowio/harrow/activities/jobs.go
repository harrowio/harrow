package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(JobAdded(&domain.Job{}))
	registerPayload(JobEdited(&domain.Job{}))
	registerPayload(JobScheduled(&domain.Schedule{}, ""))
}

func JobAdded(job *domain.Job) *domain.Activity {
	return &domain.Activity{
		Name:       "job.added",
		OccurredOn: Clock.Now(),
		Payload:    job,
		Extra:      map[string]interface{}{},
	}
}

func JobEdited(job *domain.Job) *domain.Activity {
	return &domain.Activity{
		Name:       "job.edited",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    job,
	}
}

func JobScheduled(schedule *domain.Schedule, trigger string) *domain.Activity {
	return &domain.Activity{
		Name:       "job.scheduled",
		OccurredOn: Clock.Now(),
		Extra: map[string]interface{}{
			"trigger": trigger,
		},
		Payload: schedule,
	}
}
