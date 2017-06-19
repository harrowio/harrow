package http

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/services/github"
	"github.com/harrowio/harrow/stores"
)

type oAuthHandler struct{}

func MountOAuthHandler(r *mux.Router, ctxt ServerContext) {

	oh := oAuthHandler{}

	// Collection
	root := r.PathPrefix("/oauth").Subrouter()

	github := root.PathPrefix("/github").Subrouter()

	// Setup Flow
	github.Methods("GET").Path("/authorize").Handler(HandlerFunc(ctxt, oh.GithubAuthorize)).
		Name("oauth-github-authorize")
	github.Methods("GET").Path("/signin").Handler(HandlerFunc(ctxt, oh.GithubSignin)).
		Name("oauth-github-signin")
	github.Methods("POST").Path("/callback/authorize").Handler(HandlerFunc(ctxt, oh.GithubCallbackAuthorize)).
		Name("oauth-github-callback-authorize")
	github.Methods("POST").Path("/callback/signin").Handler(HandlerFunc(ctxt, oh.GithubCallbackSignin)).
		Name("oauth-github-callback-signin")
	github.Methods("GET").Path("/deauthorize").Handler(HandlerFunc(ctxt, oh.GithubDeAuthorize)).
		Name("oauth-github-deauthorize")

	// Ping (Checks the connection)
	github.Methods("GET").Path("/ping").Handler(HandlerFunc(ctxt, oh.GithubPing)).
		Name("oauth-github-ping")

	// Queries
	github.Methods("GET").Path("/user").Handler(HandlerFunc(ctxt, oh.GithubUser)).
		Name("oauth-github-user")
	github.Methods("GET").Path("/organizations").Handler(HandlerFunc(ctxt, oh.GithubOrganizations)).
		Name("oauth-github-organizations")
	github.Methods("POST").Path("/repositories/{repoUuid}/keys").Handler(HandlerFunc(ctxt, oh.GithubDeployKey)).
		Name("oauth-github-repositories-keys")
	github.Methods("GET").Path("/repositories/{kind}/{login}").Handler(HandlerFunc(ctxt, oh.GithubRepos)).
		Name("oauth-github-repositories")

}

func (self oAuthHandler) GithubAuthorize(ctxt RequestContext) error {

	provider := ctxt.Config().OAuthConfig().Providers["github"]

	redirectUri := fmt.Sprintf(provider.RedirectUri, "authorize")

	return self.redirectToGithub(ctxt, redirectUri)
}

func (self oAuthHandler) GithubSignin(ctxt RequestContext) error {
	provider := ctxt.Config().OAuthConfig().Providers["github"]
	redirectUri := fmt.Sprintf(provider.RedirectUri, "signin")
	return self.redirectToGithub(ctxt, redirectUri)
}

func (self oAuthHandler) redirectToGithub(ctxt RequestContext, redirectUri string) error {

	provider := ctxt.Config().OAuthConfig().Providers["github"]

	// store state in redis
	state := self.makeRandomString()
	ctxt.KeyValueStore().Set(state, []byte{})

	scope := strings.Join(provider.Scope, ",")

	v := url.Values{}
	v.Add("redirect_uri", redirectUri)
	v.Add("scope", scope)
	v.Add("state", state)
	v.Set("client_id", provider.ClientId)

	ctxt.W().Header().Add("Location", fmt.Sprintf("%s?%s", provider.ProviderUrl, v.Encode()))

	return nil
}

// Used for ?state=_____ From the GH Docs:
// > An unguessable random string. It is used to protect
// > against cross-site request forgery attacks.
func (self oAuthHandler) makeRandomString() string {
	encoder := base64.URLEncoding
	stateRandBytes := make([]byte, 32)
	rand.Read(stateRandBytes)
	state := make([]byte, encoder.EncodedLen(len(stateRandBytes)))
	encoder.Encode(state, stateRandBytes)
	return string(state)
}

func (self oAuthHandler) GithubCallbackAuthorize(ctxt RequestContext) error {

	githubResponse, err := acquireToken(ctxt)
	if err != nil {
		return err
	}

	err = self.storeToken(ctxt.Tx(), ctxt.User(), githubResponse)
	if err != nil {
		return err
	}
	ctxt.EnqueueActivity(activities.UserConnectedGithub(ctxt.User()), nil)
	ctxt.W().WriteHeader(http.StatusCreated)
	return nil
}

