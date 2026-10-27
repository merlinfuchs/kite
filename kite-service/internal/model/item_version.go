package model

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"gopkg.in/guregu/null.v4"
)

// MaxItemVersions is how many versions are kept per item.
const MaxItemVersions = 25

type ItemVersionTargetType string

const (
	ItemVersionTargetTypeCommand       ItemVersionTargetType = "command"
	ItemVersionTargetTypeEventListener ItemVersionTargetType = "event_listener"
	ItemVersionTargetTypeMessage       ItemVersionTargetType = "message"
)

// ItemVersionSource is what made an item version.
type ItemVersionSource string

const (
	// The item from before its first version, made by the first change.
	ItemVersionSourceInitial ItemVersionSource = "initial"
	ItemVersionSourceEditor  ItemVersionSource = "editor"
	// A change made with an API token, e.g. by an MCP client.
	ItemVersionSourceAPIToken ItemVersionSource = "api_token"
	// A template or bundle update, or a change of its parameters.
	ItemVersionSourceBundle ItemVersionSource = "bundle"
	// The last state of a deleted item.
	ItemVersionSourceDelete ItemVersionSource = "delete"
)

// ItemVersion is a saved state of a command, event listener or message.
type ItemVersion struct {
	ID         string
	AppID      string
	TargetType ItemVersionTargetType
	TargetID   string
	Source     ItemVersionSource
	// Null for initial versions.
	CreatorUserID null.String
	// Only set in lists.
	CreatorDisplayName null.String
	// The flow of a command or event listener, the data and flow_sources of
	// a message. Empty in lists, which only load it for a single version.
	Content   json.RawMessage
	CreatedAt time.Time
	// When later changes merged into the version last updated it, if any.
	UpdatedAt null.Time
}

// contentVersion identifies the content of an item, what its version history
// keeps. Unlike updated_at, it doesn't change when an item is only enabled or
// disabled, or gets a new webhook secret. Unlike ContentHash, which compares
// items with their template, it covers every saved change, like moving blocks.
func contentVersion(content ...any) string {
	raw, _ := json.Marshal(content)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:8])
}

func (c *Command) ContentVersion() string {
	return contentVersion(c.FlowSource)
}

func (l *EventListener) ContentVersion() string {
	return contentVersion(l.FlowSource)
}

func (m *Message) ContentVersion() string {
	return contentVersion(m.Name, m.Description, m.Data, m.FlowSources)
}
