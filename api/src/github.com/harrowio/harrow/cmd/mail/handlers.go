package harrowMail

import (
	"fmt"

	"github.com/harrowio/harrow/hmail"
)

type NotificationHandler func(*Context) (*hmail.MailContext, error)

var (
	ActivityHandlers = map[string]NotificationHandler{}
)

func RegisterHandler(name string, handler NotificationHandler) {
	_, exists := ActivityHandlers[name]
	if exists {
		panic(fmt.Errorf("RegisterHandler: already registered a handler for %q", name))
	}
	ActivityHandlers[name] = handler
}
