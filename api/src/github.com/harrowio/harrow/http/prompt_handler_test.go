package http

import (
	"testing"

	"github.com/gorilla/mux"
)

func Test_PromptHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountPromptHandler(r, nil)

	spec := routingSpec{
		{"GET", "/prompts", "prompt-index"},
		{"GET", "/prompts/:key", "prompt-show"},
		{"POST", "/prompts/:key", "prompt-create"},
		{"DELETE", "/prompts/:key", "prompt-delete"},
	}

	spec.run(r, t)
}
