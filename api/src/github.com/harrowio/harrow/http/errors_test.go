package http

import (
	"encoding/json"
	"fmt"
	"testing"
)

func Test_Error_asJson_doesNotIncludeMessage_forInternalServerError(t *testing.T) {
	err := NewInternalError(fmt.Errorf("sensitive information"))
	result := struct{ Message string }{}
	if err := json.Unmarshal(err.AsJSON(), &result); err != nil {
		t.Fatal(err)
	}

	if got, want := result.Message, ""; got != want {
		t.Errorf("result.Message = %q; want %q", got, want)
	}
}
