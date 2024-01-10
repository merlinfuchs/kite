package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

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
