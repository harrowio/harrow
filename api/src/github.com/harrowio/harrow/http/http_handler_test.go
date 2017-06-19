package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/test_helpers"
	"github.com/jmoiron/sqlx"
)

// httpHandlerTest represents a single HTTP test case.  It provides
// utility methods for commonly needed functionality, such as setting up
// test data.
type httpHandlerTest struct {
	server       *httptest.Server
	tx           *sqlx.Tx
	subject      domain.Subject
	links        map[string]map[string]string
	currentUser  *domain.User
	session      *domain.Session
	context      *testContext
	world        *test_helpers.World
	URL          *url.URL
	t            *testing.T
	activities   []*domain.Activity
	responseBody []byte
	response     *http.Response
	result       interface{}
}

// NewHandlerTest creates a new test case for the given handler.
// Pass the Mount${Handler} function as te first argument, e.g:
//
//        h := NewHandlerTest(MountInvitationHandler, t)
func NewHandlerTest(fn func(*mux.Router, ServerContext), t *testing.T) *httpHandlerTest {
	db := test_helpers.GetDbConnection(t)
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	kv := test_helpers.NewMockKeyValueStore()
	ss := test_helpers.NewMockSecretKeyValueStore()
	ctxt := NewTestContext(db, tx, kv, ss, c)
	fn(r, ctxt)
	ts := httptest.NewServer(r)
	world := test_helpers.MustNewWorldUsingSecretKeyValueStore(tx, ss, t)
	tsUrl, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	result := &httpHandlerTest{
		server:      ts,
		tx:          tx,
		context:     ctxt,
		subject:     nil,
		currentUser: nil,
		world:       world,
		URL:         tsUrl,
		t:           t,
		activities:  []*domain.Activity{},
	}

	ctxt.SetActivitySink(&result.activities)
	return result
}

// Cleanup closes the test server and rolls back the test context's
// transaction.
func (test *httpHandlerTest) Cleanup() {
	test.context.Tx().Rollback()
	test.server.Close()
}

// LoginAs sets up the context for the user identified by tag.  Tag must
// name a user available on *World.
func (test *httpHandlerTest) LoginAs(tag string) {
	test.currentUser = test.world.User(tag)
	if test.currentUser == nil {
		test.t.Fatalf("No such user in world: %q", tag)
	}

	test.context.u = test.currentUser

	test.session = setupTestLoginSession(test.t, test.tx, test.currentUser)
}

// LogOut ensures that no user is set on the request context.
func (test *httpHandlerTest) LogOut() {
	test.currentUser = nil
	test.context.u = test.currentUser
	test.session = nil
}

// Subject sets the subject resource for this test.  This also sets up
// the links for the resource, so that the test case does not depend on
// hard-coded URLs any more.
func (test *httpHandlerTest) Subject(subject domain.Subject) {
	test.subject = subject
	test.links = map[string]map[string]string{}
	subject.Links(test.links, "http", test.URL.Host)
}

// User returns the currently logged-in user for this test.
func (test *httpHandlerTest) User() *domain.User { return test.currentUser }

// Url returns an arbitrary URL on the test server for this test case.
// The test server's url is prefixed to rest.
func (test *httpHandlerTest) Url(rest string) string {
	return fmt.Sprintf("http://%s%s", test.URL.Host, rest)
}

// UrlFor returns the URL for any of the test subject's links.
//
//  h.Subject(invitation)
//  h.UrlFor("self") // http://127.0.0.1:12345/invitations/a0999d1d-ce81-4833-b553-d023dd51a245
//
func (test *httpHandlerTest) UrlFor(rel string) string {
	if test.subject == nil {
		test.t.Fatalf("Subject is not set")
		return ""
	}

	link, found := test.links[rel]
	if found {
		return link["href"]
	}

	test.t.Fatalf("Subject %T has no link %q", test.subject, rel)
	return ""
}

// DoString sends a request to the test server using the given method,
// url and body.  It sends an authenticated request if a user has been
// logged in for this test case.
func (test *httpHandlerTest) DoString(method, url, body string) {
	var (
		req *http.Request
		err error
	)

	if test.currentUser != nil {
		req, err = newAuthenticatedRequest(
			test.session.Uuid,
			method,
			url,
			body,
		)
	} else {
		req, err = newRequest(
			method,
			url,
			body,
		)
	}

	if err != nil {
		test.t.Fatal(err)
	}

	test.sendRequest(req)
}

// Do sends a request to the test server, encoding params as JSON.
func (test *httpHandlerTest) Do(method, path string, params interface{}) {
	var (
		req *http.Request
		err error
	)

	if values, ok := params.(url.Values); ok {
		path = fmt.Sprintf("%s?%s", path, values.Encode())
		params = nil
	}

	if test.currentUser != nil {
		req, err = newAuthenticatedRequestJSON(
			test.session.Uuid,
			method,
			path,
			params,
		)
	} else {
		req, err = newRequestJSON(
			method,
			path,
			params,
		)
	}

	if err != nil {
		test.t.Fatal(err)
	}

	test.sendRequest(req)
}

func (test *httpHandlerTest) sendRequest(req *http.Request) {
	res, err := new(http.Client).Do(req)
	if err != nil {
		test.t.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		test.t.Fatal(err)
	}

	test.response = res
	test.responseBody = body
	test.parseResponse()
}

func (test *httpHandlerTest) parseResponse() {
	if test.result == nil || len(test.responseBody) == 0 {
		return
	}

	if err := json.Unmarshal(test.responseBody, &test.result); err != nil {
		test.t.Fatalf("%s\n%s\n", err, test.responseBody)
	}
}

// ResultTo sets the target for decoding the response JSON.  If ResultTo
// has been called, the response body of for any requests made will be
// unmarshaled into thing.
func (test *httpHandlerTest) ResultTo(thing interface{}) {
	test.result = thing
	if len(test.responseBody) > 0 {
		test.parseResponse()
	}
}
func (test *httpHandlerTest) Response() *http.Response       { return test.response }
func (test *httpHandlerTest) Tx() *sqlx.Tx                   { return test.tx }
func (test *httpHandlerTest) Config() *config.Config         { return c }
func (test *httpHandlerTest) World() *test_helpers.World     { return test.world }
func (test *httpHandlerTest) ResponseBody() []byte           { return test.responseBody }
func (test *httpHandlerTest) Activities() []*domain.Activity { return test.activities }
