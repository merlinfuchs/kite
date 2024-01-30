package config

import (
	"github.com/merlinfuchs/kite/go-sdk/internal"
	"github.com/merlinfuchs/kite/go-types/manifest"
)

var cfg map[string]interface{}

func String(key string) string {
	load()

	v := cfg[key]
	if v == nil {
		return ""
	}

	return v.(string)
}

func Int(key string) int {
	load()

	v := cfg[key]
	if v == nil {
		return 0
	}

	return v.(int)
}

func Bool(key string) bool {
	load()

	v := cfg[key]
	if v == nil {
		return false
	}

	return v.(bool)
}

func Float(key string) float64 {
	load()

	v := cfg[key]
	if v == nil {
		return 0
	}

	return v.(float64)
}

// Load loads the config from the config file
func load() {
	if cfg != nil {
		return
	}

	config, err := internal.GetConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	cfg = config
	return
}

func SetSchema(schema manifest.ConfigSchema) {
	internal.Manifest.ConfigSchema = schema
}
