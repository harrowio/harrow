package harrowMail

import (
	"testing"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
)

func TestOnOperation_returns_notification_handler_that_includes_operation_logs(t *testing.T) {
	subject := OnOperation("failed")
	context := NewContext("test@localhost")
	context.Activity = activities.OperationFailed(&domain.Operation{})
	context.Project = &domain.Project{Name: "foo"}
	context.Job = &domain.Job{Name: "bar"}
	result, _ := subject(context)
	if got := result.OperationLogs; got == nil {
		t.Fatalf("result.OperationLogs() was nil")
	}
}
