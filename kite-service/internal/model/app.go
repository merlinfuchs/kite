package model

import (
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

type AppDiscordStatus struct {
	Status        string `json:"status,omitempty"`
	ActivityType  int    `json:"activity_type,omitempty"`
	ActivityName  string `json:"activity_name,omitempty"`
	ActivityState string `json:"activity_state,omitempty"`
	ActivityURL   string `json:"activity_url,omitempty"`
}

func (s *AppDiscordStatus) Equals(other *AppDiscordStatus) bool {
	if s == nil && other == nil {
		return true
	}

	if s == nil || other == nil {
		return false
	}

	return s.Status == other.Status &&
		s.ActivityType == other.ActivityType &&
		s.ActivityName == other.ActivityName &&
		s.ActivityState == other.ActivityState &&
		s.ActivityURL == other.ActivityURL
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
