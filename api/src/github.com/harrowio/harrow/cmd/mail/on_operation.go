package harrowMail

import (
	"fmt"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/hmail"
	"github.com/harrowio/harrow/loxer"
)

func init() {
	RegisterHandler("operation.succeeded", OnOperation("Success"))
	RegisterHandler("operation.failed", OnOperation("Failed"))
	RegisterHandler("operation.timed-out", OnOperation("Timed Out"))
}

func OnOperation(actionName string) NotificationHandler {
	return func(ctxt *Context) (*hmail.MailContext, error) {
		operation, ok := ctxt.Activity.Payload.(*domain.Operation)
		if !ok {
			return nil, fmt.Errorf("unexpected payload type: %T (want %T)", ctxt.Activity.Payload, operation)
		}
		result := new(hmail.MailContext)
		result.Actor = &hmail.Actor{
			DisplayName: ctxt.Project.Name,
		}
		result.Action = &hmail.Action{
			DisplayName: actionName,
		}
		result.Object = &hmail.Object{
			DisplayName: fmt.Sprintf("%s", ctxt.Job.Name),
			Uri:         fmt.Sprintf("#/a/operations/%s", operation.Uuid),
		}

		textRenderer := loxer.NewTextRenderer()
		for _, logMessage := range operation.LogEvents {
			textRenderer.Handle(logMessage.Event())
		}

		result.OperationLogs = &hmail.OperationLogs{
			Text: textRenderer.String(),
		}

		result.Recipient = &hmail.Recipient{
			UrlHost: ctxt.UrlHost,
		}

		result.Subject = fmt.Sprintf("[%s] %s: %s",
			actionName,
			ctxt.Project.Name, ctxt.Job.Name,
		)

		return result, nil
	}
}
