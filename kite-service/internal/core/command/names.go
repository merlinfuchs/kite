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
	Name string
	// Empty is whether the user made it and its flow has nothing but its
	// entry block, so nothing depends on it yet. Commands of plugins and
	// templates never are, so they aren't renamed.
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
// It returns the names of the added commands and of the existing ones, both
// in their order.
func resolveCommandNames(existing []existingCommand, added []string) ([]string, []string, error) {
	// Conflicts within the added commands, like a template with two /ban
	// commands, aren't solved by numbers.
	if err := checkCommandNames(nil, added); err != nil {
		return nil, nil, err
	}

	names := slices.Clone(added)
	existingNames := make([]string, len(existing))
	for i, e := range existing {
		existingNames[i] = e.Name
	}
	var placed []string

	for _, group := range commandGroups(added) {
		groupNames := make([]string, len(group))
		for k, i := range group {
			groupNames[k] = names[i]
		}

		var hit []int
		blocked := false
		for j, e := range existing {
			if !conflictsAny(existingNames[j], groupNames) {
				continue
			}
			hit = append(hit, j)
			if !e.Empty {
				blocked = true
			}
		}
		for _, name := range groupNames {
			if conflictsAny(name, placed) {
				blocked = true
			}
		}

		// Numbered names never conflict with the ones they're made from, as
		// the roots differ. The added ones of later roots keep theirs.
		others := slices.Concat(existingNames, names)
		if blocked {
			free, err := freeName(groupNames, others)
			if err != nil {
				return nil, nil, err
			}
			for k, i := range group {
				names[i] = free[k]
			}
			groupNames = free
		} else if len(hit) > 0 {
			hitNames := make([]string, len(hit))
			for k, j := range hit {
				hitNames[k] = existingNames[j]
			}
			free, err := freeName(hitNames, others)
			if err != nil {
				return nil, nil, err
			}
			for k, j := range hit {
				existingNames[j] = free[k]
			}
		}
		placed = append(placed, groupNames...)
	}

	if err := checkCommandNames(existingNames, names); err != nil {
		return nil, nil, err
	}
	return names, existingNames, nil
}

// freeName gives the commands, which share their root, the lowest number for
// that root that makes them conflict with none of the others.
func freeName(names []string, others []string) ([]string, error) {
	root := []rune(commandRoot(names[0]))
	for n := 2; n <= maxNameNumber; n++ {
		suffix := fmt.Sprintf("-%d", n)
		base := string(root[:min(len(root), maxNamePartLength-len(suffix))])
		res := make([]string, len(names))
		free := true
		for i, name := range names {
			res[i] = withCommandRoot(name, base+suffix)
			if conflictsAny(res[i], others) {
				free = false
				break
			}
		}
		if free {
			return res, nil
		}
	}
	return nil, &DeployError{
		Message: fmt.Sprintf("There are too many commands named /%s.", string(root)),
		Command: string(root),
	}
}

func conflictsAny(name string, others []string) bool {
	for _, other := range others {
		if commandNameConflict(name, other) != nil {
			return true
		}
	}
	return false
}

// commandGroups returns the indexes of the names by their root, like "mod" of
// "mod ban", in the order the roots first appear.
func commandGroups(names []string) [][]int {
	var roots []string
	var groups [][]int
	for i, name := range names {
		root := commandRoot(name)
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
// need to be saved in the same transaction as the added ones. Without
// renameEmpty, only the added ones get new names.
func (m *CommandManager) ResolveCommandNames(ctx context.Context, appID string, renameEmpty bool, names ...string) ([]string, []*model.Command, error) {
	commands, pluginCommands, err := m.takenCommands(ctx, appID)
	if err != nil {
		return nil, nil, err
	}

	existing := make([]existingCommand, 0, len(commands)+len(pluginCommands))
	for _, command := range commands {
		empty := renameEmpty && !command.BundleInstallID.Valid && isEmptyCommand(command.FlowSource)
		existing = append(existing, existingCommand{Name: command.Name, Empty: empty})
	}
	for _, command := range pluginCommands {
		existing = append(existing, existingCommand{Name: command.Name})
	}

	resolved, existingNames, err := resolveCommandNames(existing, names)
	if err != nil {
		return nil, nil, err
	}

	var renamed []*model.Command
	for i, command := range commands {
		if existingNames[i] == command.Name {
			continue
		}
		updated := *command
		updated.Name = existingNames[i]
		updated.FlowSource = updated.FlowSource.WithCommandName(updated.Name)
		updated.UpdatedAt = time.Now().UTC()
		renamed = append(renamed, &updated)
	}
	return resolved, renamed, nil
}
