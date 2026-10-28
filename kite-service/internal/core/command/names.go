package command

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

// maxNameNumber is the highest number added to a name to make it free.
const maxNameNumber = 100

// maxNamePartLength is how long Discord allows a command name or the name of
// a subcommand to be.
const maxNamePartLength = 32

// resolveCommandNames picks names for the added commands, so they don't
// conflict with the existing ones or each other in a way Discord rejects.
// When the added commands of a root conflict, they get a number, like
// /giveaway-2 or /mod-2 ban, which keeps a template's subcommands together.
func resolveCommandNames(existing []string, added []string) ([]string, error) {
	// Conflicts within the added commands, like a template with two /ban
	// commands, aren't solved by numbers.
	if err := checkCommandNames(nil, added); err != nil {
		return nil, err
	}

	names := slices.Clone(added)
	var placed []string
	for _, group := range commandGroups(added) {
		groupNames := make([]string, len(group))
		for k, i := range group {
			groupNames[k] = names[i]
		}
		// Earlier roots may have taken a number already.
		taken := false
		for _, name := range groupNames {
			if conflictsAny(name, existing) || conflictsAny(name, placed) {
				taken = true
			}
		}
		if !taken {
			placed = append(placed, groupNames...)
			continue
		}

		// Numbered names never conflict with the ones they're made from, as
		// the roots differ. The added ones of later roots keep theirs.
		free, err := freeName(groupNames, slices.Concat(existing, names))
		if err != nil {
			return nil, err
		}
		for k, i := range group {
			names[i] = free[k]
		}
		placed = append(placed, free...)
	}

	if err := checkCommandNames(existing, names); err != nil {
		return nil, err
	}
	return names, nil
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

// ResolveCommandNames picks names for added commands that don't conflict
// with the app's commands, see resolveCommandNames, in their order.
func (m *CommandManager) ResolveCommandNames(ctx context.Context, appID string, names ...string) ([]string, error) {
	commands, pluginCommands, err := m.takenCommands(ctx, appID)
	if err != nil {
		return nil, err
	}

	existing := make([]string, 0, len(commands)+len(pluginCommands))
	for _, command := range commands {
		existing = append(existing, command.Name)
	}
	for _, command := range pluginCommands {
		existing = append(existing, command.Name)
	}
	return resolveCommandNames(existing, names)
}
