package model

import "time"

type Workspace struct {
	ID          string
	GuildID     string
	Type        WorkspaceType
	Name        string
	Description string
	Files       []WorkspaceFile
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WorkspaceType string

const (
	WorkspaceTypeJS   WorkspaceType = "JS"
	WorkspaceTypeFlow WorkspaceType = "FLOW"
)

type WorkspaceFile struct {
	Path    string
	Content string
}
