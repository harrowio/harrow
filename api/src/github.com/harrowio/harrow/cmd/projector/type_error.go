package projector

import (
	"fmt"

	"github.com/harrowio/harrow/domain"
)

type TypeError struct {
	Activity *domain.Activity
	Expected interface{}
}

func NewTypeError(activity *domain.Activity, expected interface{}) *TypeError {
	return &TypeError{
		Activity: activity,
		Expected: expected,
	}
}

func (self *TypeError) Error() string {
	return fmt.Sprintf("type error: activity@%d: expected payload %T, got %T", self.Activity.Id, self.Expected, self.Activity.Payload)
}
