package activities

import (
	"github.com/harrowio/harrow/domain"
)

func init() {
	registerPayload(StencilInstantiated(&domain.Stencil{}))
}

func StencilInstantiated(stencil *domain.Stencil) *domain.Activity {
	return &domain.Activity{
		Name:       "stencil.instantiated",
		Extra:      map[string]interface{}{},
		OccurredOn: Clock.Now(),
		Payload:    stencil,
	}
}
