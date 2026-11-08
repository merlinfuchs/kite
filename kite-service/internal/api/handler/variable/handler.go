package variable

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"gopkg.in/guregu/null.v4"
)

type VariableHandler struct {
	variableStore      store.VariableStore
	variableValueStore store.VariableValueStore
}

func NewVariableHandler(variableStore store.VariableStore, variableValueStore store.VariableValueStore) *VariableHandler {
	return &VariableHandler{
		variableStore:      variableStore,
		variableValueStore: variableValueStore,
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
	return wire.VariableToWire(c.Variable), nil
}

func (h *VariableHandler) HandleVariableCreate(c *handler.Context, req wire.VariableCreateRequest) (*wire.VariableCreateResponse, error) {
	if c.Features.MaxVariables != 0 {
		variableCount, err := h.variableStore.CountVariablesByApp(c.Context(), c.App.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to count variables: %w", err)
		}

		if variableCount >= c.Features.MaxVariables {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of variables (%d) reached", c.Features.MaxVariables))
		}
	}

	variable, err := h.variableStore.CreateVariable(c.Context(), &model.Variable{
		ID:        util.UniqueID(),
		Name:      req.Name,
		Scoped:    req.Scoped,
		AppID:     c.App.ID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create variable: %w", err)
	}

	return wire.VariableToWire(variable), nil
}

// PrepareImport checks that the variables can be added to the app and returns
// them with their IDs set, without saving them. Names the app already uses get
// a number, as a name can only be used once.
func (h *VariableHandler) PrepareImport(c *handler.Context, reqs []wire.VariableCreateRequest) ([]*model.Variable, error) {
	if len(reqs) == 0 {
		return nil, nil
	}

	existing, err := h.variableStore.VariablesByAppWithoutTotals(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get variables: %w", err)
	}
	if c.Features.MaxVariables != 0 && len(existing)+len(reqs) > c.Features.MaxVariables {
		return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of variables (%d) reached", c.Features.MaxVariables))
	}

	taken := make(map[string]bool, len(existing)+len(reqs))
	for _, v := range existing {
		taken[v.Name] = true
	}

	variables := make([]*model.Variable, len(reqs))
	for i, req := range reqs {
		if err := req.Validate(); err != nil {
			return nil, err
		}
		name := req.Name
		for n := 2; taken[name]; n++ {
			suffix := fmt.Sprintf("_%d", n)
			name = req.Name[:min(len(req.Name), 100-len(suffix))] + suffix
		}
		taken[name] = true

		variables[i] = &model.Variable{
			ID:        util.UniqueID(),
			Name:      name,
			Scoped:    req.Scoped,
			AppID:     c.App.ID,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
	}

	return variables, nil
}

func (h *VariableHandler) HandleVariablesImport(c *handler.Context, req wire.VariablesImportRequest) (*wire.VariablesImportResponse, error) {
	variables, err := h.PrepareImport(c, req.Variables)
	if err != nil {
		return nil, err
	}

	res := make([]*wire.Variable, len(variables))
	for i, variable := range variables {
		created, err := h.variableStore.CreateVariable(c.Context(), variable)
		if err != nil {
			return nil, fmt.Errorf("failed to create variable: %w", err)
		}
		res[i] = wire.VariableToWire(created)
	}

	return &res, nil
}

func (h *VariableHandler) HandleVariableUpdate(c *handler.Context, req wire.VariableUpdateRequest) (*wire.VariableUpdateResponse, error) {
	if req.Scoped != c.Variable.Scoped {
		// If the scoped flag changes, delete all variable values.
		err := h.variableValueStore.DeleteAllVariableValues(c.Context(), c.App.ID, c.Variable.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete variable values: %w", err)
		}
	}

	variable, err := h.variableStore.UpdateVariable(c.Context(), &model.Variable{
		ID:        c.Variable.ID,
		Name:      req.Name,
		Scoped:    req.Scoped,
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
	err := h.variableStore.DeleteVariable(c.Context(), c.Variable.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_variable", "Variable not found")
		}
		return nil, fmt.Errorf("failed to delete variable: %w", err)
	}

	return &wire.VariableDeleteResponse{}, nil
}

func (h *VariableHandler) HandleVariableValueList(c *handler.Context) (*wire.VariableValueListResponse, error) {
	limit := 25
	if raw := c.Query("limit"); raw != "" {
		var err error
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 {
			return nil, handler.ErrBadRequest("invalid_limit", "limit must be a number from 1 to 100")
		}
	}

	offset := 0
	if raw := c.Query("offset"); raw != "" {
		var err error
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 {
			return nil, handler.ErrBadRequest("invalid_offset", "offset must be a number of at least 0")
		}
	}

	search := strings.TrimSpace(c.Query("search"))
	if len(search) > wire.MaxVariableScopeLength {
		return nil, handler.ErrBadRequest("invalid_search", "search is too long")
	}

	values, total, err := h.variableValueStore.SearchVariableValues(c.Context(), c.App.ID, c.Variable.ID, search, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get variable values: %w", err)
	}

	res := make([]*wire.VariableValue, len(values))
	for i, value := range values {
		res[i] = wire.VariableValueToWire(value)
	}

	return &wire.VariableValueListResponse{
		Values: res,
		Total:  total,
	}, nil
}

func (h *VariableHandler) HandleVariableValueSet(c *handler.Context, req wire.VariableValueSetRequest) (*wire.VariableValueSetResponse, error) {
	if !c.Variable.Scoped && req.Scope != "" {
		return nil, validation.Errors{"scope": errors.New("this variable is not scoped")}
	}
	if c.Variable.Scoped && req.Scope == "" {
		return nil, validation.Errors{"scope": errors.New("cannot be blank")}
	}

	data, err := req.Thing()
	if err != nil {
		return nil, validation.Errors{"value": err}
	}

	// An empty scope is stored as NULL, the same as flows do.
	value, err := h.variableValueStore.UpdateVariableValue(c.Context(), c.App.ID, provider.VariableOperationOverwrite, model.VariableValue{
		VariableID: c.Variable.ID,
		Scope:      null.NewString(req.Scope, req.Scope != ""),
		Data:       data,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_variable", "Variable not found")
		}
		return nil, fmt.Errorf("failed to set variable value: %w", err)
	}

	return wire.VariableValueToWire(value), nil
}

func (h *VariableHandler) HandleVariableValueDelete(c *handler.Context) (*wire.VariableValueDeleteResponse, error) {
	scope := c.Query("scope")

	err := h.variableValueStore.DeleteVariableValue(c.Context(), c.App.ID, c.Variable.ID, null.NewString(scope, scope != ""))
	if err != nil {
		return nil, fmt.Errorf("failed to delete variable value: %w", err)
	}

	return &wire.VariableValueDeleteResponse{}, nil
}
