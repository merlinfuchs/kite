package function

import (
	"errors"
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

// maxFunctionsPerApp keeps an app from storing an unbounded number of
// functions. Functions only run when a flow calls them, so they don't need a
// limit per plan like commands and event listeners.
const maxFunctionsPerApp = 100

type FunctionHandler struct {
	functionStore store.FunctionStore
}

func NewFunctionHandler(functionStore store.FunctionStore) *FunctionHandler {
	return &FunctionHandler{
		functionStore: functionStore,
	}
}

func (h *FunctionHandler) HandleFunctionList(c *handler.Context) (*wire.FunctionListResponse, error) {
	functions, err := h.functionStore.FunctionsByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get functions: %w", err)
	}

	res := make([]*wire.FlowFunction, len(functions))
	for i, function := range functions {
		res[i] = wire.FunctionToWire(function)
	}

	return &res, nil
}

func (h *FunctionHandler) HandleFunctionGet(c *handler.Context) (*wire.FunctionGetResponse, error) {
	return wire.FunctionToWire(c.Function), nil
}

func (h *FunctionHandler) HandleFunctionCreate(c *handler.Context, req wire.FunctionCreateRequest) (*wire.FunctionCreateResponse, error) {
	entry, err := compileFunction(req.FlowSource)
	if err != nil {
		return nil, err
	}

	count, err := h.functionStore.CountFunctionsByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to count functions: %w", err)
	}
	if count >= maxFunctionsPerApp {
		return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of functions (%d) reached", maxFunctionsPerApp))
	}

	function, err := h.functionStore.CreateFunction(c.Context(), &model.Function{
		ID:            util.UniqueID(),
		Name:          entry.FunctionName(),
		Description:   entry.FunctionDescription(),
		AppID:         c.App.ID,
		CreatorUserID: c.Session.UserID,
		FlowSource:    req.FlowSource,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create function: %w", err)
	}

	return wire.FunctionToWire(function), nil
}

func (h *FunctionHandler) HandleFunctionsImport(c *handler.Context, req wire.FunctionsImportRequest) (*wire.FunctionsImportResponse, error) {
	// Compiled before anything is created, so an invalid function doesn't
	// leave the others behind.
	entries := make([]*flow.CompiledFlowNode, len(req.Functions))
	for i, fn := range req.Functions {
		entry, err := compileFunction(fn.FlowSource)
		if err != nil {
			return nil, err
		}
		entries[i] = entry
	}

	count, err := h.functionStore.CountFunctionsByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to count functions: %w", err)
	}
	if count+len(req.Functions) > maxFunctionsPerApp {
		return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of functions (%d) reached", maxFunctionsPerApp))
	}

	res := make([]*wire.FlowFunction, len(req.Functions))
	for i, fn := range req.Functions {
		function, err := h.functionStore.CreateFunction(c.Context(), &model.Function{
			ID:            util.UniqueID(),
			Name:          entries[i].FunctionName(),
			Description:   entries[i].FunctionDescription(),
			AppID:         c.App.ID,
			CreatorUserID: c.Session.UserID,
			FlowSource:    fn.FlowSource,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create function: %w", err)
		}
		res[i] = wire.FunctionToWire(function)
	}

	return &res, nil
}

func (h *FunctionHandler) HandleFunctionUpdate(c *handler.Context, req wire.FunctionUpdateRequest) (*wire.FunctionUpdateResponse, error) {
	entry, err := compileFunction(req.FlowSource)
	if err != nil {
		return nil, err
	}

	function, err := h.functionStore.UpdateFunction(c.Context(), &model.Function{
		ID:          c.Function.ID,
		AppID:       c.App.ID,
		Name:        entry.FunctionName(),
		Description: entry.FunctionDescription(),
		FlowSource:  req.FlowSource,
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_function", "Function not found")
		}
		return nil, fmt.Errorf("failed to update function: %w", err)
	}

	return wire.FunctionToWire(function), nil
}

func (h *FunctionHandler) HandleFunctionDelete(c *handler.Context) (*wire.FunctionDeleteResponse, error) {
	err := h.functionStore.DeleteFunction(c.Context(), c.App.ID, c.Function.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete function: %w", err)
	}

	return &wire.FunctionDeleteResponse{}, nil
}

func compileFunction(flowSource flow.FlowData) (*flow.CompiledFlowNode, error) {
	entry, err := flow.CompileFunction(flowSource)
	if err != nil {
		return nil, handler.ErrBadRequest("invalid_flow", err.Error())
	}
	return entry, nil
}
