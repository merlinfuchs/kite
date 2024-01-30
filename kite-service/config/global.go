package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/v2"
	"github.com/pelletier/go-toml"
)

type GlobalConfig struct {
	Sessions []*GlobalSessionConfig `toml:"sessions" validate:"dive"`
}

func (cfg *GlobalConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(cfg)
}

type GlobalSessionConfig struct {
	Server string `toml:"server" validate:"required"`
	Token  string `toml:"token" validate:"required"`
}

func (cfg *GlobalConfig) GetSessionForServer(server string) *GlobalSessionConfig {
	for _, session := range cfg.Sessions {
		if session.Server == server {
			return session
		}
	}

	return nil
}

func LoadGlobalConfig() (*GlobalConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("Failed to get user home dir: %v", err)
	}

	k, err := loadBase(homeDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to load base config: %v", err)
	}

	res := &GlobalConfig{}
	if err := k.UnmarshalWithConf("global", res, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal server config: %v", err)
	}

	if err := res.Validate(); err != nil {
		return nil, fmt.Errorf("Failed to validate plugin config: %v", err)
	}

	return res, nil
}

func WriteGlobalConfig(cfg *GlobalConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("Failed to get user home dir: %v", err)
	}

	raw, err := toml.Marshal(FullConfig{
		Global: cfg,
	})
	if err != nil {
		return fmt.Errorf("Failed to marshal config: %v", err)
	}

	configPath := filepath.Join(homeDir, configFile)
	if err := os.WriteFile(configPath, raw, 0644); err != nil {
		return fmt.Errorf("Failed to write config file: %v", err)
	}

	return nil
}
