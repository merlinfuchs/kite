// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0

package pgmodel

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/sqlc-dev/pqtype"
)

type Deployment struct {
	ID              string
	Key             string
	Name            string
	Description     string
	GuildID         string
	PluginVersionID sql.NullString
	WasmBytes       []byte
	Manifest        json.RawMessage
	Config          json.RawMessage
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeployedAt      sql.NullTime
}

type DeploymentLog struct {
	ID           int64
	DeploymentID string
	Level        string
	Message      string
	CreatedAt    time.Time
}

type DeploymentMetric struct {
	ID                 int64
	DeploymentID       string
	Type               string
	Metadata           pqtype.NullRawMessage
	EventType          string
	EventSuccess       bool
	EventExecutionTime int64
	EventTotalTime     int64
	CallType           string
	CallSuccess        bool
	CallTotalTime      int64
	Timestamp          time.Time
}

type Guild struct {
	ID          string
	Name        string
	Icon        sql.NullString
	Description sql.NullString
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type KvStorage struct {
	GuildID   string
	Namespace string
	Key       string
	Value     json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Plugin struct {
	ID          string
	Key         string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PluginVersion struct {
	ID                    string
	PluginID              string
	VersionMajor          int32
	VersionMinor          int32
	VersionPatch          int32
	WasmBytes             []byte
	ManifestDefaultConfig pqtype.NullRawMessage
	ManifestEvents        []string
	ManifestCommands      []string
	CreatedAt             time.Time
}

type Session struct {
	TokenHash   string
	Type        string
	UserID      string
	GuildIds    []string
	AccessToken string
	Revoked     bool
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

type SessionsPending struct {
	Code      string
	Token     sql.NullString
	CreatedAt time.Time
	ExpiresAt time.Time
}

type User struct {
	ID            string
	Username      string
	Discriminator sql.NullString
	GlobalName    sql.NullString
	Avatar        sql.NullString
	PublicFlags   int32
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Workspace struct {
	ID          string
	GuildID     string
	Name        string
	Description string
	Files       json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
