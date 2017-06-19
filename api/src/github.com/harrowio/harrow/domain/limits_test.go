package domain

import "testing"

func TestLimits_Exceeded_returns_false_if_the_trial_is_still_running(t *testing.T) {
	limits := &Limits{
		TrialDaysLeft: 14,
		TrialEnabled:  true,
		Plan: &LimitsComparedToPlan{
			UsersExceedingLimit:    1,
			ProjectsExceedingLimit: 0,
		},
	}

	if got, want := limits.Exceeded(), false; got != want {
		t.Errorf(`limits.Exceeded() = %v; want %v`, got, want)
	}

}

func TestLimits_Exceeded_returns_false_if_there_are_no_projects_or_users_exceeding_the_limit(t *testing.T) {
	limits := &Limits{
		Plan: &LimitsComparedToPlan{
			UsersExceedingLimit:    0,
			ProjectsExceedingLimit: 0,
		},
	}

	if got, want := limits.Exceeded(), false; got != want {
		t.Errorf(`limits.Exceeded() = %v; want %v`, got, want)
	}
}

func TestLimits_Exceeded_returns_true_if_there_are_users_exceeding_the_limit(t *testing.T) {
	limits := &Limits{
		Plan: &LimitsComparedToPlan{
			UsersExceedingLimit:    1,
			ProjectsExceedingLimit: 0,
		},
	}

	if got, want := limits.Exceeded(), true; got != want {
		t.Errorf(`limits.Exceeded() = %v; want %v`, got, want)
	}
}

func TestLimits_Exceeded_returns_true_if_there_are_projects_exceeding_the_limit(t *testing.T) {
	limits := &Limits{
		Plan: &LimitsComparedToPlan{
			UsersExceedingLimit:    0,
			ProjectsExceedingLimit: 1,
		},
	}

	if got, want := limits.Exceeded(), true; got != want {
		t.Errorf(`limits.Exceeded() = %v; want %v`, got, want)
	}
}
