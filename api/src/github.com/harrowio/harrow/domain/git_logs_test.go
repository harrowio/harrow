package domain

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestGitLogs_HandleEvent_unmarshalsEntryAsJSON(t *testing.T) {
	logs := NewGitLogs()
	repositoryUuid := "5bdf462d-72ee-4781-bff4-1a4b587ad15b"
	entry := &GitLogEntry{
		Author:      "John Doe",
		AuthorEmail: "john.doe@example.com",
		AuthorDate:  ISODate(time.Now()),
		Parents:     []string{"f1d2d2f924e986ac86fdf7b36c94bcdf32beec15"},
		Subject:     "Fix tests",
		Body:        "Tests were broken because of FOO",
		Commit:      "e242ed3bffccdf271b7fbaf34ed72d089537b42f",
	}
	entryAsJson, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	payload := MapPayload{
		"event":      "git-log",
		"repository": repositoryUuid,
		"entry":      string(entryAsJson),
	}

	logs.HandleEvent(payload)

	if got := logs.Repositories[repositoryUuid]; got == nil {
		t.Fatalf("logs.Repositories[repositoryUuid] is nil")
	}

	logEntries := logs.Repositories[repositoryUuid]
	if got := logEntries; got == nil {
		t.Fatalf("logEntries is nil")
	}

	logEntry := logEntries[0]
	{
		got, _ := json.MarshalIndent(logEntry, "", "  ")
		want, _ := json.MarshalIndent(entry, "", "  ")
		if !bytes.Equal(got, want) {
			t.Errorf(`logEntry = %s; want %s`, got, want)
		}
	}
}
