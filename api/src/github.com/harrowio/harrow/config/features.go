package config

import (
	"os"
	"strconv"
)

type Features struct {
	OAuthGithubImportRepository bool `json:"oauth.github.import-repository"`
	OAuthGithubAuthentication   bool `json:"oauth.github.authentication"`
	TrialEnabled                bool `json:"trial-period.enabled"`
	LimitsEnabled               bool `json:"limits.enabled"`
	GuestAccountEnabled         bool `json:"guest-account.enabled"`
	PublicProjectsEnabled       bool `json:"public-projects.enabled"`
}

func (c Config) FeaturesConfig() Features {
	var f func(string) bool = func(s string) bool {
		b, _ := strconv.ParseBool(os.Getenv(s))
		return b
	}
	return Features{
		OAuthGithubImportRepository: f("HAR_FEATURE_OAUTH_GITHUB_IMPORT_REPOSITORY"),
		OAuthGithubAuthentication:   f("HAR_FEATURE_OAUTH_GITHUB_AUTHENTICATION"),
		TrialEnabled:                f("HAR_FEATURE_TRIAL_ENABLED"),
		LimitsEnabled:               f("HAR_FEATURE_LIMITS_ENABLED"),
		GuestAccountEnabled:         f("HAR_FEATURE_GUEST_ACCOUNT_ENABLED"),
		PublicProjectsEnabled:       f("HAR_FEATURE_PUBLIC_PROJECTS_ENABLED"),
	}
}
