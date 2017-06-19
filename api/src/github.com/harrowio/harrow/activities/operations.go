package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(OperationStarted(&domain.Operation{}))
	registerPayload(OperationScheduled(&domain.Operation{}))
	registerPayload(OperationFailedFatally(&domain.Operation{}))
	registerPayload(OperationSucceeded(&domain.Operation{}))
	registerPayload(OperationFailed(&domain.Operation{}))
	registerPayload(OperationTimedOut(""))
	registerPayload(OperationCanceledDueToBilling(""))
	registerPayload(OperationCanceledByUser(""))
}

func OperationStarted(operation *domain.Operation) *domain.Activity {
	return &domain.Activity{
		Name:       "operation.started",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    operation,
	}
}

func OperationScheduled(operation *domain.Operation) *domain.Activity {
	return &domain.Activity{
		Name:       "operation.scheduled",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    operation,
	}
}

func OperationFailed(operation *domain.Operation) *domain.Activity {
	return &domain.Activity{
		Name:       "operation.failed",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    operation,
	}
}

func OperationSucceeded(operation *domain.Operation) *domain.Activity {
	return &domain.Activity{
		Name:       "operation.succeeded",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    operation,
	}
}

func OperationFailedFatally(operation *domain.Operation) *domain.Activity {
	return &domain.Activity{
		Name:       "operation.failed-fatally",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    operation,
	}
}

type OperationTimedOutPayload struct {
	Uuid string
}

func OperationTimedOut(organizationUuid string) *domain.Activity {
	return &domain.Activity{
		Name:       "operation.timed-out",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload: &OperationTimedOutPayload{
			Uuid: organizationUuid,
		},
	}
}

type OperationCanceledDueToBillingPayload struct {
	Uuid string
}

func OperationCanceledDueToBilling(operationUuid string) *domain.Activity {
	return &domain.Activity{
		Name:       "operation.canceled-due-to-billing",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    &OperationCanceledDueToBillingPayload{Uuid: operationUuid},
	}
}

type OperationCanceledByUserPayload struct {
	Uuid string
}

func OperationCanceledByUser(operationUuid string) *domain.Activity {
	return &domain.Activity{
		Name:       "operation.canceled-by-user",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    &OperationCanceledByUserPayload{Uuid: operationUuid},
	}
}
