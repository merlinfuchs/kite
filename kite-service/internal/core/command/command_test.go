package command

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

func TestCheckCommandNames(t *testing.T) {
	tests := []struct {
		names   []string
		message string
		command string
	}{
		{names: []string{"ban", "kick", "mod warn", "mod mute", "a b c", "a b d"}},
		{names: []string{"ban", "ban"}, message: "There is more than one /ban command.", command: "ban"},
		{names: []string{"mod", "mod warn"}, message: "/mod and /mod warn can't both exist, because /mod has subcommands.", command: "mod"},
		{names: []string{"mod", "mod warn now"}, message: "/mod and /mod warn now can't both exist, because /mod has subcommands.", command: "mod"},
		{names: []string{"mod warn", "mod warn"}, message: "There is more than one /mod warn command.", command: "mod"},
		{names: []string{"mod warn", "mod warn now"}, message: "/mod warn and /mod warn now can't both exist, because /mod warn has subcommands.", command: "mod"},
		{names: []string{"a b c", "a b c"}, message: "There is more than one /a b c command.", command: "a"},
		{names: []string{"ban", ""}, message: "A command has no name."},
	}

	for _, test := range tests {
		err := checkCommandNames(nil, test.names)
		if test.message == "" {
			if err != nil {
				t.Errorf("%v: unexpected error: %v", test.names, err)
			}
			continue
		}

		deployErr, ok := err.(*DeployError)
		if !ok {
			t.Errorf("%v: expected DeployError, got %v", test.names, err)
			continue
		}
		if deployErr.Message != test.message || deployErr.Command != test.command {
			t.Errorf("%v: got %q (%q), want %q (%q)", test.names, deployErr.Message, deployErr.Command, test.message, test.command)
		}
	}
}

func TestMergeCommandsKeepsOrder(t *testing.T) {
	commands := []api.CreateCommandData{
		{Name: "ping", Description: "Ping"},
		{Name: "mod", Options: discord.CommandOptions{discord.NewSubcommandOption("warn", "Warn")}},
		{Name: "mod", Options: discord.CommandOptions{discord.NewSubcommandGroupOption("role", "Role", discord.NewSubcommandOption("add", "Add"))}},
		{Name: "ban", Description: "Ban"},
		{Name: "mod", Options: discord.CommandOptions{discord.NewSubcommandGroupOption("role", "Role", discord.NewSubcommandOption("remove", "Remove"))}},
		{Name: "mod", Options: discord.CommandOptions{discord.NewSubcommandOption("mute", "Mute")}},
	}

	first, err := hashCommands("1", mergeCommands(commands))
	if err != nil {
		t.Fatal(err)
	}

	for range 20 {
		res := mergeCommands(commands)

		names := []string{}
		for _, c := range res {
			names = append(names, c.Name)
		}
		if len(names) != 3 || names[0] != "ping" || names[1] != "mod" || names[2] != "ban" {
			t.Fatalf("unexpected commands: %v", names)
		}

		options := []string{}
		for _, o := range res[1].Options {
			options = append(options, o.Name())
		}
		if len(options) != 3 || options[0] != "warn" || options[1] != "role" || options[2] != "mute" {
			t.Fatalf("unexpected options: %v", options)
		}
		if group := res[1].Options[1].(*discord.SubcommandGroupOption); len(group.Subcommands) != 2 {
			t.Fatalf("expected the role group to be merged, got %d subcommands", len(group.Subcommands))
		}

		hash, err := hashCommands("1", res)
		if err != nil {
			t.Fatal(err)
		}
		if hash != first {
			t.Fatal("hash changed for the same commands")
		}
	}
}

