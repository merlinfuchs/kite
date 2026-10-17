package leveling

import (
	"testing"
)

func TestLevelForXP(t *testing.T) {
	tests := []struct {
		xp   int64
		want int
	}{
		{xp: 0, want: 0},
		{xp: 99, want: 0},
		{xp: 100, want: 1},
		{xp: 254, want: 1},
		{xp: 255, want: 2},
		{xp: -10, want: 0},
	}

	for _, tt := range tests {
		if got := levelForXP(tt.xp); got != tt.want {
			t.Errorf("levelForXP(%d) = %d, want %d", tt.xp, got, tt.want)
		}
	}
}

func TestLevelForXPMatchesXPForLevel(t *testing.T) {
	for level := 0; level < 100; level++ {
		xp := xpForLevel(level)
		if got := levelForXP(xp); got != level {
			t.Errorf("levelForXP(xpForLevel(%d)) = %d", level, got)
		}
		if level > 0 {
			if got := levelForXP(xp - 1); got != level-1 {
				t.Errorf("levelForXP(xpForLevel(%d) - 1) = %d, want %d", level, got, level-1)
			}
		}
	}
}

func TestLevelForXPIsCapped(t *testing.T) {
	if got := levelForXP(1 << 62); got != maxLevel {
		t.Errorf("levelForXP of huge XP = %d, want %d", got, maxLevel)
	}
}

func TestProgressBar(t *testing.T) {
	tests := []struct {
		current, needed int64
		want            string
	}{
		{current: 0, needed: 100, want: "▱▱▱▱"},
		{current: 50, needed: 100, want: "▰▰▱▱"},
		{current: 100, needed: 100, want: "▰▰▰▰"},
		{current: 500, needed: 100, want: "▰▰▰▰"},
		{current: -5, needed: 100, want: "▱▱▱▱"},
		{current: 0, needed: 0, want: "▰▰▰▰"},
	}

	for _, tt := range tests {
		if got := progressBar(tt.current, tt.needed, 4); got != tt.want {
			t.Errorf("progressBar(%d, %d) = %q, want %q", tt.current, tt.needed, got, tt.want)
		}
	}
}

func TestUpdateLeaderboard(t *testing.T) {
	var entries []leaderboardEntry
	for i := 0; i < leaderboardSize; i++ {
		entries = updateLeaderboard(entries, string(rune('a'+i)), int64(100+i))
	}

	if len(entries) != leaderboardSize {
		t.Fatalf("got %d entries, want %d", len(entries), leaderboardSize)
	}
	if entries[0].UserID != "j" || entries[len(entries)-1].UserID != "a" {
		t.Fatalf("leaderboard not sorted by XP: %v", entries)
	}

	// A member with more XP pushes out the last one.
	entries = updateLeaderboard(entries, "z", 1000)
	if len(entries) != leaderboardSize || entries[0].UserID != "z" {
		t.Fatalf("new top member missing: %v", entries)
	}
	for _, e := range entries {
		if e.UserID == "a" {
			t.Fatalf("last member wasn't pushed out: %v", entries)
		}
	}

	// Updating a member doesn't add them twice.
	entries = updateLeaderboard(entries, "z", 2000)
	count := 0
	for _, e := range entries {
		if e.UserID == "z" {
			count++
		}
	}
	if count != 1 || entries[0].XP != 2000 {
		t.Fatalf("member duplicated or not updated: %v", entries)
	}

	// Zero XP removes a member.
	entries = updateLeaderboard(entries, "z", 0)
	for _, e := range entries {
		if e.UserID == "z" {
			t.Fatalf("reset member still on leaderboard: %v", entries)
		}
	}
}
