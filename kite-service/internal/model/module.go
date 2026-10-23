package model

import (
	"time"
)

// Module is a template installed into an app. The commands and event listeners
// it created have its ID as their ModuleID.
type Module struct {
	ID              string
	Name            string
	Description     string
	Enabled         bool
	AppID           string
	CreatorUserID   string
	TemplateID      string
	TemplateVersion int
	// Settings are the values the user filled in for the template's settings.
	Settings  map[string]string
	Items     []ModuleItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ModuleItemType string

const (
	ModuleItemTypeCommand       ModuleItemType = "command"
	ModuleItemTypeEventListener ModuleItemType = "event_listener"
)

// ModuleItem is something a module created. Items stay listed after the user
// deleted them, so an update doesn't add them again.
type ModuleItem struct {
	// Key identifies the item in the template, so a new version of it can be
	// matched to the item.
	Key  string         `json:"key"`
	Type ModuleItemType `json:"type"`
	ID   string         `json:"id"`
	// Hash is the ContentHash of the item's flow when it was installed. It
	// differs from the current flow's once the user edited it.
	Hash string `json:"hash"`
}
