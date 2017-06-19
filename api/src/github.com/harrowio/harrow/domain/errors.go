package domain

import (
	"fmt"
)

type ValidationError struct {
	Errors    map[string][]string `json:"errors"`
	Originals map[string]error    `json:"-"`
}

func NewValidationError(key, value string) *ValidationError {
	result := EmptyValidationError()
	result.Add(key, value)
	return result
}

func EmptyValidationError() *ValidationError {
	return &ValidationError{Errors: map[string][]string{}}
}

func (ve *ValidationError) Error() string {
	return fmt.Sprintf("harrow/domain: validation error: %v", ve.Errors)
}

func (ve *ValidationError) Get(key string) string {
	errors, found := ve.Errors[key]
	if found {
		return errors[0]
	}

	return ""
}

func (ve *ValidationError) Add(key, err string) *ValidationError {
	if key == "" {
		return ve
	}

	ve.Errors[key] = append(ve.Errors[key], err)

	return ve
}

func (ve *ValidationError) ToError() error {
	if len(ve.Errors) > 0 {
		return ve
	} else {
		return nil
	}
}

type NotFoundError struct{}

func (nfe NotFoundError) Error() string {
	return fmt.Sprintf("harrow/domain: not found")
}

func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}
