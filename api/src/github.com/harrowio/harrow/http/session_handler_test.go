package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
	"github.com/jmoiron/sqlx"
)

func Test_SessionHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountSessionHandler(r, nil)

	spec := routingSpec{
		{"POST", "/sessions", "session-create"},
		{"GET", "/sessions/:uuid", "session-show"},
		{"PATCH", "/sessions/:uuid", "session-validate"},
		{"DELETE", "/sessions/:uuid", "session-logout"},
	}

	spec.run(r, t)
}

func Test_SessionCreateHandler_Status400OnMalformedInput(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountSessionHandler, t)
	tx := ctxt.Tx()
	defer func(t *testing.T, tx *sqlx.Tx) { tx.Rollback() }(t, tx)
	defer ts.Close()

	test_helpers.MustNewWorld(tx, t)

	req, err := newRequest("POST", ts.URL+"/sessions/", `{"subject": {"emaimax@musterma.nn", "password":"changeme123"}}`)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusBadRequest {
		t.Fatal("Expected StatusBadRequest, got:", res.StatusCode, res.Body)
	}

}

func Test_SessionCreateHandler_Status404_ForPasswordlessUser(t *testing.T) {

	h := NewHandlerTest(MountSessionHandler, t)
	defer h.Cleanup()

	user := h.World().User("without_password")

	h.Do("POST", h.Url("/sessions"), &createSessionParamsWrapper{
		Subject: createSessionParams{
			Email: user.Email,
		},
	})
	if have, want := h.Response().StatusCode, http.StatusNotFound; have != want {
		t.Errorf("h.Response().StatusCode have=%d; want=%d", have, want)
	}
	h.Do("POST", h.Url("/sessions"), &createSessionParamsWrapper{
		Subject: createSessionParams{
			Email:    user.Email,
			Password: "bogus",
		},
	})
	if have, want := h.Response().StatusCode, http.StatusNotFound; have != want {
		t.Errorf("h.Response().StatusCode have=%d; want=%d", have, want)
	}

}

func Test_SessionCreateHandler_Status201OnSuccess(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountSessionHandler, t)
	tx := ctxt.Tx()
	setupTestUser(tx, t)
	defer func(t *testing.T, tx *sqlx.Tx) { tx.Rollback() }(t, tx)
	defer ts.Close()

	req, err := newRequest("POST", ts.URL+"/sessions/", `{"subject": {"email": "max@musterma.nn", "password":"changeme123"}}`)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusCreated {
		t.Fatal("Expected StatusCreated, got", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(body) == 0 {
		t.Fatal("body should be non-empty")
	}

	location, ok := res.Header["Location"]
	if !ok {
		t.Fatal("location header should be set")
	}
	if !strings.Contains(location[0], ts.URL+"/sessions/") {
		t.Fatal(location, "does not match the expected header")
	}

}

func Test_SessionHandler_Create_returnsUserBlocks(t *testing.T) {

	h := NewHandlerTest(MountSessionHandler, t)
	defer h.Cleanup()

	aDayAgo := time.Now().Add(-24 * time.Hour)
	user := h.World().User("default")
	reasonForBlock := "testing"
	block, err := user.NewBlock(reasonForBlock)
	if err != nil {
		t.Fatal(err)
	}

	block.BlockForever(aDayAgo)
	if err := stores.NewDbUserBlockStore(h.Tx()).Create(block); err != nil {
		t.Fatal(err)
	}

	result := struct {
		Subject struct {
			Blocks []*domain.UserBlock
		}
	}{}

	h.ResultTo(&result)
	h.Do("POST", h.Url("/sessions"), &halWrapper{
		Subject: &domain.User{
			Email:    user.Email,
			Password: "password-is-long-enough",
		},
	})

	if got, want := h.Response().StatusCode, http.StatusCreated; got != want {
		t.Fatalf("h.Response().StatusCode = %d; want %d")
	}

	if got, want := len(result.Subject.Blocks), 1; got != want {
		t.Fatalf("len(result.Subject.Blocks) = %d; want %d", got, want)
	}

	if got, want := result.Subject.Blocks[0].Reason, reasonForBlock; got != want {
		t.Fatalf("result.Subject.Blocks[0].Reason = %q; want %q", got, want)
	}
}

func Test_SessionCreateHandler_Status403OnFailure(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountSessionHandler, t)
	tx := ctxt.Tx()
	setupTestUser(tx, t)
	defer func(t *testing.T, tx *sqlx.Tx) { tx.Rollback() }(t, tx)
	defer ts.Close()

	req, err := newRequest("POST", ts.URL+"/sessions/", `{"subject": {"email": "max@musterma.nn", "password":"notmypassword"}}`)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNotFound {
		t.Fatal("Expected StatusNotFound, got", res.StatusCode)
	}

}

