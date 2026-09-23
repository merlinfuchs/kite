package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type MessageStore interface {
	MessagesByApp(ctx context.Context, appID string) ([]*model.Message, error)
	CountMessagesByApp(ctx context.Context, appID string) (int, error)
	Message(ctx context.Context, appID string, id string) (*model.Message, error)
	CreateMessage(ctx context.Context, variable *model.Message) (*model.Message, error)
	UpdateMessage(ctx context.Context, variable *model.Message) (*model.Message, error)
	DeleteMessage(ctx context.Context, id string) error
}

type MessageInstanceStore interface {
	MessageInstance(ctx context.Context, appID string, messageID string, instanceID uint64) (*model.MessageInstance, error)
	MessageInstancesByMessage(ctx context.Context, appID string, messageID string) ([]*model.MessageInstance, error)
	// FlowMessageInstancesByMessage returns the newest instances sent by flows,
	// leaving out ephemeral ones since those can't be edited later.
	FlowMessageInstancesByMessage(ctx context.Context, appID string, messageID string, limit int) ([]*model.MessageInstance, error)
	MessageInstanceByDiscordMessageID(ctx context.Context, appID string, discordMessageID string) (*model.MessageInstance, error)
	CreateMessageInstance(ctx context.Context, appID string, instance *model.MessageInstance) (*model.MessageInstance, error)
	UpdateMessageInstance(ctx context.Context, appID string, instance *model.MessageInstance) (*model.MessageInstance, error)
	DeleteMessageInstance(ctx context.Context, appID string, messageID string, instanceID uint64) error
	DeleteMessageInstanceByDiscordMessageID(ctx context.Context, appID string, discordMessageID string) error
}
