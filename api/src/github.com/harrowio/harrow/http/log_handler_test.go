package http

import (
	"testing"

	"github.com/harrowio/harrow/config"

	"github.com/gorilla/mux"
)

func Test_LogHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	ctxt := NewTestContext(nil, nil, nil, nil, config.GetConfig())
	MountLogHandler(r, ctxt)

	spec := routingSpec{
		{"GET", "/logs/:uuid", "log-show"},
	}

	spec.run(r, t)
}
