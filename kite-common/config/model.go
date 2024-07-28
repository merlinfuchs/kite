package config

import "github.com/go-playground/validator/v10"

type Config struct {
	Logging  LoggingConfig  `toml:"logging"`
	Database DatabaseConfig `toml:"database"`
	API      APIConfig      `toml:"api"`
	App      AppConfig      `toml:"app"`
	Discord  DiscordConfig  `toml:"discord"`
	Engine   EngineConfig   `toml:"engine"`
}

func (cfg *Config) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(cfg)
}

func LoadConfig(basePath string) (*Config, error) {
	return loadConfig[*Config](basePath)
}

type DatabaseConfig struct {
	Postgres PostgresConfig `toml:"postgres"`
}

type LoggingConfig struct {
	Filename   string `toml:"filename"`
	MaxSize    int    `toml:"max_size"`
	MaxAge     int    `toml:"max_age"`
	MaxBackups int    `toml:"max_backups"`
}

type PostgresConfig struct {
	Host     string `toml:"host" validate:"required"`
	Port     int    `toml:"port" validate:"required"`
	DBName   string `toml:"db_name" validate:"required"`
	User     string `toml:"user" validate:"required"`
	Password string `toml:"password"`
}

type APIConfig struct {
	Host          string           `toml:"host" validate:"required"`
	Port          int              `toml:"port" validate:"required"`
	PublicBaseURL string           `toml:"public_base_url" validate:"required"`
	SecureCookies bool             `toml:"secure_cookies"`
	UserLimits    UserLimitsConfig `toml:"user_limits"`
}

type AppConfig struct {
	PublicBaseURL string `toml:"public_base_url" validate:"required"`
}

type DiscordConfig struct {
	ClientID     string `toml:"client_id" validate:"required"`
	ClientSecret string `toml:"client_secret" validate:"required"`
}

type EngineConfig struct {
	MaxStackDepth int `toml:"max_stack_depth"`
	MaxOperations int `toml:"max_operations"`
	MaxActions    int `toml:"max_actions"`
}

type UserLimitsConfig struct {
	MaxAppsPerUser    int `toml:"max_apps_per_user"`
	MaxCommandsPerApp int `toml:"max_commands_per_app"`
}
