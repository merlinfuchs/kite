package plugin

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/gateway"
)

type Plugin interface {
	ID() string
	Metadata() Metadata
	Config() Config
	Events() []Event
	Commands() []Command

	Instance(ctx context.Context, appID string, config ConfigValues) (PluginInstance, error)
}

type PluginInstance interface {
	Update(ctx context.Context, config ConfigValues) error
	HandleEvent(c Context, event gateway.Event) error
	HandleCommand(c Context, event *gateway.InteractionCreateEvent) error
	HandleComponent(c Context, event *gateway.InteractionCreateEvent) error
	HandleModal(c Context, event *gateway.InteractionCreateEvent) error
	Close() error
}

type Metadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Author      string `json:"author"`
}

type Event struct {
	ID          string      `json:"id"`
	Source      EventSource `json:"source"`
	Type        EventType   `json:"type"`
	Description string      `json:"description"`
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
	ID   string                `json:"id"`
	Data api.CreateCommandData `json:"data"`
}
