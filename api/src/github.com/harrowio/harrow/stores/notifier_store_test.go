package stores

import "testing"

func Test_normalizeNotifierTypeName(t *testing.T) {
	testcases := []struct {
		Input    string
		Expected string
	}{
		{"slackNotifiers", "slack_notifiers"},
		{"fooNotifiers", "foo_notifiers"},
		{"singularNotifier", "singular_notifiers"},
		{"don't touch this", "don't touch this"},
	}

	for _, testcase := range testcases {
		if got, want := normalizeNotifierTypeName(testcase.Input), testcase.Expected; got != want {
			t.Errorf(`normalizeNotifierTypeName(%q) = %v; want %v`, testcase.Input, got, want)
		}
	}
}