func Test_SessionShowHandler404NotFoundOnNonExistentUuid(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountSessionHandler, t)
	tx := ctxt.Tx()
	defer func(t *testing.T, tx *sqlx.Tx) { tx.Rollback() }(t, tx)
	defer ts.Close()

	req, err := newRequest("GET", ts.URL+"/sessions/11111111-1111-4111-a111-111111111111", ``)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusNotFound {
		t.Fatal("Expected StatusNotFound, got", res.StatusCode)
	}
}

func Test_SessionShowHandler200OKOnCorrecttUuid(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(MountSessionHandler, t)
	tx := ctxt.Tx()
	setupTestSession(tx, t)
	defer func(t *testing.T, tx *sqlx.Tx) { tx.Rollback() }(t, tx)
	defer ts.Close()

	req, err := newRequest("GET", ts.URL+"/sessions/22222222-2222-4222-a222-222222222222", ``)
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatal("Expected StatusOK, got", res.StatusCode)
	}

	var response = struct {
		Subject domain.Session
	}{}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		t.Fatal(err)
	}

	if response.Subject.Uuid != "22222222-2222-4222-a222-222222222222" {
		t.Fatal("API UUID Incorrect")
	}

	if response.Subject.UserUuid != "11111111-1111-4111-a111-111111111111" {
		t.Fatal("API UserUuid Incorrect, got", response.Subject.UserUuid)
	}

}

func Test_SessionHandler_Validate_requiresValidTotpToken(t *testing.T) {
	h := NewHandlerTest(MountSessionHandler, t)
	defer h.Cleanup()

	u := h.World().User("default")
	u.GenerateTotpSecret()
	if err := u.EnableTotp(u.CurrentTotpToken()); err != nil {
		t.Fatal(err)
	}

	userStore := stores.NewDbUserStore(h.Tx(), h.Config())
	if err := userStore.Update(u); err != nil {
		t.Fatal(err)
	}

	sessionStore := stores.NewDbSessionStore(h.Tx())
	session := u.NewSession("go-test", "127.0.0.1")
	_, err := sessionStore.Create(session)
	if err != nil {
		t.Fatal(err)
	}

	h.Subject(session)
	h.Do("PATCH", h.UrlFor("self"), struct {
		Totp int32 `json:"totp"`
	}{
		u.CurrentTotpToken(),
	})

	if got, want := h.Response().StatusCode, http.StatusOK; got != want {
		t.Fatalf("Response().StatusCode = %d; want %d", got, want)
	}
}

func Test_SessionHandler_Validate_Returns422IfTokenIsInvalid(t *testing.T) {
	h := NewHandlerTest(MountSessionHandler, t)
	defer h.Cleanup()

	u := h.World().User("default")
	u.GenerateTotpSecret()
	if err := u.EnableTotp(u.CurrentTotpToken()); err != nil {
		t.Fatal(err)
	}

	userStore := stores.NewDbUserStore(h.Tx(), h.Config())
	if err := userStore.Update(u); err != nil {
		t.Fatal(err)
	}

	sessionStore := stores.NewDbSessionStore(h.Tx())
	session := u.NewSession("go-test", "127.0.0.1")
	_, err := sessionStore.Create(session)
	if err != nil {
		t.Fatal(err)
	}

	h.Subject(session)
	h.Do("PATCH", h.UrlFor("self"), struct {
		Totp int32 `json:"totp"`
	}{
		123456,
	})

	if got, want := h.Response().StatusCode, StatusUnprocessableEntity; got != want {
		t.Fatalf("Response().StatusCode = %d; want %d", got, want)
	}
}

func Test_SessionHandler_Show_ReturnsErrSessionExpiredIfSessionIsExpired(t *testing.T) {
	h := NewHandlerTest(MountSessionHandler, t)
	defer h.Cleanup()

	u := h.World().User("default")
	sessionStore := stores.NewDbSessionStore(h.Tx())
	session := u.NewSession("go-test", "127.0.0.1")
	_, err := sessionStore.Create(session)
	if err != nil {
		t.Fatal(err)
	}

	h.Tx().MustExec(`UPDATE sessions SET expires_at = '2010-01-01' WHERE uuid = $1`, session.Uuid)
	session, _ = sessionStore.FindByUuid(session.Uuid)
	result := &ErrorJSON{}
	h.ResultTo(result)
	h.Subject(session)
	h.Do("GET", h.UrlFor("self"), nil)

	if got, want := result.Reason, "session_expired"; got != want {
		t.Errorf("result.Reason = %q; want %q", got, want)
	}
}

func TestSessionHandler_Create_enqueuesUserLoggedInActivity(t *testing.T) {
	h := NewHandlerTest(MountSessionHandler, t)
	defer h.Cleanup()

	user := h.World().User("default")

	h.Do("POST", h.Url("/sessions"), &createSessionParamsWrapper{
		Subject: createSessionParams{
			Email:    user.Email,
			Password: "password-is-long-enough",
		},
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "user.logged-in" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "user.logged-in")
}
