package hmail

// MailContext holds all values needed in the email templates.
type MailContext struct {
	Subject       string
	Recipient     *Recipient
	Actor         *Actor
	Action        *Action
	Object        *Object
	OperationLogs *OperationLogs
}

// Actor holds information about who or what initiated a transaction.
type Actor struct {
	DisplayName string
}

// Action holds information about the action that triggered a
// transaction.
type Action struct {
	DisplayName string
	Description string
}

// Object holds information about what the object of the transaction's
// action is.
type Object struct {
	DisplayName string
	Uri         string
}

// Recipient holds personalized information about the mail's recipient.
type Recipient struct {
	DisplayName string
	Subject     string
	UrlHost     string
}

type OperationLogs struct {
	Text string
}
