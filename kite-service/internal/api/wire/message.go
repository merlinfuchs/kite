package wire

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"gopkg.in/guregu/null.v4"
)

type Message struct {
	ID            string                   `json:"id"`
	Name          string                   `json:"name"`
	Description   null.String              `json:"description"`
	AppID         string                   `json:"app_id"`
	ModuleID      null.String              `json:"module_id"`
	CreatorUserID string                   `json:"creator_user_id"`
	Data          message.MessageData      `json:"data"`
	FlowSources   map[string]flow.FlowData `json:"flow_sources"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
	Position      int                      `json:"position"`
}

type MessageGetResponse = Message

type MessageListResponse = []*Message

type MessageCreateRequest struct {
	Name        string                   `json:"name"`
	Description null.String              `json:"description"`
	Data        message.MessageData      `json:"data"`
	FlowSources map[string]flow.FlowData `json:"flow_sources"`
}

func (req *MessageCreateRequest) Sanitize() {
	req.FlowSources = usedFlowSources(&req.Data, req.FlowSources)
}

func (req MessageCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&req.Description, validation.Length(0, 255)),
	)
}

type MessageCreateResponse = Message

type MessagesImportRequest struct {
	Messages []MessageCreateRequest `json:"messages"`
}

func (req *MessagesImportRequest) Sanitize() {
	for i := range req.Messages {
		req.Messages[i].Sanitize()
	}
}

func (req MessagesImportRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Messages, validation.Required),
	)
}

type MessagesImportResponse = []*Message

type MessageUpdateRequest struct {
	Name        string                   `json:"name"`
	Description null.String              `json:"description"`
	Data        message.MessageData      `json:"data"`
	FlowSources map[string]flow.FlowData `json:"flow_sources"`
}

func (req *MessageUpdateRequest) Sanitize() {
	req.FlowSources = usedFlowSources(&req.Data, req.FlowSources)
}

func (req MessageUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&req.Description, validation.Length(0, 255)),
	)
}

type MessageUpdateResponse = Message

type MessageMoveRequest struct {
	Direction string `json:"direction"`
}

func (req MessageMoveRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Direction, validation.Required, validation.In("up", "down")),
	)
}

type MessageMoveResponse = []*Message

type MessageDeleteResponse = Empty

func MessageToWire(variable *model.Message) *Message {
	if variable == nil {
		return nil
	}

	return &Message{
		ID:            variable.ID,
		Name:          variable.Name,
		Description:   variable.Description,
		AppID:         variable.AppID,
		ModuleID:      variable.ModuleID,
		CreatorUserID: variable.CreatorUserID,
		Data:          variable.Data,
		FlowSources:   variable.FlowSources,
		CreatedAt:     variable.CreatedAt,
		UpdatedAt:     variable.UpdatedAt,
		Position:      variable.Position,
	}
}

type MessageInstance struct {
	ID               uint64                   `json:"id"`
	MessageID        string                   `json:"message_id"`
	DiscordGuildID   string                   `json:"discord_guild_id"`
	DiscordChannelID string                   `json:"discord_channel_id"`
	DiscordMessageID string                   `json:"discord_message_id"`
	FlowSources      map[string]flow.FlowData `json:"flow_sources"`
	CreatedAt        time.Time                `json:"created_at"`
	UpdatedAt        time.Time                `json:"updated_at"`
}

type MessageInstanceListResponse = []*MessageInstance

type MessageInstanceCreateRequest struct {
	DiscordGuildID   string `json:"discord_guild_id"`
	DiscordChannelID string `json:"discord_channel_id"`
}

type MessageInstanceCreateResponse = MessageInstance

type MessageInstanceUpdateRequest struct{}

type MessageInstanceUpdateResponse = MessageInstance

type MessageInstanceDeleteResponse = Empty

func MessageInstanceToWire(instance *model.MessageInstance) *MessageInstance {
	if instance == nil {
		return nil
	}

	return &MessageInstance{
		ID:               instance.ID,
		MessageID:        instance.MessageID,
		DiscordGuildID:   instance.DiscordGuildID,
		DiscordChannelID: instance.DiscordChannelID,
		DiscordMessageID: instance.DiscordMessageID,
		FlowSources:      instance.FlowSources,
		CreatedAt:        instance.CreatedAt,
		UpdatedAt:        instance.UpdatedAt,
	}
}

// usedFlowSources drops flow sources that no component in the message references.
func usedFlowSources(data *message.MessageData, flowSources map[string]flow.FlowData) map[string]flow.FlowData {
	res := make(map[string]flow.FlowData, len(flowSources))
	data.EachComponent(func(c *message.ComponentData) error {
		if flow, ok := flowSources[c.FlowSourceID]; ok {
			res[c.FlowSourceID] = flow
		}
		return nil
	})
	return res
}
