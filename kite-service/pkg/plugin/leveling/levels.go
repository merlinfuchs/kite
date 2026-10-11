package leveling

import (
	"sort"
	"strings"
)

const (
	// maxLevel caps the level so computing it from untrusted XP stays cheap.
	maxLevel = 500

	minMessageXP = 15
	maxMessageXP = 25

	// cooldownSeconds is how long a member has to wait between messages that
	// give XP, so spamming doesn't level anyone up.
	cooldownSeconds = 60

	leaderboardSize = 10
)

// xpForNextLevel is the XP needed to get from level to level+1. It's the same
// curve most leveling bots use, so members find it familiar.
func xpForNextLevel(level int) int64 {
	l := int64(level)
	return 5*l*l + 50*l + 100
}

// xpForLevel is the total XP needed to reach a level from zero.
func xpForLevel(level int) int64 {
	var total int64
	for i := 0; i < level; i++ {
		total += xpForNextLevel(i)
	}
	return total
}

// levelForXP is the level a member with the given total XP has.
func levelForXP(xp int64) int {
	level := 0
	for level < maxLevel {
		next := xpForNextLevel(level)
		if xp < next {
			break
		}
		xp -= next
		level++
	}
	return level
}

// progressBar renders how far a member is to the next level.
func progressBar(current, needed int64, width int) string {
	if needed <= 0 {
		return strings.Repeat("▰", width)
	}

	filled := int(current * int64(width) / needed)
	filled = max(0, min(filled, width))

	return strings.Repeat("▰", filled) + strings.Repeat("▱", width-filled)
}

type leaderboardEntry struct {
	UserID string `json:"user_id"`
	XP     int64  `json:"xp"`
}

// updateLeaderboard returns the leaderboard with the new XP of a member, sorted
// by XP and cut to leaderboardSize. A member with zero XP is removed.
func updateLeaderboard(entries []leaderboardEntry, userID string, xp int64) []leaderboardEntry {
	res := make([]leaderboardEntry, 0, len(entries)+1)
	for _, entry := range entries {
		if entry.UserID != userID {
			res = append(res, entry)
		}
	}

	if xp > 0 {
		res = append(res, leaderboardEntry{UserID: userID, XP: xp})
	}

	sort.SliceStable(res, func(i, j int) bool {
		return res[i].XP > res[j].XP
	})

	if len(res) > leaderboardSize {
		res = res[:leaderboardSize]
	}

	return res
}
