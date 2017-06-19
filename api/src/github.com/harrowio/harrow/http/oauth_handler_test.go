package http

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/services/github"
	"github.com/harrowio/harrow/stores"
)

func Test_OAuthHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountOAuthHandler(r, nil)

	spec := routingSpec{
		{"GET", "/oauth/github/authorize", "oauth-github-authorize"},
		{"GET", "/oauth/github/signin", "oauth-github-signin"},
		{"POST", "/oauth/github/callback/authorize", "oauth-github-callback-authorize"},
		{"POST", "/oauth/github/callback/signin", "oauth-github-callback-signin"},
		{"GET", "/oauth/github/ping", "oauth-github-ping"},
		{"GET", "/oauth/github/user", "oauth-github-user"},
		{"GET", "/oauth/github/organizations", "oauth-github-organizations"},
		{"GET", "/oauth/github/repositories/:kind/:login", "oauth-github-repositories"},
		{"POST", "/oauth/github/repositories/:repoUuid/keys", "oauth-github-repositories-keys"},
	}

	spec.run(r, t)
}

func Test_OAuthHandler_GithubAuthorize(t *testing.T) {

	os.Setenv("HAR_OAUTH_GITHUB_CLIENT_ID", "test-clientid")
	defer os.Unsetenv("HAR_OAUTH_GITHUB_CLIENT_ID")

	h := NewHandlerTest(MountOAuthHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	h.Do("GET", h.Url("/oauth/github/authorize"), nil)
	location, err := h.Response().Location()
	if err != nil {
		t.Fatal(err)
	}
	if err := testUrlParam(location, "client_id", "test-clientid"); err != nil {
		t.Error(err)
	}
	if err := testUrlParam(location, "redirect_uri", "https://test.tld/#/a/github/callback/authorize"); err != nil {
		t.Error(err)
	}
	if err := testUrlParam(location, "scope", "user,write:repo_hook,write:public_key,repo"); err != nil {
		t.Error(err)
	}
}

func Test_OAuthHandler_GithubSignin(t *testing.T) {

	h := NewHandlerTest(MountOAuthHandler, t)
	defer h.Cleanup()

	h.Do("GET", h.Url("/oauth/github/signin"), nil)
	location, err := h.Response().Location()
	if err != nil {
		t.Fatal(err)
	}
	if err := testUrlParam(location, "redirect_uri", "https://test.tld/#/a/github/callback/signin"); err != nil {
		t.Error(err)
	}
}

func Test_OAuthHandler_GithubCallbackAuthorize(t *testing.T) {
	mockAcquireToken()
	h := NewHandlerTest(MountOAuthHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	h.Do("POST", h.Url("/oauth/github/callback/authorize"), nil)
	if have, want := h.Response().StatusCode, http.StatusCreated; have != want {
		t.Errorf("h.Response().StatusCode: have=%d, want %d", have, want)
	}
	tokenStore := stores.NewDbOAuthTokenStore(h.Tx())
	_, err := tokenStore.FindByAccessToken("test-token")
	if err != nil {
		t.Errorf("Unable to find token: %s")
	}

	expectActivity(h, t, "user.user-connected-github")
}

func Test_OAuthHandler_GithubCallbackSignin_SucceedsOnMatchingToken(t *testing.T) {
	h := NewHandlerTest(MountOAuthHandler, t)
	defer h.Cleanup()
	mockAcquireToken()

	user := h.World().User("default")
	tokenStore := stores.NewDbOAuthTokenStore(h.Tx())
	_, err := tokenStore.Create(&domain.OAuthToken{
		UserUuid:    user.Uuid,
		Provider:    "github",
		Scope:       "test-scope",
		AccessToken: "test-token",
		TokenType:   "bearer",
	})
	if err != nil {
		t.Fatalf("Unable to save token: %s", err)
	}
	loadGithubUser = func(accessToken string) (*github.GithubUser, error) {
		if accessToken != "test-token" {
			t.Fatalf("Incorrect access token provided: %s", accessToken)
		}
		return &github.GithubUser{}, nil
	}

	result := &struct {
		Subject domain.Session
	}{}
	h.ResultTo(result)
	h.Do("POST", h.Url("/oauth/github/callback/signin"), nil)
	if have, want := result.Subject.Valid, true; have != want {
		t.Errorf("result.Subject.Valid have=%t, want=%t", have, want)
	}

	expectActivity(h, t, "user.logged-in")
	expectActivity(h, t, "user.logged-in-via-github")
}

func Test_OAuthHandler_GithubCallbackSignin_FailsOnMissingTokenAndNoVerifiedEmail(t *testing.T) {
	h := NewHandlerTest(MountOAuthHandler, t)
	defer h.Cleanup()
	mockAcquireToken()

	user := h.World().User("default")
	loadGithubUser = func(accessToken string) (*github.GithubUser, error) {
		if accessToken != "test-token" {
			t.Fatalf("Incorrect access token provided: %s", accessToken)
		}
		return &github.GithubUser{
			Name:  user.Name,
			Login: "testik",
			Emails: []github.GithubEmail{
				github.GithubEmail{
					Email:    user.Email,
					Verified: false,
					Primary:  true,
				},
			},
		}, nil
	}

	h.Do("POST", h.Url("/oauth/github/callback/signin"), nil)
	expectError(h, t, http.StatusBadRequest, "oauth.github.no_verified_email_found")
}

func Test_OAuthHandler_GithubCallbackSignin_FailsOnMissingTokenButExistingUser(t *testing.T) {
	h := NewHandlerTest(MountOAuthHandler, t)
	defer h.Cleanup()
	mockAcquireToken()

	user := h.World().User("default")
	loadGithubUser = func(accessToken string) (*github.GithubUser, error) {
		if accessToken != "test-token" {
			t.Fatalf("Incorrect access token provided: %s", accessToken)
		}
		return &github.GithubUser{
			Name:  user.Name,
			Login: "testik",
			Emails: []github.GithubEmail{
				github.GithubEmail{
					Email:    user.Email,
					Verified: true,
					Primary:  true,
				},
			},
		}, nil
	}

	h.Do("POST", h.Url("/oauth/github/callback/signin"), nil)
	expectError(h, t, http.StatusBadRequest, "oauth.github.existing_unlinked_user")
}

func Test_OAuthHandler_GithubCallbackSignin_CreatesAnUserWithVerifiedPrimary(t *testing.T) {
	h := NewHandlerTest(MountOAuthHandler, t)
	defer h.Cleanup()
	mockAcquireToken()

	loadGithubUser = func(accessToken string) (*github.GithubUser, error) {
		if accessToken != "test-token" {
			t.Fatalf("Incorrect access token provided: %s", accessToken)
		}
		return &github.GithubUser{
			Name:  "Foo Bar",
			Login: "testik",
			Emails: []github.GithubEmail{
				github.GithubEmail{
					Email:    "foo@bar.com",
					Verified: true,
					Primary:  true,
				},
				github.GithubEmail{
					Email:    "foo@another.com",
					Verified: true,
					Primary:  false,
				},
			},
		}, nil
	}

	result := &struct {
		Subject domain.Session
	}{}
	h.ResultTo(result)
	h.Do("POST", h.Url("/oauth/github/callback/signin"), nil)
	if have, want := result.Subject.Valid, true; have != want {
		t.Errorf("result.Subject.Valid have=%t, want=%t", have, want)
	}
	userUuid := result.Subject.UserUuid
	userStore := stores.NewDbUserStore(h.Tx(), h.Config())
	user, err := userStore.FindByUuid(userUuid)
	if err != nil {
		t.Errorf("Unable to load created user: %s", user)
	}
	if have, want := user.Name, "Foo Bar"; have != want {
		t.Errorf("user.Name: have=%q, want=%q", have, want)
	}
	if have, want := user.Email, "foo@bar.com"; have != want {
		t.Errorf("user.Email: have=%q, want=%q", have, want)
	}
	tokenStore := stores.NewDbOAuthTokenStore(h.Tx())
	token, err := tokenStore.FindByUserUuid(userUuid)
	if err != nil {
		t.Errorf("Unable to load token: %s", user)
	}
	if have, want := token.AccessToken, "test-token"; have != want {
		t.Errorf("token.AccessToken: have=%q, want=%q", have, want)
	}

	expectActivity(h, t, "user.signed-up-via-github")
	expectActivity(h, t, "user.logged-in")
	expectActivity(h, t, "user.logged-in-via-github")
}

func Test_OAuthHandler_GithubCallbackSignin_CreatesAnUserWithVerifiedSecondary(t *testing.T) {
	h := NewHandlerTest(MountOAuthHandler, t)
	defer h.Cleanup()
	mockAcquireToken()

	loadGithubUser = func(accessToken string) (*github.GithubUser, error) {
		if accessToken != "test-token" {
			t.Fatalf("Incorrect access token provided: %s", accessToken)
		}
		return &github.GithubUser{
			Name:  "Foo Bar",
			Login: "testik",
			Emails: []github.GithubEmail{
				github.GithubEmail{
					Email:    "foo@bar.com",
					Verified: false, // Primary email not verified!
					Primary:  true,
				},
				github.GithubEmail{
					Email:    "foo@another.com",
					Verified: true,
					Primary:  false,
				},
			},
		}, nil
	}

	result := &struct {
		Subject domain.Session
	}{}
	h.ResultTo(result)
	h.Do("POST", h.Url("/oauth/github/callback/signin"), nil)
	userUuid := result.Subject.UserUuid
	userStore := stores.NewDbUserStore(h.Tx(), h.Config())
	user, err := userStore.FindByUuid(userUuid)
	if err != nil {
		t.Errorf("Unable to load created user: %s", user)
	}
	if have, want := user.Email, "foo@another.com"; have != want {
		t.Errorf("user.Email: have=%q, want=%q", have, want)
	}

	expectActivity(h, t, "user.signed-up-via-github")
	expectActivity(h, t, "user.logged-in")
	expectActivity(h, t, "user.logged-in-via-github")
}

func testUrlParam(u *url.URL, key, value string) error {

	str := u.String()
	query := strings.IndexRune(str, '?')
	v, err := url.ParseQuery(str[query+1 : len(str)])
	if err != nil {
		return err
	}

	if have, want := len(v[key]), 1; have != want {
		return fmt.Errorf("len(v[%q]) have=%d, want %d", key, have, want)
	}
	if have, want := v[key][0], value; have != want {
		return fmt.Errorf("%s: have=%q, want %q", key, have, want)
	}
	return nil
}

func mockAcquireToken() {
	acquireToken = func(ctxt RequestContext) (map[string]string, error) {
		return map[string]string{
			"scope":        "test-scope",
			"token_type":   "bearer",
			"access_token": "test-token",
		}, nil
	}
}

func expectError(h *httpHandlerTest, t *testing.T, code int, reason string) {
	result := new(ErrorJSON)
	h.ResultTo(result)
	if have, want := h.Response().StatusCode, code; have != want {
		t.Errorf("h.Response().StatusCode: have=%d, want %d", have, want)
	}
	if have, want := result.Reason, reason; have != want {
		t.Errorf("result.Reason: have=%q, want %q", have, want)
	}
}

func expectActivity(h *httpHandlerTest, t *testing.T, a string) {

	for _, activity := range h.Activities() {
		if activity.Name == a {
			return
		}
	}

	t.Fatalf("Activity %q not found", a)
}
