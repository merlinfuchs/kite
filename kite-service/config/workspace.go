package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/v2"
	"github.com/pelletier/go-toml"
)

type WorkspaceConfig struct {
	Deployment *DeploymentConfig `toml:"deployment" validate:"required"`
	Module     *ModuleConfig     `toml:"module" validate:"required"`
}

func (cfg *WorkspaceConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(cfg)
}

type DeploymentConfig struct {
	Key         string `toml:"key" validate:"required,ascii"`
	Name        string `toml:"name" validate:"required"`
	Description string `toml:"description" validate:"required"`
}

type ModuleConfig struct {
	Type  string             `toml:"type" validate:"required,oneof=go rust js"`
	Build *ModuleBuildConfig `toml:"build" validate:"required"`
}
type ModuleBuildConfig struct {
	In  string `toml:"in"`
	Out string `toml:"out" validate:"required"`
}

func DefaultWorkspaceConifg() (*WorkspaceConfig, error) {
	k, err := defaultBase()
	if err != nil {
		return nil, fmt.Errorf("Failed to load base config: %v", err)
	}

	res := &WorkspaceConfig{}
	if err := k.UnmarshalWithConf("", res, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal plugin config: %v", err)
	}

	return res, nil
}

func LoadworkspaceConfig(basePath string) (*WorkspaceConfig, error) {
	k, err := loadBase(basePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load base config: %v", err)
	}

	res := &WorkspaceConfig{}
	if err := k.UnmarshalWithConf("", res, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal deployment config: %v", err)
	}

	if err := res.Validate(); err != nil {
		return nil, fmt.Errorf("Failed to validate deployment config: %v", err)
	}

	return res, nil
}

func WriteWorkspaceConfig(basePath string, cfg *WorkspaceConfig) error {
	raw, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("Failed to marshal config: %v", err)
	}

	configPath := filepath.Join(basePath, configFile)
	if err := os.WriteFile(configPath, raw, 0644); err != nil {
		return fmt.Errorf("Failed to write config file: %v", err)
	}

	return nil
}
