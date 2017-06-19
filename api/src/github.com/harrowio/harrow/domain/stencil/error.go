package stencil

import (
	"bytes"
	"fmt"
	"reflect"
)

// Error collects information about failed operations
type Error struct {
	Operations []*ErrorOperation
}

type ErrorOperation struct {
	Operation string
	Thing     interface{}
	Err       error
}

func (self *ErrorOperation) Error() string {
	if reflect.ValueOf(self.Thing).Kind() == reflect.Ptr {
		return fmt.Sprintf("%s on %T: %s", self.Operation, self.Thing, self.Err)
	} else {
		return fmt.Sprintf("%s on %v: %s", self.Operation, self.Thing, self.Err)
	}
}

func NewError() *Error {
	return &Error{
		Operations: []*ErrorOperation{},
	}
}

func (self *Error) Add(operation string, thing interface{}, err error) *Error {
	self.Operations = append(self.Operations, &ErrorOperation{
		Operation: operation,
		Thing:     thing,
		Err:       err,
	})

	return self
}

func (self *Error) Error() string {
	out := new(bytes.Buffer)
	fmt.Fprintf(out, "[\n")
	for i, operation := range self.Operations {
		fmt.Fprintf(out, "  %q", operation)
		if i < len(self.Operations)-1 {
			fmt.Fprintf(out, ",\n")
		}
	}
	fmt.Fprintf(out, "]\n")
	return out.String()
}

func (self *Error) ToError() error {
	if len(self.Operations) == 0 {
		return nil
	}

	return self
}
