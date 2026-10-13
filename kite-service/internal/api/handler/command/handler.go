package command

import (
	"errors"
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/command"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

type CommandHandler struct {
	commandStore   store.CommandStore
	commandManager *command.CommandManager
}

func NewCommandHandler(
	commandStore store.CommandStore,
	commandManager *command.CommandManager,
) *CommandHandler {
	return &CommandHandler{
		commandStore:   commandStore,
		commandManager: commandManager,
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
	if c.Features.MaxCommands != 0 {
		commandCount, err := h.commandStore.CountCommandsByApp(c.Context(), c.App.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to count commands: %w", err)
		}

		if commandCount >= c.Features.MaxCommands {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of commands (%d) reached", c.Features.MaxCommands))
		}
	}

	cmdFlow, err := compileCommand(req.FlowSource)
	if err != nil {
		return nil, err
	}

	if err := h.checkCommandNames(c, "", cmdFlow.CommandName()); err != nil {
		return nil, err
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
	if c.Features.MaxCommands != 0 {
		commandCount, err := h.commandStore.CountCommandsByApp(c.Context(), c.App.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to count commands: %w", err)
		}

		newCommandCount := commandCount + len(req.Commands)

		if newCommandCount > c.Features.MaxCommands {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of commands (%d) reached", c.Features.MaxCommands))
		}
	}

	cmdFlows := make([]*flow.CompiledFlowNode, len(req.Commands))
	names := make([]string, len(req.Commands))
	for i, cmd := range req.Commands {
		cmdFlow, err := compileCommand(cmd.FlowSource)
		if err != nil {
			return nil, err
		}
		cmdFlows[i] = cmdFlow
		names[i] = cmdFlow.CommandName()
	}

	if err := h.checkCommandNames(c, "", names...); err != nil {
		return nil, err
	}

	res := make([]*wire.Command, len(req.Commands))
	for i, cmd := range req.Commands {
		cmdFlow := cmdFlows[i]
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
	cmdFlow, err := compileCommand(req.FlowSource)
	if err != nil {
		return nil, err
	}

	// Don't block saving an existing conflict that this change didn't cause.
	if cmdFlow.CommandName() != c.Command.Name {
		if err := h.checkCommandNames(c, c.Command.ID, cmdFlow.CommandName()); err != nil {
			return nil, err
		}
	}

	command, err := h.commandStore.UpdateCommand(c.Context(), &model.Command{
		ID:          c.Command.ID,
		Name:        cmdFlow.CommandName(),
		Description: cmdFlow.CommandDescription(),
		FlowSource:  req.FlowSource,
		Enabled:     c.Command.Enabled,
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

func (h *CommandHandler) HandleCommandUpdateEnabled(c *handler.Context, req wire.CommandUpdateEnabledRequest) (*wire.CommandUpdateEnabledResponse, error) {
	command, err := h.commandStore.UpdateCommand(c.Context(), &model.Command{
		ID:          c.Command.ID,
		Name:        c.Command.Name,
		Description: c.Command.Description,
		FlowSource:  c.Command.FlowSource,
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

// HandleCommandsDeploy deploys the commands again, even if they look unchanged,
// in case they were changed outside of Kite.
func (h *CommandHandler) HandleCommandsDeploy(c *handler.Context) (*wire.CommandsDeployResponse, error) {
	if err := h.commandManager.QueueDeploy(c.Context(), c.App.ID, true); err != nil {
		return nil, fmt.Errorf("failed to queue commands deploy: %w", err)
	}

	return &wire.CommandsDeployResponse{}, nil
}

// compileCommand compiles a command that is being saved, and also rejects
// arguments Discord would reject.
func compileCommand(source flow.FlowData) (*flow.CompiledFlowNode, error) {
	cmdFlow, err := flow.CompileCommand(source)
	if err != nil {
		return nil, handler.ErrBadRequest("invalid_command", err.Error())
	}
	if err := cmdFlow.CheckCommandArguments(); err != nil {
		return nil, handler.ErrBadRequest("invalid_command", err.Error())
	}
	return cmdFlow, nil
}

// checkCommandNames rejects names Discord would reject, which would keep all
// commands of the app from being deployed.
func (h *CommandHandler) checkCommandNames(c *handler.Context, replacedCommandID string, names ...string) error {
	err := h.commandManager.CheckCommandNames(c.Context(), c.App.ID, replacedCommandID, names...)
	if err != nil {
		var deployErr *command.DeployError
		if errors.As(err, &deployErr) {
			return handler.ErrBadRequest("invalid_command_name", deployErr.Message)
		}
		return fmt.Errorf("failed to check command names: %w", err)
	}
	return nil
}
