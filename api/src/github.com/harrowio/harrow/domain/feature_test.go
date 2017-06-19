package domain

import (
	"testing"

	"github.com/harrowio/harrow/config"
)

func TestNewFeaturesFromConfig_initializes_list_of_features_via_reflection(t *testing.T) {
	features := NewFeaturesFromConfig(config.Features{
		OAuthGithubImportRepository: true,
		OAuthGithubAuthentication:   false,
	})

	if got, want := len(features) >= 2, true; got != want {
		t.Errorf(`len(features) >= 2 = %v; want %v`, got, want)
	}

	if got, want := features[0].Name, "oauth.github.import-repository"; got != want {
		t.Errorf(`features[0].Name = %v; want %v`, got, want)
	}

	if got, want := features[0].Enabled, true; got != want {
		t.Errorf(`features[0].Enabled = %v; want %v`, got, want)
	}

	if got, want := features[1].Name, "oauth.github.authentication"; got != want {
		t.Errorf(`features[1].Name = %v; want %v`, got, want)
	}

	if got, want := features[1].Enabled, false; got != want {
		t.Errorf(`features[1].Enabled = %v; want %v`, got, want)
	}

}
