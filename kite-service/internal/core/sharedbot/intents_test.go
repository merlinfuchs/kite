package sharedbot

import (
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state/store/defaultstore"
	"github.com/diamondburned/arikawa/v3/utils/httputil/httpdriver"
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
	both.messageContent.Store(true)
	both.guildMembers.Store(true)

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

	intents.messageContent.Store(true)
	if got, err := scoped.Message(ownChannel, 1); err != nil || !hasText(*got) {
		t.Errorf("cached message with the intent = %v, %v, want its text", got, err)
	}
}

// set takes the intents of the bundles the app installed.
func TestAppIntentsSet(t *testing.T) {
	var intents appIntents
	intents.set([]string{"leveling", "starboard"})
	if intents.messageContent.Load() != slices.Contains(bundleIntents["starboard"], "message_content") {
		t.Error("message content doesn't follow the bundles")
	}
	intents.set(nil)
	if intents.messageContent.Load() || intents.guildMembers.Load() {
		t.Error("intents without bundles")
	}
}

type fakeResponse struct {
	status int
	body   string
}

func (r fakeResponse) GetStatus() int         { return r.status }
func (r fakeResponse) GetHeader() http.Header { return http.Header{} }
func (r fakeResponse) GetBody() io.ReadCloser { return io.NopCloser(strings.NewReader(r.body)) }

func TestFilterResponse(t *testing.T) {
	none := &appIntents{}
	read := func(resp httpdriver.Response) string {
		raw, _ := io.ReadAll(resp.GetBody())
		return string(raw)
	}

	one := fakeResponse{http.StatusOK, `{"id":"1","content":"secret","embeds":[{"title":"secret"}]}`}
	resp, err := filterResponse("GET", "/api/v10/channels/10/messages/1", one, none)
	if err != nil || strings.Contains(read(resp), "secret") {
		t.Errorf("message without the intent kept its text: %v", err)
	}

	list := fakeResponse{http.StatusOK, `[{"id":"1","content":"secret"},{"id":"2","content":"secret"}]`}
	resp, err = filterResponse("GET", "/api/v10/channels/10/messages", list, none)
	if err != nil || strings.Contains(read(resp), "secret") {
		t.Errorf("messages without the intent kept their text: %v", err)
	}

	with := &appIntents{}
	with.messageContent.Store(true)
	if resp, _ := filterResponse("GET", "/api/v10/channels/10/messages/1", one, with); !strings.Contains(read(resp), "secret") {
		t.Error("message with the intent lost its text")
	}
	other := fakeResponse{http.StatusOK, `{"name":"secret"}`}
	if resp, _ := filterResponse("GET", "/api/v10/channels/10", other, none); !strings.Contains(read(resp), "secret") {
		t.Error("other responses changed")
	}
}
