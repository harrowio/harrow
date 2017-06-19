package http

import (
	"net/http"

	"github.com/harrowio/harrow/authz"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
)

// ServerContext provides access to objects that are needed for fulfilling
// HTTP requests.  Objects exposed are expected to be accessed by multiple
// goroutines.
type ServerContext interface {
	// Config returns the server-wide configuration
	Config() config.Config

	// Returns a stores.KeyValueStore to be used for non-sensitive data.
	KeyValueStore() stores.KeyValueStore
	// Returns a stores.KeyValueStore to be used for secret data.
	SecretKeyValueStore() stores.SecretKeyValueStore

	Log() logger.Logger
	SetLogger(logger.Logger)

	// RequestContext returns a new context with valid values for
	// all request specific variables
	RequestContext(w http.ResponseWriter, req *http.Request) (RequestContext, error)
}

// A RequestContext extends ServerContext by providing access to request
// specific variables.
type RequestContext interface {
	ServerContext

	// R returns the current HTTP request.
	R() *http.Request
	// W return the http.ResponseWriter of the current HTTP request.
	W() http.ResponseWriter

	// User returns the user who is currently logged in.
	User() *domain.User

	// Auth provides access to the authorization logic.
	Auth() authz.Service

	// Tx returns the request-scoped database transaction.
	// Multiple calls to Tx have to return the same transaction.
	Tx() *sqlx.Tx

	// CommitTx commits the current transaction, as returned by Tx.
	// This method exists for testing purposes.
	CommitTx() error

	// Rollback the Tx() in case of an error
	RollbackTx()

	// PathParameter returns the path parameter identified by key
	// for the current request.
	PathParameter(key string) string

	// EnqueueActivity records enqueues an activity for emitting it
	// when the request's database transaction has succeeded.
	//
	// If userUuid is non-nil, its value is used as id of the user
	// triggering the activity.
	EnqueueActivity(activity *domain.Activity, userUuid *string)

	// Activities returns all activities that have been enqueued
	// for later emittal.
	Activities() []*domain.Activity
}

type Handler func(RequestContext) error
