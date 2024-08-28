package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type MessageStore interface {
	MessagesByApp(ctx context.Context, appID string) ([]*model.Message, error)
	CountMessagesByApp(ctx context.Context, appID string) (int, error)
	Message(ctx context.Context, id string) (*model.Message, error)
	CreateMessage(ctx context.Context, variable *model.Message) (*model.Message, error)
	UpdateMessage(ctx context.Context, variable *model.Message) (*model.Message, error)
	DeleteMessage(ctx context.Context, id string) error
}
