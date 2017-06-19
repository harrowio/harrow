package http

import (
	"testing"

	"github.com/gorilla/mux"
)

func Test_IndexHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountIndexHandler(r, nil)

	spec := routingSpec{
		{"GET", "/", "index"},
	}

	spec.run(r, t)
}
