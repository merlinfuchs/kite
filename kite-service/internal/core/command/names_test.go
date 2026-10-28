package command

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

func TestResolveCommandNames(t *testing.T) {
	used := func(name string) existingCommand { return existingCommand{ID: "id-" + name, Name: name} }
	empty := func(name string) existingCommand { return existingCommand{ID: "id-" + name, Name: name, Empty: true} }
	long := strings.Repeat("a", 32)

	tests := []struct {
		name     string
		existing []existingCommand
		added    []string
		want     []string
		renamed  map[int]string
	}{
		{name: "free", existing: []existingCommand{used("ban")}, added: []string{"giveaway"}, want: []string{"giveaway"}},
		{name: "taken", existing: []existingCommand{used("giveaway")}, added: []string{"giveaway"}, want: []string{"giveaway-2"}},
		{name: "taken twice", existing: []existingCommand{used("giveaway"), used("giveaway-2")}, added: []string{"giveaway"}, want: []string{"giveaway-3"}},
		{name: "empty renamed", existing: []existingCommand{empty("giveaway")}, added: []string{"giveaway"}, want: []string{"giveaway"}, renamed: map[int]string{0: "giveaway-2"}},
		{name: "empty renamed past taken", existing: []existingCommand{empty("giveaway"), used("giveaway-2")}, added: []string{"giveaway"}, want: []string{"giveaway"}, renamed: map[int]string{0: "giveaway-3"}},
		{name: "empty and used", existing: []existingCommand{empty("mod ban"), used("mod kick")}, added: []string{"mod ban", "mod kick"}, want: []string{"mod-2 ban", "mod-2 kick"}},
		{name: "plugin", existing: []existingCommand{{Name: "giveaway", Empty: true}}, added: []string{"giveaway"}, want: []string{"giveaway-2"}},
		{name: "subcommands together", existing: []existingCommand{used("mod ban")}, added: []string{"mod ban", "mod kick"}, want: []string{"mod-2 ban", "mod-2 kick"}},
		{name: "subcommands join group", existing: []existingCommand{used("mod warn")}, added: []string{"mod ban"}, want: []string{"mod ban"}},
		{name: "parent taken", existing: []existingCommand{used("mod")}, added: []string{"mod ban"}, want: []string{"mod-2 ban"}},
		{name: "empty parent renamed", existing: []existingCommand{empty("mod")}, added: []string{"mod ban"}, want: []string{"mod ban"}, renamed: map[int]string{0: "mod-2"}},
		{name: "long", existing: []existingCommand{used(long)}, added: []string{long}, want: []string{long[:30] + "-2"}},
	}

	for _, test := range tests {
		got, renamed, err := resolveCommandNames(test.existing, test.added)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", test.name, err)
			continue
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s: got %v, want %v", test.name, got, test.want)
		}
		if test.renamed == nil {
			test.renamed = map[int]string{}
		}
		if !reflect.DeepEqual(renamed, test.renamed) {
			t.Errorf("%s: renamed %v, want %v", test.name, renamed, test.renamed)
		}
	}
}

func TestResolveCommandNamesWithinAdded(t *testing.T) {
	_, _, err := resolveCommandNames(nil, []string{"ban", "ban"})
	if err == nil || err.Error() != "There is more than one /ban command." {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestIsEmptyCommand(t *testing.T) {
	entry := flow.FlowNode{ID: "a", Type: flow.FlowNodeTypeEntryCommand}
	note := flow.FlowNode{ID: "b", Type: flow.FlowNodeTypeNote}
	action := flow.FlowNode{ID: "c", Type: flow.FlowNodeTypeActionResponseCreate}
	if !isEmptyCommand(flow.FlowData{Nodes: []flow.FlowNode{entry, note}}) {
		t.Error("entry and note should be empty")
	}
	if isEmptyCommand(flow.FlowData{Nodes: []flow.FlowNode{entry, action}}) {
		t.Error("entry and action shouldn't be empty")
	}
}
