package http

import (
	"testing"

	"github.com/gorilla/mux"
)

func Test_RepoPreflightHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountRepoPreflightHandler(r, nil)

	spec := routingSpec{
		{"GET", "/repo_preflight", "index"},
	}

	spec.run(r, t)
}
