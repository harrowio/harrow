package domain

import (
	"time"
)

type WorkspaceBaseImage struct {
	defaultSubject
	Uuid       string    `json:"uuid" db:"uuid"`
	Name       string    `json:"name" db:"name"`
	Type       string    `json:"type" db:"type"`
	Path       string    `json:"path" db:"path"`
	Ref        string    `json:"ref" db:"ref"`
	Repository string    `json:"repository" db:"repository"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}

func (self *WorkspaceBaseImage) AuthorizationName() string { return "workspace-base-image" }
