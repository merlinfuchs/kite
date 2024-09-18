package message

import (
	"errors"
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

type MessageHandler struct {
	messageStore         store.MessageStore
	messageInstanceStore store.MessageInstanceStore
	assetStore           store.AssetStore
	appStateManager      store.AppStateManager
	maxMessagesPerApp    int
}

func NewMessageHandler(
	messageStore store.MessageStore,
	messageInstanceStore store.MessageInstanceStore,
	assetStore store.AssetStore,
	appStateManager store.AppStateManager,
	maxMessagesPerApp int,
) *MessageHandler {
	return &MessageHandler{
		messageStore:         messageStore,
		messageInstanceStore: messageInstanceStore,
		assetStore:           assetStore,
		appStateManager:      appStateManager,
		maxMessagesPerApp:    maxMessagesPerApp,
	}
}

func (h *MessageHandler) HandleMessageList(c *handler.Context) (*wire.MessageListResponse, error) {
	messages, err := h.messageStore.MessagesByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	res := make([]*wire.Message, len(messages))
	for i, message := range messages {
		res[i] = wire.MessageToWire(message)
	}

	return &res, nil
}

func (h *MessageHandler) HandleMessageGet(c *handler.Context) (*wire.MessageGetResponse, error) {
	return wire.MessageToWire(c.Message), nil
}

func (h *MessageHandler) HandleMessageCreate(c *handler.Context, req wire.MessageCreateRequest) (*wire.MessageCreateResponse, error) {
	if h.maxMessagesPerApp != 0 {
		messageCount, err := h.messageStore.CountMessagesByApp(c.Context(), c.App.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to count messages: %w", err)
		}

		if messageCount >= h.maxMessagesPerApp {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of messages (%d) reached", h.maxMessagesPerApp))
		}
	}

	message, err := h.messageStore.CreateMessage(c.Context(), &model.Message{
		ID:            util.UniqueID(),
		Name:          req.Name,
		Description:   req.Description,
		AppID:         c.App.ID,
		CreatorUserID: c.Session.UserID,
		Data:          req.Data,
		FlowSources:   req.FlowSources,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return wire.MessageToWire(message), nil
}

func (h *MessageHandler) HandleMessageUpdate(c *handler.Context, req wire.MessageUpdateRequest) (*wire.MessageUpdateResponse, error) {
	message, err := h.messageStore.UpdateMessage(c.Context(), &model.Message{
		ID:          c.Message.ID,
		Name:        req.Name,
		Description: req.Description,
		AppID:       c.App.ID,
		Data:        req.Data,
		FlowSources: req.FlowSources,
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_message", "Message not found")
		}
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	return wire.MessageToWire(message), nil
}

func (h *MessageHandler) HandleMessageDelete(c *handler.Context) (*wire.MessageDeleteResponse, error) {
	err := h.messageStore.DeleteMessage(c.Context(), c.Message.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_message", "Message not found")
		}
		return nil, fmt.Errorf("failed to delete message: %w", err)
	}

	return &wire.MessageDeleteResponse{}, nil
}
