package builder

import (
	"fmt"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
)

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
				if len(aParts) == 1 && len(bParts) == 1 {
					return fmt.Errorf("duplicate command name: %s", aName)
				}

				if len(aParts)+len(bParts) == 3 {
					// One command has has a subcommand and the other doesn't
					return fmt.Errorf("mixed nested and unnested commands: %s, %s", aName, bName)
				}

				if aParts[1] == bParts[1] {
					if len(aParts) == 2 && len(bParts) == 2 {
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

func getFullCommandName(d *discord.CommandInteraction) string {
	fullName := d.Name
	for _, option := range d.Options {
		if option.Type == discord.SubcommandOptionType {
			fullName += " " + option.Name
			break
		} else if option.Type == discord.SubcommandGroupOptionType {
			fullName += " " + option.Name
			for _, subOption := range option.Options {
				fullName += " " + subOption.Name
			}
			break
		}
	}

	return fullName
}
