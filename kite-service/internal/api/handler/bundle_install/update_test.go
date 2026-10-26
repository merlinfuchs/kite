package bundleinstall

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	commandhandler "github.com/kitecloud/kite/kite-service/internal/api/handler/command"
	variablehandler "github.com/kitecloud/kite/kite-service/internal/api/handler/variable"
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
	installs  []*model.BundleInstall
	commands  *memoryCommandStore
	variables *memoryVariableStore
	changes   *store.BundleInstallChanges
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

func (s *memoryInstallStore) InstalledContentsByApp(ctx context.Context, appID string) (store.BundleInstallContents, error) {
	var res store.BundleInstallContents
	for _, cmd := range s.commands.commands {
		if cmd.BundleInstallID.Valid {
			res.Commands = append(res.Commands, cmd)
		}
	}
	for _, v := range s.variables.variables {
		if v.BundleInstallID.Valid {
			res.Variables = append(res.Variables, v)
		}
	}
	return res, nil
}

func (s *memoryInstallStore) UpdateBundleInstall(ctx context.Context, install *model.BundleInstall, fromUpdatedAt time.Time, changes store.BundleInstallChanges, userID string) (*model.BundleInstall, error) {
	s.changes = &changes
	return install, nil
}

// memoryBundleStore has the official bundles, or the one bundle it's given.
type memoryBundleStore struct {
	store.BundleStore
	bundle *bundle.Bundle
}

func (s *memoryBundleStore) Bundle(ctx context.Context, id string) (*model.Bundle, error) {
	b := officialBundle(id)
	if s.bundle != nil && s.bundle.ID == id {
		b = s.bundle
	}
	if b == nil {
		return nil, store.ErrNotFound
	}
	content, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	return &model.Bundle{ID: b.ID, Version: b.Version, Content: content}, nil
}

func officialBundle(id string) *bundle.Bundle {
	for _, b := range bundle.Official() {
		if b.ID == id {
			return b
		}
	}
	return nil
}

type memoryCommandStore struct {
	store.CommandStore
	commands []*model.Command
}

type memoryVariableStore struct {
	store.VariableStore
	variables []*model.Variable
}

func (s *memoryVariableStore) VariablesByAppWithoutTotals(ctx context.Context, appID string) ([]*model.Variable, error) {
	return s.variables, nil
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
	return moderationInstallOf(t, officialBundle("moderation"))
}

