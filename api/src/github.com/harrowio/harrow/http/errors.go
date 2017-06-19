package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/harrowio/harrow/domain"
)

// Error represents an error response from the API.  Its responsibility
// is to ensure that all errors returned by the API have are properly
// structured.
type Error struct {
	// msg is an English language description of the error intended
	// for consumption by humans.
	msg string

	// status is the HTTP status code that should be returned for
	// this error.
	status int

	// code is short identifier describing the error.  This identifier
	// is intended to be used by machines.
	code string

	// internal is an internal error that will be logged but not
	// exposed to the client.
	internal error

	// errors maps parameter names to error identifiers
	errors map[string][]string
}

// ErrorJSON is the serialized form of an API error ("wire format").
type ErrorJSON struct {
	// Reason is a machine readable identifier for the error.
	// Use this identifier for translating error messages.
	Reason string `json:"reason"`

	// Message is a human readable string describing this instance
	// of the error, such as the location at which a parse error
	// occurred.
	Message string `json:"message"`

	// Errors maps input parameter names to error identifiers
	// pertaining to that parameter:
	//
	//    { "field": ["parse_error"] }
	Errors map[string][]string `json:"errors,omitempty"`
}

func NewError(status int, code, msg string) *Error {
	return &Error{
		status: status,
		code:   code,
		msg:    msg,
	}
}

func (self *Error) MarshalJSON() ([]byte, error) {
	return json.MarshalIndent(&ErrorJSON{
		Reason:  self.code,
		Message: self.msg,
		Errors:  self.errors,
	}, "", "  ")
}

func (self *Error) AsJSON() []byte {
	data, err := self.MarshalJSON()
	if err != nil {
		panic(err)
	}
	return data
}

func NewInternalError(err error) *Error {
	return &Error{
		status:   http.StatusInternalServerError,
		code:     "internal",
		internal: err,
	}
}

func NewCapabilityMissingError(haveCapabilities map[string]bool, missingCapability string) *Error {
	result := &Error{
		status: 403,
		code:   "capability_missing",
		errors: map[string][]string{
			"missing_capability": []string{missingCapability},
			"have_capabilities":  []string{},
		},
	}

	for capabilityName, _ := range haveCapabilities {
		result.errors["have_capabilities"] = append(result.errors["have_capabilities"], capabilityName)
	}

	return result
}

func NewMalformedParameters(param string, err error) *Error {
	return &Error{
		status: http.StatusBadRequest,
		code:   "malformed_parameter",
		msg:    err.Error(),
		errors: map[string][]string{
			param: []string{"malformed"},
		},
	}
}

func NewValidationError(src *domain.ValidationError) *Error {
	return &Error{
		status: StatusUnprocessableEntity,
		code:   "invalid",
		errors: src.Errors,
	}
}

func NewOAuthError(provider string, reason string, err error) *Error {
	return NewError(http.StatusBadRequest, fmt.Sprintf("oauth.%s.%s", provider, reason), err.Error())
}

func (e *Error) Error() string {
	return e.msg
}

var (
	ErrRecoveredFromPanic   = NewError(500, "recovered_from_panic", "Recovered from panic")
	ErrNotFound             = NewError(404, "not_found", "Not found")
	ErrUnexpectedEOF        = NewError(400, "unexpected_eof", "Unexpected EOF")
	ErrSessionExpired       = NewError(403, "session_expired", "Session expired")
	ErrSessionNotFound      = NewError(403, "session_not_found", "Session not found")
	ErrSessionNotValid      = NewError(403, "session_not_valid", "Session not valid (validated)")
	ErrSessionInvalidated   = NewError(403, "session_invalidated", "Session invalidated")
	ErrRateLimitExceeded    = NewError(403, "rate_limit_exceeded", "Rate limit exceeded")
	ErrLimitsExceeded       = NewError(403, "limits_exceeded", "Billing plan limits exceeded")
	ErrSessionUserNotFound  = NewError(403, "session_user_not_found", "Session user not found")
	ErrSessionUuidMalformed = NewError(400, "session_uuid_malformed", "Session UUID malformed")
	ErrLoginRequired        = NewError(403, "login_required", "Login required")
	ErrUserBlocked          = NewError(403, "user_blocked", "User blocked")
)
