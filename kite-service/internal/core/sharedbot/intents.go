package sharedbot

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"sync/atomic"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/httputil/httpdriver"
	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
)

// appIntents are the privileged data the templates of an app read, see
// bundle.Bundle.SharedBotIntents. Intents are global, so the Kite bot gets
// message text and member events for every server once Discord grants them.
// Apps without a template that reads them get messages without their text,
// in events, the cache and REST responses, and no member events, like a bot
// without the intents. Scopes share them, so loading the apps again updates
// them in place.
type appIntents struct {
	messageContent atomic.Bool
	guildMembers   atomic.Bool
}

// set takes the intents of the official bundles the app installed.
func (a *appIntents) set(bundleIDs []string) {
	var content, members bool
	for _, id := range bundleIDs {
		intents := bundleIntents[id]
		content = content || slices.Contains(intents, bundle.SharedBotIntentMessageContent)
		members = members || slices.Contains(intents, bundle.SharedBotIntentGuildMembers)
	}
	a.messageContent.Store(content)
	a.guildMembers.Store(members)
}

// bundleIntents are the intents of the official bundles by ID.
var bundleIntents = func() map[string][]string {
	res := make(map[string][]string)
	for _, b := range bundle.Official() {
		res[b.ID] = b.SharedBotIntents
	}
	return res
}()

// filterEvent is the event as an app with the intents gets it, or false if
// it doesn't get it at all. bot is the Kite bot's user.
func filterEvent(e gateway.Event, intents *appIntents, bot discord.UserID) (gateway.Event, bool) {
	switch e.(type) {
	case *gateway.GuildMemberAddEvent, *gateway.GuildMemberRemoveEvent,
		*gateway.GuildMemberUpdateEvent, *gateway.GuildMembersChunkEvent,
		*guildevent.BoostRemove:
		return e, intents.guildMembers.Load()
	}
	if intents.messageContent.Load() {
		return e, true
	}

	switch e := e.(type) {
	case *gateway.MessageCreateEvent:
		if strips(e.Message, bot) {
			copied := *e
			copied.Message = withoutContent(e.Message, bot)
			return &copied, true
		}
	case *gateway.MessageUpdateEvent:
		if strips(e.Message, bot) {
			copied := *e
			copied.Message = withoutContent(e.Message, bot)
			return &copied, true
		}
	case *guildevent.MessageEdit:
		if strips(e.Message, bot) {
			copied := *e
			copied.Message = withoutContent(e.Message, bot)
			return &copied, true
		}
	case *guildevent.MessageDelete:
		if e.Message != nil && strips(*e.Message, bot) {
			copied := *e
			stripped := withoutContent(*e.Message, bot)
			copied.Message = &stripped
			return &copied, true
		}
	}
	return e, true
}

// keepsContent is whether Discord sends the text of the message also to bots
// without the Message Content intent: their own messages and the ones that
// mention them.
func keepsContent(author discord.UserID, mentions []discord.UserID, bot discord.UserID) bool {
	return author == bot || slices.Contains(mentions, bot)
}

func mentionIDs(m discord.Message) []discord.UserID {
	ids := make([]discord.UserID, len(m.Mentions))
	for i, u := range m.Mentions {
		ids[i] = u.ID
	}
	return ids
}

// strips is whether withoutContent changes the message. Without the intent
// Discord already leaves the text out, so most messages have none.
func strips(m discord.Message, bot discord.UserID) bool {
	if m.ReferencedMessage != nil && strips(*m.ReferencedMessage, bot) {
		return true
	}
	if keepsContent(m.Author.ID, mentionIDs(m), bot) {
		return false
	}
	return m.Content != "" || len(m.Embeds) > 0 || len(m.Attachments) > 0 ||
		len(m.Components) > 0 || len(m.MessageSnapshots) > 0
}

// withoutContent is the message without what Discord leaves out for bots
// without the Message Content intent, also of the message it replies to.
func withoutContent(m discord.Message, bot discord.UserID) discord.Message {
	if m.ReferencedMessage != nil {
		referenced := withoutContent(*m.ReferencedMessage, bot)
		m.ReferencedMessage = &referenced
	}
	if keepsContent(m.Author.ID, mentionIDs(m), bot) {
		return m
	}
	m.Content = ""
	m.Embeds = nil
	m.Attachments = nil
	m.Components = nil
	m.MessageSnapshots = nil
	return m
}

// contentKeys are the fields of a message Discord leaves out for bots without
// the Message Content intent.
var contentKeys = []string{"content", "embeds", "attachments", "components", "message_snapshots", "poll"}

// withoutContentJSON is withoutContent for a message as Discord sent it, so
// fields arikawa doesn't know stay as they are.
func withoutContentJSON(raw json.RawMessage, bot discord.UserID) (json.RawMessage, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if referenced, ok := m["referenced_message"]; ok && string(referenced) != "null" {
		stripped, err := withoutContentJSON(referenced, bot)
		if err != nil {
			return nil, err
		}
		m["referenced_message"] = stripped
	}

	var who struct {
		Author   discord.User   `json:"author"`
		Mentions []discord.User `json:"mentions"`
	}
	if err := json.Unmarshal(raw, &who); err != nil {
		return nil, err
	}
	mentions := make([]discord.UserID, len(who.Mentions))
	for i, u := range who.Mentions {
		mentions[i] = u.ID
	}
	if !keepsContent(who.Author.ID, mentions, bot) {
		for _, key := range contentKeys {
			delete(m, key)
		}
	}
	return json.Marshal(m)
}

// readsMessages is whether the request gets messages in a form
// filterResponse doesn't take the text out of, like pins or a search. Apps
// whose templates don't read message text can't send it.
func readsMessages(method string, segments []string) bool {
	if method != http.MethodGet || len(segments) < 3 {
		return false
	}
	switch {
	case segments[0] == "channels" && segments[2] == "pins":
		return true
	case segments[0] == "channels" && segments[2] == "messages" && len(segments) > 3 && segments[3] == "pins":
		return true
	case segments[0] == "guilds" && segments[2] == "messages":
		return true
	}
	return false
}

// filterResponse takes the text out of the messages Discord sent for a GET
// of one message or a channel's messages, like Get Message, for apps whose
// templates don't read it.
func filterResponse(method, path string, resp httpdriver.Response, intents *appIntents, bot discord.UserID) (httpdriver.Response, error) {
	if method != http.MethodGet || resp.GetStatus() != http.StatusOK || intents.messageContent.Load() {
		return resp, nil
	}
	segments := util.DiscordAPISegments(path)
	if len(segments) < 3 || segments[0] != "channels" || segments[2] != "messages" {
		return resp, nil
	}
	one := len(segments) == 4
	if one {
		if _, err := discord.ParseSnowflake(segments[3]); err != nil {
			return resp, nil
		}
	} else if len(segments) != 3 {
		return resp, nil
	}

	body := resp.GetBody()
	defer body.Close()
	raw, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	if one {
		raw, err = withoutContentJSON(raw, bot)
	} else {
		var messages []json.RawMessage
		if err = json.Unmarshal(raw, &messages); err == nil {
			for i := range messages {
				if messages[i], err = withoutContentJSON(messages[i], bot); err != nil {
					break
				}
			}
			if err == nil {
				raw, err = json.Marshal(messages)
			}
		}
	}
	if err != nil {
		return nil, err
	}
	return filteredResponse{Response: resp, body: raw}, nil
}

type filteredResponse struct {
	httpdriver.Response
	body []byte
}

func (r filteredResponse) GetBody() io.ReadCloser { return io.NopCloser(bytes.NewReader(r.body)) }
