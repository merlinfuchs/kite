package access

import "github.com/kitecloud/kite/kite-service/internal/store"

type AccessManager struct {
	appStore           store.AppStore
	commandStore       store.CommandStore
	variableStore      store.VariableStore
	messageStore       store.MessageStore
	eventListenerStore store.EventListenerStore
}

func NewAccessManager(
	appStore store.AppStore,
	commandStore store.CommandStore,
	variableStore store.VariableStore,
	messageStore store.MessageStore,
	eventListenerStore store.EventListenerStore,
) *AccessManager {
	return &AccessManager{
		appStore:           appStore,
		commandStore:       commandStore,
		variableStore:      variableStore,
		messageStore:       messageStore,
		eventListenerStore: eventListenerStore,
	}
}