func moderationInstallOf(t *testing.T, b *bundle.Bundle, variables ...*model.Variable) (*BundleInstallHandler, *memoryInstallStore) {
	require.NotNil(t, b)
	sel, err := b.Select([]string{"ban", "unban", "kick", "mute"}, map[string]string{})
	require.NoError(t, err)

	install := &model.BundleInstall{
		ID:            "install",
		AppID:         "app",
		BundleID:      b.ID,
		BundleVersion: b.Version - 1,
		Parameters:    map[string]string{},
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

	variableStore := &memoryVariableStore{variables: variables}
	installs := &memoryInstallStore{installs: []*model.BundleInstall{install}, commands: commands, variables: variableStore}
	h := NewBundleInstallHandler(
		installs,
		&memoryBundleStore{bundle: b},
		commands,
		nil,
		nil,
		variableStore,
		commandhandler.NewCommandHandler(commands, nil),
		nil,
		nil,
		variablehandler.NewVariableHandler(variableStore, nil),
		100,
	)
	return h, installs
}

func update(t *testing.T, h *BundleInstallHandler, installID string, req wire.BundleInstallUpdateRequest) (int, map[string]any) {
	t.Helper()
	return call(t, handler.TypedWithBody(h.HandleBundleInstallUpdate), installID, req)
}

func call(t *testing.T, h handler.HandlerFunc, installID string, req any) (int, map[string]any) {
	t.Helper()

	srv := handler.APIHandler(func(c *handler.Context) error {
		c.App = &model.App{ID: "app"}
		c.Session = &model.Session{UserID: "user"}
		return h(c)
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
	for _, cmd := range s.changes.Updated.Commands {
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
	version := officialBundle("moderation").Version

	h, installs := moderationInstall(t)
	code, res := update(t, h, "install", wire.BundleInstallUpdateRequest{BundleVersion: version})
	require.Equal(t, http.StatusOK, code, res)

	// Unban needs no change and the edited kick is kept.
	assert.Equal(t, []string{"command_ban"}, updatedCommands(installs))
	sel, err := officialBundle("moderation").Select([]string{"ban"}, map[string]string{})
	require.NoError(t, err)
	assert.Equal(t, hash(t, sel.Commands[0].FlowSource), hash(t, installs.changes.Updated.Commands[0].FlowSource))

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
		BundleVersion: officialBundle("moderation").Version,
		Overwrite:     []string{"kick", "mute"},
	})
	require.Equal(t, http.StatusOK, code, res)

	// The deleted mute isn't added again.
	assert.ElementsMatch(t, []string{"command_ban", "command_kick"}, updatedCommands(installs))
	assert.Equal(t, "edited=false deleted=false", itemStates(res)["kick"])
}

func TestBundleInstallUpdateErrors(t *testing.T) {
	version := officialBundle("moderation").Version

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
	assert.Nil(t, installs.changes)
}

// A new version that uses a variable gets it created.
func TestBundleInstallUpdateCreatesVariables(t *testing.T) {
	b := *officialBundle("moderation")
	b.Variables = []bundle.Variable{{ID: "bans", Name: "bans", Scoped: true}}
	b.Commands = slices.Clone(b.Commands)
	var ban map[string]any
	require.NoError(t, json.Unmarshal(b.Commands[0].FlowSource, &ban))
	ban["nodes"] = append(ban["nodes"].([]any), map[string]any{
		"id":   "count",
		"type": flow.FlowNodeTypeActionVariableSet,
		"data": map[string]any{"variable_id": "bans"},
	})
	raw, err := json.Marshal(ban)
	require.NoError(t, err)
	b.Commands[0].FlowSource = raw

	// The user made a variable with the name by hand, which isn't used.
	h, installs := moderationInstallOf(t, &b, &model.Variable{ID: "own", Name: "bans"})
	// An earlier version had it, and the user deleted it.
	install := installs.installs[0]
	install.Items = append(install.Items, model.BundleInstallItem{Key: "bans", Type: model.BundleInstallItemTypeVariable, ID: "deleted"})
	code, res := update(t, h, "install", wire.BundleInstallUpdateRequest{BundleVersion: b.Version})
	require.Equal(t, http.StatusOK, code, res)

	created := installs.changes.Created.Variables
	require.Len(t, created, 1)
	assert.Equal(t, "bans_2", created[0].Name)
	assert.True(t, created[0].Scoped)
	assert.Equal(t, "install", created[0].BundleInstallID.String)

	updated := installs.changes.Updated.Commands
	require.Len(t, updated, 1)
	nodes := updated[0].FlowSource.Nodes
	assert.Equal(t, created[0].ID, nodes[len(nodes)-1].Data.VariableID)
	assert.Equal(t, "edited=false deleted=false", itemStates(res)["bans"])
	// The created one replaces the deleted one.
	bans := 0
	for _, item := range install.Items {
		if item.Key == "bans" {
			bans++
		}
	}
	assert.Equal(t, 1, bans)
}

// moderationWithReason is Moderation with unban using a reason parameter, and
// the install's unban having it with an older label and a value.
func moderationWithReason(t *testing.T) (*BundleInstallHandler, *memoryInstallStore, *bundle.Bundle) {
	b := *officialBundle("moderation")
	b.Parameters = []bundle.Parameter{{Key: "reason", Label: "Reason", Type: flow.FlowParameterTypeText}}
	b.Commands = slices.Clone(b.Commands)
	var unban map[string]any
	require.NoError(t, json.Unmarshal(b.Commands[1].FlowSource, &unban))
	unban["nodes"] = append(unban["nodes"].([]any), map[string]any{
		"id":   "log",
		"type": flow.FlowNodeTypeActionLog,
		"data": map[string]any{"log_message": "{{params.reason}}"},
	})
	raw, err := json.Marshal(unban)
	require.NoError(t, err)
	b.Commands[1].FlowSource = raw

	h, installs := moderationInstallOf(t, &b)
	for _, cmd := range h.commandStore.(*memoryCommandStore).commands {
		if cmd.ID == "command_unban" {
			cmd.FlowSource.Parameters = []flow.FlowParameter{{Key: "reason", Label: "Old Label", Type: flow.FlowParameterTypeText, Value: "kept"}}
		}
	}
	return h, installs, &b
}

func updatedCommand(s *memoryInstallStore, id string) *model.Command {
	for _, cmd := range s.changes.Updated.Commands {
		if cmd.ID == id {
			return cmd
		}
	}
	return nil
}

// A new label of a parameter is an update, and the value is kept.
func TestBundleInstallUpdateParameterDefinition(t *testing.T) {
	h, installs, b := moderationWithReason(t)
	code, res := update(t, h, "install", wire.BundleInstallUpdateRequest{BundleVersion: b.Version})
	require.Equal(t, http.StatusOK, code, res)

	unban := updatedCommand(installs, "command_unban")
	require.NotNil(t, unban)
	assert.Equal(t, []flow.FlowParameter{{Key: "reason", Label: "Reason", Type: flow.FlowParameterTypeText, Value: "kept"}}, unban.FlowSource.Parameters)
}

// A value the update sets replaces the one the flow has.
func TestBundleInstallUpdateSetsParameterValue(t *testing.T) {
	h, installs, b := moderationWithReason(t)
	code, res := update(t, h, "install", wire.BundleInstallUpdateRequest{
		BundleVersion: b.Version,
		Parameters:    map[string]string{"reason": "new"},
	})
	require.Equal(t, http.StatusOK, code, res)
	assert.Equal(t, "new", updatedCommand(installs, "command_unban").FlowSource.ParameterValues()["reason"])
}

// A value that doesn't fit the new type of a parameter is cleared instead of
// failing the update.
func TestBundleInstallUpdateParameterType(t *testing.T) {
	h, installs, b := moderationWithReason(t)
	b.Parameters[0].Type = flow.FlowParameterTypeChannel
	code, res := update(t, h, "install", wire.BundleInstallUpdateRequest{BundleVersion: b.Version})
	require.Equal(t, http.StatusOK, code, res)
	assert.Equal(t, "", updatedCommand(installs, "command_unban").FlowSource.ParameterValues()["reason"])
}

// Adding a template the app has goes into its install, which has to be up to
// date, and can't add what it still has.
func TestBundleInstallCreateExisting(t *testing.T) {
	h, installs := moderationInstall(t)
	version := officialBundle("moderation").Version
	create := handler.TypedWithBody(h.HandleBundleInstallCreate)

	code, res := call(t, create, "", wire.BundleInstallCreateRequest{BundleID: "moderation", BundleVersion: version, Items: []string{"mute"}})
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "bundle_install_outdated", res["error"].(map[string]any)["code"])

	installs.installs[0].BundleVersion = version
	code, res = call(t, create, "", wire.BundleInstallCreateRequest{BundleID: "moderation", BundleVersion: version, Items: []string{"ban"}})
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "already_installed", res["error"].(map[string]any)["code"])
	assert.Nil(t, installs.changes)
}
