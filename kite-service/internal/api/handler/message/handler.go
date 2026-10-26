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
}

func NewMessageHandler(
	messageStore store.MessageStore,
	messageInstanceStore store.MessageInstanceStore,
	assetStore store.AssetStore,
	appStateManager store.AppStateManager,
) *MessageHandler {
	return &MessageHandler{
		messageStore:         messageStore,
		messageInstanceStore: messageInstanceStore,
		assetStore:           assetStore,
		appStateManager:      appStateManager,
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
	if c.Features.MaxMessages != 0 {
		messageCount, err := h.messageStore.CountMessagesByApp(c.Context(), c.App.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to count messages: %w", err)
		}

		if messageCount >= c.Features.MaxMessages {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of messages (%d) reached", c.Features.MaxMessages))
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

func (h *MessageHandler) HandleMessagesImport(c *handler.Context, req wire.MessagesImportRequest) (*wire.MessagesImportResponse, error) {
	messages, err := h.PrepareImport(c, req.Messages)
	if err != nil {
		return nil, err
	}

	if err := h.CreateMessages(c, messages); err != nil {
		return nil, err
	}

	res := make([]*wire.Message, len(messages))
	for i, message := range messages {
		res[i] = wire.MessageToWire(message)
	}

	return &res, nil
}

// PrepareImport checks that the messages can be added to the app and returns
// them with their IDs set, without saving them.
func (h *MessageHandler) PrepareImport(c *handler.Context, reqs []wire.MessageCreateRequest) ([]*model.Message, error) {
	if len(reqs) == 0 {
		return nil, nil
	}

	if c.Features.MaxMessages != 0 {
		messageCount, err := h.messageStore.CountMessagesByApp(c.Context(), c.App.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to count messages: %w", err)
		}

		newMessageCount := messageCount + len(reqs)

		if newMessageCount > c.Features.MaxMessages {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of messages (%d) reached", c.Features.MaxMessages))
		}
	}

	messages := make([]*model.Message, len(reqs))
	for i, msg := range reqs {
		messages[i] = &model.Message{
			ID:            util.UniqueID(),
			Name:          msg.Name,
			Description:   msg.Description,
			AppID:         c.App.ID,
			CreatorUserID: c.Session.UserID,
			Data:          msg.Data,
			FlowSources:   msg.FlowSources,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}
	}

	return messages, nil
}

// CreateMessages saves the messages from PrepareImport.
func (h *MessageHandler) CreateMessages(c *handler.Context, messages []*model.Message) error {
	for i, message := range messages {
		created, err := h.messageStore.CreateMessage(c.Context(), message)
		if err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}
		messages[i] = created
	}

	return nil
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
	}, c.ItemChange(req.EditorSession))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_message", "Message not found")
		}
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	return wire.MessageToWire(message), nil
}

func (h *MessageHandler) HandleMessageDelete(c *handler.Context) (*wire.MessageDeleteResponse, error) {
	err := h.messageStore.DeleteMessage(c.Context(), c.Message.ID, c.Session.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_message", "Message not found")
		}
		return nil, fmt.Errorf("failed to delete message: %w", err)
	}

	return &wire.MessageDeleteResponse{}, nil
}
