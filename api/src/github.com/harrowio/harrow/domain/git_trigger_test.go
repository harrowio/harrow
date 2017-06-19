package domain

import "testing"

func TestGitTrigger_Match_matchesBasedOnChangeTypeRepositoryUuidAndMatchRef(t *testing.T) {
	repositoryUuid := "a54eb784-22b7-4928-ae7f-fc3842152e02"
	otherRepositoryUuid := "dc6637eb-81a6-4579-acfd-3e226679d379"
	creatorUuid := "5b0662fb-20f7-44a2-990b-90415568f66e"
	changedRef := &ChangedRepositoryRef{
		RepositoryUuid: repositoryUuid,
		Symbolic:       "refs/heads/master",
		OldHash:        "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15",
		NewHash:        "e242ed3bffccdf271b7fbaf34ed72d089537b42f",
	}
	activity := &Activity{
		Name:       "repository-metadata.ref-changed",
		Payload:    changedRef,
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
	}

	testcases := []struct {
		repositoryUuid string
		matchRef       string
		changeType     string
		shouldMatch    bool
	}{
		{"", ".", "add", false},
		{"", ".", "change", true},
		{"", ".", "remove", false},
		{repositoryUuid, ".", "change", true},
		{otherRepositoryUuid, ".", "change", false},
		{"", "master", "change", true},
		{"", "m.ster", "change", true},
		{"", "unicorns", "change", false},
	}

	for _, testcase := range testcases {
		trigger := NewGitTrigger("test", creatorUuid).
			ForChangeType(testcase.changeType).
			MatchingRef(testcase.matchRef).
			InRepository(testcase.repositoryUuid)

		if got, want := trigger.Match(activity), testcase.shouldMatch; got != want {
			t.Logf("repository=%q matchRef=%q changeType=%q",
				testcase.repositoryUuid,
				testcase.matchRef,
				testcase.changeType,
			)

			t.Errorf(`trigger.Match(activity) = %v; want %v`, got, want)
		}
	}

}
