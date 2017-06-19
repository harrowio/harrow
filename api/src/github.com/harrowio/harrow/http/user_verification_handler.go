package http

import (
	"time"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

func MountUserVerificationHandler(r *mux.Router, ctxt ServerContext) {

	h := &userVerificationHandler{}

	r.PathPrefix("/verify-email").Subrouter().
		Methods("POST").Handler(HandlerFunc(ctxt, h.RerequestVerificationEmail)).
		Name("verify-email")
}

type userVerificationHandler struct {
}

func (self userVerificationHandler) RerequestVerificationEmail(ctxt RequestContext) error {
	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	activityStore := stores.NewDbActivityStore(ctxt.Tx())
	userBlockStore := stores.NewDbUserBlockStore(ctxt.Tx())

	blocks, err := userBlockStore.FindAllByUserUuid(ctxt.User().Uuid)
	if err != nil {
		ctxt.Log().Warn().Msgf("userBlockStore.FindAllByUserUuid(%q): %s", ctxt.User().Uuid, err)
	} else if len(blocks) == 0 {
		return domain.NewValidationError("email", "already_verified")
	}

	userHistory := domain.NewUserHistory(ctxt.User().Uuid, time.Now())
	if err := activityStore.AllByUser(ctxt.User().Uuid, userHistory.HandleActivity); err != nil {
		return err
	}

	fairNumberOfTries := 3

	if userHistory.VerificationEmailsRequestedToday() >= fairNumberOfTries {
		return ErrRateLimitExceeded
	}

	ctxt.EnqueueActivity(activities.UserRequestedVerificationEmail(ctxt.User()), nil)

	return nil
}
