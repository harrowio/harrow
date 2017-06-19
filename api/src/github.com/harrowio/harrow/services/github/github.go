package github

import (
	"github.com/harrowio/harrow/config"

	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	c      *config.Config
	Accept = "application/vnd.github.v3+json"
)

var (
	ErrCommunication = errors.New(`{"errors":["github.communication"]}`)
	ErrBadStatus     = errors.New(`{"errors":["github.bad_status"]}`)
	ErrUnprocessable = errors.New(`{"errors":["github.unprocessable"]}`)
)

func init() {
	c = config.GetConfig()
}

func Ping(token string) ([]byte, error) {

	// see: https://developer.github.com/v3/rate_limit/
	// "Note: Accessing this endpoint does not count against your rate limit."
	req, err := http.NewRequest("GET", "https://api.github.com/rate_limit", nil)
	if err != nil {
		panic(err)
	}

	return makeRequest(req, token)
}

func Deauth(token, clientId, clientSecret string) error {

	url := fmt.Sprintf("https://api.github.com/applications/%s/tokens/%s", clientId, token)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	// for this request, we need to authenticate with the application clientId
	// and client secret.
	// see: https://developer.github.com/v3/oauth_authorizations/#revoke-an-authorization-for-an-application
	req.SetBasicAuth(clientId, clientSecret)
	req.Header.Set("Accept", Accept)

	resp, err := client.Do(req)
	if err != nil {
		return ErrCommunication
	}
	resp.Body.Close()

	if resp.StatusCode != 204 {
		return ErrBadStatus
	}

	return nil
}

func User(token string) ([]byte, error) {

	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		panic(err)
	}

	return makeRequest(req, token)
}

func Emails(token string) ([]byte, error) {

	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		panic(err)
	}

	return makeRequest(req, token)
}

type GithubUser struct {
	Name   string
	Login  string
	Emails []GithubEmail
}

type GithubEmail struct {
	Email    string
	Primary  bool
	Verified bool
}

func LoadGithubUser(token string) (*GithubUser, error) {
	body, err := User(token)
	if err != nil {
		return nil, err
	}

	user := new(GithubUser)
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, ErrUnprocessable
	}

	user.Emails = make([]GithubEmail, 0)
	body, err = Emails(token)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &user.Emails)

	return user, nil
}

type createKeyReq struct {
	Title string `json:"title"`
	Key   string `json:"key"`
}

func CreateDeployKey(username, repo, title, key, token string) ([]byte, error) {

	j, err := json.Marshal(&createKeyReq{Title: title, Key: key})
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/keys", username, repo)
	req, err := http.NewRequest("POST", url, bytes.NewReader(j))
	if err != nil {
		return nil, err
	}
	return makeRequest(req, token)

}

func OrganizationsForAuthenticatedUser(token string) ([]byte, error) {

	req, err := http.NewRequest("GET", "https://api.github.com/user/orgs", nil)
	if err != nil {
		panic(err)
	}
	return makeRequest(req, token)

}

func RepositoriesForUsername(username, token string) ([]byte, error) {

	url := fmt.Sprintf("https://api.github.com/users/%s/repos", username)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	return makeRequest(req, token)
}

func RepositoriesForOrg(org, token string) ([]byte, error) {

	url := fmt.Sprintf("https://api.github.com/orgs/%s/repos", org)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	return makeRequest(req, token)
}

func makeRequest(req *http.Request, oauthToken string) ([]byte, error) {

	client := &http.Client{}

	req.Header.Set("Authorization", "token "+oauthToken)
	req.Header.Set("Accept", Accept)

	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrCommunication
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrCommunication
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, ErrBadStatus
	}

	return body, nil

}
