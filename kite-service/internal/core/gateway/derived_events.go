package gateway

import (
	"sync"
	"time"

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
	edits := newEditTimes()

	pre := handler.New()
	pre.AddSyncHandler(func(e gateway.Event) {
		derived := deriveEvents(session.Cabinet, edits, e)
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
func deriveEvents(c cache, edits *editTimes, e gateway.Event) []guildevent.Event {
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
	case *gateway.MessageUpdateEvent:
		// Message Update listeners only run for servers.
		if !e.GuildID.IsValid() || !isEdit(c, edits, e) {
			return nil
		}
		return []guildevent.Event{&guildevent.MessageEdit{MessageUpdateEvent: *e}}
	}
	return nil
}

// isEdit reports whether a MESSAGE_UPDATE is an edit, by the time of the last
// edit changing. Link previews and pins keep it.
func isEdit(c cache, edits *editTimes, e *gateway.MessageUpdateEvent) bool {
	if !e.EditedTimestamp.IsValid() {
		return false
	}
	edited := e.EditedTimestamp.Time()
	if old, err := c.Message(e.ChannelID, e.ID); err == nil {
		return !old.EditedTimestamp.Time().Equal(edited)
	}
	return edits.changed(e.ID, edited)
}

// maxEditTimes bounds editTimes, which runs for every app. It mainly needs to
// cover the seconds between an edit and its link preview. Emptying it at worst
// counts an update of a message that isn't cached as an edit once more.
const maxEditTimes = 100

// editTimes has the time of the last edit of messages the cache doesn't have,
// as it only keeps the latest messages of a channel and doesn't add older
// ones when they're updated.
type editTimes struct {
	mu    sync.Mutex
	times map[discord.MessageID]time.Time
}

func newEditTimes() *editTimes {
	return &editTimes{times: make(map[discord.MessageID]time.Time)}
}

// changed records the time of the last edit of a message, and reports whether
// it's different from the one recorded before. A message without one counts as
// edited, as it may have been edited before Kite saw it.
func (t *editTimes) changed(id discord.MessageID, edited time.Time) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if last, ok := t.times[id]; ok && last.Equal(edited) {
		return false
	}
	if len(t.times) >= maxEditTimes {
		// A new map frees the memory, clear() would keep it.
		t.times = make(map[discord.MessageID]time.Time)
	}
	t.times[id] = edited
	return true
}
