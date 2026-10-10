package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type CreditLimitStore interface {
	CreditLimitsByApp(ctx context.Context, appID string) ([]*model.CreditLimit, error)
	CreditLimit(ctx context.Context, appID string, id string) (*model.CreditLimit, error)
	CountCreditLimitsByApp(ctx context.Context, appID string) (int, error)
	// CreateCreditLimit returns ErrAlreadyExists if the app already has a limit
	// for the same scope, target and period.
	CreateCreditLimit(ctx context.Context, limit *model.CreditLimit) (*model.CreditLimit, error)
	UpdateCreditLimit(ctx context.Context, limit *model.CreditLimit) (*model.CreditLimit, error)
	DeleteCreditLimit(ctx context.Context, appID string, id string) error

	// CreditLimitSettings returns ErrNotFound if the app never saved any.
	CreditLimitSettings(ctx context.Context, appID string) (*model.CreditLimitSettings, error)
	UpsertCreditLimitSettings(ctx context.Context, settings *model.CreditLimitSettings) (*model.CreditLimitSettings, error)
}