func (self oAuthHandler) GithubCallbackSignin(ctxt RequestContext) error {
	githubResponse, err := acquireToken(ctxt)
	if err != nil {
		return err
	}

	ghUser, err := loadGithubUser(githubResponse["access_token"])
	if err != nil {
		return err
	}

	userStore := stores.NewDbUserStore(ctxt.Tx(), c)
	oauthTokenStore := stores.NewDbOAuthTokenStore(ctxt.Tx())
	token, _ := oauthTokenStore.FindByAccessToken(githubResponse["access_token"])
	var user *domain.User
	if token != nil {
		u, err := userStore.FindByUuid(token.UserUuid)
		if err != nil {
			ctxt.Log().Warn().Msgf("unable to find user %q indicated by token %q\n", token.UserUuid, token.Uuid)
		} else {
			user = u
		}
	}
	if user == nil {
		// Try to find a user that matches a verified GitHub email address
		for _, e := range ghUser.Emails {
			if !e.Verified {
				continue
			}
			u, _ := userStore.FindByEmailAddress(e.Email)
			if u != nil {
				// Used to display a helpful message explaining how to link existing users
				return NewOAuthError("github", "existing_unlinked_user", errors.New("No token found, but matching User found"))
			}
		}
		// Neither token nor matching User by email address found, create a new User
		u, err := self.newUser(ctxt, ghUser, githubResponse)
		if err != nil {
			return err
		}
		user = u
		ctxt.EnqueueActivity(activities.UserSignedUpViaGithub(user), nil)
	}
	ctxt.EnqueueActivity(activities.UserLoggedInViaGithub(user), nil)
	return login(ctxt, user)
}

func (self oAuthHandler) newUser(ctxt RequestContext, ghUser *github.GithubUser, ghResponse map[string]string) (*domain.User, error) {
	userStore := stores.NewDbUserStore(ctxt.Tx(), c)
	var email string
	// Find a verified email, prefer primary
	for _, e := range ghUser.Emails {
		if !e.Verified {
			continue
		}
		if e.Primary {
			email = e.Email
			break
		}
		if email == "" {
			email = e.Email
		}
	}
	if email == "" {
		return nil, NewOAuthError("github", "no_verified_email_found", errors.New("No verified email address found"))
	}
	user := &domain.User{
		Email:           email,
		GhUsername:      &ghUser.Login,
		Name:            ghUser.Name,
		UrlHost:         ctxt.R().Host,
		WithoutPassword: true,
	}
	uuid, err := userStore.Create(user)
	if err != nil {
		return nil, NewOAuthError("github", "unable_to_create_user", fmt.Errorf("Unable to create user: %s", err))
	}
	user, err = userStore.FindByUuid(uuid)
	if err != nil {
		return nil, NewOAuthError("github", "unable_to_load_user", fmt.Errorf("Unable to load user: %s", err))
	}
	err = self.storeToken(ctxt.Tx(), user, ghResponse)
	if err != nil {
		return nil, NewOAuthError("github", "unable_to_store_token", fmt.Errorf("Unable to store token: %s", err))
	}
	ctxt.Log().Info().Msgf("new github signup: email=%q name=%q\n", user.Name, user.Email)
	return user, nil
}

//
// Forward the request to Github, here's where we handover our client
// secret, and trade it in for a bearer token for the user's account
//
var acquireToken = func(ctxt RequestContext) (map[string]string, error) {

	provider := c.OAuthConfig().Providers["github"]

	// Guard against request replay, etc

	state := ctxt.R().FormValue("state")
	exists, err := ctxt.KeyValueStore().Exists(state)
	if err != nil {
		return nil, NewOAuthError("github", "unable_to_check_state", fmt.Errorf("Unable to check state: %s", err))
	}
	if !exists {
		return nil, NewOAuthError("github", "state_mismatch", fmt.Errorf("Unknown state %q", state))
	}

	client := &http.Client{}

	v := url.Values{}
	v.Add("code", ctxt.R().FormValue("code"))
	v.Add("client_id", provider.ClientId)
	v.Add("client_secret", provider.ClientSecret)

	url := fmt.Sprintf("https://github.com/login/oauth/access_token?%s", v.Encode())
	requ, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, NewOAuthError("github", "communication", err)
	}

	requ.Header.Set("Accept", "application/json")
	resp, err := client.Do(requ)
	if err != nil {
		return nil, NewOAuthError("github", "communication", err)
	}
	defer resp.Body.Close()

	if got, want := resp.StatusCode, 200; got != want {
		ctxt.Log().Debug().Msgf("got status %d from github, but expected %d :-(\n", got, want)
		b, _ := ioutil.ReadAll(resp.Body)
		ctxt.Log().Debug().Msgf("the github response body was: %s\n", string(b))
		return nil, NewOAuthError("github", "bad_status", fmt.Errorf("response.StatusCode = %d; want %d", got, want))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, NewOAuthError("github", "unreadable", err)
	}

	//
	// Unmarshal the response from Github, if we get this far we can expect the data returned
	// to look something like {"data": {"access_token":"...", "token_type":"...", "scope":"..."}, "status":200, ....})
	//
	var githubResponse map[string]string

	err = json.Unmarshal(body, &githubResponse)
	if err != nil {
		return nil, NewOAuthError("github", "unprocessable", err)
	}

	return githubResponse, nil
}

var loadGithubUser = func(accessToken string) (*github.GithubUser, error) {
	return github.LoadGithubUser(accessToken)
}

