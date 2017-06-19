package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(ScheduleDeleted(&domain.Schedule{}))
	registerPayload(ScheduleEdited(&domain.Schedule{}))
}

func ScheduleDeleted(schedule *domain.Schedule) *domain.Activity {
	return &domain.Activity{
		Name:       "schedule.deleted",
		OccurredOn: Clock.Now(),
		Payload:    schedule,
		Extra:      map[string]interface{}{},
	}
}

func ScheduleEdited(schedule *domain.Schedule) *domain.Activity {
	return &domain.Activity{
		Name:       "schedule.edited",
		OccurredOn: Clock.Now(),
		Payload:    schedule,
		Extra:      map[string]interface{}{},
	}
}
