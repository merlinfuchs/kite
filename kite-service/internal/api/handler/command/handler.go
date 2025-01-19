package command

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

type CommandHandler struct {
	commandStore      store.CommandStore
	maxCommandsPerApp int
}

func NewCommandHandler(commandStore store.CommandStore, maxCommandsPerApp int) *CommandHandler {
	return &CommandHandler{
		commandStore:      commandStore,
		maxCommandsPerApp: maxCommandsPerApp,
	}
}

func (h *CommandHandler) HandleCommandList(c *handler.Context) (*wire.CommandListResponse, error) {
	commands, err := h.commandStore.CommandsByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get commands: %w", err)
	}

	res := make([]*wire.Command, len(commands))
	for i, command := range commands {
		res[i] = wire.CommandToWire(command)
	}

	return &res, nil
}

func (h *CommandHandler) HandleCommandGet(c *handler.Context) (*wire.CommandGetResponse, error) {
	return wire.CommandToWire(c.Command), nil
}

func (h *CommandHandler) HandleCommandCreate(c *handler.Context, req wire.CommandCreateRequest) (*wire.CommandCreateResponse, error) {
	if h.maxCommandsPerApp != 0 {
		commandCount, err := h.commandStore.CountCommandsByApp(c.Context(), c.App.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to count commands: %w", err)
		}

		if commandCount >= h.maxCommandsPerApp {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of commands (%d) reached", h.maxCommandsPerApp))
		}
	}

	cmdFlow, err := flow.CompileCommand(req.FlowSource)
	if err != nil {
		return nil, fmt.Errorf("failed to compile command: %w", err)
	}

	command, err := h.commandStore.CreateCommand(c.Context(), &model.Command{
		ID:            util.UniqueID(),
		Name:          cmdFlow.CommandName(),
		Description:   cmdFlow.CommandDescription(),
		AppID:         c.App.ID,
		CreatorUserID: c.Session.UserID,
		FlowSource:    req.FlowSource,
		Enabled:       req.Enabled,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create command: %w", err)
	}

	return wire.CommandToWire(command), nil
}

func (h *CommandHandler) HandleCommandsImport(c *handler.Context, req wire.CommandsImportRequest) (*wire.CommandsImportResponse, error) {
	if h.maxCommandsPerApp != 0 {
		commandCount, err := h.commandStore.CountCommandsByApp(c.Context(), c.App.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to count commands: %w", err)
		}

		newCommandCount := commandCount + len(req.Commands)

		if newCommandCount > h.maxCommandsPerApp {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of commands (%d) reached", h.maxCommandsPerApp))
		}
	}

	res := make([]*wire.Command, len(req.Commands))
	for i, cmd := range req.Commands {
		cmdFlow, err := flow.CompileCommand(cmd.FlowSource)
		if err != nil {
			return nil, fmt.Errorf("failed to compile command: %w", err)
		}

		command, err := h.commandStore.CreateCommand(c.Context(), &model.Command{
			ID:            util.UniqueID(),
			Name:          cmdFlow.CommandName(),
			Description:   cmdFlow.CommandDescription(),
			AppID:         c.App.ID,
			CreatorUserID: c.Session.UserID,
			FlowSource:    cmd.FlowSource,
			Enabled:       cmd.Enabled,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create command: %w", err)
		}

		res[i] = wire.CommandToWire(command)
	}

	return &res, nil
}

func (h *CommandHandler) HandleCommandUpdate(c *handler.Context, req wire.CommandUpdateRequest) (*wire.CommandUpdateResponse, error) {
	cmdFlow, err := flow.CompileCommand(req.FlowSource)
	if err != nil {
		return nil, fmt.Errorf("failed to compile command: %w", err)
	}

	command, err := h.commandStore.UpdateCommand(c.Context(), &model.Command{
		ID:          c.Command.ID,
		Name:        cmdFlow.CommandName(),
		Description: cmdFlow.CommandDescription(),
		FlowSource:  req.FlowSource,
		Enabled:     req.Enabled,
		UpdatedAt:   time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_command", "Command not found")
		}
		return nil, fmt.Errorf("failed to update command: %w", err)
	}

	return wire.CommandToWire(command), nil
}

func (h *CommandHandler) HandleCommandDelete(c *handler.Context) (*wire.CommandDeleteResponse, error) {
	err := h.commandStore.DeleteCommand(c.Context(), c.Command.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_command", "Command not found")
		}
		return nil, fmt.Errorf("failed to delete command: %w", err)
	}

	return &wire.CommandDeleteResponse{}, nil
}
