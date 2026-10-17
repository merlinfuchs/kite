package modlogs

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
)

func TestMessageCacheEvictsOldest(t *testing.T) {
	cache := newMessageCache(2)

	cache.Set(1, cachedMessage{Content: "one"})
	cache.Set(2, cachedMessage{Content: "two"})
	cache.Set(3, cachedMessage{Content: "three"})

	if _, ok := cache.Get(1); ok {
		t.Error("oldest message wasn't evicted")
	}
	for _, id := range []discord.MessageID{2, 3} {
		if _, ok := cache.Get(id); !ok {
			t.Errorf("message %d missing", id)
		}
	}
}

func TestMessageCacheUpdateKeepsPosition(t *testing.T) {
	cache := newMessageCache(2)

	cache.Set(1, cachedMessage{Content: "one"})
	cache.Set(2, cachedMessage{Content: "two"})
	cache.Set(1, cachedMessage{Content: "one edited"})

	if msg, _ := cache.Get(1); msg.Content != "one edited" {
		t.Errorf("content = %q, want the edit", msg.Content)
	}

	// Editing doesn't refresh the message, so it's still evicted first.
	cache.Set(3, cachedMessage{Content: "three"})
	if _, ok := cache.Get(1); ok {
		t.Error("edited message should still be the oldest")
	}
	if _, ok := cache.Get(2); !ok {
		t.Error("message 2 was evicted instead of 1")
	}
}

func TestMessageCacheDelete(t *testing.T) {
	cache := newMessageCache(2)
	cache.Set(1, cachedMessage{Content: "one"})

	msg, ok := cache.Delete(1)
	if !ok || msg.Content != "one" {
		t.Fatalf("Delete = %v, %v", msg, ok)
	}
	if _, ok := cache.Delete(1); ok {
		t.Error("deleted message is still cached")
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("truncate short = %q", got)
	}
	if got := truncate("héllo wörld", 5); got != "héll…" {
		t.Errorf("truncate long = %q", got)
	}
}
