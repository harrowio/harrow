package git

import (
	"bytes"
	"fmt"
)

type ValidationError map[string]string

func (self ValidationError) Error() string {
	out := new(bytes.Buffer)
	fmt.Fprintf(out, "git: url validation error")
	for field, msg := range self {
		fmt.Fprintf(out, "\n- %s: %s\n", field, msg)
	}
	return out.String()
}

func (self ValidationError) IsZero() bool {
	return len(self) == 0
}

func (self ValidationError) MergeOverwrite(m error) {
	if m == nil {
		return
	}
	if vE, ok := m.(ValidationError); ok {
		for k, v := range vE {
			self[k] = v
		}
	}
}
