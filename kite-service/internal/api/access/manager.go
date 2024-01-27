package access

import (
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

type AccessManager struct {
	state store.DiscordStateStore
}

func New(state store.DiscordStateStore) *AccessManager {
	return &AccessManager{
		state: state,
	}
}
