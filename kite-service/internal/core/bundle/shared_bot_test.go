package bundle

import (
	"encoding/json"
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

	"action_expression_evaluate":     true,
	"action_member_ban":              true,
	"action_member_kick":             true,
	"action_member_role_add":         true,
	"action_member_timeout":          true,
	"action_member_unban":            true,
	"action_message_create":          true,
	"action_message_delete":          true,
	"action_message_edit":            true,
	"action_message_list":            true,
	"action_message_reaction_create": true,
	"action_random_generate":         true,
	"action_response_create":         true,
	"action_variable_get":            true,
	"action_variable_list":           true,
	"action_variable_set":            true,
}

// sharedBotEvents are the events the Kite bot gets with its intents, see
// sharedbot.Intents. Messages come without their text, unless the bundle
// reads it, see SharedBotIntents.
var sharedBotEvents = map[string]bool{
	"message_create":          true,
	"message_update":          true,
	"message_delete":          true,
	"message_reaction_add":    true,
	"message_reaction_remove": true,
}

// readsContent matches what reads the text of messages in a flow, like
// event.message.content in a placeholder or an event filter on it. Get
// Message is there for the text too. A bundle that reads it without saying so
// gets empty text on the Kite bot.
var readsContent = regexp.MustCompile(`\.(content|embeds|attachments|components)\b|"action_message_get"|"event_filter_target":"message_content"`)

// readsMembers is whether a bundle listening to the event needs member
// events, like for apps with a bot of their own.
func readsMembers(eventType string) bool {
	return model.AppGatewayRequirements{EventListenerTypes: []model.EventListenerType{model.EventListenerType(eventType)}}.NeedsGuildMembers()
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

		var content, members bool
		for _, raw := range bundleFlows(b) {
			content = content || readsContent.Match(raw)

			var f flow.FlowData
			require.NoError(t, json.Unmarshal(raw, &f), b.ID)

			for _, node := range f.Nodes {
				typ := string(node.Type)
				allowed := sharedBotBlocks[node.Type] ||
					strings.HasPrefix(typ, "control_") ||
					strings.HasPrefix(typ, "option_")
				require.True(t, allowed, "%s uses %s, which the Kite bot can't run", b.ID, typ)

				if node.Type == flow.FlowNodeTypeEntryEvent {
					reads := readsMembers(node.Data.EventType)
					members = members || reads
					require.True(t, sharedBotEvents[node.Data.EventType] || reads, "%s listens to %s, which the Kite bot doesn't get", b.ID, node.Data.EventType)
				}
			}
		}

		// Apps only get message text and member events with a bundle that
		// says it reads them, so a bundle has to name exactly what it reads.
		var want []string
		if content {
			want = append(want, SharedBotIntentMessageContent)
		}
		if members {
			want = append(want, SharedBotIntentGuildMembers)
		}
		require.ElementsMatch(t, want, b.SharedBotIntents, "%s: shared_bot_intents must name what it reads", b.ID)
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