func TestDiscordDeployError(t *testing.T) {
	cmdData := []api.CreateCommandData{
		{Name: "ping", Description: "Ping"},
		{
			Name:        "ban",
			Description: "Ban",
			Options: discord.CommandOptions{
				&discord.UserOption{OptionName: "user", Description: "User"},
				&discord.StringOption{OptionName: "reason", Description: "", Choices: []discord.StringChoice{{Name: "Spam", Value: "spam"}}},
			},
		},
	}

	tests := []struct {
		name    string
		err     *httputil.HTTPError
		message string
		command string
	}{
		{
			name: "option field",
			err: &httputil.HTTPError{
				Status:  400,
				Code:    50035,
				Message: "Invalid Form Body",
				Errors:  []byte(`{"1": {"options": {"1": {"description": {"_errors": [{"code": "BASE_TYPE_BAD_LENGTH", "message": "Must be between 1 and 100 in length."}]}}}}}`),
			},
			message: "Description of /ban reason: Must be between 1 and 100 in length.",
			command: "ban",
		},
		{
			name: "choice",
			err: &httputil.HTTPError{
				Status: 400,
				Code:   50035,
				Errors: []byte(`{"1": {"options": {"1": {"choices": {"0": {"name": {"_errors": [{"code": "X", "message": "Too long"}]}}}}}}}`),
			},
			message: "Name of /ban reason: Too long",
			command: "ban",
		},
		{
			name: "command field",
			err: &httputil.HTTPError{
				Status:  400,
				Code:    50035,
				Message: "Invalid Form Body",
				Errors:  []byte(`{"0": {"name": {"_errors": [{"code": "APPLICATION_COMMAND_INVALID_NAME", "message": "Command name is invalid"}]}}}`),
			},
			message: "Name of /ping: Command name is invalid",
			command: "ping",
		},
		{
			name: "unknown index",
			err: &httputil.HTTPError{
				Status: 400,
				Code:   50035,
				Errors: []byte(`{"_errors": [{"code": "X", "message": "Too many commands"}]}`),
			},
			message: "Too many commands",
		},
		{
			name:    "no form errors",
			err:     &httputil.HTTPError{Status: 400, Code: 30032, Message: "Maximum number of application commands reached (100)"},
			message: "Maximum number of application commands reached (100)",
		},
		{
			name:    "invalid token",
			err:     &httputil.HTTPError{Status: 401, Message: "401: Unauthorized"},
			message: "Discord didn't accept the bot token. Update it in the settings.",
		},
	}

	for _, test := range tests {
		err := discordDeployError(test.err, cmdData)
		if err.Message != test.message || err.Command != test.command {
			t.Errorf("%s: got %q (%q), want %q (%q)", test.name, err.Message, err.Command, test.message, test.command)
		}
	}
}

func TestOverwriteCommandsKeepsEntryPoint(t *testing.T) {
	var puts []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case "GET":
			w.Write([]byte(`[{"id":"1","type":4,"name":"launch","handler":2},{"id":"2","type":1,"name":"ping"}]`))
		case "PUT":
			body, _ := io.ReadAll(r.Body)
			puts = append(puts, string(body))
			if len(puts) == 1 {
				w.WriteHeader(400)
				w.Write([]byte(`{"code":50240,"message":"You cannot remove this app's Entry Point command in a bulk update operation."}`))
				return
			}
			w.Write([]byte(`[]`))
		}
	}))
	defer srv.Close()

	endpoint := api.EndpointApplications
	api.EndpointApplications = srv.URL + "/applications/"
	defer func() { api.EndpointApplications = endpoint }()

	client := api.NewClient("Bot token")
	err := overwriteCommands(client, 1, []api.CreateCommandData{{Name: "ping", Description: "Ping"}})
	if err != nil {
		t.Fatal(err)
	}

	if len(puts) != 2 {
		t.Fatalf("expected a retry, got %d requests", len(puts))
	}
	if !strings.Contains(puts[1], `"type":4`) || !strings.Contains(puts[1], `"name":"ping"`) || strings.Count(puts[1], `"name":"ping"`) != 1 {
		t.Fatalf("retry should keep only the entry point and our commands: %s", puts[1])
	}
}

type fakeCommandStore struct {
	store.CommandStore
	commands []*model.Command
}

func (s fakeCommandStore) CommandsByApp(ctx context.Context, appID string) ([]*model.Command, error) {
	return s.commands, nil
}

type fakePluginInstanceStore struct {
	store.PluginInstanceStore
}

func (fakePluginInstanceStore) PluginInstancesByApp(ctx context.Context, appID string) ([]*model.PluginInstance, error) {
	return nil, nil
}

func commandFlow(name string) flow.FlowData {
	return flow.FlowData{
		Nodes: []flow.FlowNode{
			{ID: "entry", Type: flow.FlowNodeTypeEntryCommand, Data: flow.FlowNodeData{Name: name, Description: "Test"}},
		},
	}
}

func TestAppCommandsSkipsDisabled(t *testing.T) {
	m := &CommandManager{
		commandStore: fakeCommandStore{commands: []*model.Command{
			{Name: "ping", Enabled: true, FlowSource: commandFlow("ping")},
			// Would conflict with the enabled /ping.
			{Name: "ping", Enabled: false, FlowSource: commandFlow("ping")},
			// Would fail to compile.
			{Name: "broken", Enabled: false},
		}},
		pluginInstanceStore: fakePluginInstanceStore{},
	}

	res, err := m.appCommands(context.Background(), "app")
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 || res[0].Name != "ping" {
		t.Fatalf("expected only the enabled /ping, got %v", res)
	}
}
