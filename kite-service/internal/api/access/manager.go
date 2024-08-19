package access

import "github.com/kitecloud/kite/kite-service/internal/store"

type AccessManager struct {
	appStore      store.AppStore
	commandStore  store.CommandStore
	variableStore store.VariableStore
}

func NewAccessManager(
	appStore store.AppStore,
	commandStore store.CommandStore,
	variableStore store.VariableStore,
) *AccessManager {
	return &AccessManager{
		appStore:      appStore,
		commandStore:  commandStore,
		variableStore: variableStore,
	}
}
