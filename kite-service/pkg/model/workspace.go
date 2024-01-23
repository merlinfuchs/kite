package model

import "time"

type Workspace struct {
	ID          string
	GuildID     string
	Name        string
	Description string
	Files       []WorkspaceFile
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WorkspaceFile struct {
	Path    string
	Content string
}
