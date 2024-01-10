package config

import (
	"fmt"
	"os"
	"path/filepath"

	_ "embed"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

const configFile = "kite.toml"

//go:embed default.toml
var defaultConfig []byte

func LoadFullConfig(basePath string) (*FullConfig, error) {
	k, err := loadBase(basePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load base config: %v", err)
	}

	res := &FullConfig{}
	if err := k.UnmarshalWithConf("", res, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal full config: %v", err)
	}

	if err := res.Validate(); err != nil {
		return nil, fmt.Errorf("Failed to validate plugin config: %v", err)
	}

	return res, nil
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

	fmt.Println(k.String("plugin.name"))

	res := &PluginConfig{}
	if err := k.UnmarshalWithConf("plugin", res, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal plugin config: %v", err)
	}

	return res, nil
}

func defaultBase() (*koanf.Koanf, error) {
	k := koanf.New(".")
	parser := toml.Parser()

	if err := k.Load(rawbytes.Provider(defaultConfig), parser); err != nil {
		return nil, fmt.Errorf("Failed to load default config: %v", err)
	}

	return k, nil
}

func loadBase(basePath string) (*koanf.Koanf, error) {
	k := koanf.New(".")
	parser := toml.Parser()

	if err := k.Load(rawbytes.Provider(defaultConfig), parser); err != nil {
		return nil, fmt.Errorf("Failed to load default config: %v", err)
	}

	configPath := filepath.Join(basePath, configFile)
	if err := k.Load(file.Provider(configPath), parser); err != nil {
		return nil, fmt.Errorf("Failed to load config file: %v", err)
	}

	return k, nil
}

func ConfigExists(basePath string) bool {
	configPath := filepath.Join(basePath, configFile)
	_, err := os.Stat(configPath)
	return err == nil
}
