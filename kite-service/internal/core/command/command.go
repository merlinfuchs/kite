package command

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"golang.org/x/sync/errgroup"
	"gopkg.in/guregu/null.v4"
)

const (
	deployTimeout      = 15 * time.Second
	maxDeploysPerBatch = 20
	maxParallelDeploys = 5
)

// QueueDeploy deploys the app's commands again in a few seconds. Changes to
// commands and plugins do so by themselves. Forcing sends them even if they
// look unchanged, in case they were changed outside of Kite.
func (m *CommandManager) QueueDeploy(ctx context.Context, appID string, force bool) error {
	return m.appStore.QueueAppCommandsDeploy(ctx, appID, force)
}

// Run deploys the commands of apps whose commands changed to Discord.
func (m *CommandManager) Run(ctx context.Context) {
	go func() {
		for ctx.Err() == nil {
			deployed := m.deployDueApps(ctx)
			// Keep going while there's a backlog.
			if deployed == maxDeploysPerBatch {
				continue
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
		}
	}()
}

func (m *CommandManager) deployDueApps(ctx context.Context) int {
	appIDs, err := m.appStore.AppIDsWithDueCommandsDeploy(ctx, maxDeploysPerBatch)
	if err != nil {
		if ctx.Err() == nil {
			slog.Error("Failed to get apps with due commands deploy", slog.String("error", err.Error()))
		}
		return 0
	}

	var g errgroup.Group
	g.SetLimit(maxParallelDeploys)
	for _, appID := range appIDs {
		g.Go(func() error {
			if err := m.deployCommands(ctx, appID); err != nil {
				slog.Error(
					"Failed to deploy commands",
					slog.String("app_id", appID),
					slog.String("error", err.Error()),
				)
			}
			return nil
		})
	}
	g.Wait()

	return len(appIDs)
}

func (m *CommandManager) deployCommands(ctx context.Context, appID string) error {
	app, err := m.appStore.App(ctx, appID)
	if err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}
	if !app.CommandsDeployAt.Valid {
		return nil
	}

	// The result is stored with ctx, so it's stored even if the deploy timed out.
	deployCtx, cancel := context.WithTimeout(ctx, deployTimeout)
	defer cancel()

	res := store.AppCommandsDeployResult{ID: appID, DeployAt: app.CommandsDeployAt.Time}

	hash, err := m.sendCommands(deployCtx, app)
	var deployErr *DeployError
	switch {
	case errors.As(err, &deployErr):
		res.Error = deployErr
	case err != nil:
		// The commands may well be fine, so try again later.
		delayErr := m.appStore.DelayAppCommandsDeploy(ctx, appID, &DeployError{
			Message: "Discord didn't respond. Kite tries again in a few minutes.",
		})
		if delayErr != nil {
			return fmt.Errorf("failed to delay commands deploy: %w", delayErr)
		}
		return err
	default:
		res.DeployedHash = null.StringFrom(hash)
	}

	if err := m.appStore.FinishAppCommandsDeploy(ctx, res); err != nil {
		return fmt.Errorf("failed to finish commands deploy: %w", err)
	}
	return nil
}

// sendCommands sends the app's commands to Discord, unless they're unchanged
// since the last time, and returns their hash.
func (m *CommandManager) sendCommands(ctx context.Context, app *model.App) (string, error) {
	commands, err := m.appCommands(ctx, app.ID)
	if err != nil {
		return "", err
	}

	cmdData := mergeCommands(commands)

	hash, err := hashCommands(app.DiscordID, cmdData)
	if err != nil {
		return "", fmt.Errorf("failed to hash commands: %w", err)
	}

	// Most changes only change what a command does, which needs no deploy.
	if app.CommandsDeployedHash.Valid && app.CommandsDeployedHash.String == hash {
		return hash, nil
	}

	appID, err := strconv.ParseUint(app.DiscordID, 10, 64)
	if err != nil {
		return "", &DeployError{Message: "The app has an invalid Discord ID."}
	}

	token, err := m.tokenCrypt.DecryptString(app.DiscordToken)
	if err != nil {
		return "", &DeployError{Message: "Kite couldn't read the bot token. Update it in the settings."}
	}

	client := api.NewClient("Bot " + token).WithContext(ctx)

	err = overwriteCommands(client, discord.AppID(appID), cmdData)
	if err != nil {
		var httpErr *httputil.HTTPError
		if errors.As(err, &httpErr) && isRejection(httpErr.Status) {
			return "", discordDeployError(httpErr, cmdData)
		}
		return "", fmt.Errorf("failed to deploy commands: %w", err)
	}

	return hash, nil
}

