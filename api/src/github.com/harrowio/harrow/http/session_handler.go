package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

func ReadSessionCreateParams(r io.Reader) (*createSessionParams, error) {
	decoder := json.NewDecoder(r)
	var w createSessionParamsWrapper
	err := decoder.Decode(&w)
	if err != nil {
		return nil, err
	}
	return &w.Subject, nil
}

func ReadSessionPatchParams(r io.Reader) (*patchSessionParams, error) {
	decoder := json.NewDecoder(r)
	var p patchSessionParams
	err := decoder.Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type createSessionParamsWrapper struct {
	Subject createSessionParams
}

type createSessionParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type patchSessionParams struct {
	Totp int32 `json:"totp"`
}

func MountSessionHandler(r *mux.Router, ctxt ServerContext) {

	sh := sessionHandler{}

	// Collection
	root := r.PathPrefix("/sessions").Subrouter()
	root.Methods("POST").Handler(HandlerFunc(ctxt, sh.Create)).
		Name("session-create")

	// Item
	item := root.PathPrefix("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, sh.Show)).
		Name("session-show")
	item.Methods("PATCH").Handler(HandlerFunc(ctxt, sh.Validate)).
		Name("session-validate")
	item.Methods("DELETE").Handler(HandlerFunc(ctxt, sh.Logout)).
		Name("session-logout")
}

type sessionHandler struct {
}

func (self sessionHandler) Show(ctxt RequestContext) (err error) {

	var sessionUuid string = ctxt.PathParameter("uuid")

	sessionStore := stores.NewDbSessionStore(ctxt.Tx())

	session, err := sessionStore.FindByUuid(sessionUuid)
	if err != nil {
		return err
	}

	if session.IsExpired() {
		return ErrSessionExpired
	}

	writeAsJson(ctxt, session)
	return err
}

func (self sessionHandler) Create(ctxt RequestContext) (err error) {

	c := ctxt.Config()
	userStore := stores.NewDbUserStore(ctxt.Tx(), &c)

	params, err := ReadSessionCreateParams(ctxt.R().Body)
	if err != nil {
		return err
	}

	userUuid, err := userStore.FindUuidByEmailAddressAndPassword(params.Email, params.Password)
	if err != nil {
		return err
	}

	user, err := userStore.FindByUuid(userUuid)
	if err != nil {
		return err
	}

	return login(ctxt, user)
}

func (self sessionHandler) Logout(ctxt RequestContext) (err error) {

	var sessionUuid string = ctxt.PathParameter("uuid")

	sessionStore := stores.NewDbSessionStore(ctxt.Tx())

	session, err := sessionStore.FindByUuid(sessionUuid)
	if err != nil {
		return err
	}

	if session.Valid {
		session.LogOut()
		if err := sessionStore.MarkAsLoggedOut(session); err != nil {
			return err
		}
	}

	ctxt.W().WriteHeader(http.StatusNoContent)

	return nil
}

func (self sessionHandler) Validate(ctxt RequestContext) (err error) {

	sessionUuid := ctxt.PathParameter("uuid")

	c := ctxt.Config()
	userStore := stores.NewDbUserStore(ctxt.Tx(), &c)
	sessionStore := stores.NewDbSessionStore(ctxt.Tx())

	params, err := ReadSessionPatchParams(ctxt.R().Body)
	if err != nil {
		return err
	}

	session, err := sessionStore.FindByUuid(sessionUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().Can(domain.CapabilityValidate, session); !allowed {
		return err
	}

	user, err := userStore.FindByUuid(session.UserUuid)
	if err != nil {
		return err
	}

	if !user.IsValidTotpToken(params.Totp) {
		return domain.ErrTotpTokenNotValid
	}

	if err := sessionStore.MarkAsValid(session); err != nil {
		return err
	}

	session, err = sessionStore.FindByUuid(sessionUuid)
	if err != nil {
		return err
	}

	writeAsJson(ctxt, session)

	return err
}