func (self oAuthHandler) storeToken(tx *sqlx.Tx, user *domain.User, githubResponse map[string]string) error {
	store := stores.NewDbOAuthTokenStore(tx)
	existingToken, err := store.FindByUserUuid(user.Uuid)

	// If there is a token in the db, the user possibly authorized earlier,
	// removed the authorization from the provider, and is trying to authorize
	// again. In this case, the access token will be different.  In any case,
	// delete the existing token and create a new one
	if existingToken != nil {
		store.DeleteByUuid(existingToken.Uuid)
	}

	// Create the new token
	token := &domain.OAuthToken{
		Provider:    "github",
		UserUuid:    user.Uuid,
		Scope:       githubResponse["scope"],
		TokenType:   githubResponse["token_type"],
		AccessToken: githubResponse["access_token"],
	}

	_, err = store.Create(token)
	if err != nil {
		return err
	}

	return nil
}

func (self oAuthHandler) GithubDeAuthorize(ctxt RequestContext) error {

	provider := ctxt.Config().OAuthConfig().Providers["github"]

	store := stores.NewDbOAuthTokenStore(ctxt.Tx())

	token, err := store.FindByProviderAndUserUuid("github", ctxt.User().Uuid)
	if err != nil {
		return err
	}

	err = github.Deauth(token.AccessToken, provider.ClientId, provider.ClientSecret)
	if err != nil {
		return NewOAuthError("github", "deauthorize", err)
	}
	return store.DeleteByUuid(token.Uuid)
}

func (self oAuthHandler) GithubOrganizations(ctxt RequestContext) error {

	store := stores.NewDbOAuthTokenStore(ctxt.Tx())

	token, err := store.FindByProviderAndUserUuid("github", ctxt.User().Uuid)
	if err != nil {
		return err
	}

	body, err := github.OrganizationsForAuthenticatedUser(token.AccessToken)
	if err != nil {
		return NewOAuthError("github", "organizations_for_authenticated_user", err)
	}

	ctxt.W().Write(body)
	return nil
}

func (self oAuthHandler) GithubPing(ctxt RequestContext) error {

	store := stores.NewDbOAuthTokenStore(ctxt.Tx())

	response := struct {
		Status string `json:"status"`
	}{
		Status: "up",
	}

	token, err := store.FindByProviderAndUserUuid("github", ctxt.User().Uuid)
	if err != nil {
		response.Status = "down"
	} else {
		_, err = github.Ping(token.AccessToken)
		if err != nil {
			response.Status = "down"
		}
	}

	return json.NewEncoder(ctxt.W()).Encode(&response)
}

func (self oAuthHandler) GithubUser(ctxt RequestContext) error {

	store := stores.NewDbOAuthTokenStore(ctxt.Tx())

	token, err := store.FindByProviderAndUserUuid("github", ctxt.User().Uuid)
	if err != nil {
		return err
	}

	body, err := github.User(token.AccessToken)
	if err != nil {
		return NewOAuthError("github", "user", err)
	}

	ctxt.W().Write(body)

	return err

}

func (self oAuthHandler) GithubRepos(ctxt RequestContext) error {

	kind := mux.Vars(ctxt.R())["kind"]
	login := mux.Vars(ctxt.R())["login"]

	store := stores.NewDbOAuthTokenStore(ctxt.Tx())

	token, err := store.FindByProviderAndUserUuid("github", ctxt.User().Uuid)
	if err != nil {
		return err
	}

	var body []byte
	switch kind {
	case "users":
		body, err = github.RepositoriesForUsername(login, token.AccessToken)
	case "orgs":
		body, err = github.RepositoriesForOrg(login, token.AccessToken)
	default:
		err = fmt.Errorf("Kind must be users or orgs, but was %s", kind)
	}
	if err != nil {
		return NewOAuthError("github", "repos", err)
	}

	ctxt.W().Write(body)

	return err

}

func (self oAuthHandler) GithubDeployKey(ctxt RequestContext) error {

	repoUuid := mux.Vars(ctxt.R())["repoUuid"]

	tokenStore := stores.NewDbOAuthTokenStore(ctxt.Tx())
	token, err := tokenStore.FindByProviderAndUserUuid("github", ctxt.User().Uuid)
	if err != nil {
		return err
	}

	repoStore := stores.NewDbRepositoryStore(ctxt.Tx())
	repo, err := repoStore.FindByUuid(repoUuid)
	if err != nil {
		return err
	}

	projectStore := stores.NewDbProjectStore(ctxt.Tx())
	project, err := projectStore.FindByUuid(repo.ProjectUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(project); !allowed {
		return err
	}

	repositoryCredentialStore := stores.NewRepositoryCredentialStore(ctxt.SecretKeyValueStore(), ctxt.Tx())
	rc, err := repositoryCredentialStore.FindByRepositoryUuid(repo.Uuid)
	if err != nil {
		return NewOAuthError("github", "repo_for_deploy_key", err)
	}
	sshRc, err := domain.AsSshRepositoryCredential(rc)
	if err != nil {
		return NewOAuthError("github", "repo_for_deploy_key", err)
	}

	body, err := github.CreateDeployKey(repo.GithubLogin, repo.GithubRepo, sshRc.Name, sshRc.PublicKey, token.AccessToken)
	if err != nil {
		return NewOAuthError("github", "create_deploy_key", err)
	}

	ctxt.W().Write(body)
	return nil

}
