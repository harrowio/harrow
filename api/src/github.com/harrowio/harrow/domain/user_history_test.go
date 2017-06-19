package domain_test

import (
	"testing"
	"time"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
)

func TestUserHistory_HandleActivity_counts_how_often_the_user_requested_the_verification_today(t *testing.T) {
	now := time.Date(2016, 3, 23, 16, 51, 7, 0, time.UTC)
	userUuid := "5dd3b99a-1243-491c-8713-6efbf42138c9"
	user := &domain.User{Uuid: userUuid}
	history := domain.NewUserHistory(userUuid, now)

	yesterday := activities.UserRequestedVerificationEmail(user)
	yesterday.OccurredOn = now.Add(-1 * 24 * time.Hour)
	today := activities.UserRequestedVerificationEmail(user)
	today.OccurredOn = now

	happened := []*domain.Activity{yesterday, today, today}

	for _, activity := range happened {
		activity.ContextUserUuid = &userUuid
		if err := history.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	if got, want := history.VerificationEmailsRequestedToday(), 2; got != want {
		t.Errorf(`history.VerificationEmailsRequestedToday() = %v; want %v`, got, want)
	}
}

func TestUserHistory_HandleActivity_only_processes_activities_for_this_user(t *testing.T) {
	now := time.Date(2016, 3, 23, 16, 51, 7, 0, time.UTC)
	userUuid := "5dd3b99a-1243-491c-8713-6efbf42138c9"
	otherUserUuid := "bc8b61a1-38bd-4b9b-8f18-c15585c405aa"
	user := &domain.User{Uuid: userUuid}
	history := domain.NewUserHistory(userUuid, now)

	thisUser := activities.UserRequestedVerificationEmail(user)
	thisUser.OccurredOn = now
	thisUser.ContextUserUuid = &userUuid
	otherUser := activities.UserRequestedVerificationEmail(&domain.User{
		Uuid: otherUserUuid,
	})
	otherUser.OccurredOn = now
	otherUser.ContextUserUuid = &otherUserUuid

	happened := []*domain.Activity{otherUser, thisUser}

	for _, activity := range happened {
		if err := history.HandleActivity(activity); err != nil {
			t.Fatal(err)
		}
	}

	if got, want := history.VerificationEmailsRequestedToday(), 1; got != want {
		t.Errorf(`history.VerificationEmailsRequestedToday() = %v; want %v`, got, want)

	}
}

func TestUserHistory_HandleActivity_reports_user_as_active_for_a_given_day(t *testing.T) {
	now := time.Date(2016, 3, 23, 16, 51, 7, 0, time.UTC)
	userUuid := "5dd3b99a-1243-491c-8713-6efbf42138c9"
	user := &domain.User{Uuid: userUuid}
	history := domain.NewUserHistory(userUuid, now)

	yesterday := activities.UserReportedAsActive(user)
	yesterday.OccurredOn = now.Add(-1 * 24 * time.Hour)
	yesterday.ContextUserUuid = &userUuid
	today := activities.UserReportedAsActive(user)
	today.OccurredOn = now
	today.ContextUserUuid = &userUuid

	if err := history.HandleActivity(yesterday); err != nil {
		t.Fatal(err)
	}

	if got, want := history.IsActive(), false; got != want {
		t.Errorf(`history.IsActive() = %v; want %v`, got, want)
	}

	if err := history.HandleActivity(today); err != nil {
		t.Fatal(err)
	}

	if got, want := history.IsActive(), true; got != want {
		t.Errorf(`history.IsActive() = %v; want %v`, got, want)
	}
}

func TestUserHistory_Segment_returns_cold_if_user_was_never_active(t *testing.T) {
	now := time.Date(2016, 3, 23, 16, 51, 7, 0, time.UTC)
	userUuid := "5dd3b99a-1243-491c-8713-6efbf42138c9"
	history := domain.NewUserHistory(userUuid, now)

	if got, want := history.Segment(), domain.SegmentCold; got != want {
		t.Errorf(`history.Segment() = %v; want %v`, got, want)
	}
}

func TestUserHistory_Segment_returns_limbo_if_user_was_active_at_least_twice(t *testing.T) {
	now := time.Date(2016, 3, 23, 16, 51, 7, 0, time.UTC)
	userUuid := "5dd3b99a-1243-491c-8713-6efbf42138c9"
	user := &domain.User{Uuid: userUuid}
	history := domain.NewUserHistory(userUuid, now)

	beforeYesterday := activities.UserReportedAsActive(user)
	beforeYesterday.OccurredOn = now.Add(-2 * 24 * time.Hour)
	beforeYesterday.ContextUserUuid = &userUuid
	today := activities.UserReportedAsActive(user)
	today.OccurredOn = now
	today.ContextUserUuid = &userUuid

	if err := history.HandleActivity(beforeYesterday); err != nil {
		t.Fatal(err)
	}
	if err := history.HandleActivity(today); err != nil {
		t.Fatal(err)
	}

	if got, want := history.Segment(), domain.SegmentLimbo; got != want {
		t.Errorf(`history.Segment() = %v; want %v`, got, want)
	}
}

func TestUserHistory_Segment_returns_limbo_if_user_verified_email_address(t *testing.T) {
	now := time.Date(2016, 3, 23, 16, 51, 7, 0, time.UTC)
	userUuid := "5dd3b99a-1243-491c-8713-6efbf42138c9"
	user := &domain.User{Uuid: userUuid}
	history := domain.NewUserHistory(userUuid, now)

	yesterday := activities.UserEmailVerified(user)
	yesterday.OccurredOn = now.Add(-1 * 24 * time.Hour)
	yesterday.ContextUserUuid = &userUuid

	if err := history.HandleActivity(yesterday); err != nil {
		t.Fatal(err)
	}

	if got, want := history.Segment(), domain.SegmentLimbo; got != want {
		t.Errorf(`history.Segment() = %v; want %v`, got, want)
	}
}

func TestUserHistory_Segment_returns_returned_if_user_was_active_today_and_yesterday(t *testing.T) {
	now := time.Date(2016, 3, 23, 16, 51, 7, 0, time.UTC)
	userUuid := "5dd3b99a-1243-491c-8713-6efbf42138c9"
	user := &domain.User{Uuid: userUuid}
	history := domain.NewUserHistory(userUuid, now)

	yesterday := activities.UserReportedAsActive(user)
	yesterday.OccurredOn = now.Add(-1 * 24 * time.Hour)
	yesterday.ContextUserUuid = &userUuid
	today := activities.UserReportedAsActive(user)
	today.OccurredOn = now
	today.ContextUserUuid = &userUuid

	if err := history.HandleActivity(yesterday); err != nil {
		t.Fatal(err)
	}
	if err := history.HandleActivity(today); err != nil {
		t.Fatal(err)
	}

	if got, want := history.Segment(), domain.SegmentReturned; got != want {
		t.Errorf(`history.Segment() = %v; want %v`, got, want)
	}
}

func TestUserHistory_Segment_returns_failed_if_user_added_a_job_but_never_ran_it(t *testing.T) {
	now := time.Date(2016, 3, 23, 16, 51, 7, 0, time.UTC)
	userUuid := "5dd3b99a-1243-491c-8713-6efbf42138c9"
	history := domain.NewUserHistory(userUuid, now)

	jobUuid := "9d904b42-369b-4946-b633-fec8f07334c8"
	job := &domain.Job{Uuid: jobUuid}

	jobAdded := activities.JobAdded(job)
	jobAdded.ContextUserUuid = &userUuid

	if err := history.HandleActivity(jobAdded); err != nil {
		t.Fatal(err)
	}

	if got, want := history.Segment(), domain.SegmentFailed; got != want {
		t.Errorf(`history.Segment() = %v; want %v`, got, want)
	}

}

func TestUserHistory_Segment_returns_loner_if_user_added_a_job_and_ran_it_but_never_invited_collaborators(t *testing.T) {
	now := time.Date(2016, 3, 23, 16, 51, 7, 0, time.UTC)
	userUuid := "5dd3b99a-1243-491c-8713-6efbf42138c9"
	history := domain.NewUserHistory(userUuid, now)

	jobUuid := "2ec72b7d-ac5d-4086-b5c0-264dcbd8f41b"
	job := &domain.Job{
		Uuid:     jobUuid,
		TaskUuid: "643d6821-00a5-4326-b746-d0dafbb937c8",
	}
	schedule := &domain.Schedule{
		Uuid:    "81fa2851-a9be-4707-ba3c-7282ec6cd98c",
		JobUuid: jobUuid,
	}
	jobAdded := activities.JobAdded(job)
	jobAdded.ContextUserUuid = &userUuid
	jobScheduled := activities.JobScheduled(schedule, "test")
	jobScheduled.ContextUserUuid = &userUuid

	if err := history.HandleActivity(jobAdded); err != nil {
		t.Fatal(err)
	}

	if err := history.HandleActivity(jobScheduled); err != nil {
		t.Fatal(err)
	}

	if got, want := history.Segment(), domain.SegmentLoner; got != want {
		t.Errorf(`history.Segment() = %v; want %v`, got, want)
	}
}

func TestUserHistory_Segment_returns_ideal_if_user_added_a_job_and_ran_it_and_invited_collaborators(t *testing.T) {
	now := time.Date(2016, 3, 23, 16, 51, 7, 0, time.UTC)
	userUuid := "5dd3b99a-1243-491c-8713-6efbf42138c9"
	history := domain.NewUserHistory(userUuid, now)

	jobUuid := "2ec72b7d-ac5d-4086-b5c0-264dcbd8f41b"
	job := &domain.Job{
		Uuid:     jobUuid,
		TaskUuid: "643d6821-00a5-4326-b746-d0dafbb937c8",
	}
	schedule := &domain.Schedule{
		Uuid:    "81fa2851-a9be-4707-ba3c-7282ec6cd98c",
		JobUuid: jobUuid,
	}

	jobAdded := activities.JobAdded(job)
	jobAdded.ContextUserUuid = &userUuid
	jobScheduled := activities.JobScheduled(schedule, "test")
	jobScheduled.ContextUserUuid = &userUuid
	invitationCreated := activities.InvitationCreated(&domain.Invitation{})
	invitationCreated.ContextUserUuid = &userUuid

	if err := history.HandleActivity(jobAdded); err != nil {
		t.Fatal(err)
	}

	if err := history.HandleActivity(jobScheduled); err != nil {
		t.Fatal(err)
	}

	if err := history.HandleActivity(invitationCreated); err != nil {
		t.Fatal(err)
	}

	if got, want := history.Segment(), domain.SegmentIdeal; got != want {
		t.Errorf(`history.Segment() = %v; want %v`, got, want)
	}
}

func TestUserHistory_PreviousSegment_returns_last_segment_reported_by_entered_segment(t *testing.T) {
	now := time.Date(2016, 5, 25, 12, 20, 13, 0, time.UTC)
	userUuid := "4ed2d649-a576-4f02-a5eb-e555450f3da6"
	history := domain.NewUserHistory(userUuid, now)
	happened := []*domain.Activity{
		activities.UserEnteredSegment(domain.SegmentFailed),
		activities.UserLeftSegment(domain.SegmentFailed),
		activities.UserEnteredSegment(domain.SegmentLoner),
	}

	for _, activity := range happened {
		activity.ContextUserUuid = &userUuid
		if err := history.HandleActivity(activity); err != nil {
			t.Fatalf("%s: %s\n", activity.Name, err)
		}
	}

	if got, want := history.PreviousSegment(), domain.SegmentLoner; got != want {
		t.Errorf(`history.PreviousSegment() = %v; want %v`, got, want)
	}
}

func TestUserHistory_PreviousSegment_returns_segment_limbo_if_user_has_not_entered_segment_before(t *testing.T) {
	now := time.Date(2016, 5, 25, 12, 20, 13, 0, time.UTC)
	userUuid := "4ed2d649-a576-4f02-a5eb-e555450f3da6"
	history := domain.NewUserHistory(userUuid, now)

	if got, want := history.PreviousSegment(), domain.SegmentLimbo; got != want {
		t.Errorf(`history.PreviousSegment() = %v; want %v`, got, want)
	}
}
