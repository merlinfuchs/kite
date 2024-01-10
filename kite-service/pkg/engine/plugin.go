package engine

import "github.com/merlinfuchs/kite/kite-service/pkg/plugin"

type LoadedPlugin struct {
	Plugin   *plugin.Plugin
	GuildIDs map[string]struct{}
}
