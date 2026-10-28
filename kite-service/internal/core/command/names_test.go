package command

import (
	"reflect"
	"strings"
	"testing"
)

func TestResolveCommandNames(t *testing.T) {
	long := strings.Repeat("a", 32)

	tests := []struct {
		name     string
		existing []string
		added    []string
		want     []string
	}{
		{name: "free", existing: []string{"ban"}, added: []string{"giveaway"}, want: []string{"giveaway"}},
		{name: "taken", existing: []string{"giveaway"}, added: []string{"giveaway"}, want: []string{"giveaway-2"}},
		{name: "taken twice", existing: []string{"giveaway", "giveaway-2"}, added: []string{"giveaway"}, want: []string{"giveaway-3"}},
		{name: "subcommands together", existing: []string{"mod ban"}, added: []string{"mod ban", "mod kick"}, want: []string{"mod-2 ban", "mod-2 kick"}},
		{name: "subcommands join group", existing: []string{"mod warn"}, added: []string{"mod ban"}, want: []string{"mod ban"}},
		{name: "parent taken", existing: []string{"mod"}, added: []string{"mod ban"}, want: []string{"mod-2 ban"}},
		{name: "later added names kept", existing: []string{"giveaway"}, added: []string{"giveaway", "giveaway-2"}, want: []string{"giveaway-3", "giveaway-2"}},
		{name: "long", existing: []string{long}, added: []string{long}, want: []string{long[:30] + "-2"}},
	}

	for _, test := range tests {
		got, err := resolveCommandNames(test.existing, test.added)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", test.name, err)
			continue
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s: got %v, want %v", test.name, got, test.want)
		}
	}
}

func TestResolveCommandNamesWithinAdded(t *testing.T) {
	// Not /ban-2, which the user never saw.
	_, err := resolveCommandNames([]string{"ban"}, []string{"ban", "ban"})
	if err == nil || err.Error() != "There is more than one /ban command." {
		t.Errorf("unexpected error: %v", err)
	}
}
