package activities

import (
	"encoding/json"
	"reflect"
	"sync"

	"github.com/harrowio/harrow/domain"
)

var (
	payloadConstructorsLock = &sync.Mutex{}
	payloadConstructors     = map[string]func() interface{}{}
)

func registerPayload(activity *domain.Activity) {
	payloadConstructorsLock.Lock()
	defer payloadConstructorsLock.Unlock()

	typ := reflect.TypeOf(activity.Payload).Elem()
	payloadConstructors[activity.Name] = func() interface{} {
		return reflect.New(typ).Interface()
	}
}

func UnmarshalPayload(activity *domain.Activity, rawPayload []byte) error {
	payloadFn, found := payloadConstructors[activity.Name]
	payload := (interface{})(nil)
	if found {
		payload = payloadFn()
	}

	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return err
	}

	activity.Payload = payload

	return nil
}

func UnmarshalJSON(data []byte) (*domain.Activity, error) {
	result := struct {
		domain.Activity
		Payload json.RawMessage `json:"payload"`
	}{}

	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	if err := UnmarshalPayload(&result.Activity, result.Payload); err != nil {
		return nil, err
	}

	return &result.Activity, nil
}
