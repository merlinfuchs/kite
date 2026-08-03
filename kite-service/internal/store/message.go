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

// MessageInstanceStore scopes every operation by app.
//
// message_instances rows carry no app_id, so they are only tenant-safe when
// looked up through their message. The engine reaches these from Discord
// interactions, where the caller chose the message being acted on.
type MessageInstanceStore interface {
	MessageInstance(ctx context.Context, appID string, messageID string, instanceID uint64) (*model.MessageInstance, error)
	MessageInstancesByMessage(ctx context.Context, appID string, messageID string, includeHidden bool) ([]*model.MessageInstance, error)
	MessageInstanceByDiscordMessageID(ctx context.Context, appID string, discordMessageID string) (*model.MessageInstance, error)
	CreateMessageInstance(ctx context.Context, instance *model.MessageInstance) (*model.MessageInstance, error)
	UpdateMessageInstance(ctx context.Context, appID string, instance *model.MessageInstance) (*model.MessageInstance, error)
	DeleteMessageInstance(ctx context.Context, appID string, messageID string, instanceID uint64) error
	DeleteMessageInstanceByDiscordMessageID(ctx context.Context, appID string, discordMessageID string) error
}
