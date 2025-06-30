package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type MessageStore interface {
	PublicMessagesByApp(ctx context.Context, appID string) ([]*model.Message, error)
	CountPublicMessagesByApp(ctx context.Context, appID string) (int, error)
	MessagesByCommand(ctx context.Context, commandID string) ([]*model.Message, error)
	MessagesByEventListener(ctx context.Context, eventListenerID string) ([]*model.Message, error)
	Message(ctx context.Context, id string) (*model.Message, error)
	CreateMessage(ctx context.Context, variable *model.Message) (*model.Message, error)
	UpdateMessage(ctx context.Context, variable *model.Message) (*model.Message, error)
	DeleteMessage(ctx context.Context, id string) error
	DeleteMessagesByCommand(ctx context.Context, commandID string) error
	DeleteMessagesByEventListener(ctx context.Context, eventListenerID string) error
}

type MessageInstanceStore interface {
	MessageInstance(ctx context.Context, messageID string, instanceID uint64) (*model.MessageInstance, error)
	MessageInstancesByMessage(ctx context.Context, messageID string, includeHidden bool) ([]*model.MessageInstance, error)
	MessageInstanceByDiscordMessageID(ctx context.Context, discordMessageID string) (*model.MessageInstance, error)
	CreateMessageInstance(ctx context.Context, instance *model.MessageInstance) (*model.MessageInstance, error)
	UpdateMessageInstance(ctx context.Context, instance *model.MessageInstance) (*model.MessageInstance, error)
	DeleteMessageInstance(ctx context.Context, messageID string, instanceID uint64) error
	DeleteMessageInstanceByDiscordMessageID(ctx context.Context, discordMessageID string) error
}
