package bundle

import (
	"encoding/json"
	"strings"
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

		var raws []json.RawMessage
		for _, c := range sel.Commands {
			raws = append(raws, c.FlowSource)
		}
		for _, l := range sel.EventListeners {
			raws = append(raws, l.FlowSource)
		}
		for _, m := range sel.Messages {
			var data message.MessageData
			require.NoError(t, json.Unmarshal(m.Data, &data), b.ID)
			raws = append(raws, m.Data)
			for _, f := range m.FlowSources {
				raws = append(raws, f)
			}
		}
		for _, raw := range raws {
			assert.NotContains(t, string(raw), "$setting.", b.ID)
			var f flow.FlowData
			require.NoError(t, json.Unmarshal(raw, &f), b.ID)
		}
	}
}

func TestSelect(t *testing.T) {
	verification := Get("verification")
	require.NotNil(t, verification)
	keys := allKeys(verification)

	_, err := verification.Select(keys, map[string]string{"channel_id": "123456789012345678"})
	assert.ErrorContains(t, err, "is required")

	_, err = verification.Select(keys, map[string]string{"role_id": "@everyone"})
	assert.ErrorContains(t, err, "must be the ID of a role")

	_, err = verification.Select(keys, map[string]string{"role_id": "123456789012345678", "other": "x"})
	assert.ErrorContains(t, err, "unknown setting")

	_, err = verification.Select([]string{"nope"}, map[string]string{})
	assert.Error(t, err)

	sel, err := verification.Select(keys, map[string]string{"role_id": "123456789012345678"})
	require.NoError(t, err)
	for _, f := range sel.Messages[0].FlowSources {
		assert.Contains(t, string(f), `"role_target":"123456789012345678"`)
	}
}

func TestSelectLeavesOutEmptyOptionalSettings(t *testing.T) {
	askAI := Get("ask-ai")
	require.NotNil(t, askAI)

	sel, err := askAI.Select([]string{askAI.Commands[0].ID}, map[string]string{})
	require.NoError(t, err)
	assert.NotContains(t, string(sel.Commands[0].FlowSource), "system_prompt")

	sel, err = askAI.Select([]string{askAI.Commands[0].ID}, map[string]string{"system_prompt": "Be nice. $setting.system_prompt"})
	require.NoError(t, err)
	// A value isn't searched for settings again.
	assert.Equal(t, 1, strings.Count(string(sel.Commands[0].FlowSource), "$setting.system_prompt"))
}
