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
	// Remove unused flow sources
	newFlowSources := make(map[string]flow.FlowData, len(req.FlowSources))
	for _, row := range req.Data.Components {
		for _, comp := range row.Components {
			flow, ok := req.FlowSources[comp.FlowSourceID]
			if ok {
				newFlowSources[comp.FlowSourceID] = flow
			}
		}
	}

	req.FlowSources = newFlowSources
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
	for _, message := range req.Messages {
		message.Sanitize()
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
	// Remove unused flow sources
	newFlowSources := make(map[string]flow.FlowData, len(req.FlowSources))
	for _, row := range req.Data.Components {
		for _, comp := range row.Components {
			flow, ok := req.FlowSources[comp.FlowSourceID]
			if ok {
				newFlowSources[comp.FlowSourceID] = flow
			}
		}
	}

	req.FlowSources = newFlowSources
}

func (req MessageUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&req.Description, validation.Length(0, 255)),
	)
}

type MessageUpdateResponse = Message

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
