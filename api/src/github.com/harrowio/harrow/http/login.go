package http

import (
	"net/http"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

func login(ctxt RequestContext, user *domain.User) (err error) {
	userBlockStore := stores.NewDbUserBlockStore(ctxt.Tx())
	sessionStore := stores.NewDbSessionStore(ctxt.Tx())

	userBlocks, err := userBlockStore.FindAllByUserUuid(user.Uuid)
	if err != nil {
		return err
	}

	session := &domain.Session{
		UserUuid:      user.Uuid,
		UserAgent:     ctxt.R().UserAgent(),
		ClientAddress: ctxt.R().Header.Get("X-Forwarded-For"),
	}

	if allowed, err := ctxt.Auth().CanCreate(session); !allowed {
		return err
	}

	invalidatedSessions, err := sessionStore.InvalidateAllButMostRecentSessionForUser(user.Uuid)
	if err != nil {
		return err
	}
	ctxt.Log().Debug().Msgf("invalidated %d sessions for user %q\n", invalidatedSessions, user.Uuid)
	if invalidatedSessions > 0 {
		ctxt.EnqueueActivity(activities.UserAccountUsedInTooManyPlaces(user), &user.Uuid)
	}

	sessionUuid, err := sessionStore.Create(session)
	if err != nil {
		return err
	}

	session, err = sessionStore.FindByUuid(sessionUuid)
	if err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.UserLoggedIn(user), &user.Uuid)

	ctxt.W().Header().Add("Location", urlForSubject(ctxt.R(), session))
	ctxt.W().WriteHeader(http.StatusCreated)

	result := struct {
		*domain.Session
		Blocks []*domain.UserBlock `json:"blocks"`
	}{
		Session: session,
		Blocks:  userBlocks,
	}
	writeAsJson(ctxt, &result)

	return err
}
