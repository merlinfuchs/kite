package variable

import (
	"errors"
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

type VariableHandler struct {
	variableStore      store.VariableStore
	variableValueStore store.VariableValueStore
	maxVariablesPerApp int
}

func NewVariableHandler(variableStore store.VariableStore, variableValueStore store.VariableValueStore, maxVariablesPerApp int) *VariableHandler {
	return &VariableHandler{
		variableStore:      variableStore,
		variableValueStore: variableValueStore,
		maxVariablesPerApp: maxVariablesPerApp,
	}
}

func (h *VariableHandler) HandleVariableList(c *handler.Context) (*wire.VariableListResponse, error) {
	variables, err := h.variableStore.VariablesByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get variables: %w", err)
	}

	res := make([]*wire.Variable, len(variables))
	for i, variable := range variables {
		res[i] = wire.VariableToWire(variable)
	}

	return &res, nil
}

func (h *VariableHandler) HandleVariableGet(c *handler.Context) (*wire.VariableGetResponse, error) {
	return wire.VariableToWire(c.Variabe), nil
}

func (h *VariableHandler) HandleVariableCreate(c *handler.Context, req wire.VariableCreateRequest) (*wire.VariableCreateResponse, error) {
	if h.maxVariablesPerApp != 0 {
		variableCount, err := h.variableStore.CountVariablesByApp(c.Context(), c.App.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to count variables: %w", err)
		}

		if variableCount >= h.maxVariablesPerApp {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of variables (%d) reached", h.maxVariablesPerApp))
		}
	}

	variable, err := h.variableStore.CreateVariable(c.Context(), &model.Variable{
		ID:        util.UniqueID(),
		Name:      req.Name,
		Type:      req.Type,
		Scope:     req.Scope,
		AppID:     c.App.ID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create variable: %w", err)
	}

	return wire.VariableToWire(variable), nil
}

func (h *VariableHandler) HandleVariableUpdate(c *handler.Context, req wire.VariableUpdateRequest) (*wire.VariableUpdateResponse, error) {
	if req.Scope != c.Variabe.Scope || req.Type != c.Variabe.Type {
		// If the scope or type changes, we have to delete all variable values
		err := h.variableValueStore.DeleteAllVariableValues(c.Context(), c.Variabe.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete variable values: %w", err)
		}
	}

	variable, err := h.variableStore.UpdateVariable(c.Context(), &model.Variable{
		ID:        c.Variabe.ID,
		Name:      req.Name,
		Type:      req.Type,
		Scope:     req.Scope,
		AppID:     c.App.ID,
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_variable", "Variable not found")
		}
		return nil, fmt.Errorf("failed to update variable: %w", err)
	}

	return wire.VariableToWire(variable), nil
}

func (h *VariableHandler) HandleVariableDelete(c *handler.Context) (*wire.VariableDeleteResponse, error) {
	err := h.variableStore.DeleteVariable(c.Context(), c.Variabe.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_variable", "Variable not found")
		}
		return nil, fmt.Errorf("failed to delete variable: %w", err)
	}

	return &wire.VariableDeleteResponse{}, nil
}
