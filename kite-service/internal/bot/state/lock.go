package state

import "github.com/merlinfuchs/dismod/distype"

type memberLockKey struct {
	guildID distype.Snowflake
	userID  distype.Snowflake
}
