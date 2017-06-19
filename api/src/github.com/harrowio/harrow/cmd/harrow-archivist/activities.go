package harrowArchivist

import (
	"encoding/json"
	"fmt"

	"github.com/harrowio/harrow/domain"
)

type Activities interface {
	FindActivityByNameAndPayloadUuid(name, payloadUuid string) (*domain.Activity, error)
}

type ActivitiesInMemory struct {
	data map[string]*domain.Activity
}

func NewActivitiesInMemory() *ActivitiesInMemory {
	return &ActivitiesInMemory{
		data: map[string]*domain.Activity{},
	}
}

func (self *ActivitiesInMemory) Add(activity *domain.Activity) *ActivitiesInMemory {
	key := self.keyFor(activity)
	self.data[key] = activity
	return self
}

func (self *ActivitiesInMemory) keyFor(activity *domain.Activity) string {
	serializedPayload, err := json.Marshal(activity.Payload)
	if err != nil {
		panic(err)
	}

	payload := struct {
		Uuid string `json:"uuid"`
	}{}
	if err := json.Unmarshal(serializedPayload, &payload); err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s:%s", activity.Name, payload.Uuid)
}

func (self *ActivitiesInMemory) FindActivityByNameAndPayloadUuid(name string, payloadUuid string) (*domain.Activity, error) {
	key := fmt.Sprintf("%s:%s", name, payloadUuid)
	activity, found := self.data[key]
	if !found {
		return nil, new(domain.NotFoundError)
	}

	return activity, nil
}
