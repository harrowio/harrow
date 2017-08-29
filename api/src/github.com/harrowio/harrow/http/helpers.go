package http

import (
	"net/http"

	"strings"

	"github.com/harrowio/harrow/authz"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/limits"
)

// requestBaseUri takes an http.Request and returns the canonical base. This
// function is used to make the api available via api.somehost.com, or
// somehost.com/api/, this function will always return the correctly formatted
// base URI based on the `x-forwarded-request-uri` header value, and the
// *http.Request.Host.
func requestBaseUri(r *http.Request) string {
	parts := []string{r.Host}

	fwdUri := r.Header.Get(http.CanonicalHeaderKey("x-forwarded-request-uri"))
	if len(fwdUri) > 0 {
		if strings.HasPrefix(fwdUri, "/api/") {
			parts = append(parts, "/api")
		}
	}

	return strings.Join(parts, "")
}

// urlForSubject takes a *http.Request and a domain.Subject, and using the
// helper functions requestScheme() and requestBaseUri() cooperates with the
// domain.Subject to render it's URL according to the incoming request
// properties.
func urlForSubject(r *http.Request, subject domain.Subject) string {
	return subject.OwnUrl(requestScheme(r), requestBaseUri(r))
}

// linksForSubject takes a *http.Request and a domain.Subject and returns a
// complex map of links for that subject. In cooperation between the
// *http.Request and domain.Subject, the incoming request properties are
// honored, and the map is populated with links on the same protocol scheme and
// host (with path).
func linksForSubject(auth authz.Service, r *http.Request, subject domain.Subject) map[string]map[string]string {
	links := map[string]map[string]string{}
	subject.Links(links, requestScheme(r), requestBaseUri(r))
	capabilitiesBySubject := auth.CapabilitiesBySubject()
	for authSubject, verbs := range capabilitiesBySubject {
		for _, relname := range []string{authSubject, pluralOf(authSubject)} {
			link, found := links[relname]

			if found {
				addCapabilityVerbsToLink(verbs, link)
			}
		}
	}

	authNamer, ok := subject.(authz.Subject)
	if ok {
		verbs := capabilitiesBySubject[authNamer.AuthorizationName()]
		if verbs == nil {
			verbs = []string{}
		}
		link := links["self"]
		if link != nil {
			addCapabilityVerbsToLink(verbs, link)
		}
	}
	return links
}

func pluralOf(noun string) string {
	// good enough for now
	switch noun[len(noun)-1] {
	case 'y':
		return noun[0:len(noun)-1] + "ies"
	default:
		return noun + "s"
	}
}

func addCapabilityVerbsToLink(verbs []string, link map[string]string) {
	for _, verb := range verbs {
		method := methodForVerb(verb)
		link[verb] = method
	}
}

func methodForVerb(verb string) string {
	switch verb {
	case "read":
		return "GET"
	case "archive":
		return "DELETE"
	case "edit":
		return "PUT"
	case "create":
		return "POST"
	default:
		return "POST"
	}

}

// applyFilterTemplate uses the current request context
// (requestCurrentUserUuid, requestTx) to filter the outgoing response using a
// template if necessary.
func applyFilterTemplate(r *http.Request, subj domain.Subject, o []byte) []byte {
	switch s := subj.(type) {
	case *domain.User:
		// TODO: LH Bring this back, however it needs the http.Context object now
		// return applyDomainUserTemplate(r, s, o)
		_ = s
		return o // remove me when the above TODO is fixed
	default:
		return o
	}
}

func NewLimitsFromContext(ctxt RequestContext) *limits.Client {
	cfg := ctxt.Config()
	limits := limits.NewDefaultClient(&cfg)
	limits.SetLogger(ctxt.Log())
	return limits
}
