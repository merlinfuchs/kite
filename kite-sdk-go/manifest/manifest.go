package manifest

import "github.com/merlinfuchs/kite/kite-sdk-go/event"

type Manifest struct {
	Events          []event.EventType `json:"events"`
	DiscordCommands []DiscordCommand  `json:"discord_commands"`
	ConfigSchema    ConfigSchema      `json:"config_schema"`
}
