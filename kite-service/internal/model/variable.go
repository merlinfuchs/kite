package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

type Variable struct {
	ID          string
	Scope       VariableScope
	Name        string
	Type        string
	AppID       string
	ModuleID    null.String
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TotalValues null.Int
}

type VariableScope string

const (
	VariableScopeGlobal  VariableScope = "global"
	VariableScopeGuild   VariableScope = "guild"
	VariableScopeChannel VariableScope = "channel"
	VariableScopeUser    VariableScope = "user"
	VariableScopeMember  VariableScope = "member"
)

type VariableValue struct {
	ID         uint64
	VariableID string
	Scope      null.String
	Value      []byte
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type VariableValueScope struct {
	GuildID   string
	ChannelID string
	UserID    string
}

func (s VariableValueScope) Resolve(scope VariableScope) null.String {
	switch scope {
	case VariableScopeGlobal:
		return null.String{}
	case VariableScopeGuild:
		return null.StringFrom(s.GuildID)
	case VariableScopeChannel:
		return null.StringFrom(s.ChannelID)
	case VariableScopeUser:
		return null.StringFrom(s.UserID)
	case VariableScopeMember:
		return null.StringFrom(s.GuildID + s.UserID)
	}

	return null.String{}
}
