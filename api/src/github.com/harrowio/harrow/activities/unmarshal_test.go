package activities

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/harrowio/harrow/domain"
)

func TestUnmarshalJSON_restoresPayloadType(t *testing.T) {
	activity := JobAdded(&domain.Job{})
	marshaled, err := json.Marshal(activity)
	if err != nil {
		t.Fatal(err)
	}
	unmarshaled, err := UnmarshalJSON(marshaled)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := reflect.TypeOf(unmarshaled.Payload).Name(), reflect.TypeOf(&domain.Job{}).Name(); got != want {
		t.Errorf("unmarshaled.Payload.(type) = %s; want %s", got, want)
	}
}