// errCodeEntryPointRemoved is Discord's error for leaving out an app's Entry
// Point command, which launches its Activity, from a bulk update.
const errCodeEntryPointRemoved = 50240

// overwriteCommands replaces the app's commands in Discord. Kite doesn't manage
// Entry Point commands, so an app that has one keeps it.
func overwriteCommands(client *api.Client, appID discord.AppID, cmdData []api.CreateCommandData) error {
	_, err := client.BulkOverwriteCommands(appID, cmdData)
	var httpErr *httputil.HTTPError
	if !errors.As(err, &httpErr) || httpErr.Code != errCodeEntryPointRemoved {
		return err
	}

	// The library doesn't know Entry Point commands, so they're passed on as
	// Discord sent them.
	url := api.EndpointApplications + appID.String() + "/commands"
	var existing []json.RawMessage
	if err := client.RequestJSON(&existing, "GET", url); err != nil {
		return fmt.Errorf("failed to get commands: %w", err)
	}

	body := make([]any, 0, len(cmdData)+1)
	for _, command := range cmdData {
		body = append(body, command)
	}
	for _, raw := range existing {
		var command struct {
			Type discord.CommandType `json:"type"`
		}
		if err := json.Unmarshal(raw, &command); err == nil && command.Type == entryPointCommandType {
			body = append(body, raw)
		}
	}

	return client.RequestJSON(nil, "PUT", url, httputil.WithJSONBody(body))
}

// entryPointCommandType is the type of Entry Point commands, which the library
// has no constant for.
const entryPointCommandType discord.CommandType = 4

