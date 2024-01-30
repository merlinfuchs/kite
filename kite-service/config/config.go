package config

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

//go:embed default.toml
var defaultConfig []byte

type FullConfig struct {
	Global *GlobalConfig `toml:"global,omitempty"`
	Server *ServerConfig `toml:"server,omitempty"`
	Plugin *PluginConfig `toml:"plugin,omitempty"`
}

func (cfg *FullConfig) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(cfg)
}

const configFile = "kite.toml"

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
		var pathError *fs.PathError
		if !errors.As(err, &pathError) {
			return nil, fmt.Errorf("Failed to load config file: %v", err)
		}
	}

	return k, nil
}

func ConfigExists(basePath string) bool {
	configPath := filepath.Join(basePath, configFile)
	_, err := os.Stat(configPath)
	return err == nil
}
