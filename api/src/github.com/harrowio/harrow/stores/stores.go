// Package stores provides storage objects, they all have a trivial interface
// for finding by UUID, updateing and deleting
package stores

import (
	"github.com/harrowio/harrow/domain"

	"database/sql"
	"strings"

	"github.com/lib/pq"
)

// resolveErrType takes what may be nil, an error, or a *pq.Error
// figures out which it is, and if we can consider that a violation in
// that we can catch in the UI, and otherwise, returns it.
// It relies on code downstream doing a similar sanity check
func resolveErrType(err error) error {

	if err == nil {
		return err
	}

	// This block only really happens if *really* there were no rows,
	// or if we've used QueryRow, and violated a constraint, see:
	//  * https://code.google.com/p/go/issues/detail?id=6651
	// and:
	//  * https://github.com/lib/pq/issues/77#issuecomment-24874659
	if err == sql.ErrNoRows {
		return err
	}

	if e, ok := err.(*pq.Error); ok {
		if e.Code.Name() == "unique_violation" {
			return validationErrorFromPqError(e)
		}
		if e.Code.Name() == "foreign_key_violation" {
			return validationErrorFromPqError(e)
		}
	}

	return err

}

func normalBrackets(c rune) bool {
	return c == '(' || c == ')'
}

func validationErrorFromPqError(err *pq.Error) *domain.ValidationError {
	var ve = &domain.ValidationError{Errors: make(map[string][]string)}
	ve.Errors[strings.FieldsFunc(err.Detail, normalBrackets)[1]] = []string{err.Code.Name()}
	return ve
}
