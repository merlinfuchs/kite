package plugin

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/gateway"
)

type Plugin interface {
	ID() string
	IsDefault() bool
	Metadata() Metadata
	Config() Config

	Instance(ctx context.Context, appID string, config ConfigValues) (PluginInstance, error)
}

type PluginInstance interface {
	Events() []Event
	Commands() []Command

	Update(c Context, config ConfigValues) error
	HandleEvent(c Context, event gateway.Event) error
}

type Metadata struct {
	Name        string
	Description string
	Icon        string
	Author      string
}

type Event struct {
	ID          string
	Source      EventSource
	Type        EventType
	Description string
}

type EventSource string

const (
	EventSourceDiscord EventSource = "discord"
)

type EventType string

const (
	EventTypeMessageCreate EventType = "message_create"
)

type Command struct {
	ID   string
	Data api.CreateCommandData
}
