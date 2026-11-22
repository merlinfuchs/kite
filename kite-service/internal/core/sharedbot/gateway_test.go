package sharedbot

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

// The Kite bot only asks for the privileged intents Discord granted, and for
// what the templates that can run with them need.
func TestIntents(t *testing.T) {
	none := Intents(0)
	if none.Has(gateway.IntentMessageContent) || none.Has(gateway.IntentGuildMembers) {
		t.Errorf("intents without grants = %d, want no privileged ones", none)
	}
	if !none.Has(gateway.IntentGuilds | gateway.IntentGuildMessages) {
		t.Errorf("intents without grants = %d, want guilds and messages", none)
	}

	all := Intents(model.Intents(gateway.IntentMessageContent | gateway.IntentGuildMembers))
	if !all.Has(gateway.IntentMessageContent | gateway.IntentGuildMembers | gateway.IntentGuildMessageReactions) {
		t.Errorf("intents with grants = %d, want content, members and the reactions of Starboard", all)
	}
}
