package bundle

import (
	"encoding/json"
	"github.com/diamondburned/arikawa/v3/gateway"
	"regexp"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/stretchr/testify/require"
)

// sharedBotBlocks are the blocks bundles for the Kite bot can use. Others are
// left out on purpose, like DMs, HTTP requests, code, the bot's status and
// voice, or because no bundle for the Kite bot needs them yet. Add a block
// once it's checked that it works in the view of the Kite bot an app gets,
// see the sharedbot package.
var sharedBotBlocks = map[flow.FlowNodeType]bool{
	"entry_command":          true,
	"entry_component_button": true,
	"entry_event":            true,

	"action_channel_get":             true,
	"action_expression_evaluate":     true,
	"action_member_ban":              true,
	"action_member_kick":             true,
	"action_member_role_add":         true,
	"action_member_timeout":          true,
	"action_member_unban":            true,
	"action_message_create":          true,
	"action_message_delete":          true,
	"action_message_edit":            true,
	"action_message_get":             true,
	"action_message_list":            true,
	"action_message_reaction_create": true,
	"action_random_generate":         true,
	"action_response_create":         true,
	"action_variable_delete":         true,
	"action_variable_get":            true,
	"action_variable_list":           true,
	"action_variable_set":            true,
}

// sharedBotIntents are the intents the Kite bot can get, with the
// privileged ones once Discord grants them. Templates that need others, like
// voice states or DMs, can't run on it.
const sharedBotIntents = model.Intents(gateway.IntentGuildMessages | gateway.IntentGuildMessageReactions | gateway.IntentGuildMembers | gateway.IntentMessageContent)

// readsContent matches what reads the text of messages in a flow, like
// event.message.content in a placeholder or an event filter on it. Get
// Message is there for the text too. A bundle that reads it without saying so
// gets empty text on the Kite bot.
var readsContent = regexp.MustCompile(`\.(content|embeds|attachments|components)\b|"action_message_get"|"event_filter_target":\s*"message_content"`)

// neededIntents are the intents the bundle's flows need: what its event
// listeners need, like for apps with a bot of their own, and Message Content
// if a flow reads the text of messages.
func neededIntents(t *testing.T, b *Bundle) model.Intents {
	var reqs model.AppGatewayRequirements
	var content bool
	for _, raw := range bundleFlows(b) {
		content = content || readsContent.Match(raw)

		var f flow.FlowData
		require.NoError(t, json.Unmarshal(raw, &f), b.ID)
		for _, node := range f.Nodes {
			if node.Type == flow.FlowNodeTypeEntryEvent {
				reqs.EventListenerTypes = append(reqs.EventListenerTypes, model.EventListenerType(node.Data.EventType))
			}
		}
	}
	intents := reqs.Intents()
	if content {
		intents |= model.Intents(gateway.IntentMessageContent)
	}
	return intents
}

func TestBundleIntents(t *testing.T) {
	for _, b := range Official() {
		require.Equal(t, neededIntents(t, b).Names(), b.Intents.Names(), "%s: intents must be what its flows need", b.ID)
	}
}

// bundleFlows are the flows of the bundle's commands, event listeners and
// messages.
func bundleFlows(b *Bundle) []json.RawMessage {
	var flows []json.RawMessage
	for _, c := range b.Commands {
		flows = append(flows, c.FlowSource)
	}
	for _, l := range b.EventListeners {
		flows = append(flows, l.FlowSource)
	}
	for _, m := range b.Messages {
		for _, f := range m.FlowSources {
			flows = append(flows, f)
		}
	}
	return flows
}

func TestSharedBotBundles(t *testing.T) {
	for _, b := range Official() {
		if !b.SharedBot {
			continue
		}

		for _, l := range b.EventListeners {
			require.Equal(t, "discord", l.Source, "%s: %s", b.ID, l.ID)
		}

		require.True(t, sharedBotIntents.Has(b.Intents), "%s needs intents the Kite bot can't get: %v", b.ID, b.Intents.Names())

		for _, raw := range bundleFlows(b) {
			var f flow.FlowData
			require.NoError(t, json.Unmarshal(raw, &f), b.ID)

			for _, node := range f.Nodes {
				typ := string(node.Type)
				allowed := sharedBotBlocks[node.Type] ||
					strings.HasPrefix(typ, "control_") ||
					strings.HasPrefix(typ, "option_")
				require.True(t, allowed, "%s uses %s, which the Kite bot can't run", b.ID, typ)
			}
		}
	}
}

// readsContent finds the templates that read message text, also ones that
// aren't on the Kite bot yet.
func TestReadsContent(t *testing.T) {
	want := map[string]bool{
		"ai-automod":     true,
		"counting":       true,
		"starboard":      true,
		"leveling":       false,
		"sticky-message": false,
		"verification":   false,
		"welcomer":       false,
	}
	for _, b := range Official() {
		reads, ok := want[b.ID]
		if !ok {
			continue
		}
		delete(want, b.ID)

		var got bool
		for _, raw := range bundleFlows(b) {
			got = got || readsContent.Match(raw)
		}
		require.Equal(t, reads, got, b.ID)
	}
	require.Empty(t, want, "templates not found")
}

func TestSharedBotGranted(t *testing.T) {
	content := model.Intents(gateway.IntentMessageContent)
	members := model.Intents(gateway.IntentGuildMembers)
	b := &Bundle{Intents: content | model.Intents(gateway.IntentGuildMessages)}
	require.False(t, b.SharedBotGranted(0))
	require.False(t, b.SharedBotGranted(members))
	require.True(t, b.SharedBotGranted(members|content))
	require.True(t, (&Bundle{Intents: model.Intents(gateway.IntentGuildMessages)}).SharedBotGranted(0))
}
