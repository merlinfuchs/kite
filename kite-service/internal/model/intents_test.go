package model

import (
	"encoding/json"
	"testing"

	"github.com/diamondburned/arikawa/v3/gateway"
)

func TestIntentsJSON(t *testing.T) {
	var i Intents
	if err := json.Unmarshal([]byte(`["message_content","guild_messages"]`), &i); err != nil {
		t.Fatal(err)
	}
	if gateway.Intents(i) != gateway.IntentGuildMessages|gateway.IntentMessageContent {
		t.Errorf("intents = %d", i)
	}
	raw, _ := json.Marshal(i)
	if string(raw) != `["guild_messages","message_content"]` {
		t.Errorf("JSON = %s, want the names in the order of their bits", raw)
	}
	if err := json.Unmarshal([]byte(`["presence"]`), &i); err == nil {
		t.Error("unknown intent parsed")
	}
}
