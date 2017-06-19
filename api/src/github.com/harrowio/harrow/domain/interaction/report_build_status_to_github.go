package interaction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/git"
	"github.com/harrowio/harrow/logger"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type OAuthTokenFinder interface {
	FindOAuthTokenForRepository(repositoryUUID string) (string, error)
}

type RepositoryFinder interface {
	FindRepository(repositoryUUID string) (*domain.Repository, error)
}

type ReportBuildStatusToGitHub struct {
	httpClient   HTTPClient
	tokens       OAuthTokenFinder
	repositories RepositoryFinder
	log          logger.Logger

	CommitRef      string
	RepositoryUUID string
	State          string
	TargetURL      string
}

func NewReportBuildStatusToGitHub(repositoryUUID string, commitRef string, state string, targetURL string, log logger.Logger, httpClient HTTPClient, tokens OAuthTokenFinder, repositories RepositoryFinder) *ReportBuildStatusToGitHub {
	return &ReportBuildStatusToGitHub{
		httpClient:   httpClient,
		tokens:       tokens,
		repositories: repositories,
		log:          log,

		RepositoryUUID: repositoryUUID,
		CommitRef:      commitRef,
		State:          state,
		TargetURL:      targetURL,
	}
}

func (self *ReportBuildStatusToGitHub) Execute() error {
	oAuthToken, err := self.tokens.FindOAuthTokenForRepository(self.RepositoryUUID)
	if err != nil {
		if domain.IsNotFound(err) {
			return fmt.Errorf("No OAuth token for repository %s", self.RepositoryUUID)
		}
		return err
	}

	repository, err := self.repositories.FindRepository(self.RepositoryUUID)
	if err != nil {
		if domain.IsNotFound(err) {
			return fmt.Errorf("No such repository: %s", self.RepositoryUUID)
		}
		return err
	}

	request, err := self.createStatusRequest(oAuthToken, repository)
	if err != nil {
		return err
	}
	self.log.Info().Msgf("repository %s %s %s", self.RepositoryUUID, request.Method, request.URL.Path)
	response, err := self.httpClient.Do(request)
	if err != nil {
		return err
	}

	body, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode >= 400 {
		return fmt.Errorf("Error %d (%s):\n%s\n", response.StatusCode, response.Status, body)
	}

	return nil
}

func (self *ReportBuildStatusToGitHub) createStatusRequest(oAuthToken string, repository *domain.Repository) (*http.Request, error) {
	gitURL, err := git.Parse(repository.Url)
	if err != nil {
		return nil, err
	}

	if gitURL.Host != "github.com" {
		return nil, fmt.Errorf("%s not hosted on github.com", gitURL)
	}

	segments := strings.Split(gitURL.Path, "/")
	if len(segments) < 3 {
		return nil, fmt.Errorf("%s expected to contain at least three segments", gitURL.Path)
	}
	gitHubUsername := segments[1]
	gitHubRepository := strings.TrimSuffix(segments[2], ".git")

	gitHubURL := &url.URL{
		Scheme: "https",
		Host:   "api.github.com",
		Path: fmt.Sprintf("repos/%s/%s/statuses/%s",
			gitHubUsername,
			gitHubRepository,
			self.CommitRef,
		),
	}

	payloadData := map[string]string{
		"state":   self.State,
		"context": "Harrow",
	}
	if self.TargetURL != "" {
		_, err := url.Parse(self.TargetURL)
		if err != nil {
			return nil, fmt.Errorf("invalid TargetURL: %q: %s", self.TargetURL, err)
		}
		payloadData["target_url"] = self.TargetURL
	}

	payload, err := json.Marshal(payloadData)
	if err != nil {
		return nil, err
	}

	requestBody := bytes.NewBuffer(payload)
	request, err := http.NewRequest("POST", gitHubURL.String(), requestBody)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", fmt.Sprintf("token %s", oAuthToken))

	return request, nil
}
