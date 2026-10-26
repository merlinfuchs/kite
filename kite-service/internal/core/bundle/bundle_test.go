package bundle

import (
	"encoding/json"
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func allKeys(b *Bundle) []string {
	var keys []string
	for _, c := range b.Commands {
		keys = append(keys, c.ID)
	}
	for _, l := range b.EventListeners {
		keys = append(keys, l.ID)
	}
	for _, m := range b.Messages {
		keys = append(keys, m.ID)
	}
	for _, v := range b.Variables {
		keys = append(keys, v.ID)
	}
	return keys
}

// Every item of every bundle installs once its parameters are filled in, and
// its flows get the values of the parameters they use.
func TestEveryBundleInstalls(t *testing.T) {
	require.NotEmpty(t, official)
	for _, b := range Official() {
		values := map[string]string{}
		for _, p := range b.Parameters {
			if p.Type == flow.FlowParameterTypeChannel || p.Type == flow.FlowParameterTypeRole {
				values[p.Key] = "123456789012345678"
			} else {
				values[p.Key] = "text"
			}
		}

		sel, err := b.Select(allKeys(b), values)
		require.NoError(t, err, b.ID)

		for _, f := range sel.Flows() {
			params := f.ParameterValues()
			for _, key := range f.ReferencedParameters() {
				assert.Equal(t, values[key], params[key], b.ID)
			}
		}

		raw, err := json.Marshal(sel)
		require.NoError(t, err)
		// Bundles can't have attachments, the assets aren't stored with them.
		assert.NotContains(t, string(raw), "asset_id", b.ID)
	}
}

func get(t *testing.T, id string) *Bundle {
	t.Helper()
	for _, b := range Official() {
		if b.ID == id {
			return b
		}
	}
	t.Fatalf("no bundle %s", id)
	return nil
}

// A bundle is the same after it was stored.
func TestParse(t *testing.T) {
	for _, b := range Official() {
		content, err := json.Marshal(b)
		require.NoError(t, err)
		parsed, err := Parse(content)
		require.NoError(t, err)
		stored, err := json.Marshal(parsed)
		require.NoError(t, err)
		assert.JSONEq(t, string(content), string(stored))
		assert.Equal(t, allKeys(b), allKeys(parsed))
	}
}

func TestSelect(t *testing.T) {
	verification := get(t, "verification")
	keys := allKeys(verification)
	valid := map[string]string{"role_id": "123456789012345678", "channel_id": "223456789012345678"}

	_, err := verification.Select(keys, map[string]string{"channel_id": "123456789012345678"})
	var invalidErr *InvalidError
	assert.ErrorAs(t, err, &invalidErr)
	assert.ErrorContains(t, err, "Verified Role is required")

	// The channel is only used to post the message.
	_, err = verification.Select(keys, map[string]string{"role_id": "123456789012345678"})
	assert.ErrorContains(t, err, "Channel is required")

	for _, role := range []string{"@everyone", "0", "null", "-1"} {
		_, err = verification.Select(keys, map[string]string{"role_id": role, "channel_id": "223456789012345678"})
		assert.ErrorContains(t, err, "must be the ID of a role", role)
	}

	_, err = verification.Select(keys, map[string]string{"role_id": "123456789012345678", "other": "x"})
	assert.ErrorContains(t, err, "unknown parameter")

	_, err = verification.Select([]string{"nope", "other"}, valid)
	assert.ErrorContains(t, err, "no items nope, other")

	// The blocks keep using the parameter, which gets the value.
	sel, err := verification.Select(keys, valid)
	require.NoError(t, err)
	for _, f := range sel.Messages[0].FlowSources {
		raw, err := json.Marshal(f.Nodes)
		require.NoError(t, err)
		assert.Contains(t, string(raw), `"role_target":"{{params.role_id}}"`)
		assert.Equal(t, []flow.FlowParameter{{
			Key:         "role_id",
			Label:       "Verified Role",
			Description: verification.Parameters[1].Description,
			Type:        flow.FlowParameterTypeRole,
			Required:    true,
			Value:       "123456789012345678",
		}}, f.Parameters)
	}
}

// An optional parameter without a value is empty, which blocks treat like a
// field that isn't set.
func TestSelectOptionalParameter(t *testing.T) {
	askAI := get(t, "ask-ai")

	sel, err := askAI.Select([]string{askAI.Commands[0].ID}, map[string]string{})
	require.NoError(t, err)
	params := sel.Commands[0].FlowSource.Parameters
	require.Len(t, params, 1)
	assert.Equal(t, "system_prompt", params[0].Key)
	assert.Equal(t, "", params[0].Value)
}

// The changelog of official templates goes with them into the store.
func TestOfficialChangelog(t *testing.T) {
	for _, o := range official {
		for _, n := range o.changelog {
			assert.NotEmpty(t, n.Notes, o.bundle.ID)
			assert.LessOrEqual(t, n.Version, o.bundle.Version, o.bundle.ID)
		}
		if o.bundle.ID == "verification" {
			require.Len(t, o.changelog, 1)
			assert.Equal(t, 2, o.changelog[0].Version)
		}
	}
}

// Only selected messages that allow it can be changed while adding a bundle,
// and they remember the data of the bundle.
func TestSelectionCustomize(t *testing.T) {
	b := &Bundle{Messages: []Message{
		{ID: "sticky", Name: "Sticky", Data: json.RawMessage(`{"content":"Default"}`), Customize: true},
		{ID: "panel", Name: "Panel", Data: json.RawMessage(`{"content":"Panel"}`)},
		{ID: "other", Name: "Other", Data: json.RawMessage(`{}`), Customize: true},
	}}
	custom := message.MessageData{Content: "Mine"}

	sel, err := b.Select([]string{"sticky", "panel"}, map[string]string{})
	require.NoError(t, err)
	require.NoError(t, sel.Customize(map[string]message.MessageData{"sticky": custom}))
	assert.Equal(t, custom, sel.Messages[0].Data)
	assert.Equal(t, &message.MessageData{Content: "Default"}, sel.Messages[0].Default)
	assert.Nil(t, sel.Messages[1].Default)

	for _, key := range []string{"panel", "other", "nope"} {
		sel, err := b.Select([]string{"sticky", "panel"}, map[string]string{})
		require.NoError(t, err)
		assert.ErrorContains(t, sel.Customize(map[string]message.MessageData{key: custom}), "the message "+key+" can't be changed")
	}
}
