package http

import (
	"testing"

	"github.com/gorilla/mux"
)

func Test_ActivitiesHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountActivitiesHandler(r, nil)

	spec := routingSpec{
		{"PUT", "/activities/:id/read-status", "activities-read-status"},
	}

	spec.run(r, t)
}
