package domain

const (
	EventNone      = "no events"
	EventNoHandler = "no event handler registered"

	EventOperationStarted   = "operations.started"
	EventOperationFailed    = "operations.failed"
	EventOperationSucceeded = "operations.succeeded"
	EventOperationTimedOut  = "operations.timed_out"
	EventOperationScheduled = "operations.scheduled"
)
