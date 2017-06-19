package scheduler

import "github.com/harrowio/harrow/domain"

type SerializedNotification struct {
	Table string           `json:"table"`
	New   *domain.Schedule `json:"new"`
	Old   *domain.Schedule `json:"old"`
}
