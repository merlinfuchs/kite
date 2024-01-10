package engine

import (
	"sync"

	"github.com/merlinfuchs/kite/kite-service/pkg/plugin"
)

type PluginEngine struct {
	sync.RWMutex

	Plugins []LoadedPlugin
}

func New() *PluginEngine {
	return &PluginEngine{}
}

func (e *PluginEngine) LoadPlugin(plugin *plugin.Plugin, guildIDs []string) error {
	e.Lock()
	defer e.Unlock()

	gids := make(map[string]struct{}, len(guildIDs))
	for _, gid := range guildIDs {
		gids[gid] = struct{}{}
	}

	e.Plugins = append(e.Plugins, LoadedPlugin{
		Plugin:   plugin,
		GuildIDs: gids,
	})
	return nil
}
