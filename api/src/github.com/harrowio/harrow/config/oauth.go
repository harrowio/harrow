package config

import (
	"os"
	"strings"
)

type OAuthConfig struct {
	Providers map[string]OAuthProviderConfig
}

type OAuthProviderConfig struct {
	ClientId     string
	ClientSecret string
	ProviderUrl  string
	RedirectUri  string
	Scope        []string
}

func (c Config) OAuthConfig() OAuthConfig {
	return OAuthConfig{
		Providers: map[string]OAuthProviderConfig{
			"github": OAuthProviderConfig{
				ClientId:     os.Getenv("HAR_OAUTH_GITHUB_CLIENT_ID"),
				ClientSecret: os.Getenv("HAR_OAUTH_GITHUB_CLIENT_SECRET"),
				ProviderUrl:  getEnvWithDefault("HAR_OAUTH_GITHUB_PROVIDER_URL", "https://test.tld/#/a/github/callback/"),
				RedirectUri:  getEnvWithDefault("HAR_OAUTH_GITHUB_REDIRECT_URI", "https://test.tld/#/a/github/callback/%s"),
				Scope:        strings.Split(os.Getenv("HAR_OAUTH_GITHUB_SCOPE"), ","),
			},
		},
	}
}
