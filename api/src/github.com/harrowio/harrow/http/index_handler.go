package http

import (
	"fmt"

	"github.com/gorilla/mux"
)

type indexHandler struct {
}

func MountIndexHandler(r *mux.Router, ctxt ServerContext) {
	ih := &indexHandler{}
	r.Methods("GET").Path("/").Handler(HandlerFunc(ctxt, ih.Index)).
		Name("index")
}

func (ih indexHandler) Index(ctxt RequestContext) error {

	ctxt.W().Write([]byte(fmt.Sprintf(`
{
  "organizationsUrl": "https://%s/organizations",
  "sessionsUrl": "https://%s/sessions",
  "usersUrl": "https://%s/users",
  "_links": {
    "documentation": "https://github.com/harrowio/harrow"
  }
}
`,
		requestBaseUri(ctxt.R()),
		requestBaseUri(ctxt.R()),
		requestBaseUri(ctxt.R()),
	)))
	return nil
}
