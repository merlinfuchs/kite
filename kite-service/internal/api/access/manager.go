package access

import "github.com/kitecloud/kite/kite-service/internal/store"

type AccessManager struct {
	appStore     store.AppStore
	commandStore store.CommandStore
}

func NewAccessManager(appStore store.AppStore, commandStore store.CommandStore) *AccessManager {
	return &AccessManager{
		appStore:     appStore,
		commandStore: commandStore,
	}
}
