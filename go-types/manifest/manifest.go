package manifest

import "github.com/merlinfuchs/kite/go-types/event"

type Manifest struct {
	Events          []event.EventType `json:"events"`
	DiscordCommands []DiscordCommand  `json:"discord_commands"`
}
