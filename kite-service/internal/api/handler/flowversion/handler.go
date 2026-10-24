package flowversion

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

// MaxVersions is how many versions are kept per flow. Auto-save makes a
// version every few seconds of editing, so older ones are dropped.
const MaxVersions = 50

type FlowVersionHandler struct {
	flowVersionStore store.FlowVersionStore
}

func NewFlowVersionHandler(flowVersionStore store.FlowVersionStore) *FlowVersionHandler {
	return &FlowVersionHandler{
		flowVersionStore: flowVersionStore,
	}
}

// Target is the command or event listener a flow belongs to, one of the IDs
// is set.
type Target struct {
	CommandID       null.String
	EventListenerID null.String
}

func CommandTarget(commandID string) Target {
	return Target{CommandID: null.StringFrom(commandID)}
}

func EventListenerTarget(eventListenerID string) Target {
	return Target{EventListenerID: null.StringFrom(eventListenerID)}
}

// Record adds the saved flow to the version history. The flow from before the
// save is added first if the flow has no versions yet, so the first save
// after this feature can be undone too. It's called after the save succeeded
// and only logs errors, a missing version shouldn't fail the save.
func (h *FlowVersionHandler) Record(
	ctx context.Context,
	appID string,
	target Target,
	userID string,
	previous flow.FlowData,
	previousAt time.Time,
	saved flow.FlowData,
	autoSaved bool,
) {
	if err := h.record(ctx, appID, target, userID, previous, previousAt, saved, autoSaved); err != nil {
		slog.With("error", err).With("app_id", appID).Error("Failed to record flow version")
	}
}

func (h *FlowVersionHandler) record(
	ctx context.Context,
	appID string,
	target Target,
	userID string,
	previous flow.FlowData,
	previousAt time.Time,
	saved flow.FlowData,
	autoSaved bool,
) error {
	// Saving without changes, e.g. pressing Ctrl+S twice, adds nothing.
	unchanged, err := sameFlow(previous, saved)
	if err != nil {
		return fmt.Errorf("failed to compare flows: %w", err)
	}
	if unchanged {
		return nil
	}

	count, err := h.count(ctx, target)
	if err != nil {
		return fmt.Errorf("failed to count flow versions: %w", err)
	}

	if count == 0 {
		err := h.flowVersionStore.CreateFlowVersion(ctx, &model.FlowVersion{
			ID:              util.UniqueID(),
			AppID:           appID,
			CommandID:       target.CommandID,
			EventListenerID: target.EventListenerID,
			FlowSource:      previous,
			CreatedAt:       previousAt,
		})
		if err != nil {
			return fmt.Errorf("failed to create initial flow version: %w", err)
		}
	}

	err = h.flowVersionStore.CreateFlowVersion(ctx, &model.FlowVersion{
		ID:              util.UniqueID(),
		AppID:           appID,
		CommandID:       target.CommandID,
		EventListenerID: target.EventListenerID,
		FlowSource:      saved,
		AutoSaved:       autoSaved,
		CreatorUserID:   null.StringFrom(userID),
		CreatedAt:       time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("failed to create flow version: %w", err)
	}

	if target.CommandID.Valid {
		err = h.flowVersionStore.DeleteOldFlowVersionsByCommand(ctx, target.CommandID.String, MaxVersions)
	} else {
		err = h.flowVersionStore.DeleteOldFlowVersionsByEventListener(ctx, target.EventListenerID.String, MaxVersions)
	}
	if err != nil {
		return fmt.Errorf("failed to delete old flow versions: %w", err)
	}

	return nil
}

func (h *FlowVersionHandler) count(ctx context.Context, target Target) (int, error) {
	if target.CommandID.Valid {
		return h.flowVersionStore.CountFlowVersionsByCommand(ctx, target.CommandID.String)
	}
	return h.flowVersionStore.CountFlowVersionsByEventListener(ctx, target.EventListenerID.String)
}

func sameFlow(a, b flow.FlowData) (bool, error) {
	rawA, err := json.Marshal(a)
	if err != nil {
		return false, err
	}
	rawB, err := json.Marshal(b)
	if err != nil {
		return false, err
	}
	return bytes.Equal(rawA, rawB), nil
}

func (h *FlowVersionHandler) HandleCommandFlowVersionList(c *handler.Context) (*wire.FlowVersionListResponse, error) {
	versions, err := h.flowVersionStore.FlowVersionsByCommand(c.Context(), c.App.ID, c.Command.ID, MaxVersions)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow versions: %w", err)
	}

	return versionsToWire(versions), nil
}

func (h *FlowVersionHandler) HandleCommandFlowVersionGet(c *handler.Context) (*wire.FlowVersionGetResponse, error) {
	version, err := h.flowVersion(c)
	if err != nil {
		return nil, err
	}

	if version.CommandID.String != c.Command.ID {
		return nil, handler.ErrNotFound("unknown_flow_version", "Flow version not found")
	}

	return versionToGetResponse(version), nil
}

func (h *FlowVersionHandler) HandleEventListenerFlowVersionList(c *handler.Context) (*wire.FlowVersionListResponse, error) {
	versions, err := h.flowVersionStore.FlowVersionsByEventListener(c.Context(), c.App.ID, c.EventListener.ID, MaxVersions)
	if err != nil {
		return nil, fmt.Errorf("failed to get flow versions: %w", err)
	}

	return versionsToWire(versions), nil
}

func (h *FlowVersionHandler) HandleEventListenerFlowVersionGet(c *handler.Context) (*wire.FlowVersionGetResponse, error) {
	version, err := h.flowVersion(c)
	if err != nil {
		return nil, err
	}

	if version.EventListenerID.String != c.EventListener.ID {
		return nil, handler.ErrNotFound("unknown_flow_version", "Flow version not found")
	}

	return versionToGetResponse(version), nil
}

func (h *FlowVersionHandler) flowVersion(c *handler.Context) (*model.FlowVersion, error) {
	version, err := h.flowVersionStore.FlowVersion(c.Context(), c.App.ID, c.Param("versionID"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_flow_version", "Flow version not found")
		}
		return nil, fmt.Errorf("failed to get flow version: %w", err)
	}

	return version, nil
}

func versionsToWire(versions []*model.FlowVersion) *wire.FlowVersionListResponse {
	res := make([]*wire.FlowVersion, len(versions))
	for i, version := range versions {
		res[i] = wire.FlowVersionToWire(version)
	}
	return &res
}

func versionToGetResponse(version *model.FlowVersion) *wire.FlowVersionGetResponse {
	return &wire.FlowVersionGetResponse{
		ID:            version.ID,
		AutoSaved:     version.AutoSaved,
		CreatorUserID: version.CreatorUserID,
		FlowSource:    version.FlowSource,
		CreatedAt:     version.CreatedAt,
	}
}
