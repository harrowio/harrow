package domain

import (
	"strings"
	"testing"
)

func Test_gravatarUrl_ignoresCaseOfInput(t *testing.T) {
	upper, lower, mixed := "ABC", "abc", "aBc"

	expected := newGravatarUrl(lower).String()
	if actual := newGravatarUrl(upper).String(); actual != expected {
		t.Errorf("Expected %q to yield %q", upper, expected)
	}
	if actual := newGravatarUrl(mixed).String(); actual != expected {
		t.Errorf("Expected %q to yield %q", mixed, expected)
	}
}

func Test_gravatarUrl_usesHTTPS(t *testing.T) {
	url := newGravatarUrl("vagrant@localhost").String()
	prefix := "https://secure.gravatar.com"
	if !strings.HasPrefix(url, prefix) {
		t.Fatalf("Expected %q to start with %q", url, prefix)
	}
}
