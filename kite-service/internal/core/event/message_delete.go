package event

import (
	"context"
	"log/slog"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

func (h *EventHandlerWrapper) HandleMessageDeleteEvent(appID string, session *state.State, event *gateway.MessageDeleteEvent) {
	err := h.messageInstanceStore.DeleteMessageInstanceByDiscordMessageID(context.Background(), event.ID.String())
	if err != nil && err != store.ErrNotFound {
		slog.With("error", err).Error("failed to delete message instance from message delete event")
	}
}
