package command

import (
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

type CommandManager struct {
	appStore            store.AppStore
	logStore            store.LogStore
	commandStore        store.CommandStore
	pluginInstanceStore store.PluginInstanceStore
	pluginRegistry      *plugin.Registry
	tokenCrypt          *util.SymmetricCrypt
}

func NewCommandManager(
	appStore store.AppStore,
	commandStore store.CommandStore,
	pluginInstanceStore store.PluginInstanceStore,
	pluginRegistry *plugin.Registry,
	tokenCrypt *util.SymmetricCrypt,
) *CommandManager {
	return &CommandManager{
		appStore:            appStore,
		commandStore:        commandStore,
		pluginInstanceStore: pluginInstanceStore,
		pluginRegistry:      pluginRegistry,
		tokenCrypt:          tokenCrypt,
	}
}
