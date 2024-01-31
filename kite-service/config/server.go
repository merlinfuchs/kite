package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/v2"
)

type ServerConfig struct {
	Host      string               `toml:"host" validate:"required"`
	Port      int                  `toml:"port" validate:"required"`
	Log       ServerLogConfig      `toml:"log"`
	PublicURL string               `toml:"public_url" validate:"required"`
	SDK       ServerSDKConfig      `toml:"sdk" validate:"required"`
	App       ServerAppConfig      `toml:"app" validate:"required"`
	Postgres  ServerPostgresConfig `toml:"postgres" validate:"required"`
	Discord   ServerDiscordConfig  `toml:"discord" validate:"required"`
}

func (cfg *ServerConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(cfg)
}

func (cfg *ServerConfig) AuthCallbackURL() string {
	return cfg.PublicURL + "/v1/auth/callback"
}

func (cfg *ServerConfig) AuthCLICallbackURL() string {
	return cfg.PublicURL + "/v1/auth/cli/callback"
}

type ServerSDKConfig struct {
	JSScriptPath      string `toml:"js_script_path" validate:"required"`
	JSDeclarationPath string `toml:"js_declaration_path" validate:"required"`
}

type ServerAppConfig struct {
	PublicURL string `toml:"public_url" validate:"required"`
}

func (cfg *ServerAppConfig) AuthCallbackURL() string {
	return cfg.PublicURL
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

func LoadServerConfig(basePath string) (*ServerConfig, error) {
	k, err := loadBase(basePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load base config: %v", err)
	}

	res := &ServerConfig{}
	if err := k.UnmarshalWithConf("server", res, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal server config: %v", err)
	}

	if err := res.Validate(); err != nil {
		return nil, fmt.Errorf("Failed to validate plugin config: %v", err)
	}

	return res, nil
}
