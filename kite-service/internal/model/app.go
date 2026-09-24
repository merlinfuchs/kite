package model

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/diamondburned/arikawa/v3/utils/ws"
	"gopkg.in/guregu/null.v4"
)

type App struct {
	ID             string
	Name           string
	Description    null.String
	Enabled        bool
	DisabledReason null.String
	OwnerUserID    string
	CreatorUserID  string
	DiscordToken   string
	DiscordID      string
	DiscordStatus  *AppDiscordStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type AppCredentials struct {
	DiscordID    string
	DiscordToken string
}

// AppGatewayRequirements describes what an app actually consumes from the
// Discord gateway, so it can identify with only the intents it needs instead
// of every intent it is permitted to have.
type AppGatewayRequirements struct {
	// EventListenerTypes are the types of the app's enabled Discord event
	// listeners.
	EventListenerTypes []EventListenerType
	// PluginEventTypes are the Discord event types the app's enabled plugin
	// instances subscribe to.
	PluginEventTypes []ws.EventType
	// HasMessageInstances reports whether the app has sent messages that Kite
	// still tracks. Those rows are cleaned up in response to MESSAGE_DELETE,
	// which requires the guild messages intent.
	//
	// This is a transitional dependency: once message instance cleanup no
	// longer relies on gateway events, it can be dropped and these apps will
	// stop receiving message traffic they otherwise have no use for.
	HasMessageInstances bool
}

// NeedsGuildMessages reports whether the app consumes guild message events.
func (r AppGatewayRequirements) NeedsGuildMessages() bool {
	if r.HasMessageInstances {
		return true
	}

	for _, t := range r.EventListenerTypes {
		switch t {
		case EventListenerTypeDiscordMessageCreate,
			EventListenerTypeDiscordMessageUpdate,
			EventListenerTypeDiscordMessageDelete:
			return true
		}
	}

	for _, t := range r.PluginEventTypes {
		switch t {
		case "MESSAGE_CREATE", "MESSAGE_UPDATE", "MESSAGE_DELETE":
			return true
		}
	}

	return false
}

// NeedsGuildMembers reports whether the app consumes guild member events.
func (r AppGatewayRequirements) NeedsGuildMembers() bool {
	for _, t := range r.EventListenerTypes {
		switch t {
		case EventListenerTypeDiscordGuildMemberAdd,
			EventListenerTypeDiscordGuildMemberRemove:
			return true
		}
	}

	for _, t := range r.PluginEventTypes {
		switch t {
		case "GUILD_MEMBER_ADD", "GUILD_MEMBER_REMOVE", "GUILD_MEMBER_UPDATE":
			return true
		}
	}

	return false
}

// NeedsGuildMessageReactions reports whether the app consumes reaction events.
// No event listener type covers reactions today, so this is driven entirely by
// plugins such as starboard.
func (r AppGatewayRequirements) NeedsGuildMessageReactions() bool {
	for _, t := range r.PluginEventTypes {
		switch t {
		case "MESSAGE_REACTION_ADD", "MESSAGE_REACTION_REMOVE":
			return true
		}
	}

	return false
}

// AppDiscordStatus holds the statuses an app can show. ActiveID picks the one
// shown when rotation is off, otherwise the gateway cycles through all of them.
type AppDiscordStatus struct {
	Statuses      []AppDiscordStatusEntry `json:"statuses,omitempty"`
	ActiveID      string                  `json:"active_id,omitempty"`
	RotateEnabled bool                    `json:"rotate_enabled,omitempty"`
}

type AppDiscordStatusEntry struct {
	ID            string `json:"id"`
	Label         string `json:"label,omitempty"`
	Status        string `json:"status,omitempty"`
	ActivityType  int    `json:"activity_type,omitempty"`
	ActivityName  string `json:"activity_name,omitempty"`
	ActivityState string `json:"activity_state,omitempty"`
	ActivityURL   string `json:"activity_url,omitempty"`
}

// UnmarshalJSON also accepts the old format, which stored a single status
// directly on the object, and turns it into a status list with one entry.
func (s *AppDiscordStatus) UnmarshalJSON(data []byte) error {
	type current AppDiscordStatus
	var v struct {
		current
		AppDiscordStatusEntry
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*s = AppDiscordStatus(v.current)

	legacy := v.AppDiscordStatusEntry
	if len(s.Statuses) == 0 && legacy != (AppDiscordStatusEntry{}) {
		legacy.ID = "default"
		s.Statuses = []AppDiscordStatusEntry{legacy}
		s.ActiveID = legacy.ID
	}
	return nil
}

func (s *AppDiscordStatus) Equals(other *AppDiscordStatus) bool {
	if s == nil && other == nil {
		return true
	}

	if s == nil || other == nil {
		return false
	}

	return s.ActiveID == other.ActiveID &&
		s.RotateEnabled == other.RotateEnabled &&
		slices.Equal(s.Statuses, other.Statuses)
}

// ActiveEntry returns the entry matching ActiveID, falling back to the first
// one if it doesn't match any.
func (s *AppDiscordStatus) ActiveEntry() *AppDiscordStatusEntry {
	if s == nil || len(s.Statuses) == 0 {
		return nil
	}

	for i := range s.Statuses {
		if s.Statuses[i].ID == s.ActiveID {
			return &s.Statuses[i]
		}
	}

	return &s.Statuses[0]
}

// Rotates reports whether the app cycles through more than one status.
func (s *AppDiscordStatus) Rotates() bool {
	return s != nil && s.RotateEnabled && len(s.Statuses) > 1
}

// RotationEntry returns the entry to show at the given time when rotating.
// It advances once per minute, so every gateway agrees on the current entry
// without keeping any state.
func (s *AppDiscordStatus) RotationEntry(t time.Time) *AppDiscordStatusEntry {
	return &s.Statuses[int(t.Unix()/60)%len(s.Statuses)]
}

type AppCollaboratorRole string

const (
	AppCollaboratorRoleOwner AppCollaboratorRole = "owner"
	AppCollaboratorRoleAdmin AppCollaboratorRole = "admin"
)

func (r AppCollaboratorRole) CanDeleteApp() bool {
	return r == AppCollaboratorRoleOwner
}

func (r AppCollaboratorRole) CanManageCollaborators() bool {
	return r == AppCollaboratorRoleOwner
}

type AppCollaborator struct {
	AppID     string
	UserID    string
	User      *User
	Role      AppCollaboratorRole
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AppEntity struct {
	ID   string
	Type AppEntityType
	Name string
}

type AppEntityType string

const (
	AppEntityTypeCommand       AppEntityType = "command"
	AppEntityTypeMessage       AppEntityType = "message"
	AppEntityTypeEventListener AppEntityType = "event_listener"
	AppEntityTypeVariable      AppEntityType = "variable"
)
