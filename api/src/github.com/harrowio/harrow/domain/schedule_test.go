package domain

import (
	"testing"
	"time"

	"github.com/harrowio/harrow/clock"
)

func Test_oneTimeSchedule_NextOperation_InTheFuture(t *testing.T) {
	now := time.Now().UTC()
	ref := now.Add(24 * time.Hour)
	Clock = clock.At(now)
	defer func() { Clock = clock.System }()

	timespec := "now + 1 day"
	s, err := NewSchedulable(&Schedule{Timespec: &timespec, TimezoneName: "UTC"})
	if err != nil {
		t.Fatalf("NewSchedulable: %s\n", err)
	}

	err = s.Validate()
	if err != nil {
		t.Fatalf("Expected schedule to be valid, error: %s\n", err)
	}

	nextTime, err := s.NextOperation(nil)
	if err != nil {
		t.Fatalf("Expected to get a next operation. Got error: %s\n", err)
	}
	nextStr := nextTime.Format(time.RFC3339)
	refStr := ref.Format(time.RFC3339)
	if refStr != nextStr {
		t.Fatalf("Expected %s to match reference time %s", nextStr, refStr)
	}
}

func Test_oneTimeSchedule_NextOperation_InThePast(t *testing.T) {
	// the beginning of time, definitely in the past
	timespec := "1970-01-01"
	_, err := NewSchedulable(&Schedule{Timespec: &timespec, TimezoneName: "UTC"})
	if err == nil {
		t.Fatalf("Expected schedule to be invalid.\n")
	}
}

func Test_recurringSchedule_NextOperation(t *testing.T) {
	cronspec := "*/5 * * * *"
	s, err := NewSchedulable(&Schedule{Cronspec: &cronspec, TimezoneName: "UTC"})
	if err != nil {
		t.Fatalf("NewSchedulable: %s\n", err)
	}

	err = s.Validate()
	if err != nil {
		t.Fatalf("Expected schedule to be valid, got: %s\n", err)
	}

	now, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
	refTime, _ := time.Parse(time.RFC3339, "2006-01-02T15:05:00+07:00")
	op, err := s.NextOperation(&now)
	if err != nil {
		t.Fatalf("Expected next operation to be successful, got: %s\n", err)
	}

	if !op.Equal(refTime) {
		t.Fatalf("Expected %s to match reference time %s\n", op, refTime)
	}
}

func Test_oneTimeSchedule_OperationParameters_setsReasonToUserIfItIsScheduledForNow(t *testing.T) {
	timespec := "now"
	s, err := NewSchedulable(&Schedule{Timespec: &timespec, TimezoneName: "UTC"})
	if got, want := err, (error)(nil); got != want {
		t.Fatalf(`err = %v; want %v`, got, want)
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	params := s.OperationParameters()
	if got, want := params.Reason, OperationTriggeredByUser; got != want {
		t.Errorf(`params.Reason = %v; want %v`, got, want)
	}
}

func Test_oneTimeSchedule_OperationParameters_doesNotOverrideExistingReason(t *testing.T) {
	timespec := "now"
	existingParams := NewOperationParameters()
	existingParams.Reason = OperationTriggeredByWebhook
	s, err := NewSchedulable(&Schedule{
		Timespec:     &timespec,
		TimezoneName: "UTC",
		Parameters:   existingParams,
	})

	if got, want := err, (error)(nil); got != want {
		t.Fatalf(`err = %v; want %v`, got, want)
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	params := s.OperationParameters()
	if got, want := params.Reason, OperationTriggeredByWebhook; got != want {
		t.Errorf(`params.Reason = %v; want %v`, got, want)
	}
}

func Test_oneTimeSchedule_OperationParameters_setsReasonToScheduledIfItIsNotScheduledForNow(t *testing.T) {
	timespec := "now + 1 minute"
	s, err := NewSchedulable(&Schedule{Timespec: &timespec, TimezoneName: "UTC"})
	if got, want := err, (error)(nil); got != want {
		t.Fatalf(`err = %v; want %v`, got, want)
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
	params := s.OperationParameters()
	if got, want := params.Reason, OperationTriggeredBySchedule; got != want {
		t.Errorf(`params.Reason = %v; want %v`, got, want)
	}

}

func Test_recurringSchedule_OperationParameters_setsReasonToScheduled(t *testing.T) {
	cronspec := "*/5 * * * *"
	s, err := NewSchedulable(&Schedule{Cronspec: &cronspec, TimezoneName: "UTC"})
	if got, want := err, (error)(nil); got != want {
		t.Fatalf(`err = %v; want %v`, got, want)
	}
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}

	params := s.OperationParameters()
	if got, want := params.Reason, OperationTriggeredBySchedule; got != want {
		t.Errorf(`params.Reason = %v; want %v`, got, want)
	}
}
