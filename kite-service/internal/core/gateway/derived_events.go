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
// event. AutoMod actions don't need the cache, but the name of their rule is
// fetched from Discord, so they're worked out when they're dispatched.
func trackDerivedEvents(session *state.State, dispatch func(gateway.Event)) {
	var mu sync.Mutex
	pending := make(map[gateway.Event][]guildevent.Event)
	edits := newEditTimes()
	rules := newRuleNames(session.GetAutoModerationRule)

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
		if e, ok := e.(*gateway.AutoModerationActionExecutionEvent); ok {
			action := *e
			go func() {
				dispatch(&guildevent.AutoModAction{
					AutoModerationActionExecutionEvent: action,
					RuleName:                           rules.name(action.GuildID, action.RuleID),
				})
			}()
			return
		}

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

// ruleNameTTL is how long the name of an AutoMod rule is kept. Rules can be
// renamed, and there's no event for it without another intent.
const ruleNameTTL = 10 * time.Minute

// maxRuleNames bounds ruleNames. Servers have at most a few dozen rules.
const maxRuleNames = 100

// ruleNames keeps the names of AutoMod rules, so a raid that makes AutoMod act
// on every message doesn't fetch the rule for each of them.
type ruleNames struct {
	mu    sync.Mutex
	names map[discord.AutoModerationRuleID]ruleName
	fetch func(discord.GuildID, discord.AutoModerationRuleID) (discord.AutoModerationRule, error)
	now   func() time.Time
}

type ruleName struct {
	name      string
	fetchedAt time.Time
}

func newRuleNames(fetch func(discord.GuildID, discord.AutoModerationRuleID) (discord.AutoModerationRule, error)) *ruleNames {
	return &ruleNames{
		names: make(map[discord.AutoModerationRuleID]ruleName),
		fetch: fetch,
		now:   time.Now,
	}
}

// name returns the name of a rule, empty if it can't be fetched. A failed
// fetch is kept like a name, so a missing permission doesn't make every action
// fetch again. The lock is held while fetching, so actions that come at once
// fetch the rule once.
func (r *ruleNames) name(guildID discord.GuildID, ruleID discord.AutoModerationRuleID) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if n, ok := r.names[ruleID]; ok && r.now().Sub(n.fetchedAt) < ruleNameTTL {
		return n.name
	}

	var name string
	if rule, err := r.fetch(guildID, ruleID); err == nil {
		name = rule.Name
	}

	if len(r.names) >= maxRuleNames {
		// A new map frees the memory, clear() would keep it.
		r.names = make(map[discord.AutoModerationRuleID]ruleName)
	}
	r.names[ruleID] = ruleName{name: name, fetchedAt: r.now()}
	return name
}
