package http

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/net/netutil"

	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/git"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
	"github.com/rs/zerolog"

	"net/http"

	"github.com/gorilla/mux"

	"github.com/jmoiron/sqlx"
)

var c *config.Config
var OS git.System

func init() {
	c = config.GetConfig()
	OS = git.NewOperatingSystem(c.FilesystemConfig().GitTempDir)
}

func MountAll(r *mux.Router, ctxt ServerContext) {
	MountIndexHandler(r, ctxt)

	MountActivitiesHandler(r, ctxt)
	MountBillingPlanHandler(r, ctxt)
	MountCapistranoHandler(r, ctxt)
	MountEnvironmentHandler(r, ctxt)
	MountEmailNotifierHandler(r, ctxt)
	MountDeliveryHandler(r, ctxt)
	MountInvitationHandler(r, ctxt)
	MountGitTriggerHandler(r, ctxt)
	MountFeaturesHandler(r, ctxt)
	MountJobHandler(r, ctxt)
	MountJobNotifierHandler(r, ctxt)
	MountLogHandler(r, ctxt)
	MountNotificationRuleHandler(r, ctxt)
	MountOAuthHandler(r, ctxt)
	MountOperationHandler(r, ctxt)
	MountOrganizationHandler(r, ctxt)
	MountProjectHandler(r, ctxt)
	MountProjectMemberHandler(r, ctxt)
	MountPromptHandler(r, ctxt)
	MountRepoHandler(r, ctxt)
	MountRepoPreflightHandler(r, ctxt)
	MountPasswordResetHandler(r, ctxt)
	MountScheduleHandler(r, ctxt)
	MountSlackNotifierHandler(r, ctxt)
	MountSecretHandler(r, ctxt)
	MountSessionHandler(r, ctxt)
	MountStencilHandler(r, ctxt)
	MountScriptEditorHandler(r, ctxt)
	MountTaskHandler(r, ctxt)
	MountUserHandler(r, ctxt)
	MountUserVerificationHandler(r, ctxt)
	MountWebhookHandler(r, ctxt)
}

func ListenAndServe(l zerolog.Logger, db *sqlx.DB, bus activity.Sink, kv stores.KeyValueStore, ss stores.SecretKeyValueStore, c *config.Config) {

	ctxt := NewStandardContext(db, c, kv, ss)
	ctxt.SetLogger(l)

	r := mux.NewRouter()
	activitySink = NewBusSink(bus)

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() != "/inform" {
			l.Warn().Msgf("404 invoked: %s", r.URL.String())
		}
	})

	MountAll(r, ctxt)

	listener, err := net.Listen("tcp", c.HttpConfig().String())
	if err != nil {
		l.Fatal().Msgf("Listen Failed: %v", err)
	}

	ctxt.Log().Info().Msgf("listening on %s", c.HttpConfig())

	ctxt.Log().Info().Msgf("setting max simultanious connections to %d", c.HttpConfig().MaxSimultaneousConns)
	listener = netutil.LimitListener(listener, c.HttpConfig().MaxSimultaneousConns)

	server := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	ctxt.Log().Info().Msg("starting server")
	server.Serve(listener)

}

func requestScheme(r *http.Request) string {
	var scheme = "http"
	if len(r.URL.Scheme) > 0 {
		scheme = r.URL.Scheme
	}
	if x, ok := r.Header["X-Forwarded-Proto"]; ok {
		scheme = x[0]
	}
	return scheme
}

// returns either
// - User, nil when the Session is valid
// - nil, error when an error occured
// - nil, nil when the Session is invalid, or the sessionId is empty
func CurrentUser(c *config.Config, tx *sqlx.Tx, sessionUuid string) (*domain.User, error) {

	sessionStore := stores.NewDbSessionStore(tx)
	userStore := stores.NewDbUserStore(tx, c)
	if len(sessionUuid) == 0 {
		// Nothing given, no need to panic, might be
		// simply an unauthenticated request, but we
		// can't continue here.
		return nil, nil
	}
	if !uuidhelper.IsValid(sessionUuid) {
		return nil, ErrSessionUuidMalformed
	}
	session, err := sessionStore.FindByUuid(sessionUuid)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	if session.IsExpired() {
		return nil, ErrSessionExpired
	}

	if session.IsInvalidated() {
		return nil, ErrSessionInvalidated
	}

	if session.Validate() != nil {
		return nil, nil
	}

	user, err := userStore.FindByUuid(session.UserUuid)
	if err != nil {
		return nil, ErrSessionUserNotFound
	}

	query := fmt.Sprintf("SET LOCAL harrow.context_user_uuid TO '%s'", user.Uuid)
	tx.MustExec(query)
	return user, nil
}
