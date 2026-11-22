package access

import (
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/core/sharedbot"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type AccessManager struct {
	appStore            store.AppStore
	commandStore        store.CommandStore
	variableStore       store.VariableStore
	messageStore        store.MessageStore
	eventListenerStore  store.EventListenerStore
	pluginInstanceStore store.PluginInstanceStore
	planManager         *plan.PlanManager
	userStore           store.UserStore
	// sharedMembers decide who can change apps on the Kite bot, nil without
	// it.
	sharedMembers *sharedbot.Members
}

func NewAccessManager(
	appStore store.AppStore,
	commandStore store.CommandStore,
	variableStore store.VariableStore,
	messageStore store.MessageStore,
	eventListenerStore store.EventListenerStore,
	pluginInstanceStore store.PluginInstanceStore,
	planManager *plan.PlanManager,
	userStore store.UserStore,
	sharedMembers *sharedbot.Members,
) *AccessManager {
	return &AccessManager{
		appStore:            appStore,
		commandStore:        commandStore,
		variableStore:       variableStore,
		messageStore:        messageStore,
		eventListenerStore:  eventListenerStore,
		pluginInstanceStore: pluginInstanceStore,
		planManager:         planManager,
		userStore:           userStore,
		sharedMembers:       sharedMembers,
	}
}
