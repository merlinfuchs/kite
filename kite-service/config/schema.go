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
	Host         string               `toml:"host" validate:"required"`
	Port         int                  `toml:"port" validate:"required"`
	Log          ServerLogConfig      `toml:"log"`
	PublicURL    string               `toml:"public_url" validate:"required"`
	AppPublicURL string               `toml:"app_public_url" validate:"required"`
	Postgres     ServerPostgresConfig `toml:"postgres" validate:"required"`
	Discord      ServerDiscordConfig  `toml:"discord" validate:"required"`
}

func (cfg *ServerConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(cfg)
}

type ServerLogConfig struct {
	Filename   string `toml:"filename"`
	MaxSize    int    `toml:"max_size"`
	MaxAge     int    `toml:"max_age"`
	MaxBackups int    `toml:"max_backups"`
}

type ServerPostgresConfig struct {
	Host     string `toml:"host" validate:"required"`
	Port     int    `toml:"port" validate:"required"`
	DBName   string `toml:"db_name" validate:"required"`
	User     string `toml:"user" validate:"required"`
	Password string `toml:"password"`
}

type ServerDiscordConfig struct {
	Token        string `toml:"token" validate:"required"`
	ClientID     string `toml:"client_id" validate:"required"`
	ClientSecret string `toml:"client_secret" validate:"required"`
}

type PluginConfig struct {
	Key           string                `toml:"key" validate:"required,ascii"`
	Name          string                `toml:"name" validate:"required"`
	Description   string                `toml:"description" validate:"required"`
	Type          string                `toml:"type" validate:"required,oneof=go rust js"`
	Build         *PluginBuildConfig    `toml:"build" validate:"required"`
	DefaultConfig map[string]string     `toml:"default_config"`
	Commands      []PluginCommandConfig `toml:"commands" validate:"dive"`
	Events        []string              `toml:"events"`
}

func (cfg *PluginConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(cfg)
}

type PluginBuildConfig struct {
	In  string `toml:"in"`
	Out string `toml:"out" validate:"required"`
}

type PluginCommandConfig struct {
	Type                     string                      `toml:"type" validate:"required,ascii,oneof=chat user message"`
	Name                     string                      `toml:"name" validate:"required,ascii"`
	Description              string                      `toml:"description" validate:"required,ascii"`
	DefaultMemberPermissions []string                    `toml:"default_member_permissions"`
	DMPermission             bool                        `toml:"dm_permission"`
	NSFW                     bool                        `toml:"nsfw"`
	Options                  []PluginCommandOptionConfig `toml:"options" validate:"dive"`
}

type PluginCommandOptionConfig struct {
	Type        string                              `toml:"type" validate:"required,oneof=string int bool user channel role mentionable float attachment"`
	Name        string                              `toml:"name" validate:"required,ascii"`
	Description string                              `toml:"description" validate:"required,ascii"`
	Required    bool                                `toml:"required"`
	MinValue    int                                 `toml:"min_value"`
	MaxValue    int                                 `toml:"max_value"`
	MinLength   int                                 `toml:"min_length"`
	MaxLength   int                                 `toml:"max_length"`
	Choices     []PluginCommandArgumentChoiceConfig `toml:"choices" validate:"dive"`
	Options     []PluginCommandOptionConfig         `toml:"options" validate:"dive"`
}

type PluginCommandArgumentChoiceConfig struct {
	Name  string `toml:"name" validate:"required"`
	Value string `toml:"value" validate:"required"`
}
