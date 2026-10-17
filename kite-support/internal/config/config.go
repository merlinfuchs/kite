package config

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

const ConfigFile = "kite-support.toml"
const EnvPrefix = "KITE_SUPPORT_"

//go:embed default.toml
var defaultConfig []byte

type Validate interface {
	Validate() error
}

func loadConfig[T Validate](basePath string) (T, error) {
	var res T

	k, err := loadBase(basePath)
	if err != nil {
		return res, fmt.Errorf("failed to load base config: %w", err)
	}

	if err := k.UnmarshalWithConf("", &res, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return res, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := res.Validate(); err != nil {
		return res, fmt.Errorf("failed to validate config: %w", err)
	}

	return res, nil
}

func loadBase(basePath string) (*koanf.Koanf, error) {
	k := koanf.New(".")
	parser := toml.Parser()

	if err := k.Load(rawbytes.Provider(defaultConfig), parser); err != nil {
		return nil, fmt.Errorf("failed to load default config: %w", err)
	}

	configPath := filepath.Join(basePath, ConfigFile)
	if err := k.Load(file.Provider(configPath), parser); err != nil {
		var pathErr *fs.PathError
		if !errors.As(err, &pathErr) {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	envProvider := env.Provider(EnvPrefix, ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, EnvPrefix)), "__", ".")
	})
	if err := k.Load(envProvider, nil); err != nil {
		return nil, fmt.Errorf("failed to load env config: %w", err)
	}

	return k, nil
}
