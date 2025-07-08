package engine

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

type Command struct {
	cmd  *model.Command
	flow *flow.CompiledFlowNode
	env  Env
}

func NewCommand(
	cmd *model.Command,
	env Env,
) (*Command, error) {
	flow, err := flow.CompileCommand(cmd.FlowSource)
	if err != nil {
		slog.Error(
			"Failed to compile command flow",
			slog.String("app_id", cmd.AppID),
			slog.String("command_id", cmd.ID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to compile command flow: %w", err)
	}

	return &Command{
		cmd:  cmd,
		flow: flow,
		env:  env,
	}, nil
}

func (c *Command) HandleEvent(appID string, session *state.State, event gateway.Event) {
	links := entityLinks{
		CommandID: null.NewString(c.cmd.ID, true),
	}

	c.env.executeFlowEvent(
		context.Background(),
		c.cmd.AppID,
		c.flow,
		session,
		event,
		links,
		nil,
	)
}

func (a *App) DeployCommands(ctx context.Context) error {
	a.Lock()
	a.hasUndeployedChanges = false

	var lastUpdatedAt time.Time

	commandNames := make([]string, 0, len(a.commands))
	commands := make([]api.CreateCommandData, 0, len(a.commands))
	for _, command := range a.commands {
		cmd := command.cmd

		if cmd.UpdatedAt.After(lastUpdatedAt) {
			lastUpdatedAt = cmd.UpdatedAt
		}

		node, err := flow.CompileCommand(cmd.FlowSource)
		if err != nil {
			go a.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to compile command flow: %v", err))
			return nil
		}

		data := node.CommandData()
		commands = append(commands, api.CreateCommandData{
			Name:                     data.Name,
			Description:              data.Description,
			Options:                  data.Options,
			DefaultMemberPermissions: data.DefaultMemberPermissions,
			Contexts:                 node.CommandContexts(),
			IntegrationTypes:         node.CommandIntegrations(),
		})
		commandNames = append(commandNames, node.CommandName())
	}

	a.Unlock()

	if err := validateCommandNames(commandNames); err != nil {
		go a.createLogEntry(model.LogLevelError, fmt.Sprintf("invalid command names: %v", err))
		return nil
	}

	commands, err := mergeCommands(commands)
	if err != nil {
		go a.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to merge commands: %v", err))
		return nil
	}

	app, err := a.stores.AppStore.App(ctx, a.id)
	if err != nil {
		return fmt.Errorf("failed to get app: %w", err)
	}

	appId, err := strconv.ParseUint(app.DiscordID, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse app ID: %w", err)
	}

	client := api.NewClient("Bot " + app.DiscordToken).WithContext(ctx)

	_, err = client.BulkOverwriteCommands(discord.AppID(appId), commands)
	if err != nil {
		go a.createLogEntry(model.LogLevelError, fmt.Sprintf("Failed to deploy commands: %v", err))
		return nil
	}

	err = a.stores.CommandStore.UpdateCommandsLastDeployedAt(ctx, a.id, lastUpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update last deployed at: %w", err)
	}

	go a.createLogEntry(model.LogLevelInfo, "Successfully deployed commands")
	return nil
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
