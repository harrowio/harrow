package hmail

import "net/textproto"

// Mail defines the payload of the message sent to the postal-worker.
type Mail struct {
	Headers textproto.MIMEHeader
	From    string
	To      []string
	Data    *MailContext

	RoutingKey string `json:"-"`
}
