package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/v2"
	"github.com/pelletier/go-toml"
)

type PluginConfig struct {
	Key         string             `toml:"key" validate:"required,ascii"`
	Name        string             `toml:"name" validate:"required"`
	Description string             `toml:"description" validate:"required"`
	Type        string             `toml:"type" validate:"required,oneof=go rust js"`
	Build       *PluginBuildConfig `toml:"build" validate:"required"`
}

func (cfg *PluginConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(cfg)
}

type PluginBuildConfig struct {
	In  string `toml:"in"`
	Out string `toml:"out" validate:"required"`
}

func LoadPluginConfig(basePath string) (*PluginConfig, error) {
	k, err := loadBase(basePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load base config: %v", err)
	}

	res := &PluginConfig{}
	if err := k.UnmarshalWithConf("plugin", res, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal plugin config: %v", err)
	}

	if err := res.Validate(); err != nil {
		return nil, fmt.Errorf("Failed to validate plugin config: %v", err)
	}

	return res, nil
}

func DefaultPluginConifg() (*PluginConfig, error) {
	k, err := defaultBase()
	if err != nil {
		return nil, fmt.Errorf("Failed to load base config: %v", err)
	}

	res := &PluginConfig{}
	if err := k.UnmarshalWithConf("plugin", res, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal plugin config: %v", err)
	}

	return res, nil
}

func WritePluginConfig(basePath string, cfg *PluginConfig) error {
	raw, err := toml.Marshal(FullConfig{
		Plugin: cfg,
	})
	if err != nil {
		return fmt.Errorf("Failed to marshal config: %v", err)
	}

	configPath := filepath.Join(basePath, configFile)
	if err := os.WriteFile(configPath, raw, 0644); err != nil {
		return fmt.Errorf("Failed to write config file: %v", err)
	}

	return nil
}
