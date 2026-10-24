package bundle

import (
	"encoding/json"
	"strings"
	"testing"

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
	return keys
}

// Every item of every bundle installs once its settings are filled in.
func TestEveryBundleInstalls(t *testing.T) {
	require.NotEmpty(t, bundles())
	for _, b := range bundles() {
		values := map[string]string{}
		for _, s := range b.Settings {
			if s.Type == SettingTypeChannel || s.Type == SettingTypeRole {
				values[s.Key] = "123456789012345678"
			} else {
				values[s.Key] = "text"
			}
		}

		sel, err := b.Select(allKeys(b), values)
		require.NoError(t, err, b.ID)

		raw, err := json.Marshal(sel)
		require.NoError(t, err)
		assert.NotContains(t, string(raw), "$setting.", b.ID)
	}
}

func TestSelect(t *testing.T) {
	verification := Get("verification")
	require.NotNil(t, verification)
	keys := allKeys(verification)
	valid := map[string]string{"role_id": "123456789012345678", "channel_id": "223456789012345678"}

	_, err := verification.Select(keys, map[string]string{"channel_id": "123456789012345678"})
	assert.ErrorContains(t, err, "Verified Role is required")

	// The channel is only used to post the message.
	_, err = verification.Select(keys, map[string]string{"role_id": "123456789012345678"})
	assert.ErrorContains(t, err, "Channel is required")

	for _, role := range []string{"@everyone", "0", "null", "-1"} {
		_, err = verification.Select(keys, map[string]string{"role_id": role, "channel_id": "223456789012345678"})
		assert.ErrorContains(t, err, "must be the ID of a role", role)
	}

	_, err = verification.Select(keys, map[string]string{"role_id": "123456789012345678", "other": "x"})
	assert.ErrorContains(t, err, "unknown setting")

	_, err = verification.Select([]string{"nope", "other"}, valid)
	assert.ErrorContains(t, err, "no items nope, other")

	sel, err := verification.Select(keys, valid)
	require.NoError(t, err)
	for _, f := range sel.Messages[0].FlowSources {
		raw, err := json.Marshal(f)
		require.NoError(t, err)
		assert.Contains(t, string(raw), `"role_target":"123456789012345678"`)
	}
}

func TestSelectLeavesOutEmptyOptionalSettings(t *testing.T) {
	askAI := Get("ask-ai")
	require.NotNil(t, askAI)

	sel, err := askAI.Select([]string{askAI.Commands[0].ID}, map[string]string{})
	require.NoError(t, err)
	raw, err := json.Marshal(sel.Commands[0].FlowSource)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "system_prompt")

	sel, err = askAI.Select([]string{askAI.Commands[0].ID}, map[string]string{"system_prompt": "Be nice. $setting.system_prompt"})
	require.NoError(t, err)
	raw, err = json.Marshal(sel.Commands[0].FlowSource)
	require.NoError(t, err)
	// A value isn't searched for settings again.
	assert.Equal(t, 1, strings.Count(string(raw), "$setting.system_prompt"))
}
