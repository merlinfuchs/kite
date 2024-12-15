// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package pgmodel

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type App struct {
	ID             string
	Name           string
	Description    pgtype.Text
	Enabled        bool
	OwnerUserID    string
	CreatorUserID  string
	DiscordToken   string
	DiscordID      string
	CreatedAt      pgtype.Timestamp
	UpdatedAt      pgtype.Timestamp
	DiscordStatus  []byte
	DisabledReason pgtype.Text
}

type Asset struct {
	ID            string
	Name          string
	ContentHash   string
	ContentType   string
	ContentSize   int32
	AppID         string
	ModuleID      pgtype.Text
	CreatorUserID string
	CreatedAt     pgtype.Timestamp
	UpdatedAt     pgtype.Timestamp
	ExpiresAt     pgtype.Timestamp
}

type Collaborator struct {
	UserID    string
	AppID     string
	Role      string
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}

type Command struct {
	ID             string
	Name           string
	Description    string
	Enabled        bool
	AppID          string
	ModuleID       pgtype.Text
	CreatorUserID  string
	FlowSource     []byte
	CreatedAt      pgtype.Timestamp
	UpdatedAt      pgtype.Timestamp
	LastDeployedAt pgtype.Timestamp
}

type EventListener struct {
	ID            string
	Source        string
	Type          string
	Description   string
	Enabled       bool
	AppID         string
	ModuleID      pgtype.Text
	CreatorUserID string
	Filter        []byte
	FlowSource    []byte
	CreatedAt     pgtype.Timestamp
	UpdatedAt     pgtype.Timestamp
}

type Log struct {
	ID        int64
	AppID     string
	Message   string
	Level     string
	CreatedAt pgtype.Timestamp
}

type Message struct {
	ID            string
	Name          string
	Description   pgtype.Text
	Data          []byte
	FlowSources   []byte
	AppID         string
	ModuleID      pgtype.Text
	CreatorUserID string
	CreatedAt     pgtype.Timestamp
	UpdatedAt     pgtype.Timestamp
}

type MessageInstance struct {
	ID               int64
	MessageID        string
	Hidden           bool
	Ephemeral        bool
	DiscordGuildID   string
	DiscordChannelID string
	DiscordMessageID string
	FlowSources      []byte
	CreatedAt        pgtype.Timestamp
	UpdatedAt        pgtype.Timestamp
}

type Module struct {
	ID            string
	Name          string
	Description   string
	Enabled       bool
	AppID         string
	CreatorUserID string
	Resources     []byte
	CreatedAt     pgtype.Timestamp
	UpdatedAt     pgtype.Timestamp
}

type Session struct {
	KeyHash   string
	UserID    string
	CreatedAt pgtype.Timestamp
	ExpiresAt pgtype.Timestamp
}

type UsageRecord struct {
	ID              int64
	Type            string
	AppID           string
	CommandID       pgtype.Text
	EventListenerID pgtype.Text
	MessageID       pgtype.Text
	CreditsUsed     int32
	CreatedAt       pgtype.Timestamp
}

type User struct {
	ID              string
	Email           string
	DisplayName     string
	DiscordID       string
	DiscordUsername string
	DiscordAvatar   pgtype.Text
	CreatedAt       pgtype.Timestamp
	UpdatedAt       pgtype.Timestamp
}

type Variable struct {
	ID        string
	Name      string
	Scoped    bool
	AppID     string
	ModuleID  pgtype.Text
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}

type VariableValue struct {
	ID         int64
	VariableID string
	Scope      pgtype.Text
	Value      []byte
	CreatedAt  pgtype.Timestamp
	UpdatedAt  pgtype.Timestamp
}
