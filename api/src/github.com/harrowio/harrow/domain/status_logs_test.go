package domain

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/harrowio/harrow/clock"
)

var (
	ImplementsScanner sql.Scanner   = NewStatusLogs()
	ImplementsValuer  driver.Valuer = NewStatusLogs()
)

func TestStatusLogs_Log_appendsAnEntryToTheLogs(t *testing.T) {
	logs := NewStatusLogs()
	entry := NewStatusLogEntry("git.clone-started")
	entry.OccurredOn = time.Now()

	n := logs.Len()
	logs.Log(entry)

	if got, want := logs.Len(), n+1; got != want {
		t.Errorf(`logs.Len() = %v; want %v`, got, want)
	}
}

func TestStatusLogs_Newest_returnsTheNewestEntry_ifMessagesArriveOutOfOrder(t *testing.T) {
	now := time.Now()
	logs := NewStatusLogs()
	entries := []*StatusLogEntry{
		NewStatusLogEntry("git.clone-finished"),
		NewStatusLogEntry("git.clone-started"),
	}

	entries[0].OccurredOn = now
	entries[1].OccurredOn = now.Add(-1 * time.Minute)

	for _, entry := range entries {
		logs.Log(entry)
	}

	newest := logs.Newest()
	if got := newest; got == nil {
		t.Fatalf("newest is nil")
	}

	if got, want := newest.OccurredOn, now; got != want {
		t.Errorf(`newest.OccurredOn = %v; want %v`, got, want)
	}
}

func TestStatusLogs_Newest_returnsTheNewestEntry_ifMessagesArriveInOrder(t *testing.T) {
	now := time.Now()
	logs := NewStatusLogs()
	entries := []*StatusLogEntry{
		NewStatusLogEntry("git.clone-started"),
		NewStatusLogEntry("git.clone-finished"),
	}

	entries[0].OccurredOn = now.Add(-1 * time.Minute)
	entries[1].OccurredOn = now

	for _, entry := range entries {
		logs.Log(entry)
	}

	newest := logs.Newest()
	if got := newest; got == nil {
		t.Fatalf("newest is nil")
	}

	if got, want := newest.OccurredOn, now; got != want {
		t.Errorf(`newest.OccurredOn = %v; want %v`, got, want)
	}
}

// Sorting entries every time we call newest would do a lot of
// unnecessary work in the default case of messages arriving in order.
func TestStatusLogs_Newest_DoesNotSortEntries(t *testing.T) {
	now := time.Now()
	logs := NewStatusLogs()
	entries := []*StatusLogEntry{
		NewStatusLogEntry("git.clone-started"),
		NewStatusLogEntry("git.clone-finished"),
	}

	entries[0].OccurredOn = now.Add(-1 * time.Minute)
	entries[1].OccurredOn = now

	for _, entry := range entries {
		logs.Log(entry)
	}

	logs.Swap(0, 1)
	logs.Newest()
	logs.Swap(0, 1)

	if got, want := logs.Entries[0].Type, "git.clone-started"; got != want {
		t.Errorf(`logs.Entries[0].Type = %v; want %v`, got, want)
	}
}

func TestStatusLogs_ScanIsTheInverseOfValue(t *testing.T) {
	now := time.Now()
	logs := NewStatusLogs()
	entries := []*StatusLogEntry{
		NewStatusLogEntry("git.clone-started"),
		NewStatusLogEntry("git.clone-finished"),
	}

	entries[0].OccurredOn = now.Add(-1 * time.Minute)
	entries[1].OccurredOn = now

	for _, entry := range entries {
		logs.Log(entry)
	}

	serialized, err := logs.Value()
	if err != nil {
		t.Fatal(err)
	}

	scanned := NewStatusLogs()
	if err := scanned.Scan(serialized); err != nil {
		t.Fatal(err)
	}

	if got, want := scanned.Len(), logs.Len(); got != want {
		t.Fatalf(`scanned.Len() = %v; want %v`, got, want)
	}

	if got, want := scanned.Entries[0].Type, logs.Entries[0].Type; got != want {
		t.Errorf(`scanned.Entries[0].Type = %v; want %v`, got, want)
	}

	if got, want := scanned.Entries[1].Type, logs.Entries[1].Type; got != want {
		t.Errorf(`scanned.Entries[1].Type = %v; want %v`, got, want)
	}
}

func TestStatusLogs_HandleEvent_logEntryForNewStatusEvent(t *testing.T) {
	now := time.Date(2015, 10, 7, 17, 55, 0, 0, time.UTC)
	Clock = clock.At(now)
	defer func() {
		Clock = clock.System
	}()

	logs := NewStatusLogs()
	payload := MapPayload{
		"event":   "status",
		"subject": "the-subject",
		"body":    "the-body",
		"type":    "git.clone-started",
	}
	logs.HandleEvent(payload)

	expectedEntry := NewStatusLogEntry("git.clone-started")
	expectedEntry.OccurredOn = now
	expectedEntry.Subject = "the-subject"
	expectedEntry.Body = "the-body"

	newest := logs.Newest()
	if got := newest; got == nil {
		t.Fatalf("newest is nil")
	}

	{
		got, _ := json.MarshalIndent(newest, "", "  ")
		want, _ := json.MarshalIndent(expectedEntry, "", "  ")
		if !bytes.Equal(got, want) {
			t.Errorf(`newest = %s; want %s`, got, want)
		}
	}
}

func TestStatusLogs_HandleEvent_OnlyLogsStatusEvents(t *testing.T) {
	logs := NewStatusLogs()
	payload := MapPayload{
		"event": "something-else",
	}
	n := logs.Len()
	logs.HandleEvent(payload)

	if got, want := logs.Len(), n; got != want {
		t.Errorf(`logs.Len() = %v; want %v`, got, want)
	}
}

func TestStatusLogs_HandleEvent_setsTimestampOnEntryBasedOnClock(t *testing.T) {
	now := time.Date(2015, 10, 7, 17, 55, 0, 0, time.UTC)
	Clock = clock.At(now)
	defer func() {
		Clock = clock.System
	}()

	logs := NewStatusLogs()
	payload := MapPayload{
		"event":   "status",
		"subject": "the-subject",
		"body":    "the-body",
		"type":    "git.clone-started",
	}
	logs.HandleEvent(payload)
	newest := logs.Newest()

	if got, want := newest.OccurredOn, now; got != want {
		t.Errorf(`newest.OccurredOn = %v; want %v`, got, want)
	}
}
