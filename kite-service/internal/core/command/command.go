package command

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

const (
	// Commands wait this long after a change, so quick successive changes are
	// deployed together.
	deployDelay = 3 * time.Second
	// Discord limits how often an app can overwrite its commands.
	deployInterval = 30 * time.Second
	// How long to wait before trying again when Discord or the network failed.
	deployRetryInterval = 1 * time.Minute
	deployTimeout       = 15 * time.Second
	maxDeploysPerTick   = 20
	maxParallelDeploys  = 5
)

// DeployError is a problem with an app's commands that its user has to fix.
type DeployError struct {
	Message string
	// Command is the name of the root command the error is about, if known.
	Command string
}

func (e *DeployError) Error() string {
	return e.Message
}

// MarkCommandsChanged queues the app's commands to be deployed in the
// background. Forcing deploys them even if they look unchanged, in case they
// were changed outside of Kite.
func (m *CommandManager) MarkCommandsChanged(ctx context.Context, appID string, force bool) error {
	return m.appStore.MarkAppCommandsChanged(ctx, appID, time.Now().UTC(), force)
}

// Run deploys the commands of apps whose commands changed to Discord.
func (m *CommandManager) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				m.deployPendingApps(ctx)
			}
		}
	}()
}

func (m *CommandManager) deployPendingApps(ctx context.Context) {
	appIDs, err := m.appStore.AppIDsWithPendingCommandsDeploy(ctx, time.Now().Add(-deployDelay), maxDeploysPerTick)
	if err != nil {
		slog.Error("Failed to get apps with pending commands deploy", slog.String("error", err.Error()))
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxParallelDeploys)
	for _, appID := range appIDs {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()

			if err := m.deployCommands(ctx, appID); err != nil {
				slog.Error(
					"Failed to deploy commands",
					slog.String("app_id", appID),
					slog.String("error", err.Error()),
				)
			}
		}()
	}
	wg.Wait()
}

func (m *CommandManager) deployCommands(ctx context.Context, appID string) error {
	// The result is stored with ctx, so it's stored even if the deploy timed out.
	deployCtx, cancel := context.WithTimeout(ctx, deployTimeout)
	defer cancel()

	startedAt := time.Now().UTC()
	res := store.AppCommandsDeployResult{
		ID:        appID,
		StartedAt: startedAt,
		RetryAt:   startedAt.Add(deployInterval),
	}

	err := m.deployCommandsOnce(deployCtx, appID, &res)
	if err != nil {
		var deployErr *DeployError
		if !errors.As(err, &deployErr) {
			// Try again later, the commands may well be fine.
			if delayErr := m.appStore.DelayAppCommandsDeploy(ctx, appID, time.Now().Add(deployRetryInterval)); delayErr != nil {
				return fmt.Errorf("failed to delay commands deploy: %w", delayErr)
			}
			return err
		}

		res.Error = null.StringFrom(deployErr.Message)
		res.ErrorCommand = null.NewString(deployErr.Command, deployErr.Command != "")
	}

	if err := m.appStore.FinishAppCommandsDeploy(ctx, res); err != nil {
		return fmt.Errorf("failed to finish commands deploy: %w", err)
	}
	return nil
}

func (m *CommandManager) deployCommandsOnce(ctx context.Context, appID string, res *store.AppCommandsDeployResult) error {
	app, err := m.appStore.App(ctx, appID)
	if err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}

	commands, err := m.appCommands(ctx, appID)
	if err != nil {
		return err
	}

	cmdData := mergeCommands(commands)

	hash, err := hashCommands(app.DiscordID, cmdData)
	if err != nil {
		return fmt.Errorf("failed to hash commands: %w", err)
	}

	// Most changes only change what a command does, which needs no deploy.
	if app.CommandsDeployedHash.Valid && app.CommandsDeployedHash.String == hash {
		res.DeployedHash = null.StringFrom(hash)
		return nil
	}

	appId, err := strconv.ParseUint(app.DiscordID, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse app ID: %w", err)
	}

	token, err := m.tokenCrypt.DecryptString(app.DiscordToken)
	if err != nil {
		return fmt.Errorf("failed to decrypt token: %w", err)
	}

	client := api.NewClient("Bot " + token).WithContext(ctx)

	_, err = client.BulkOverwriteCommands(discord.AppID(appId), cmdData)
	if err != nil {
		var httpErr *httputil.HTTPError
		if errors.As(err, &httpErr) && httpErr.Status >= 400 && httpErr.Status < 500 && httpErr.Status != http.StatusTooManyRequests {
			return discordDeployError(httpErr, cmdData)
		}
		return fmt.Errorf("failed to deploy commands: %w", err)
	}

	res.DeployedHash = null.StringFrom(hash)

	err = m.commandStore.UpdateCommandsLastDeployedAt(ctx, appID, res.StartedAt)
	if err != nil {
		return fmt.Errorf("failed to update last deployed at: %w", err)
	}

	err = m.pluginInstanceStore.UpdatePluginInstancesLastDeployedAt(ctx, appID, res.StartedAt)
	if err != nil {
		return fmt.Errorf("failed to update last deployed at: %w", err)
	}

	return nil
}

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

	pluginInstances, err := m.pluginInstanceStore.PluginInstancesByApp(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin instances by app: %w", err)
	}

	commandNames := make([]string, 0, len(commands))
	res := make([]api.CreateCommandData, 0, len(commands))
	for _, command := range commands {
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

	for _, pluginInstance := range pluginInstances {
		plugin := m.pluginRegistry.Plugin(pluginInstance.PluginID)
		if plugin == nil {
			slog.Warn("Unknown plugin in deploy manager", slog.String("plugin_id", pluginInstance.PluginID))
			continue
		}

		for _, command := range plugin.Commands() {
			if slices.Contains(pluginInstance.EnabledResourceIDs, command.ID) {
				commandNames = append(commandNames, command.ID)
				res = append(res, command.Data)
			}
		}
	}

	if err := validateCommandNames(commandNames); err != nil {
		return nil, err
	}

	return res, nil
}

func validateCommandNames(commandNames []string) error {
	for a, aName := range commandNames {
		for _, bName := range commandNames[a+1:] {
			if err := commandNameConflict(aName, bName); err != nil {
				return err
			}
		}
	}

	if slices.Contains(commandNames, "") {
		return &DeployError{Message: "A command has no name."}
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
		return &DeployError{
			Message: fmt.Sprintf(
				"/%s and /%s can't both exist, because /%s has subcommands.",
				aName, bName, strings.Join(aParts[:shared], " "),
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

	pluginInstances, err := m.pluginInstanceStore.PluginInstancesByApp(ctx, appID)
	if err != nil {
		return fmt.Errorf("failed to get plugin instances by app: %w", err)
	}

	existing := make([]string, 0, len(commands))
	for _, command := range commands {
		if command.ID != replacedCommandID {
			existing = append(existing, command.Name)
		}
	}
	for _, pluginInstance := range pluginInstances {
		plugin := m.pluginRegistry.Plugin(pluginInstance.PluginID)
		if plugin == nil {
			continue
		}
		for _, command := range plugin.Commands() {
			if slices.Contains(pluginInstance.EnabledResourceIDs, command.ID) {
				existing = append(existing, command.ID)
			}
		}
	}

	for i, name := range names {
		for _, other := range slices.Concat(existing, names[:i]) {
			if err := commandNameConflict(name, other); err != nil {
				return err
			}
		}
	}
	return nil
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
