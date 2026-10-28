package command

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

// existingCommand is a command of the app that added commands can conflict
// with.
type existingCommand struct {
	// ID is empty for commands of plugins, which can't be renamed.
	ID   string
	Name string
	// Empty is whether its flow has nothing but its entry block, so nothing
	// depends on it yet.
	Empty bool
}

// maxNameNumber is the highest number added to a name to make it free.
const maxNameNumber = 100

// maxNamePartLength is how long Discord allows a command name or the name of
// a subcommand to be.
const maxNamePartLength = 32

// resolveCommandNames picks names for the added commands, so they don't
// conflict with the existing ones or each other in a way Discord rejects.
// When all existing commands the added commands of a root conflict with are
// empty, those are renamed instead, so a template keeps the names its pages
// use. Otherwise the added commands of the root get a number, like
// /giveaway-2 or /mod-2 ban, which keeps a template's subcommands together.
// It returns the names of the added commands in their order, and the new
// names of the renamed existing commands by their index.
func resolveCommandNames(existing []existingCommand, added []string) ([]string, map[int]string, error) {
	names := slices.Clone(added)
	existingNames := make([]string, len(existing))
	for i, e := range existing {
		existingNames[i] = e.Name
	}
	renamed := map[int]string{}
	var placed []string

	conflicts := func(name string, others []string) bool {
		for _, other := range others {
			if commandNameConflict(name, other) != nil {
				return true
			}
		}
		return false
	}

	for _, root := range commandRoots(added) {
		var group []int
		for i, name := range added {
			if commandRoot(name) == root {
				group = append(group, i)
			}
		}

		var hit []int
		blocked := false
		for _, i := range group {
			for j, e := range existing {
				if commandNameConflict(names[i], existingNames[j]) == nil {
					continue
				}
				if e.ID == "" || !e.Empty {
					blocked = true
				}
				if !slices.Contains(hit, j) {
					hit = append(hit, j)
				}
			}
			if conflicts(names[i], placed) {
				blocked = true
			}
		}

		groupNames := func() []string {
			res := make([]string, len(group))
			for k, i := range group {
				res[k] = names[i]
			}
			return res
		}

		if !blocked {
			for _, j := range hit {
				others := slices.Concat(slices.Delete(slices.Clone(existingNames), j, j+1), placed, groupNames())
				name, err := freeName([]string{existingNames[j]}, others)
				if err != nil {
					return nil, nil, err
				}
				existingNames[j] = name[0]
				renamed[j] = name[0]
			}
		} else {
			free, err := freeName(groupNames(), slices.Concat(existingNames, placed))
			if err != nil {
				return nil, nil, err
			}
			for k, i := range group {
				names[i] = free[k]
			}
		}
		placed = append(placed, groupNames()...)
	}

	// Conflicts within the added commands, like a template with two /ban
	// commands, aren't solved by numbers.
	if err := checkCommandNames(existingNames, names); err != nil {
		return nil, nil, err
	}
	return names, renamed, nil
}

// freeName gives the commands, which share their root, the lowest number for
// that root that makes them conflict with none of the others.
func freeName(names []string, others []string) ([]string, error) {
	root := commandRoot(names[0])
	for n := 2; n <= maxNameNumber; n++ {
		suffix := fmt.Sprintf("-%d", n)
		base := []rune(root)
		base = base[:min(len(base), maxNamePartLength-len(suffix))]
		res := make([]string, len(names))
		free := true
		for i, name := range names {
			res[i] = withCommandRoot(name, string(base)+suffix)
			for _, other := range others {
				if commandNameConflict(res[i], other) != nil {
					free = false
				}
			}
		}
		if free {
			return res, nil
		}
	}
	return nil, &DeployError{
		Message: fmt.Sprintf("There are too many commands named /%s.", root),
		Command: root,
	}
}

// commandRoots returns the roots of the names, like "mod" of "mod ban", in
// the order they first appear.
func commandRoots(names []string) []string {
	var res []string
	for _, name := range names {
		if root := commandRoot(name); !slices.Contains(res, root) {
			res = append(res, root)
		}
	}
	return res
}

func commandRoot(name string) string {
	root, _, _ := strings.Cut(name, " ")
	return root
}

func withCommandRoot(name string, root string) string {
	if _, rest, ok := strings.Cut(name, " "); ok {
		return root + " " + rest
	}
	return root
}

// isEmptyCommand reports whether the flow has nothing but its entry block,
// like a command the user just created.
func isEmptyCommand(source flow.FlowData) bool {
	for _, node := range source.Nodes {
		if node.Type != flow.FlowNodeTypeEntryCommand && node.Type != flow.FlowNodeTypeNote {
			return false
		}
	}
	return true
}

// ResolveCommandNames picks names for added commands that don't conflict
// with the app's commands, see resolveCommandNames. It returns the names of
// the added commands in their order, and the app's commands it renamed, which
// need to be saved with the added ones.
func (m *CommandManager) ResolveCommandNames(ctx context.Context, appID string, names ...string) ([]string, []*model.Command, error) {
	commands, err := m.commandStore.CommandsByApp(ctx, appID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get commands by app: %w", err)
	}
	pluginCommands, err := m.pluginCommands(ctx, appID)
	if err != nil {
		return nil, nil, err
	}

	existing := make([]existingCommand, 0, len(commands)+len(pluginCommands))
	for _, command := range commands {
		existing = append(existing, existingCommand{ID: command.ID, Name: command.Name, Empty: isEmptyCommand(command.FlowSource)})
	}
	for _, command := range pluginCommands {
		existing = append(existing, existingCommand{Name: command.Name})
	}

	resolved, renamedNames, err := resolveCommandNames(existing, names)
	if err != nil {
		return nil, nil, err
	}

	var renamed []*model.Command
	for i, name := range renamedNames {
		updated := *commands[i]
		updated.Name = name
		updated.FlowSource = updated.FlowSource.WithCommandName(name)
		updated.UpdatedAt = time.Now().UTC()
		renamed = append(renamed, &updated)
	}
	// Map order is random, and the order they're saved in shouldn't be.
	slices.SortFunc(renamed, func(a, b *model.Command) int { return strings.Compare(a.ID, b.ID) })
	return resolved, renamed, nil
}
