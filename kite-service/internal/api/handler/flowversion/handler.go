package flowversion

import (
	"errors"
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

// FlowVersionHandler lists the version history of flows. The stores record
// the versions when a flow is changed.
type FlowVersionHandler struct {
	flowVersionStore store.FlowVersionStore
}

func NewFlowVersionHandler(flowVersionStore store.FlowVersionStore) *FlowVersionHandler {
	return &FlowVersionHandler{
		flowVersionStore: flowVersionStore,
	}
}

func (h *FlowVersionHandler) HandleCommandFlowVersionList(c *handler.Context) (*wire.FlowVersionListResponse, error) {
	return h.list(c, c.Command.ID)
}

func (h *FlowVersionHandler) HandleCommandFlowVersionGet(c *handler.Context) (*wire.FlowVersionGetResponse, error) {
	return h.get(c, c.Command.ID)
}

func (h *FlowVersionHandler) HandleEventListenerFlowVersionList(c *handler.Context) (*wire.FlowVersionListResponse, error) {
	return h.list(c, c.EventListener.ID)
}

func (h *FlowVersionHandler) HandleEventListenerFlowVersionGet(c *handler.Context) (*wire.FlowVersionGetResponse, error) {
	return h.get(c, c.EventListener.ID)
}

func (h *FlowVersionHandler) list(c *handler.Context, targetID string) (*wire.FlowVersionListResponse, error) {
	versions, err := h.flowVersionStore.FlowVersionsByTarget(c.Context(), c.App.ID, targetID, model.MaxFlowVersions)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow versions: %w", err)
	}

	res := make([]*wire.FlowVersion, len(versions))
	for i, version := range versions {
		res[i] = wire.FlowVersionToWire(version)
	}
	return &res, nil
}

func (h *FlowVersionHandler) get(c *handler.Context, targetID string) (*wire.FlowVersionGetResponse, error) {
	version, err := h.flowVersionStore.FlowVersion(c.Context(), c.App.ID, c.Param("versionID"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_flow_version", "Flow version not found")
		}
		return nil, fmt.Errorf("failed to get flow version: %w", err)
	}
	if version.TargetID != targetID {
		return nil, handler.ErrNotFound("unknown_flow_version", "Flow version not found")
	}

	return &wire.FlowVersionGetResponse{
		ID:            version.ID,
		Source:        string(version.Source),
		CreatorUserID: version.CreatorUserID,
		FlowSource:    version.FlowSource,
		CreatedAt:     version.CreatedAt,
	}, nil
}
