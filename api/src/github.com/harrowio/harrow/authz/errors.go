package authz

import "fmt"

type Error struct {
	// internal points to any non-authorization related error that occurred, such as failure to load an object from the database.
	internal error
	reason   string

	HaveCapabilities  map[string]bool
	MissingCapability string
}

func NewMissingCapabilityError(haveCapabilities map[string]bool, missingCapability string) *Error {
	return &Error{
		reason:            "capability_missing",
		HaveCapabilities:  haveCapabilities,
		MissingCapability: missingCapability,
	}
}

func (e *Error) Internal() error {
	return e.internal
}

func (e *Error) Error() string {
	return fmt.Sprintf("authorization error: %s", e.reason)
}

func (e *Error) Reason() string {
	return e.reason
}
