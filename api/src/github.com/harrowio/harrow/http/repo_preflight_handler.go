package http

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/git"
)

type repoPreflightHandler struct {
}

func MountRepoPreflightHandler(r *mux.Router, ctxt ServerContext) {
	ih := &repoPreflightHandler{}
	root := r.PathPrefix("/repo_preflight").Subrouter()
	root.Methods("GET").Handler(HandlerFunc(ctxt, ih.Index)).
		Name("index")
}

func (rph repoPreflightHandler) Index(ctxt RequestContext) error {

	if user := ctxt.User(); user == nil {
		return ErrLoginRequired
	}

	repoUrl := strings.TrimSpace(ctxt.R().URL.Query().Get("url"))
	if repoUrl == "" {
		return NewMalformedParameters("url", fmt.Errorf("empty"))
	}

	u, err := url.Parse(repoUrl)
	if err != nil {
		return NewMalformedParameters("url", fmt.Errorf("Not a valid URL"))
	}

	repo, _ := git.NewRepository(repoUrl) // can't err, we already parsed URL

	result := struct {
		Accessible bool     `json:"accessible"`
		Url        *url.URL `json:"url"`
	}{Url: u, Accessible: repo.IsAccessible()}

	var o []byte
	ctxt.W().Header().Set("Content-Type", "application/json")
	if b, err := json.MarshalIndent(result, "", "  "); err != nil {
		return NewInternalError(fmt.Errorf("Unable to marshal json (%s)", err))
	} else {
		o = b
	}

	ctxt.W().Write(o)
	return nil
}
