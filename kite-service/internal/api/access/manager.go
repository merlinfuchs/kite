package access

import (
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

type AccessManager struct {
	apps store.AppStore
}

func New(apps store.AppStore) *AccessManager {
	return &AccessManager{
		apps: apps,
	}
}
