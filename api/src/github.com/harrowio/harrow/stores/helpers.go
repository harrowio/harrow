package stores

import (
	"bytes"
	"fmt"

	"github.com/harrowio/harrow/uuidhelper"
)

// concatenate uuids suitable for use in an IN clause.
func pqConcatUuids(uuids []string) (string, error) {
	result := new(bytes.Buffer)
	for i, uuid := range uuids {
		if !uuidhelper.IsValid(uuid) {
			return "", fmt.Errorf("malformed uuid: %q", uuid)
		}

		fmt.Fprintf(result, "'%s'", uuid)
		if i < len(uuids)-1 {
			fmt.Fprintf(result, ", ")
		}
	}

	return result.String(), nil
}
