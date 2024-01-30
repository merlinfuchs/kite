package host

import (
	"context"
)

func (h HostEnvironment) GetConfig(ctx context.Context) (map[string]interface{}, error) {
	res := h.Manifest.ConfigSchema.DefaultConfig()
	for k, v := range h.Config {
		res[k] = v
	}

	return res, nil
}
