package message

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

func (h *MessageHandler) HandleMessageInstanceList(c *handler.Context) (*wire.MessageInstanceListResponse, error) {
	instances, err := h.messageInstanceStore.MessageInstancesByMessage(c.Context(), c.Message.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message instances: %w", err)
	}

	res := make([]*wire.MessageInstance, len(instances))
	for i, instance := range instances {
		res[i] = wire.MessageInstanceToWire(instance)
	}

	return &res, nil
}

func (h *MessageHandler) HandleMessageInstanceCreate(c *handler.Context, req wire.MessageInstanceCreateRequest) (*wire.MessageInstanceCreateResponse, error) {
	// TODO: send message
	discordMessageID := "123"

	instance, err := h.messageInstanceStore.CreateMessageInstance(c.Context(), &model.MessageInstance{
		MessageID:        c.Message.ID,
		DiscordGuildID:   req.DiscordGuildID,
		DiscordChannelID: req.DiscordChannelID,
		DiscordMessageID: discordMessageID,
		FlowSources:      c.Message.FlowSources,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create message instance: %w", err)
	}

	return wire.MessageInstanceToWire(instance), nil
}

func (h *MessageHandler) HandleMessageInstanceUpdate(c *handler.Context) (*wire.MessageInstanceUpdateResponse, error) {
	instanceID, _ := strconv.ParseUint(c.Param("instanceID"), 10, 64)

	instance, err := h.messageInstanceStore.MessageInstance(c.Context(), c.Message.ID, instanceID)
	if err != nil {
		if err == store.ErrNotFound {
			return nil, handler.ErrNotFound("message_instance_not_found", "message instance not found")
		}
		return nil, fmt.Errorf("failed to get message instance: %w", err)
	}

	// TODO: update discord message
	fmt.Println(instance.DiscordChannelID)

	instance, err = h.messageInstanceStore.UpdateMessageInstance(c.Context(), &model.MessageInstance{
		ID:          instance.ID,
		MessageID:   instance.MessageID,
		FlowSources: c.Message.FlowSources,
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update message instance: %w", err)
	}

	return wire.MessageInstanceToWire(instance), nil
}

func (h *MessageHandler) HandleMessageInstanceDelete(c *handler.Context) (*wire.MessageInstanceDeleteResponse, error) {
	instanceID, _ := strconv.ParseUint(c.Param("instanceID"), 10, 64)

	err := h.messageInstanceStore.DeleteMessageInstance(c.Context(), c.Message.ID, instanceID)
	if err != nil {
		if err == store.ErrNotFound {
			return nil, handler.ErrNotFound("message_instance_not_found", "message instance not found")
		}
		return nil, fmt.Errorf("failed to get message instance: %w", err)
	}

	return &wire.MessageInstanceDeleteResponse{}, nil
}
