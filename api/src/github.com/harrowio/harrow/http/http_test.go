package http

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/harrowio/harrow/authz"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/stores"
	helpers "github.com/harrowio/harrow/test_helpers"
	"github.com/harrowio/harrow/uuidhelper"

	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

func newRequest(method string, url string, body string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header = map[string][]string{
		"Accept":       {"application/json"},
		"User-Agent":   {"GoLang TestCase"},
		"Content-Type": {"application/json"},
	}
	return req, nil
}

func newRequestJSON(method string, url string, object interface{}) (*http.Request, error) {
	body, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	return newRequest(method, url, string(body))
}

func newAuthenticatedRequest(sessionUuid string, method string, url string, body string) (*http.Request, error) {

	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header = map[string][]string{
		"Accept":                {"application/json"},
		"User-Agent":            {"GoLang TestCase"},
		"Content-Type":          {"application/json"},
		"X-Harrow-Session-Uuid": {sessionUuid},
	}

	return req, nil
}

func newAuthenticatedRequestJSON(sessionUuid string, method string, url string, obj interface{}) (*http.Request, error) {
	body, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return newAuthenticatedRequest(sessionUuid, method, url, string(body))
}

func setupTestUser(tx *sqlx.Tx, t *testing.T) *domain.User {
	u := &domain.User{
		Uuid:     "11111111-1111-4111-a111-111111111111",
		Name:     "Max Mustermann",
		Email:    "max@musterma.nn",
		Password: "changeme123",
	}
	if userUuid, err := stores.NewDbUserStore(tx, helpers.GetConfig(t)).Create(u); err != nil {
		t.Fatal(err)
	} else {
		u.Uuid = userUuid
	}
	return u
}

func setupOtherTestUser(tx *sqlx.Tx, t *testing.T) *domain.User {
	u := &domain.User{
		Uuid:     "22222222-2222-4222-a222-222222222222",
		Name:     "Erika Mustermann",
		Email:    "erika@musterma.nn",
		Password: "changeme123",
	}
	if userUuid, err := stores.NewDbUserStore(tx, helpers.GetConfig(t)).Create(u); err != nil {
		t.Fatal(err)
	} else {
		u.Uuid = userUuid
	}
	return u
}

func setupTestLoginSession(t *testing.T, tx *sqlx.Tx, u *domain.User) *domain.Session {

	sessionStore := stores.NewDbSessionStore(tx)

	s := &domain.Session{
		UserUuid:      u.Uuid,
		UserAgent:     "TestCase",
		ClientAddress: "::1",
	}

	sessionUuid, err := sessionStore.Create(s)
	if err != nil {
		t.Fatal(err)
	}

	s, err = sessionStore.FindByUuid(sessionUuid)
	if err != nil {
		t.Fatal(err)
	}

	return s
}

func setupTestSession(tx *sqlx.Tx, t *testing.T) {
	setupTestUser(tx, t)
	s := &domain.Session{
		Uuid:          "22222222-2222-4222-a222-222222222222",
		UserUuid:      "11111111-1111-4111-a111-111111111111",
		UserAgent:     "TestCase",
		ClientAddress: "::1",
	}
	if _, err := stores.NewDbSessionStore(tx).Create(s); err != nil {
		t.Fatal(err)
	}
}

type testContext struct {
	u          *domain.User
	c          *config.Config
	kv         stores.KeyValueStore
	ss         stores.SecretKeyValueStore
	db         *sqlx.DB
	w          http.ResponseWriter
	r          *http.Request
	tx         *sqlx.Tx
	activities *[]*domain.Activity
	// need to be pointers so both ServerContext and RequestContext can
	// point to the same values
	txRolledBack, txCommitted *bool
	authz                     *helpers.MockAuthzService
	log                       logger.Logger
}

func (tc *testContext) Log() logger.Logger {
	if tc.log == nil {
		tc.log = logger.Discard
	}
	return tc.log
}

func (tc *testContext) SetLogger(l logger.Logger) {
	tc.log = l
}

func (tc *testContext) R() *http.Request {
	return tc.r
}

func (tc *testContext) W() http.ResponseWriter {
	return tc.w
}

func (tc *testContext) KeyValueStore() stores.KeyValueStore {
	return tc.kv
}

func (tc *testContext) SecretKeyValueStore() stores.SecretKeyValueStore {
	return tc.ss
}

func (tc *testContext) Config() config.Config {
	return *tc.c
}

func (tc *testContext) User() *domain.User {
	return tc.u
}

func (tc *testContext) Auth() authz.Service {
	return tc.authz
}

func (tc *testContext) PathParameter(key string) string {
	return mux.Vars(tc.r)[key]
}

func (tc *testContext) Tx() *sqlx.Tx {
	return tc.tx
}

func (tc *testContext) CommitTx() error {
	// no-op, just record that we committed
	*tc.txCommitted = !*tc.txRolledBack
	return nil
}

func (tc *testContext) RollbackTx() {
	// no-op, just record that we rolled back
	*tc.txRolledBack = !*tc.txCommitted
}

