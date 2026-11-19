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
// it doesn't get it at all.
func filterEvent(e gateway.Event, intents *appIntents) (gateway.Event, bool) {
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
		if hasContent(e.Message) {
			copied := *e
			copied.Message = withoutContent(e.Message)
			return &copied, true
		}
	case *gateway.MessageUpdateEvent:
		if hasContent(e.Message) {
			copied := *e
			copied.Message = withoutContent(e.Message)
			return &copied, true
		}
	case *guildevent.MessageEdit:
		if hasContent(e.Message) {
			copied := *e
			copied.Message = withoutContent(e.Message)
			return &copied, true
		}
	case *guildevent.MessageDelete:
		if e.Message != nil && hasContent(*e.Message) {
			copied := *e
			stripped := withoutContent(*e.Message)
			copied.Message = &stripped
			return &copied, true
		}
	}
	return e, true
}

// hasContent is whether the message has anything withoutContent removes.
// Without the intent Discord already leaves it out, so most messages have
// none.
func hasContent(m discord.Message) bool {
	return m.Content != "" || len(m.Embeds) > 0 || len(m.Attachments) > 0 ||
		len(m.Components) > 0 || len(m.MessageSnapshots) > 0 ||
		(m.ReferencedMessage != nil && hasContent(*m.ReferencedMessage))
}

// withoutContent is the message without what Discord leaves out for bots
// without the Message Content intent, also of the message it replies to or
// forwards.
func withoutContent(m discord.Message) discord.Message {
	m.Content = ""
	m.Embeds = nil
	m.Attachments = nil
	m.Components = nil
	m.MessageSnapshots = nil
	if m.ReferencedMessage != nil {
		referenced := withoutContent(*m.ReferencedMessage)
		m.ReferencedMessage = &referenced
	}
	return m
}

// filterResponse takes the text out of the messages Discord sent for a GET
// of one or several messages, like Get Message, for apps whose templates
// don't read it.
func filterResponse(method, path string, resp httpdriver.Response, intents *appIntents) (httpdriver.Response, error) {
	if method != http.MethodGet || resp.GetStatus() != http.StatusOK || intents.messageContent.Load() {
		return resp, nil
	}
	segments := util.DiscordAPISegments(path)
	if len(segments) < 3 || segments[0] != "channels" || segments[2] != "messages" {
		return resp, nil
	}

	body := resp.GetBody()
	defer body.Close()
	raw, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	var stripped any
	if len(segments) == 3 {
		var messages []discord.Message
		if err := json.Unmarshal(raw, &messages); err != nil {
			return nil, err
		}
		for i := range messages {
			messages[i] = withoutContent(messages[i])
		}
		stripped = messages
	} else {
		var m discord.Message
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, err
		}
		stripped = withoutContent(m)
	}
	raw, err = json.Marshal(stripped)
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
