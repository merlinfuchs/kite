package gateway

import (
	"sync"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/handler"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
)

// trackDerivedEvents makes session dispatch the events of the guildevent
// package. They read the cache from before Discord's event updated it, so
// they're worked out in the PreHandler. They're dispatched once the cache is
// updated, so their flows see the same cache as the flows of the Discord
// event.
func trackDerivedEvents(session *state.State, dispatch func(gateway.Event)) {
	var mu sync.Mutex
	pending := make(map[gateway.Event][]guildevent.Event)

	pre := handler.New()
	pre.AddSyncHandler(func(e gateway.Event) {
		derived := deriveEvents(session.Cabinet, e)
		if len(derived) == 0 {
			return
		}
		mu.Lock()
		pending[e] = derived
		mu.Unlock()
	})
	session.PreHandler = pre

	session.AddSyncHandler(func(e gateway.Event) {
		mu.Lock()
		derived, ok := pending[e]
		delete(pending, e)
		mu.Unlock()
		if !ok {
			return
		}

		// Like the handler of Discord's events, so a slow event listener
		// can't hold up the connection.
		go func() {
			for _, d := range derived {
				dispatch(d)
			}
		}()
	})
}

// cache is the part of the state cache deriving events reads.
type cache interface {
	Message(discord.ChannelID, discord.MessageID) (*discord.Message, error)
}

// deriveEvents returns the events Kite derives from a Discord event. c is the
// cache from before the event updated it.
func deriveEvents(c cache, e gateway.Event) []guildevent.Event {
	switch e := e.(type) {
	case *gateway.MessageDeleteEvent:
		// Message Delete listeners only run for servers.
		if !e.GuildID.IsValid() {
			return nil
		}
		// The cache removes the message right after, and Discord only sends
		// its IDs. Bulk deletes don't run Message Delete listeners, so they
		// aren't looked up.
		del := &guildevent.MessageDelete{MessageDeleteEvent: *e}
		if msg, err := c.Message(e.ChannelID, e.ID); err == nil {
			del.Message = msg
		}
		return []guildevent.Event{del}
	}
	return nil
}
