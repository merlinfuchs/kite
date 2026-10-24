package bundleinstall

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	commandhandler "github.com/kitecloud/kite/kite-service/internal/api/handler/command"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

// The memory stores only have what updates use, the other methods panic.
type memoryInstallStore struct {
	store.BundleInstallStore
	installs []*model.BundleInstall
	updated  *store.BundleInstallContents
}

func (s *memoryInstallStore) BundleInstall(ctx context.Context, appID string, id string) (*model.BundleInstall, error) {
	for _, install := range s.installs {
		if install.ID == id && install.AppID == appID {
			return install, nil
		}
	}
	return nil, store.ErrNotFound
}

func (s *memoryInstallStore) BundleInstallsByApp(ctx context.Context, appID string) ([]*model.BundleInstall, error) {
	return s.installs, nil
}

func (s *memoryInstallStore) UpdateBundleInstall(ctx context.Context, install *model.BundleInstall, contents store.BundleInstallContents) (*model.BundleInstall, error) {
	s.updated = &contents
	return install, nil
}

type memoryCommandStore struct {
	store.CommandStore
	commands []*model.Command
}

func (s *memoryCommandStore) CommandsByApp(ctx context.Context, appID string) ([]*model.Command, error) {
	return s.commands, nil
}

type memoryEventListenerStore struct{ store.EventListenerStore }

func (s *memoryEventListenerStore) EventListenersByApp(ctx context.Context, appID string) ([]*model.EventListener, error) {
	return nil, nil
}

type memoryMessageStore struct{ store.MessageStore }

func (s *memoryMessageStore) MessagesByApp(ctx context.Context, appID string) ([]*model.Message, error) {
	return nil, nil
}

func hash(t *testing.T, f flow.FlowData) string {
	t.Helper()
	h, err := f.ContentHash()
	require.NoError(t, err)
	return h
}

// withNode is the flow with another block, like an older version or an edit.
func withNode(f flow.FlowData, id string) flow.FlowData {
	f.Nodes = append(append([]flow.FlowNode{}, f.Nodes...), flow.FlowNode{ID: id, Type: flow.FlowNodeTypeActionLog})
	return f
}

// moderationInstall is an install of the previous version of Moderation. Ban
// is unchanged since, unban is already like the new version, kick was edited
// and mute deleted.
func moderationInstall(t *testing.T) (*BundleInstallHandler, *memoryInstallStore) {
	b := bundle.Get("moderation")
	require.NotNil(t, b)
	sel, err := b.Select([]string{"ban", "unban", "kick", "mute"}, map[string]string{})
	require.NoError(t, err)

	install := &model.BundleInstall{
		ID:            "install",
		AppID:         "app",
		BundleID:      b.ID,
		BundleVersion: b.Version - 1,
		Settings:      map[string]string{},
	}
	commands := &memoryCommandStore{}
	for _, cmd := range sel.Commands {
		installed := withNode(cmd.FlowSource, "old")
		current := installed
		switch cmd.ID {
		case "unban":
			installed, current = cmd.FlowSource, cmd.FlowSource
		case "kick":
			current = withNode(installed, "edit")
		}

		id := "command_" + cmd.ID
		install.Items = append(install.Items, model.BundleInstallItem{
			Key:  cmd.ID,
			Type: model.BundleInstallItemTypeCommand,
			ID:   id,
			Hash: hash(t, installed),
		})
		if cmd.ID == "mute" {
			continue
		}
		compiled, err := flow.CompileCommand(cmd.FlowSource)
		require.NoError(t, err)
		commands.commands = append(commands.commands, &model.Command{
			ID:              id,
			Name:            compiled.CommandName(),
			AppID:           "app",
			BundleInstallID: null.StringFrom(install.ID),
			FlowSource:      current,
			Enabled:         true,
		})
	}

	installs := &memoryInstallStore{installs: []*model.BundleInstall{install}}
	h := NewBundleInstallHandler(
		installs,
		commands,
		&memoryEventListenerStore{},
		&memoryMessageStore{},
		commandhandler.NewCommandHandler(commands, nil),
		nil,
		nil,
	)
	return h, installs
}

func update(t *testing.T, h *BundleInstallHandler, installID string, req wire.BundleInstallUpdateRequest) (int, map[string]any) {
	t.Helper()

	srv := handler.APIHandler(func(c *handler.Context) error {
		c.App = &model.App{ID: "app"}
		return handler.TypedWithBody(h.HandleBundleInstallUpdate)(c)
	})

	body, err := json.Marshal(req)
	require.NoError(t, err)
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")
	r.SetPathValue("installID", installID)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, r)

	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	return rec.Code, res
}

func updatedCommands(s *memoryInstallStore) []string {
	var names []string
	for _, cmd := range s.updated.Commands {
		names = append(names, cmd.ID)
	}
	return names
}

func itemStates(res map[string]any) map[string]string {
	states := map[string]string{}
	for _, raw := range res["data"].(map[string]any)["items"].([]any) {
		item := raw.(map[string]any)
		states[item["key"].(string)] = fmt.Sprintf("edited=%v deleted=%v", item["edited"], item["deleted"])
	}
	return states
}

func TestBundleInstallUpdate(t *testing.T) {
	version := bundle.Get("moderation").Version

	h, installs := moderationInstall(t)
	code, res := update(t, h, "install", wire.BundleInstallUpdateRequest{BundleVersion: version})
	require.Equal(t, http.StatusOK, code, res)

	// Unban needs no change and the edited kick is kept.
	assert.Equal(t, []string{"command_ban"}, updatedCommands(installs))
	sel, err := bundle.Get("moderation").Select([]string{"ban"}, map[string]string{})
	require.NoError(t, err)
	assert.Equal(t, hash(t, sel.Commands[0].FlowSource), hash(t, installs.updated.Commands[0].FlowSource))

	assert.Equal(t, float64(version), res["data"].(map[string]any)["bundle_version"])
	assert.Equal(t, map[string]string{
		"ban":   "edited=false deleted=false",
		"unban": "edited=false deleted=false",
		"kick":  "edited=true deleted=false",
		"mute":  "edited=false deleted=true",
	}, itemStates(res))
}

func TestBundleInstallUpdateOverwrite(t *testing.T) {
	h, installs := moderationInstall(t)
	code, res := update(t, h, "install", wire.BundleInstallUpdateRequest{
		BundleVersion: bundle.Get("moderation").Version,
		Overwrite:     []string{"kick", "mute"},
	})
	require.Equal(t, http.StatusOK, code, res)

	// The deleted mute isn't added again.
	assert.ElementsMatch(t, []string{"command_ban", "command_kick"}, updatedCommands(installs))
	assert.Equal(t, "edited=false deleted=false", itemStates(res)["kick"])
}

func TestBundleInstallUpdateErrors(t *testing.T) {
	version := bundle.Get("moderation").Version

	h, installs := moderationInstall(t)
	code, res := update(t, h, "install", wire.BundleInstallUpdateRequest{BundleVersion: version + 1})
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "bundle_version_changed", res["error"].(map[string]any)["code"])

	code, res = update(t, h, "other", wire.BundleInstallUpdateRequest{BundleVersion: version})
	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "unknown_bundle_install", res["error"].(map[string]any)["code"])

	installs.installs[0].BundleVersion = version
	code, res = update(t, h, "install", wire.BundleInstallUpdateRequest{BundleVersion: version})
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "bundle_install_up_to_date", res["error"].(map[string]any)["code"])
	assert.Nil(t, installs.updated)
}
