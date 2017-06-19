package http

import (
	"crypto/hmac"
	"encoding/base32"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

func MountPasswordResetHandler(r *mux.Router, ctxt ServerContext) {

	h := &passwordResetHandler{}

	r.PathPrefix("/forgot-password").Subrouter().
		Methods("POST").Handler(HandlerFunc(ctxt, h.Forgot)).
		Name("password-forgot")

	r.PathPrefix("/reset-password").Subrouter().
		Methods("POST").Handler(HandlerFunc(ctxt, h.Reset)).
		Name("password-reset")
}

type passwordResetHandler struct {
}

func (self passwordResetHandler) Forgot(ctxt RequestContext) (err error) {

	body, err := ioutil.ReadAll(ctxt.R().Body)
	if err != nil {
		return err
	}
	params := &struct {
		Email string
	}{}
	err = json.Unmarshal(body, params)
	if err != nil {
		return err
	}

	userStore := stores.NewDbUserStore(ctxt.Tx(), c)
	user, err := userStore.FindByEmailAddress(params.Email)

	if err == nil {
		ctxt.EnqueueActivity(activities.UserRequestedPasswordReset(user), &user.Uuid)
	}

	// do not return error to not leak information about what email addresses
	// we know of
	ctxt.W().WriteHeader(http.StatusCreated)

	return nil
}

func (self passwordResetHandler) Reset(ctxt RequestContext) (err error) {

	userStore := stores.NewDbUserStore(ctxt.Tx(), c)

	body, err := ioutil.ReadAll(ctxt.R().Body)
	if err != nil {
		return err
	}
	params := &struct {
		Email    string
		Mac      string
		Password string
	}{}
	err = json.Unmarshal(body, params)
	if err != nil {
		return err
	}
	if len(params.Password) == 0 {
		return domain.NewValidationError("password", "invalid")
	}
	user, err := userStore.FindByEmailAddress(params.Email)

	if err != nil {
		return err
	}
	mac, err := base32.StdEncoding.DecodeString(params.Mac)
	if err != nil {
		return err
	}
	if !hmac.Equal(mac, user.HMAC([]byte(c.HttpConfig().UserHmacSecret))) {
		return new(domain.NotFoundError)
	}
	user.Password = params.Password
	err = userStore.Update(user)
	if err != nil {
		return err
	}

	if err := stores.NewDbSessionStore(ctxt.Tx()).ExpireAllSessionsForUserUuid(user.Uuid); err != nil {
		ctxt.Log().Error().Msgf("failed to invalidate sessions for user: %s", err)
	}

	ctxt.EnqueueActivity(activities.UserResetPassword(user), nil)

	writeAsJson(ctxt, user)
	return nil
}
