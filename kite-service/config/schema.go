package config

import (
	"github.com/go-playground/validator/v10"
)

type FullConfig struct {
	Server *ServerConfig `toml:"server"`
	Plugin *PluginConfig `toml:"plugin"`
}

func (cfg *FullConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(cfg)
}

type ServerConfig struct {
	Host    string               `toml:"host" validate:"required"`
	Port    int                  `toml:"port" validate:"required"`
	Discord *ServerDiscordConfig `toml:"discord" validate:"required"`
	Plugins []ServerPluginConfig `toml:"plugins" validate:"dive"`
}

func (cfg *ServerConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(cfg)
}

type ServerDiscordConfig struct {
	Token    string `toml:"token" validate:"required"`
	ClientID string `toml:"client_id" validate:"required"`
}

type ServerPluginConfig struct {
	Path     string            `toml:"path" validate:"required"`
	GuildIDs []string          `toml:"guild_ids"`
	Config   map[string]string `toml:"config"`
}

type PluginConfig struct {
	Name          string                `toml:"name" validate:"required,ascii"`
	Description   string                `toml:"description" validate:"required"`
	Type          string                `toml:"type" validate:"required,oneof=go rust"`
	Build         *PluginBuildConfig    `toml:"build" validate:"required"`
	DefaultConfig map[string]string     `toml:"default_config"`
	Commands      []PluginCommandConfig `toml:"commands" validate:"dive"`
	Events        []string              `toml:"events"`
	Permissions   []string              `toml:"permissions"`
}

func (cfg *PluginConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(cfg)
}

type PluginBuildConfig struct {
	Out string `toml:"out" validate:"required"`
}

type PluginCommandConfig struct {
	Type                     string                        `toml:"type" validate:"required,ascii,oneof=chat user message"`
	Name                     string                        `toml:"name" validate:"required,ascii"`
	Description              string                        `toml:"description" validate:"required,ascii"`
	DefaultMemberPermissions []string                      `toml:"default_member_permissions"`
	DMPermission             bool                          `toml:"dm_permission"`
	NSFW                     bool                          `toml:"nsfw"`
	SubCommands              []PluginSubCommandConfig      `toml:"sub_commands" validate:"dive"`
	Arguments                []PluginCommandArgumentConfig `toml:"arguments" validate:"dive"`
}

type PluginSubCommandConfig struct {
	Name        string                        `toml:"name" validate:"required,ascii"`
	Description string                        `toml:"description" validate:"required,ascii"`
	SubCommands []PluginCommandConfig         `toml:"sub_commands" validate:"dive"`
	Arguments   []PluginCommandArgumentConfig `toml:"arguments" validate:"dive"`
}

type PluginCommandArgumentConfig struct {
	Type        string                              `toml:"type" validate:"required,oneof=string int bool user channel role mentionable float attachment"`
	Name        string                              `toml:"name" validate:"required,ascii"`
	Description string                              `toml:"description" validate:"required,ascii"`
	Required    bool                                `toml:"required"`
	MinValue    int                                 `toml:"min_value"`
	MaxValue    int                                 `toml:"max_value"`
	MinLength   int                                 `toml:"min_length"`
	MaxLength   int                                 `toml:"max_length"`
	Choices     []PluginCommandArgumentChoiceConfig `toml:"choices" validate:"dive"`
}

type PluginCommandArgumentChoiceConfig struct {
	Name  string `toml:"name" validate:"required"`
	Value string `toml:"value" validate:"required"`
}
