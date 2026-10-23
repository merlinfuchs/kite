package gateway

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
)

const (
	eventTypeForumPostCreate ws.EventType = "FORUM_POST_CREATE"
	eventTypeForumPostUpdate ws.EventType = "FORUM_POST_UPDATE"
	eventTypeForumPostDelete ws.EventType = "FORUM_POST_DELETE"
)

type forumPostGatewayEvent struct {
	event  gateway.Event
	post   discord.Channel
	typeID ws.EventType
}

func (e *forumPostGatewayEvent) EventType() ws.EventType {
	return e.typeID
}

func (e *forumPostGatewayEvent) Op() ws.OpCode {
	return e.event.Op()
}

func (e *forumPostGatewayEvent) OriginalEvent() gateway.Event {
	return e.event
}

func (e *forumPostGatewayEvent) ForumPost() discord.Channel {
	return e.post
}

func (e *forumPostGatewayEvent) IsForumPostEvent() bool {
	return true
}

func forumPostEvent(session *state.State, event gateway.Event) (gateway.Event, bool) {
	var post discord.Channel
	var eventType ws.EventType

	switch e := event.(type) {
	case *gateway.ThreadCreateEvent:
		post = e.Channel
		eventType = eventTypeForumPostCreate
	case *gateway.ThreadUpdateEvent:
		post = e.Channel
		eventType = eventTypeForumPostUpdate
	case *gateway.ThreadDeleteEvent:
		post = discord.Channel{
			ID:       e.ID,
			GuildID:  e.GuildID,
			Type:     e.Type,
			ParentID: e.ParentID,
		}
		eventType = eventTypeForumPostDelete
	default:
		return nil, false
	}

	if post.Type != discord.GuildPublicThread || !isForumChannel(session, post.ParentID) {
		return nil, false
	}

	return &forumPostGatewayEvent{
		event:  event,
		post:   post,
		typeID: eventType,
	}, true
}

func isForumChannel(session *state.State, channelID discord.ChannelID) bool {
	if session == nil || session.Cabinet == nil || channelID == 0 {
		return false
	}

	channel, err := session.Cabinet.Channel(channelID)
	return err == nil && channel.Type == discord.GuildForum
}
