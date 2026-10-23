package model

import (
	"time"
)

// BundleInstall is a bundle installed into an app. Bundles are the templates
// for now. The commands and event listeners it created have its ID as their
// BundleInstallID.
type BundleInstall struct {
	ID            string
	Name          string
	Description   string
	Enabled       bool
	AppID         string
	CreatorUserID string
	BundleID      string
	BundleVersion int
	// Settings are the values the user filled in for the bundle's settings.
	Settings  map[string]string
	Items     []BundleInstallItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BundleInstallItemType string

const (
	BundleInstallItemTypeCommand       BundleInstallItemType = "command"
	BundleInstallItemTypeEventListener BundleInstallItemType = "event_listener"
	BundleInstallItemTypeMessage       BundleInstallItemType = "message"
)

// BundleInstallItem is something an install created. Items stay listed after
// the user deleted them, so an update doesn't add them again.
type BundleInstallItem struct {
	// Key identifies the item in the bundle, so a new version of it can be
	// matched to the item.
	Key  string                `json:"key"`
	Type BundleInstallItemType `json:"type"`
	ID   string                `json:"id"`
	// Hash is the ContentHash of the item's flow when it was installed, or of
	// the content and flows of a message. It differs from the current one once
	// the user edited it.
	Hash string `json:"hash"`
}
