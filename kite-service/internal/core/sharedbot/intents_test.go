package sharedbot

import (
	"context"
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

const botUser = discord.UserID(botApp)

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
			Author:  discord.User{ID: 6},
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
	e, ok := filterEvent(created, none, botUser)
	if !ok || hasText(e.(*gateway.MessageCreateEvent).Message) {
		t.Errorf("message without the intent = %v, %v, want it without text", e, ok)
	}
	if e.(*gateway.MessageCreateEvent).Author.ID != 5 {
		t.Error("the message lost its author")
	}
	if !hasText(created.Message) {
		t.Error("the event of the bot was changed")
	}
	if e, ok := filterEvent(created, both, botUser); !ok || !hasText(e.(*gateway.MessageCreateEvent).Message) {
		t.Error("message with the intent lost its text")
	}

	edited := &guildevent.MessageEdit{MessageUpdateEvent: gateway.MessageUpdateEvent{Message: textMessage()}}
	if e, ok := filterEvent(edited, none, botUser); !ok || hasText(e.(*guildevent.MessageEdit).Message) {
		t.Error("edit without the intent kept its text")
	}
	cached := textMessage()
	deleted := &guildevent.MessageDelete{Message: &cached}
	if e, ok := filterEvent(deleted, none, botUser); !ok || hasText(*e.(*guildevent.MessageDelete).Message) {
		t.Error("deleted message without the intent kept its text")
	}

	joined := &gateway.GuildMemberAddEvent{GuildID: ownGuild}
	if _, ok := filterEvent(joined, none, botUser); ok {
		t.Error("member event without the intent wasn't dropped")
	}
	if _, ok := filterEvent(joined, both, botUser); !ok {
		t.Error("member event with the intent was dropped")
	}

	reaction := &gateway.MessageReactionAddEvent{GuildID: ownGuild}
	if e, ok := filterEvent(reaction, none, botUser); !ok || e != reaction {
		t.Error("other events changed")
	}
}

// Like Discord, the bot's own messages and the ones that mention it keep
// their text.
func TestFilterEventKeepsOwnAndMentions(t *testing.T) {
	none := &appIntents{}

	own := textMessage()
	own.Author.ID = botUser
	own.ReferencedMessage = nil
	if e, _ := filterEvent(&gateway.MessageCreateEvent{Message: own}, none, botUser); !hasText(e.(*gateway.MessageCreateEvent).Message) {
		t.Error("the bot's own message lost its text")
	}

	mention := textMessage()
	mention.Mentions = []discord.GuildUser{{User: discord.User{ID: botUser}}}
	e, _ := filterEvent(&gateway.MessageCreateEvent{Message: mention}, none, botUser)
	m := e.(*gateway.MessageCreateEvent).Message
	if m.Content == "" {
		t.Error("a message mentioning the bot lost its text")
	}
	if m.ReferencedMessage.Content != "" {
		t.Error("the message it replies to kept its text")
	}
}

func TestScopedMessagesWithoutContent(t *testing.T) {
	cabinet := defaultstore.New()
	cabinet.ChannelSet(&discord.Channel{ID: ownChannel, GuildID: ownGuild}, false)
	m := textMessage()
	cabinet.MessageSet(&m, false)

	intents := &appIntents{}
	scoped := scopedCabinet(cabinet, ownGuild, intents, botUser)
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

	one := fakeResponse{http.StatusOK, `{"id":"1","author":{"id":"5"},"content":"secret","embeds":[{"title":"secret"}],"unknown_field":1}`}
	resp, err := filterResponse("GET", "/api/v10/channels/10/messages/1", one, none, botUser)
	body := read(resp)
	if err != nil || strings.Contains(body, "secret") {
		t.Errorf("message without the intent kept its text: %v", err)
	}
	if !strings.Contains(body, "unknown_field") {
		t.Error("fields the filter doesn't know were dropped")
	}

	list := fakeResponse{http.StatusOK, `[{"id":"1","author":{"id":"5"},"content":"secret"},{"id":"2","author":{"id":"5"},"content":"secret"}]`}
	resp, err = filterResponse("GET", "/api/v10/channels/10/messages", list, none, botUser)
	if err != nil || strings.Contains(read(resp), "secret") {
		t.Errorf("messages without the intent kept their text: %v", err)
	}

	own := fakeResponse{http.StatusOK, `{"id":"1","author":{"id":"` + botUser.String() + `"},"content":"secret"}`}
	if resp, _ := filterResponse("GET", "/api/v10/channels/10/messages/1", own, none, botUser); !strings.Contains(read(resp), "secret") {
		t.Error("the bot's own message lost its text")
	}

	with := &appIntents{}
	with.messageContent.Store(true)
	if resp, _ := filterResponse("GET", "/api/v10/channels/10/messages/1", one, with, botUser); !strings.Contains(read(resp), "secret") {
		t.Error("message with the intent lost its text")
	}

	reactions := fakeResponse{http.StatusOK, `[{"id":"5","username":"secret"}]`}
	if resp, err := filterResponse("GET", "/api/v10/channels/10/messages/1/reactions/%F0%9F%91%8D", reactions, none, botUser); err != nil || !strings.Contains(read(resp), "secret") {
		t.Errorf("other routes under messages changed: %v", err)
	}
	other := fakeResponse{http.StatusOK, `{"name":"secret"}`}
	if resp, _ := filterResponse("GET", "/api/v10/channels/10", other, none, botUser); !strings.Contains(read(resp), "secret") {
		t.Error("other responses changed")
	}
}

// Pins and searches return messages filterResponse doesn't filter, so apps
// without the intent can't get them.
func TestScopeRefusesOtherMessageReads(t *testing.T) {
	s := testScope(t)
	for _, path := range []string{
		"/channels/" + ownChannel.String() + "/pins",
		"/channels/" + ownChannel.String() + "/messages/pins",
		"/guilds/" + ownGuild.String() + "/messages/search",
	} {
		if err := s.check(context.Background(), http.MethodGet, path); err == nil {
			t.Errorf("GET %s was allowed without the intent", path)
		}
	}
	s.intents.messageContent.Store(true)
	if err := s.check(context.Background(), http.MethodGet, "/channels/"+ownChannel.String()+"/pins"); err != nil {
		t.Errorf("pins with the intent: %v", err)
	}
}
