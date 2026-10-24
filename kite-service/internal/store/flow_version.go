package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type FlowVersionStore interface {
	CreateFlowVersion(ctx context.Context, version *model.FlowVersion) error
	FlowVersion(ctx context.Context, appID string, id string) (*model.FlowVersion, error)
	// FlowVersionsByCommand returns the newest versions first, without their flow.
	FlowVersionsByCommand(ctx context.Context, appID string, commandID string, limit int) ([]*model.FlowVersion, error)
	// FlowVersionsByEventListener returns the newest versions first, without their flow.
	FlowVersionsByEventListener(ctx context.Context, appID string, eventListenerID string, limit int) ([]*model.FlowVersion, error)
	CountFlowVersionsByCommand(ctx context.Context, commandID string) (int, error)
	CountFlowVersionsByEventListener(ctx context.Context, eventListenerID string) (int, error)
	// DeleteOldFlowVersionsByCommand deletes all but the newest keep versions.
	DeleteOldFlowVersionsByCommand(ctx context.Context, commandID string, keep int) error
	// DeleteOldFlowVersionsByEventListener deletes all but the newest keep versions.
	DeleteOldFlowVersionsByEventListener(ctx context.Context, eventListenerID string, keep int) error
}
