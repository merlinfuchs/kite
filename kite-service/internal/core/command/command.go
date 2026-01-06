package command

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

func (m *CommandManager) DeployCommandsForApp(ctx context.Context, appID string) error {
	deploymentStartedAt := time.Now().UTC()

	commands, err := m.appCommands(ctx, appID)
	if err != nil {
		return fmt.Errorf("failed to get commands for app: %w", err)
	}

	cmdData, err := mergeCommands(commands)
	if err != nil {
		return fmt.Errorf("failed to merge commands: %w", err)
	}

	app, err := m.appStore.App(ctx, appID)
	if err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}

	if !app.Enabled {
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
		return fmt.Errorf("failed to deploy commands: %w", err)
	}

	err = m.commandStore.UpdateCommandsLastDeployedAt(ctx, appID, deploymentStartedAt)
	if err != nil {
		return fmt.Errorf("failed to update last deployed at: %w", err)
	}

	err = m.pluginInstanceStore.UpdatePluginInstancesLastDeployedAt(ctx, appID, deploymentStartedAt)
	if err != nil {
		return fmt.Errorf("failed to update last deployed at: %w", err)
	}

	return nil
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
			return nil, fmt.Errorf("failed to compile command flow: %w", err)
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
		return nil, fmt.Errorf("invalid command names: %w", err)
	}

	return res, nil
}

func validateCommandNames(commandNames []string) error {
	for a, aName := range commandNames {
		if len(aName) == 0 {
			return fmt.Errorf("empty command name")
		}

		aParts := strings.Split(aName, " ")

		for b, bName := range commandNames {
			if a == b {
				continue
			}

			if len(bName) == 0 {
				return fmt.Errorf("empty command name")
			}

			bParts := strings.Split(bName, " ")

			if aParts[0] == bParts[0] {
				if len(aParts) == 1 || len(bParts) == 1 {
					return fmt.Errorf("duplicate command name: %s", aName)
				}

				if len(aParts)+len(bParts) == 3 {
					// One command has has a subcommand and the other doesn't
					return fmt.Errorf("mixed nested and unnested commands: %s, %s", aName, bName)
				}

				if aParts[1] == bParts[1] {
					if len(aParts) == 2 || len(bParts) == 2 {
						return fmt.Errorf("duplicate subcommand name: %s", aName)
					}

					if len(aParts)+len(bParts) == 5 {
						// One nested subcommand has a subcommand and the other doesn't
						return fmt.Errorf("mixed nested and unnested subcommands: %s, %s", aName, bName)
					}

					if aParts[2] == bParts[2] {
						return fmt.Errorf("duplicate subcommand name: %s", aName)
					}
				}
			}
		}
	}

	return nil
}

func mergeCommands(commands []api.CreateCommandData) ([]api.CreateCommandData, error) {
	rootCMDs := make(map[string]*api.CreateCommandData)

	// Merge root commands
	for _, command := range commands {
		// TODO: think about how to handle different configs for root cmd
		if c, ok := rootCMDs[command.Name]; ok {
			c.Options = append(c.Options, command.Options...)
		} else {
			rootCMDs[command.Name] = &command
		}
	}

	// Merge sub command groups
	for _, command := range rootCMDs {
		groups := make(map[string]*discord.SubcommandGroupOption)
		args := make([]discord.CommandOption, 0, len(command.Options))

		for _, option := range command.Options {
			if g, ok := option.(*discord.SubcommandGroupOption); ok {
				if group, ok := groups[g.Name()]; ok {
					group.Subcommands = append(group.Subcommands, g.Subcommands...)
				} else {
					groups[g.Name()] = g
				}
			} else {
				args = append(args, option)
			}
		}

		command.Options = args
		for _, group := range groups {
			command.Options = append(command.Options, group)
		}
	}

	res := make([]api.CreateCommandData, 0, len(rootCMDs))
	for _, command := range rootCMDs {
		res = append(res, *command)
	}

	return res, nil
}
