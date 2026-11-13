package command

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

// maxNameNumber is the highest number added to a name to make it free.
const maxNameNumber = 100

// maxNamePartLength is how long Discord allows a command name or the name of
// a subcommand to be.
const maxNamePartLength = 32

// maxContextMenuCommands is how many User commands, and how many Message
// commands, Discord allows an app.
const maxContextMenuCommands = 15

// Key is a command's name and its type. Discord allows commands of different
// types to have the same name, and only slash commands have subcommands.
type Key struct {
	Type discord.CommandType
	Name string
}

// NewKey returns the key of a command. Commands without a type are slash
// commands.
func NewKey(t discord.CommandType, name string) Key {
	if t == 0 {
		t = discord.ChatInputCommand
	}
	return Key{Type: t, Name: name}
}

// ModelKey returns the key of a saved command.
func ModelKey(command *model.Command) Key {
	return NewKey(command.FlowSource.CommandType(), command.Name)
}

// FlowKey returns the key of a compiled command.
func FlowKey(node *flow.CompiledFlowNode) Key {
	return NewKey(node.CommandType(), node.CommandName())
}

func (k Key) isSlash() bool {
	return k.Type == discord.ChatInputCommand
}

// typeName is how Discord names the type of a context menu command.
func (k Key) typeName() string {
	if k.Type == discord.MessageCommand {
		return "Message"
	}
	return "User"
}

// String is the command as the user knows it, like /mod ban or the User
// command "Report".
func (k Key) String() string {
	if k.isSlash() {
		return "/" + k.Name
	}
	return fmt.Sprintf("the %s command %q", k.typeName(), k.Name)
}

// deployError is a problem with the command, which the dashboard shows next
// to it.
func (k Key) deployError(message string) *DeployError {
	res := &DeployError{Message: message, Command: k.root()}
	switch k.Type {
	case discord.UserCommand:
		res.CommandType = string(flow.CommandTypeUser)
	case discord.MessageCommand:
		res.CommandType = string(flow.CommandTypeMessage)
	}
	return res
}

// root is the command Discord has, like "mod" of /mod ban. Only slash
// commands have subcommands.
func (k Key) root() string {
	if !k.isSlash() {
		return k.Name
	}
	root, _, _ := strings.Cut(k.Name, " ")
	return root
}

// withRoot returns the key with its root replaced.
func (k Key) withRoot(root string) Key {
	if k.isSlash() {
		if _, rest, ok := strings.Cut(k.Name, " "); ok {
			k.Name = root + " " + rest
			return k
		}
	}
	k.Name = root
	return k
}

// resolveCommandNames picks names for the added commands, so they don't
// conflict with the existing ones or each other in a way Discord rejects.
// When the added commands of a root conflict, they get a number, like
// /giveaway-2 or /mod-2 ban, which keeps a template's subcommands together.
func resolveCommandNames(existing []Key, added []Key) ([]string, error) {
	// Conflicts within the added commands, like a template with two /ban
	// commands, aren't solved by numbers.
	if err := checkCommandNames(nil, added); err != nil {
		return nil, err
	}

	keys := slices.Clone(added)
	var placed []Key
	for _, group := range commandGroups(added) {
		groupKeys := make([]Key, len(group))
		for k, i := range group {
			groupKeys[k] = keys[i]
		}
		// Earlier roots may have taken a number already.
		taken := false
		for _, key := range groupKeys {
			if conflictsAny(key, existing) || conflictsAny(key, placed) {
				taken = true
			}
		}
		if !taken {
			placed = append(placed, groupKeys...)
			continue
		}

		// Numbered names never conflict with the ones they're made from, as
		// the roots differ. The added ones of later roots keep theirs.
		free, err := freeName(groupKeys, slices.Concat(existing, keys))
		if err != nil {
			return nil, err
		}
		for k, i := range group {
			keys[i] = free[k]
		}
		placed = append(placed, free...)
	}

	if err := checkCommandNames(existing, keys); err != nil {
		return nil, err
	}

	names := make([]string, len(keys))
	for i, key := range keys {
		names[i] = key.Name
	}
	return names, nil
}

// freeName gives the commands, which share their root, the lowest number for
// that root that makes them conflict with none of the others.
func freeName(keys []Key, others []Key) ([]Key, error) {
	root := []rune(keys[0].root())
	// Context menu commands are named like "Report 2".
	format := " %d"
	if keys[0].isSlash() {
		format = "-%d"
	}
	for n := 2; n <= maxNameNumber; n++ {
		suffix := fmt.Sprintf(format, n)
		base := string(root[:min(len(root), maxNamePartLength-len(suffix))])
		res := make([]Key, len(keys))
		free := true
		for i, key := range keys {
			res[i] = key.withRoot(base + suffix)
			if conflictsAny(res[i], others) {
				free = false
				break
			}
		}
		if free {
			return res, nil
		}
	}
	rootKey := Key{Type: keys[0].Type, Name: string(root)}
	return nil, rootKey.deployError(fmt.Sprintf("There are too many commands like %s.", rootKey))
}

func conflictsAny(key Key, others []Key) bool {
	for _, other := range others {
		if commandNameConflict(key, other) != nil {
			return true
		}
	}
	return false
}

// commandGroups returns the indexes of the commands by their type and root,
// like "mod" of /mod ban, in the order the roots first appear.
func commandGroups(keys []Key) [][]int {
	var roots []Key
	var groups [][]int
	for i, key := range keys {
		root := Key{Type: key.Type, Name: key.root()}
		k := slices.Index(roots, root)
		if k == -1 {
			roots = append(roots, root)
			groups = append(groups, nil)
			k = len(groups) - 1
		}
		groups[k] = append(groups[k], i)
	}
	return groups
}

// ResolveCommandNames picks names for added commands that don't conflict
// with the app's commands, see resolveCommandNames, in their order.
func (m *CommandManager) ResolveCommandNames(ctx context.Context, appID string, keys ...Key) ([]string, error) {
	taken, deployed, err := m.takenCommands(ctx, appID, "")
	if err != nil {
		return nil, err
	}
	if err := checkCommandLimits(deployed, keys); err != nil {
		return nil, err
	}
	return resolveCommandNames(taken, keys)
}
