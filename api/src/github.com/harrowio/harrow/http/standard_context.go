package http

import (
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/authz"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
)

type standardContext struct {
	c  *config.Config
	d  *sqlx.DB
	kv stores.KeyValueStore
	ss stores.SecretKeyValueStore

	t          *sqlx.Tx
	u          *domain.User
	r          *http.Request
	activities []*domain.Activity
	w          http.ResponseWriter
	auth       authz.Service

	log logger.Logger
}

func (sc *standardContext) R() *http.Request {
	return sc.r
}

func (sc *standardContext) W() http.ResponseWriter {
	return sc.w
}

func (sc *standardContext) KeyValueStore() stores.KeyValueStore {
	return sc.kv
}

func (sc *standardContext) SecretKeyValueStore() stores.SecretKeyValueStore {
	return sc.ss
}

func (sc *standardContext) Config() config.Config {
	return *sc.c
}

func (sc *standardContext) User() *domain.User {
	return sc.u
}

func (sc *standardContext) Auth() authz.Service {
	return sc.auth
}

func (sc *standardContext) EnqueueActivity(activity *domain.Activity, userUuid *string) {
	sc.activities = append(sc.activities, activity)
	if userUuid != nil {
		contextUserUuid := *userUuid
		activity.ContextUserUuid = &contextUserUuid
	} else if user := sc.User(); user != nil {
		contextUserUuid := user.Uuid
		activity.ContextUserUuid = &contextUserUuid
	}
}

func (sc *standardContext) Log() logger.Logger {
	return sc.log
}

func (sc *standardContext) SetLogger(l logger.Logger) {
	sc.log = l
}

func (sc *standardContext) Activities() []*domain.Activity {
	return sc.activities
}

func (sc *standardContext) PathParameter(key string) string {

	return mux.Vars(sc.R())[key]
}

func (sc *standardContext) Tx() *sqlx.Tx {
	return sc.t
}

func (sc *standardContext) RequestContext(w http.ResponseWriter, req *http.Request) (RequestContext, error) {
	requestContext := &standardContext{}
	*requestContext = *sc

	requestContext.w = w
	requestContext.r = req
	requestContext.t = requestContext.newTx()
	if err := requestContext.loadUser(); err != nil {
		return nil, err
	}
	requestContext.auth = authz.NewService(requestContext.t, requestContext.User(), sc.c)

	return requestContext, nil
}

func (sc *standardContext) newTx() *sqlx.Tx {

	// if err := sc.d.Ping(); err != nil {
	// 	fmt.Fprintf(os.Stderr, "ERROR PINGING DATABASE, WILL PROBABLY PANIC\n")
	// }

	if sc.t == nil {
		fmt.Fprintf(os.Stderr, "had no tx in progess, starting one\n")
		sc.t = sc.d.MustBegin()
		fmt.Fprintf(os.Stderr, "have tx, will continue\n")
	} else {
		fmt.Fprintf(os.Stderr, "reusing prexisting tx\n")
	}

	return sc.t
}

var (
	AllowedWithInvalidatedSession = regexp.MustCompile(`/(sessions/|api-features)`)
)

func (sc *standardContext) loadUser() error {

	if sc.u == nil {
		sessionUuid := sc.R().Header.Get(http.CanonicalHeaderKey("x-harrow-session-uuid"))
		c := sc.Config()
		user, err := CurrentUser(&c, sc.Tx(), sessionUuid)
		if err != nil {
			if err == ErrSessionInvalidated {
				if !AllowedWithInvalidatedSession.MatchString(sc.R().URL.Path) {
					return err
				}
			} else {
				return err
			}
		}
		// could be nil on an unauthenticated request
		sc.u = user
	}

	return nil
}
func (sc *standardContext) CommitTx() error {
	return sc.Tx().Commit()
}

func (sc *standardContext) RollbackTx() {
	sc.Tx().Rollback()
}

// NewContext returns a new context with some sane defaults, over
// write them if you like
func NewStandardContext(db *sqlx.DB, c *config.Config, kv stores.KeyValueStore, ss stores.SecretKeyValueStore) ServerContext {
	sc := &standardContext{d: db, c: c, kv: kv, ss: ss}
	sc.SetLogger(logger.Discard)
	return sc
}

// NewStandardContextTx returns a new context initialized with the given
// DB connection, transaction and configuration.
func NewStandardContextTx(db *sqlx.DB, tx *sqlx.Tx, c *config.Config, kv stores.KeyValueStore, ss stores.SecretKeyValueStore) ServerContext {
	sc := &standardContext{d: db, t: tx, c: c, kv: kv, ss: ss}
	sc.SetLogger(logger.Discard)
	return sc
}
