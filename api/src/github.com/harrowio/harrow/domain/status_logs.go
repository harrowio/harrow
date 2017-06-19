package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

type StatusLogs struct {
	Entries []*StatusLogEntry `json:"entries"`
}

func (self *StatusLogs) Len() int {
	return len(self.Entries)
}

func (self *StatusLogs) Swap(i, j int) {
	self.Entries[i], self.Entries[j] = self.Entries[j], self.Entries[i]
}

// Less compares status log entries by occurrence date, ascending.
func (self *StatusLogs) Less(i, j int) bool {
	return self.Entries[i].OccurredOn.Before(self.Entries[j].OccurredOn)
}

type StatusLogEntry struct {
	OccurredOn time.Time `json:"occurredOn"`
	Type       string    `json:"type"`
	Subject    string    `json:"subject"`
	Body       string    `json:"body"`
}

// NewStatusLogEntry returns a new empty status log entry of the given
// type.
func NewStatusLogEntry(entryType string) *StatusLogEntry {
	return &StatusLogEntry{
		Type: entryType,
	}
}

// NewStatusLogs returns an empty list of status logs.
func NewStatusLogs() *StatusLogs {
	return &StatusLogs{
		Entries: []*StatusLogEntry{},
	}
}

// Log adds entry to the end of this log and sorts entries by
// ocurrence date.
func (self *StatusLogs) Log(entry *StatusLogEntry) *StatusLogs {
	self.Entries = append(self.Entries, entry)
	if len(self.Entries) == 1 {
		return self
	}

	mostRecent := self.Entries[len(self.Entries)-2:]
	if mostRecent[0].OccurredOn.After(mostRecent[1].OccurredOn) {
		sort.Sort(self)
	}

	return self
}

// Newest returns the newest entry in this log going by the entries'
// occurrence date.
func (self *StatusLogs) Newest() *StatusLogEntry {
	return self.Entries[len(self.Entries)-1]
}

// Value serializes this log as a JSON array.
func (self *StatusLogs) Value() (driver.Value, error) {
	marshaled, err := json.Marshal(self.Entries)
	if err != nil {
		return nil, err
	}
	return driver.Value(marshaled), nil
}

// Scan deserializes value as a JSON array.
func (self *StatusLogs) Scan(data interface{}) error {
	if self == nil {
		*self = StatusLogs{}
	}

	src := []byte{}
	switch raw := data.(type) {
	case []byte:
		src = raw
	default:
		return fmt.Errorf("StatusLogs: cannot scan from %T", data)
	}

	return json.Unmarshal(src, &self.Entries)
}

// HandleEvent adds new entries to the logs
func (self *StatusLogs) HandleEvent(payload EventPayload) {
	if payload.Get("event") != "status" {
		return
	}
	entryType := payload.Get("type")
	newEntry := NewStatusLogEntry(entryType)
	newEntry.Subject = payload.Get("subject")
	newEntry.Body = payload.Get("body")
	newEntry.OccurredOn = Clock.Now()
	self.Log(newEntry)
}
