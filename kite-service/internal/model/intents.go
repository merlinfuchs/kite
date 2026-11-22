package model

import (
	"encoding/json"
	"fmt"
	"math/bits"

	"github.com/diamondburned/arikawa/v3/gateway"
)

// Intents are gateway intents that read and write as Discord's names in
// lowercase, like "message_content", for files people edit, like templates.
type Intents gateway.Intents

// PrivilegedIntents need Discord's approval for bots that reach more than
// 10,000 users, like the Kite bot.
const PrivilegedIntents = Intents(gateway.IntentGuildMembers | gateway.IntentGuildPresences | gateway.IntentMessageContent)

// intentNames are the intents templates can need, by name. Guilds is left
// out, as every bot has it.
var intentNames = map[string]gateway.Intents{
	"guild_members":           gateway.IntentGuildMembers,
	"guild_moderation":        gateway.IntentGuildModeration,
	"guild_voice_states":      gateway.IntentGuildVoiceStates,
	"guild_messages":          gateway.IntentGuildMessages,
	"guild_message_reactions": gateway.IntentGuildMessageReactions,
	"direct_messages":         gateway.IntentDirectMessages,
	"message_content":         gateway.IntentMessageContent,
}

// ParseIntents returns the intents of the names.
func ParseIntents(names []string) (Intents, error) {
	var res Intents
	for _, name := range names {
		intent, ok := intentNames[name]
		if !ok {
			return 0, fmt.Errorf("unknown intent %q", name)
		}
		res |= Intents(intent)
	}
	return res, nil
}

// Names are the names of the intents, in the order of their bits.
func (i Intents) Names() []string {
	var names []string
	for rest := uint64(i); rest != 0; rest &= rest - 1 {
		bit := gateway.Intents(1) << bits.TrailingZeros64(rest)
		for name, intent := range intentNames {
			if intent == bit {
				names = append(names, name)
			}
		}
	}
	return names
}

func (i Intents) Has(other Intents) bool {
	return i&other == other
}

func (i Intents) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Names())
}

func (i *Intents) UnmarshalJSON(data []byte) error {
	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		return err
	}
	parsed, err := ParseIntents(names)
	if err != nil {
		return err
	}
	*i = parsed
	return nil
}