func (tc *testContext) SetActivitySink(activities *[]*domain.Activity) {
	tc.activities = activities
}

func (tc *testContext) EnqueueActivity(activity *domain.Activity, userUuid *string) {
	if tc.activities == nil {
		return
	}
	if userUuid != nil {
		contextUserUuid := *userUuid
		activity.ContextUserUuid = &contextUserUuid
	}
	*tc.activities = append(*tc.activities, activity)
}

func (tc *testContext) Activities() []*domain.Activity {
	if tc.activities == nil {
		return []*domain.Activity{}
	}
	return *tc.activities
}

func (tc *testContext) RequestContext(w http.ResponseWriter, req *http.Request) (RequestContext, error) {
	return &testContext{
		db: tc.db,
		tx: tc.tx,
		c:  tc.c,
		kv: tc.kv,
		ss: tc.ss,

		u:     tc.u,
		authz: tc.authz,
		r:     req,
		w:     w,

		activities: tc.activities,

		txCommitted:  tc.txCommitted,
		txRolledBack: tc.txRolledBack,
	}, nil
}

func NewTestContext(db *sqlx.DB, tx *sqlx.Tx, kv stores.KeyValueStore, ss stores.SecretKeyValueStore, c *config.Config) *testContext {

	tc := &testContext{
		db: db,
		tx: tx,
		c:  c,
		kv: kv,
		ss: ss,

		u:            &domain.User{Uuid: "108007ef-3221-4186-9fcd-63d0c9280c8a"},
		authz:        helpers.NewMockAuthzService(),
		txCommitted:  new(bool),
		txRolledBack: new(bool),
	}

	return tc
}

// a testContext that returns an error from CommitTx()
type commitFailingTestContext struct {
	*testContext
}

func (tc *commitFailingTestContext) CommitTx() error {
	return errors.New("Failed to commit")
}

func NewCommitFailingTestContext(db *sqlx.DB, tx *sqlx.Tx, kv stores.KeyValueStore, ss stores.SecretKeyValueStore, c *config.Config) *commitFailingTestContext {
	return &commitFailingTestContext{
		testContext: NewTestContext(db, tx, kv, ss, c),
	}
}

func (tc *commitFailingTestContext) RequestContext(w http.ResponseWriter, req *http.Request) (RequestContext, error) {
	rc, err := tc.testContext.RequestContext(w, req)
	if err != nil {
		return nil, err
	}
	testContext, ok := rc.(*testContext)
	if !ok {
		return nil, errors.New("could not cast RequestContext to *testContext")
	}
	return &commitFailingTestContext{testContext: testContext}, nil
}

type routingTestCase struct {
	Method string
	Path   string
	Name   string
}

type authzTestContext struct {
	u  *domain.User
	c  *config.Config
	kv stores.KeyValueStore
	ss stores.SecretKeyValueStore
	db *sqlx.DB
	w  http.ResponseWriter
	r  *http.Request
	tx *sqlx.Tx
	// need to be pointers so both ServerContext and RequestContext can
	// point to the same values
	txRolledBack, txCommitted *bool
	activities                *[]*domain.Activity
	authz                     authz.Service
	log                       logger.Logger
}

func (tc *authzTestContext) Log() logger.Logger {
	if tc.log == nil {
		tc.log = logger.Discard
	}
	return tc.log
}

func (tc *authzTestContext) SetLogger(l logger.Logger) {
	tc.log = l
}

func (tc *authzTestContext) R() *http.Request {
	return tc.r
}

func (tc *authzTestContext) W() http.ResponseWriter {
	return tc.w
}

func (tc *authzTestContext) KeyValueStore() stores.KeyValueStore {
	return tc.kv
}

func (tc *authzTestContext) SecretKeyValueStore() stores.SecretKeyValueStore {
	return tc.ss
}

func (tc *authzTestContext) Config() config.Config {
	return *tc.c
}

func (tc *authzTestContext) User() *domain.User {
	return tc.u
}

func (tc *authzTestContext) Auth() authz.Service {
	return tc.authz
}

func (tc *authzTestContext) PathParameter(key string) string {
	return mux.Vars(tc.r)[key]
}

func (tc *authzTestContext) Tx() *sqlx.Tx {
	return tc.tx
}

func (tc *authzTestContext) EnqueueActivity(activity *domain.Activity, userUuid *string) {
	if tc.activities == nil {
		return
	}
	if userUuid != nil {
		contextUserUuid := *userUuid
		activity.ContextUserUuid = &contextUserUuid
	}
	*tc.activities = append(*tc.activities, activity)
}

func (tc *authzTestContext) Activities() []*domain.Activity {
	if tc.activities == nil {
		return []*domain.Activity{}
	}
	return *tc.activities
}

func (tc *authzTestContext) CommitTx() error {
	// no-op, just record that we committed
	*tc.txCommitted = !*tc.txRolledBack
	return nil
}

func (tc *authzTestContext) RollbackTx() {
	// no-op, just record that we rolled back
	*tc.txRolledBack = !*tc.txCommitted
}

