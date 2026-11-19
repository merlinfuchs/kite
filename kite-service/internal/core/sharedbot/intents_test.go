package sharedbot

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state/store/defaultstore"
	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
)

func textMessage() discord.Message {
	return discord.Message{
		ID:          1,
		ChannelID:   ownChannel,
		GuildID:     ownGuild,
		Author:      discord.User{ID: 5},
		Content:     "secret",
		Embeds:      []discord.Embed{{Title: "secret"}},
		Attachments: []discord.Attachment{{Filename: "secret.png"}},
		ReferencedMessage: &discord.Message{
			ID:      2,
			Content: "secret",
		},
	}
}

func hasText(m discord.Message) bool {
	return m.Content != "" || len(m.Embeds) > 0 || len(m.Attachments) > 0 ||
		(m.ReferencedMessage != nil && hasText(*m.ReferencedMessage))
}

// Apps without a template that reads them get messages without their text
// and no member events.
func TestFilterEvent(t *testing.T) {
	none := &appIntents{}
	both := &appIntents{}
	both.set([]string{"test-content", "test-members"})

	created := &gateway.MessageCreateEvent{Message: textMessage()}
	e, ok := filterEvent(created, none)
	if !ok || hasText(e.(*gateway.MessageCreateEvent).Message) {
		t.Errorf("message without the intent = %v, %v, want it without text", e, ok)
	}
	if e.(*gateway.MessageCreateEvent).Author.ID != 5 {
		t.Error("the message lost its author")
	}
	if !hasText(created.Message) {
		t.Error("the event of the bot was changed")
	}
	if e, ok := filterEvent(created, both); !ok || !hasText(e.(*gateway.MessageCreateEvent).Message) {
		t.Error("message with the intent lost its text")
	}

	edited := &guildevent.MessageEdit{MessageUpdateEvent: gateway.MessageUpdateEvent{Message: textMessage()}}
	if e, ok := filterEvent(edited, none); !ok || hasText(e.(*guildevent.MessageEdit).Message) {
		t.Error("edit without the intent kept its text")
	}
	cached := textMessage()
	deleted := &guildevent.MessageDelete{Message: &cached}
	if e, ok := filterEvent(deleted, none); !ok || hasText(*e.(*guildevent.MessageDelete).Message) {
		t.Error("deleted message without the intent kept its text")
	}

	joined := &gateway.GuildMemberAddEvent{GuildID: ownGuild}
	if _, ok := filterEvent(joined, none); ok {
		t.Error("member event without the intent wasn't dropped")
	}
	if _, ok := filterEvent(joined, both); !ok {
		t.Error("member event with the intent was dropped")
	}

	reaction := &gateway.MessageReactionAddEvent{GuildID: ownGuild}
	if e, ok := filterEvent(reaction, none); !ok || e != reaction {
		t.Error("other events changed")
	}
}

func TestScopedMessagesWithoutContent(t *testing.T) {
	cabinet := defaultstore.New()
	cabinet.ChannelSet(&discord.Channel{ID: ownChannel, GuildID: ownGuild}, false)
	m := textMessage()
	cabinet.MessageSet(&m, false)

	intents := &appIntents{}
	scoped := scopedCabinet(cabinet, ownGuild, intents)
	got, err := scoped.Message(ownChannel, 1)
	if err != nil || hasText(*got) {
		t.Errorf("cached message without the intent = %v, %v, want it without text", got, err)
	}
	list, err := scoped.Messages(ownChannel)
	if err != nil || len(list) != 1 || hasText(list[0]) {
		t.Errorf("cached messages without the intent = %v, %v", list, err)
	}

	intents.set([]string{"test-content"})
	if got, err := scoped.Message(ownChannel, 1); err != nil || !hasText(*got) {
		t.Errorf("cached message with the intent = %v, %v, want its text", got, err)
	}
}

func init() {
	bundleIntents["test-content"] = []string{bundle.SharedBotIntentMessageContent}
	bundleIntents["test-members"] = []string{bundle.SharedBotIntentGuildMembers}
}
