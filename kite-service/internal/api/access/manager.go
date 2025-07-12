package access

import "github.com/kitecloud/kite/kite-service/internal/store"

type AccessManager struct {
	appStore            store.AppStore
	commandStore        store.CommandStore
	variableStore       store.VariableStore
	messageStore        store.MessageStore
	eventListenerStore  store.EventListenerStore
	pluginInstanceStore store.PluginInstanceStore
}

func NewAccessManager(
	appStore store.AppStore,
	commandStore store.CommandStore,
	variableStore store.VariableStore,
	messageStore store.MessageStore,
	eventListenerStore store.EventListenerStore,
	pluginInstanceStore store.PluginInstanceStore,
) *AccessManager {
	return &AccessManager{
		appStore:            appStore,
		commandStore:        commandStore,
		variableStore:       variableStore,
		messageStore:        messageStore,
		eventListenerStore:  eventListenerStore,
		pluginInstanceStore: pluginInstanceStore,
	}
}
