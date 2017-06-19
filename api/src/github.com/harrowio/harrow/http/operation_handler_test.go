package http

import (
	"testing"

	"github.com/gorilla/mux"
)

func Test_OperationHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountOperationHandler(r, nil)

	spec := routingSpec{
		{"GET", "/operations/:uuid", "operation-show"},
	}

	spec.run(r, t)
}
