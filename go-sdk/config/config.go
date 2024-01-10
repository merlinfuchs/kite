package config

import "github.com/merlinfuchs/kite/go-sdk/internal"

var cfg map[string]string

// Load loads the config from the config file
func Load() (map[string]string, error) {
	if cfg != nil {
		return cfg, nil
	}

	config, err := internal.GetConfig()
	if err != nil {
		return nil, err
	}

	cfg = config
	return cfg, nil
}