func (tc *authzTestContext) RequestContext(w http.ResponseWriter, req *http.Request) (RequestContext, error) {
	rc := &authzTestContext{
		db: tc.db,
		tx: tc.tx,
		c:  tc.c,
		kv: tc.kv,
		ss: tc.ss,

		activities: tc.activities,

		r: req,
		w: w,

		txCommitted:  tc.txCommitted,
		txRolledBack: tc.txRolledBack,
	}
	err := rc.loadUser()
	if err != nil {
		return nil, err
	}

	return rc, nil
}

func (tc *authzTestContext) loadUser() error {
	c := tc.Config()
	sessionStore := stores.NewDbSessionStore(tc.Tx())
	userStore := stores.NewDbUserStore(tc.Tx(), &c)
	sessionUuid := tc.R().Header.Get(http.CanonicalHeaderKey("x-harrow-session-uuid"))
	if len(sessionUuid) == 0 {
		return nil
	}

	if !uuidhelper.IsValid(sessionUuid) {
		return ErrSessionUuidMalformed
	}
	session, err := sessionStore.FindByUuid(sessionUuid)
	if err != nil {
		return ErrSessionNotFound
	}
	if session.Validate() != nil {
		return ErrSessionNotValid
	}
	tc.u, err = userStore.FindByUuid(session.UserUuid)
	if err != nil {
		return ErrSessionUserNotFound
	}

	query := fmt.Sprintf("SET LOCAL harrow.context_user_uuid TO '%s'", tc.u.Uuid)
	tc.tx.MustExec(query)
	tc.authz = authz.NewService(tc.tx, tc.u, tc.c)

	return nil
}

func (tc *authzTestContext) SetActivitySink(activities *[]*domain.Activity) {
	tc.activities = activities
}

func NewAuthzTestContext(db *sqlx.DB, tx *sqlx.Tx, kv stores.KeyValueStore, ss stores.SecretKeyValueStore, c *config.Config) *authzTestContext {

	tc := &authzTestContext{
		db:           db,
		tx:           tx,
		c:            c,
		kv:           kv,
		ss:           ss,
		txCommitted:  new(bool),
		txRolledBack: new(bool),
	}

	return tc
}

type routingSpec []*routingTestCase

func (spec routingSpec) run(r *mux.Router, t *testing.T) {
	for _, testcase := range spec {
		match := mux.RouteMatch{}
		req, err := http.NewRequest(testcase.Method, "http://example.com"+testcase.Path, nil)
		if err != nil {
			t.Errorf("Failed to build request: %s", err)
			continue
		}

		if !r.Match(req, &match) {
			t.Errorf("Route %#v didn't match.", testcase)
			continue
		}

		if match.Route.GetName() != testcase.Name {
			t.Errorf("Expected %s %s to match %s, got %s",
				testcase.Method,
				testcase.Path,
				testcase.Name,
				match.Route.GetName(),
			)
		}
	}
}

func createTestUser(t *testing.T, tx *sqlx.Tx) *domain.User {
	userStore := stores.NewDbUserStore(tx, helpers.GetConfig(t))
	user := &domain.User{
		Uuid:     "11111111-1111-4111-a111-111111111111",
		Name:     "Max Mustermann",
		Email:    "max@musterma.nn",
		Password: "changeme123",
	}
	userUuid, err := userStore.Create(user)
	if err != nil {
		t.Fatal(err)
	}
	user, err = userStore.FindByUuid(userUuid)
	if err != nil {
		t.Fatal(err)
	}
	return user
}

func setupHandlerTestServer(fn func(*mux.Router, ServerContext), t *testing.T) (*httptest.Server, *testContext) {
	db := helpers.GetDbConnection(t)
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	r := mux.NewRouter()
	kv := helpers.NewMockKeyValueStore()

	ss := helpers.NewMockSecretKeyValueStore()
	ctxt := NewTestContext(db, tx, kv, ss, c)
	fn(r, ctxt)
	activities := []*domain.Activity{}
	ctxt.activities = &activities
	ts := httptest.NewServer(r)
	return ts, ctxt
}

func setupAuthzHandlerTestServer(fn func(*mux.Router, ServerContext), t *testing.T) (*httptest.Server, *authzTestContext) {
	db := helpers.GetDbConnection(t)
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	r := mux.NewRouter()
	kv := helpers.NewMockKeyValueStore()
	ss := helpers.NewMockSecretKeyValueStore()
	ctxt := NewAuthzTestContext(db, tx, kv, ss, c)
	fn(r, ctxt)
	ts := httptest.NewServer(r)
	return ts, ctxt
}

// sets up a test server that uses commitFailingTestContext
func setupCommitFailingHandlerTestServer(fn func(*mux.Router, ServerContext), t *testing.T) (*httptest.Server, *commitFailingTestContext) {
	db := helpers.GetDbConnection(t)
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	r := mux.NewRouter()
	kv := helpers.NewMockKeyValueStore()
	ss := helpers.NewMockSecretKeyValueStore()
	ctxt := NewCommitFailingTestContext(db, tx, kv, ss, c)
	fn(r, ctxt)
	ts := httptest.NewServer(r)
	return ts, ctxt
}
