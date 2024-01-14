package model

import "time"

type Plugin struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	WASMSize    int       `json:"wasm_size"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PluginDeployment struct {
	ID        string            `json:"id"`
	PluginID  string            `json:"plugin_id"`
	GuildID   string            `json:"guild_id"`
	Config    map[string]string `json:"config"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}