func hashCommands(discordAppID string, cmdData []api.CreateCommandData) (string, error) {
	raw, err := json.Marshal(cmdData)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write([]byte(discordAppID))
	h.Write(raw)
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (m *CommandManager) appCommands(ctx context.Context, appID string) ([]api.CreateCommandData, error) {
	commands, err := m.commandStore.CommandsByApp(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get commands by app: %w", err)
	}

	pluginCommands, err := m.pluginCommands(ctx, appID)
	if err != nil {
		return nil, err
	}

	commandNames := make([]string, 0, len(commands))
	res := make([]api.CreateCommandData, 0, len(commands))
	for _, command := range commands {
		// Disabled commands aren't run, so they shouldn't show up in Discord.
		if !command.Enabled {
			continue
		}

		node, err := flow.CompileCommand(command.FlowSource)
		if err != nil {
			return nil, &DeployError{
				Message: fmt.Sprintf("The command can't be compiled: %v", err),
				Command: strings.Split(command.Name, " ")[0],
			}
		}

		data := node.CommandData()
		res = append(res, api.CreateCommandData{
			Name:                     data.Name,
			Description:              data.Description,
			Options:                  data.Options,
			DefaultMemberPermissions: data.DefaultMemberPermissions,
			Contexts:                 node.CommandContexts(),
			IntegrationTypes:         node.CommandIntegrations(),
		})
		commandNames = append(commandNames, node.CommandName())
	}

	for _, command := range pluginCommands {
		commandNames = append(commandNames, command.Name)
		res = append(res, command)
	}

	if err := checkCommandNames(nil, commandNames); err != nil {
		return nil, err
	}

	return res, nil
}

// pluginCommands returns the commands the app's plugins add.
func (m *CommandManager) pluginCommands(ctx context.Context, appID string) ([]api.CreateCommandData, error) {
	pluginInstances, err := m.pluginInstanceStore.PluginInstancesByApp(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin instances by app: %w", err)
	}

	var res []api.CreateCommandData
	for _, pluginInstance := range pluginInstances {
		plugin := m.pluginRegistry.Plugin(pluginInstance.PluginID)
		if plugin == nil {
			slog.Warn("Unknown plugin in deploy manager", slog.String("plugin_id", pluginInstance.PluginID))
			continue
		}

		for _, command := range plugin.Commands() {
			if slices.Contains(pluginInstance.EnabledResourceIDs, command.ID) {
				res = append(res, command.Data)
			}
		}
	}
	return res, nil
}

// checkCommandNames reports whether the added names conflict with the
// existing ones or each other in a way Discord rejects.
func checkCommandNames(existing []string, added []string) error {
	for i, name := range added {
		if name == "" {
			return &DeployError{Message: "A command has no name."}
		}
		for _, other := range slices.Concat(existing, added[:i]) {
			if err := commandNameConflict(name, other); err != nil {
				return err
			}
		}
	}
	return nil
}

// commandNameConflict reports whether Discord rejects two commands with these
// names. Discord only allows subcommands on commands that do nothing
// themselves.
func commandNameConflict(aName, bName string) *DeployError {
	if aName == "" || bName == "" {
		return nil
	}

	aParts := strings.Split(aName, " ")
	bParts := strings.Split(bName, " ")

	shared := 0
	for shared < len(aParts) && shared < len(bParts) && aParts[shared] == bParts[shared] {
		shared++
	}

	switch {
	case shared == len(aParts) && shared == len(bParts):
		return &DeployError{
			Message: fmt.Sprintf("There is more than one /%s command.", aName),
			Command: aParts[0],
		}
	case shared == len(aParts) || shared == len(bParts):
		parent, child := aName, bName
		if len(bParts) < len(aParts) {
			parent, child = bName, aName
		}
		return &DeployError{
			Message: fmt.Sprintf(
				"/%s and /%s can't both exist, because /%s has subcommands.",
				parent, child, parent,
			),
			Command: aParts[0],
		}
	}
	return nil
}

// CheckCommandNames reports whether commands with these names conflict with
// each other or the other commands of the app. The command with the ID
// replacedCommandID is left out, because it's being replaced.
func (m *CommandManager) CheckCommandNames(ctx context.Context, appID string, replacedCommandID string, names ...string) error {
	commands, err := m.commandStore.CommandsByApp(ctx, appID)
	if err != nil {
		return fmt.Errorf("failed to get commands by app: %w", err)
	}

	pluginCommands, err := m.pluginCommands(ctx, appID)
	if err != nil {
		return err
	}

	existing := make([]string, 0, len(commands)+len(pluginCommands))
	for _, command := range commands {
		if command.ID != replacedCommandID {
			existing = append(existing, command.Name)
		}
	}
	for _, command := range pluginCommands {
		existing = append(existing, command.Name)
	}

	return checkCommandNames(existing, names)
}

// mergeCommands merges the subcommands of the same root command into one
// command. The order stays the same for the same commands, so their hash does
// too.
func mergeCommands(commands []api.CreateCommandData) []api.CreateCommandData {
	res := make([]api.CreateCommandData, 0, len(commands))
	rootIndexes := make(map[string]int)

	// Merge root commands
	for _, command := range commands {
		// TODO: think about how to handle different configs for root cmd
		if i, ok := rootIndexes[command.Name]; ok {
			res[i].Options = append(res[i].Options, command.Options...)
		} else {
			rootIndexes[command.Name] = len(res)
			// Don't append to the options of the given command.
			command.Options = slices.Clone(command.Options)
			res = append(res, command)
		}
	}

	// Merge sub command groups
	for i := range res {
		command := &res[i]
		groups := make(map[string]*discord.SubcommandGroupOption)
		options := make([]discord.CommandOption, 0, len(command.Options))

		for _, option := range command.Options {
			if g, ok := option.(*discord.SubcommandGroupOption); ok {
				if group, ok := groups[g.Name()]; ok {
					group.Subcommands = append(group.Subcommands, g.Subcommands...)
					continue
				}
				merged := *g
				merged.Subcommands = slices.Clone(g.Subcommands)
				groups[g.Name()] = &merged
				option = &merged
			}
			options = append(options, option)
		}

		command.Options = options
	}

	return res
}
