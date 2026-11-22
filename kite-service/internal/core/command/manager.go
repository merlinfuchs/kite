package command

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
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
	// sharedBot is the Kite bot, which the commands of shared apps are
	// deployed with. Nil without one.
	sharedBot SharedBot
}

// SharedBot is the Kite bot, see the sharedbot package.
type SharedBot interface {
	Client(ctx context.Context) *api.Client
	AppID(ctx context.Context) (discord.AppID, error)
}

func NewCommandManager(
	appStore store.AppStore,
	commandStore store.CommandStore,
	pluginInstanceStore store.PluginInstanceStore,
	pluginRegistry *plugin.Registry,
	tokenCrypt *util.SymmetricCrypt,
	sharedBot SharedBot,
) *CommandManager {
	return &CommandManager{
		appStore:            appStore,
		commandStore:        commandStore,
		pluginInstanceStore: pluginInstanceStore,
		pluginRegistry:      pluginRegistry,
		tokenCrypt:          tokenCrypt,
		sharedBot:           sharedBot,
	}
}
