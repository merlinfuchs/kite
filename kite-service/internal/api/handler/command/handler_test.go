package command

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/core/command"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memoryCommandStore struct {
	// Methods the tests don't use panic.
	store.CommandStore

	commands []*model.Command
}

func (s *memoryCommandStore) CommandsByApp(ctx context.Context, appID string) ([]*model.Command, error) {
	return s.commands, nil
}

func (s *memoryCommandStore) CountCommandsByApp(ctx context.Context, appID string) (int, error) {
	return len(s.commands), nil
}

func (s *memoryCommandStore) CreateCommand(ctx context.Context, cmd *model.Command) (*model.Command, error) {
	s.commands = append(s.commands, cmd)
	return cmd, nil
}

func (s *memoryCommandStore) UpdateCommand(ctx context.Context, cmd *model.Command) (*model.Command, error) {
	for i, existing := range s.commands {
		if existing.ID == cmd.ID {
			s.commands[i] = cmd
		}
	}
	return cmd, nil
}

type noPluginInstanceStore struct {
	store.PluginInstanceStore
}

func (noPluginInstanceStore) PluginInstancesByApp(ctx context.Context, appID string) ([]*model.PluginInstance, error) {
	return nil, nil
}

func commandFlow(name string, commandType flow.CommandType) string {
	return fmt.Sprintf(`{
		"nodes": [{"id": "1", "type": "entry_command", "data": {"name": %q, "description": "Test", "command_type": %q}}],
		"edges": []
	}`, name, commandType)
}

func storedCommand(t *testing.T, id string, name string, commandType flow.CommandType) *model.Command {
	t.Helper()

	var source flow.FlowData
	require.NoError(t, json.Unmarshal([]byte(commandFlow(name, commandType)), &source))
	return &model.Command{ID: id, AppID: "app", Name: name, FlowSource: source}
}

func newTestHandler(commands ...*model.Command) (*CommandHandler, *memoryCommandStore) {
	s := &memoryCommandStore{commands: commands}
	manager := command.NewCommandManager(nil, s, noPluginInstanceStore{}, plugin.NewRegistry(), nil)
	return NewCommandHandler(s, manager), s
}

func request(t *testing.T, cmd *model.Command, maxContextMenus int, body string, h handler.HandlerFunc) (int, map[string]any) {
	t.Helper()

	srv := handler.APIHandler(func(c *handler.Context) error {
		c.App = &model.App{ID: "app"}
		c.Session = &model.Session{UserID: "user"}
		c.Features = model.Features{MaxContextMenuCommands: maxContextMenus}
		c.Command = cmd
		return h(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	return rec.Code, res
}

func errorCode(res map[string]any) string {
	return res["error"].(map[string]any)["code"].(string)
}

func TestCommandCreateContextMenuLimit(t *testing.T) {
	h, _ := newTestHandler(
		storedCommand(t, "1", "Report", flow.CommandTypeUser),
		storedCommand(t, "2", "ping", flow.CommandTypeChatInput),
	)
	create := handler.TypedWithBody(h.HandleCommandCreate)

	// Slash commands don't count.
	code, res := request(t, nil, 1, `{"flow_source": `+commandFlow("info", flow.CommandTypeChatInput)+`}`, create)
	require.Equal(t, http.StatusOK, code, res)

	code, res = request(t, nil, 1, `{"flow_source": `+commandFlow("Quote", flow.CommandTypeMessage)+`}`, create)
	require.Equal(t, http.StatusBadRequest, code, res)
	assert.Equal(t, "resource_limit", errorCode(res))
}

func TestCommandImportContextMenuLimit(t *testing.T) {
	h, _ := newTestHandler(storedCommand(t, "1", "Report", flow.CommandTypeUser))
	importCommands := handler.TypedWithBody(h.HandleCommandsImport)

	body := `{"commands": [{"flow_source": ` + commandFlow("Quote", flow.CommandTypeMessage) + `}, {"flow_source": ` + commandFlow("Ban", flow.CommandTypeUser) + `}]}`
	code, res := request(t, nil, 2, body, importCommands)
	require.Equal(t, http.StatusBadRequest, code, res)
	assert.Equal(t, "resource_limit", errorCode(res))
}

// Changing the type of a command makes it count towards the limit.
func TestCommandUpdateContextMenuLimit(t *testing.T) {
	report := storedCommand(t, "1", "Report", flow.CommandTypeUser)
	ping := storedCommand(t, "2", "ping", flow.CommandTypeChatInput)
	h, _ := newTestHandler(report, ping)
	update := handler.TypedWithBody(h.HandleCommandUpdate)

	code, res := request(t, ping, 1, `{"flow_source": `+commandFlow("Ping", flow.CommandTypeMessage)+`}`, update)
	require.Equal(t, http.StatusBadRequest, code, res)
	assert.Equal(t, "resource_limit", errorCode(res))

	// Saving a context menu that is already counted is fine.
	code, res = request(t, report, 1, `{"flow_source": `+commandFlow("Report User", flow.CommandTypeUser)+`}`, update)
	require.Equal(t, http.StatusOK, code, res)
}

func TestCommandCreateNamesPerType(t *testing.T) {
	h, _ := newTestHandler(storedCommand(t, "1", "Report", flow.CommandTypeUser))
	create := handler.TypedWithBody(h.HandleCommandCreate)

	code, res := request(t, nil, 0, `{"flow_source": `+commandFlow("Report", flow.CommandTypeMessage)+`}`, create)
	require.Equal(t, http.StatusOK, code, res)

	code, res = request(t, nil, 0, `{"flow_source": `+commandFlow("Report", flow.CommandTypeUser)+`}`, create)
	require.Equal(t, http.StatusBadRequest, code, res)
	assert.Equal(t, "invalid_command_name", errorCode(res))
}

func TestCommandCreateUnknownType(t *testing.T) {
	h, _ := newTestHandler()
	create := handler.TypedWithBody(h.HandleCommandCreate)

	code, res := request(t, nil, 0, `{"flow_source": `+commandFlow("ping", "chat")+`}`, create)
	require.Equal(t, http.StatusBadRequest, code, res)
}
