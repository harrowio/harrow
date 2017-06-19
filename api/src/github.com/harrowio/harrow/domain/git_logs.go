package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type GitLogs struct {
	Repositories map[string][]*GitLogEntry `json:"repositories"`
}

type GitLogEntry struct {
	Commit      string   `json:"commit"`
	Author      string   `json:"author"`
	AuthorEmail string   `json:"authorEmail"`
	AuthorDate  ISODate  `json:"authorDate"`
	Parents     []string `json:"parents"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
}

func NewGitLogs() *GitLogs {
	return &GitLogs{
		Repositories: map[string][]*GitLogEntry{},
	}
}

func (self *GitLogs) Trim(maxLength int) *GitLogs {
	trimmed := map[string][]*GitLogEntry{}
	for repositoryUuid, commits := range self.Repositories {
		if len(commits) > maxLength {
			trimmed[repositoryUuid] = commits[0:maxLength]
		} else {
			trimmed[repositoryUuid] = commits
		}
	}

	self.Repositories = trimmed

	return self
}

func (self *GitLogs) HandleEvent(payload EventPayload) {
	if payload.Get("event") != "git-log" {
		return
	}
	repositoryId := payload.Get("repository")
	entryString := payload.Get("entry")
	entry := &GitLogEntry{}

	if err := json.Unmarshal([]byte(entryString), entry); err != nil {
		return
	}

	self.Repositories[repositoryId] = append(self.Repositories[repositoryId], entry)
}

func (self *GitLogs) Value() (driver.Value, error) {
	data, err := json.Marshal(self)
	return data, err
}

func (self *GitLogs) Scan(data interface{}) error {
	src := []byte{}
	switch raw := data.(type) {
	case []byte:
		src = raw
	default:
		return fmt.Errorf("GitLogs: cannot scan from %T", data)
	}

	if err := json.Unmarshal(src, self); err != nil {
		return err
	}
	if self.Repositories == nil {
		self.Repositories = make(map[string][]*GitLogEntry)
	}
	return nil
}
