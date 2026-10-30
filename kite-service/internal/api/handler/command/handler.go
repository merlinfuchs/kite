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

// checkContextMenuLimit rejects the added commands if the app would have more
// context menu commands than its plan allows. The command with the ID
// replacedCommandID is left out, because it's being replaced.
func (h *CommandHandler) checkContextMenuLimit(c *handler.Context, replacedCommandID string, added ...command.CommandName) error {
	count := 0
	for _, name := range added {
		if name.IsContextMenu() {
			count++
		}
	}
	if count == 0 || c.Features.MaxContextMenuCommands == 0 {
		return nil
	}

	commands, err := h.commandStore.CommandsByApp(c.Context(), c.App.ID)
	if err != nil {
		return fmt.Errorf("failed to get commands: %w", err)
	}

	for _, cmd := range commands {
		if cmd.ID != replacedCommandID && command.StoredCommandName(cmd).IsContextMenu() {
			count++
		}
	}

	if count > c.Features.MaxContextMenuCommands {
		return handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of context menu commands (%d) reached", c.Features.MaxContextMenuCommands))
	}
	return nil
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

	name := command.NodeCommandName(cmdFlow)
	if err := h.checkCommandNames(c, "", name); err != nil {
		return nil, err
	}

	if err := h.checkContextMenuLimit(c, "", name); err != nil {
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
	commands, err := h.PrepareImport(c, req.Commands)
	if err != nil {
		return nil, err
	}

	if err := h.CreateCommands(c, commands); err != nil {
		return nil, err
	}

	res := make([]*wire.Command, len(commands))
	for i, command := range commands {
		res[i] = wire.CommandToWire(command)
	}

	return &res, nil
}

// PrepareImport checks that the commands can be added to the app and returns
// them with their IDs set, without saving them. Commands whose names are
// taken get a number, like /giveaway-2.
func (h *CommandHandler) PrepareImport(c *handler.Context, reqs []wire.CommandCreateRequest) ([]*model.Command, error) {
	if len(reqs) == 0 {
		return nil, nil
	}

	if c.Features.MaxCommands != 0 {
		commandCount, err := h.commandStore.CountCommandsByApp(c.Context(), c.App.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to count commands: %w", err)
		}

		newCommandCount := commandCount + len(reqs)

		if newCommandCount > c.Features.MaxCommands {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of commands (%d) reached", c.Features.MaxCommands))
		}
	}

	commands := make([]*model.Command, len(reqs))
	names := make([]command.CommandName, len(reqs))
	for i, cmd := range reqs {
		cmdFlow, err := compileCommand(cmd.FlowSource)
		if err != nil {
			return nil, err
		}
		names[i] = command.NodeCommandName(cmdFlow)
		commands[i] = &model.Command{
			ID:            util.UniqueID(),
			Name:          cmdFlow.CommandName(),
			Description:   cmdFlow.CommandDescription(),
			AppID:         c.App.ID,
			CreatorUserID: c.Session.UserID,
			FlowSource:    cmd.FlowSource,
			Enabled:       cmd.Enabled,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}
	}

	if err := h.checkContextMenuLimit(c, "", names...); err != nil {
		return nil, err
	}

	var slashNames []string
	var slashIndexes []int
	for i, name := range names {
		if !name.IsContextMenu() {
			slashNames = append(slashNames, name.Name)
			slashIndexes = append(slashIndexes, i)
		}
	}

	resolved, err := h.commandManager.ResolveCommandNames(c.Context(), c.App.ID, slashNames...)
	if err != nil {
		return nil, commandNamesError(err)
	}
	for i, name := range resolved {
		index := slashIndexes[i]
		if name != commands[index].Name {
			commands[index].Name = name
			commands[index].FlowSource = commands[index].FlowSource.WithCommandName(name)
		}
		names[index].Name = name
	}

	if err := h.checkCommandNames(c, "", names...); err != nil {
		return nil, err
	}

	return commands, nil
}

// CreateCommands saves the commands from PrepareImport.
func (h *CommandHandler) CreateCommands(c *handler.Context, commands []*model.Command) error {
	for i, command := range commands {
		created, err := h.commandStore.CreateCommand(c.Context(), command)
		if err != nil {
			return fmt.Errorf("failed to create command: %w", err)
		}
		commands[i] = created
	}

	return nil
}

func (h *CommandHandler) HandleCommandUpdate(c *handler.Context, req wire.CommandUpdateRequest) (*wire.CommandUpdateResponse, error) {
	updated, err := h.PrepareUpdate(c, c.Command, req.FlowSource)
	if err != nil {
		return nil, err
	}

	change := c.ItemChange(req.EditorSession)
	change.BaseVersion = req.BaseVersion
	command, err := h.commandStore.UpdateCommand(c.Context(), updated, change)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_command", "Command not found")
		}
		if errors.Is(err, store.ErrOutdated) {
			return nil, handler.ErrItemOutdated()
		}
		return nil, fmt.Errorf("failed to update command: %w", err)
	}

	return wire.CommandToWire(command), nil
}

// PrepareUpdate checks that the command can be changed to the flow and returns
// it changed, without saving it.
func (h *CommandHandler) PrepareUpdate(c *handler.Context, existingCommand *model.Command, flowSource flow.FlowData) (*model.Command, error) {
	cmdFlow, err := compileCommand(flowSource)
	if err != nil {
		return nil, err
	}

	// Don't block saving an existing conflict that this change didn't cause.
	name := command.NodeCommandName(cmdFlow)
	oldName := command.StoredCommandName(existingCommand)
	if name != oldName {
		if err := h.checkCommandNames(c, existingCommand.ID, name); err != nil {
			return nil, err
		}
	}

	if name.IsContextMenu() && !oldName.IsContextMenu() {
		if err := h.checkContextMenuLimit(c, existingCommand.ID, name); err != nil {
			return nil, err
		}
	}

	updated := *existingCommand
	updated.Name = cmdFlow.CommandName()
	updated.Description = cmdFlow.CommandDescription()
	updated.FlowSource = flowSource
	updated.UpdatedAt = time.Now().UTC()
	return &updated, nil
}

func (h *CommandHandler) HandleCommandUpdateEnabled(c *handler.Context, req wire.CommandUpdateEnabledRequest) (*wire.CommandUpdateEnabledResponse, error) {
	command, err := h.commandStore.UpdateCommandEnabled(c.Context(), c.Command.ID, req.Enabled, time.Now().UTC())
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_command", "Command not found")
		}
		return nil, fmt.Errorf("failed to update command: %w", err)
	}

	return wire.CommandToWire(command), nil
}

func (h *CommandHandler) HandleCommandDelete(c *handler.Context) (*wire.CommandDeleteResponse, error) {
	err := h.commandStore.DeleteCommand(c.Context(), c.Command.ID, c.Session.UserID)
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
func (h *CommandHandler) checkCommandNames(c *handler.Context, replacedCommandID string, names ...command.CommandName) error {
	return commandNamesError(h.commandManager.CheckCommandNames(c.Context(), c.App.ID, replacedCommandID, names...))
}

// commandNamesError turns a conflict of command names into an error for the
// user.
func commandNamesError(err error) error {
	if err == nil {
		return nil
	}
	var deployErr *command.DeployError
	if errors.As(err, &deployErr) {
		return handler.ErrBadRequest("invalid_command_name", deployErr.Message)
	}
	return fmt.Errorf("failed to check command names: %w", err)
}
