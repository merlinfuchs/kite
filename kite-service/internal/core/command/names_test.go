package command

import (
	"reflect"
	"strings"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
)

func TestResolveCommandNames(t *testing.T) {
	long := strings.Repeat("a", 32)
	user := func(name string) Key { return NewKey(discord.UserCommand, name) }

	tests := []struct {
		name     string
		existing []Key
		added    []Key
		want     []string
	}{
		{name: "free", existing: slashKeys("ban"), added: slashKeys("giveaway"), want: []string{"giveaway"}},
		{name: "taken", existing: slashKeys("giveaway"), added: slashKeys("giveaway"), want: []string{"giveaway-2"}},
		{name: "taken twice", existing: slashKeys("giveaway", "giveaway-2"), added: slashKeys("giveaway"), want: []string{"giveaway-3"}},
		{name: "subcommands together", existing: slashKeys("mod ban"), added: slashKeys("mod ban", "mod kick"), want: []string{"mod-2 ban", "mod-2 kick"}},
		{name: "subcommands join group", existing: slashKeys("mod warn"), added: slashKeys("mod ban"), want: []string{"mod ban"}},
		{name: "parent taken", existing: slashKeys("mod"), added: slashKeys("mod ban"), want: []string{"mod-2 ban"}},
		{name: "later added names kept", existing: slashKeys("giveaway"), added: slashKeys("giveaway", "giveaway-2"), want: []string{"giveaway-3", "giveaway-2"}},
		{name: "long", existing: slashKeys(long), added: slashKeys(long), want: []string{long[:30] + "-2"}},
		{name: "other type", existing: slashKeys("report"), added: []Key{user("report")}, want: []string{"report"}},
		{name: "context menu taken", existing: []Key{user("User Info")}, added: []Key{user("User Info")}, want: []string{"User Info 2"}},
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
	_, err := resolveCommandNames(slashKeys("ban"), slashKeys("ban", "ban"))
	if err == nil || err.Error() != "There is more than one /ban command." {
		t.Errorf("unexpected error: %v", err)
	}
}
