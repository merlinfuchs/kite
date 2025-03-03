package plugin

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

type Plugin interface {
	ID() string
	Metadata() Metadata
	Version() string
	Config() Config

	Instance(config ConfigValues) (PluginInstance, error)
}

type PluginInstance interface {
	Events() []Event
	Commands() []Command

	Update(ctx Context) (UpdateResult, error)
	HandleEvent(ctx Context, event gateway.Event) error
}

type UpdateResult struct {
	CommandsChanged bool
	EventsChanged   bool
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
	Data discord.Command
}
