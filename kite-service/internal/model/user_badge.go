package model

import (
	"time"

	"gopkg.in/guregu/null.v4"
)

// UserBadgeKind is a badge users show on their creator page and in the
// dashboard.
type UserBadgeKind string

const (
	// UserBadgeStaff is for the Kite team: the admins in the config, and
	// others like the staff of the Discord server who admins give it by hand.
	UserBadgeStaff UserBadgeKind = "staff"
	// UserBadgeContributor is for people who helped build Kite. Admins give it
	// by hand.
	UserBadgeContributor UserBadgeKind = "contributor"
	// UserBadgeEarlySupporter is for users who paid for a plan before Kite
	// Plus launched, also after they cancelled.
	UserBadgeEarlySupporter UserBadgeKind = "early_supporter"
	// UserBadgeSubscriber is for users who pay for a plan now.
	UserBadgeSubscriber UserBadgeKind = "subscriber"
	// UserBadgeEarlyCreator is for users who listed a template in the
	// marketplace before the badges launched.
	UserBadgeEarlyCreator UserBadgeKind = "early_creator"
	// UserBadgeCreator is for users with a template in the marketplace, with
	// a Milestone once enough other users added them.
	UserBadgeCreator UserBadgeKind = "creator"
)

// GrantableUserBadges are the badges admins give by hand.
var GrantableUserBadges = []UserBadgeKind{UserBadgeStaff, UserBadgeContributor}

type UserBadge struct {
	Kind UserBadgeKind
	// Plan is the title of the plan of a subscriber.
	Plan string
	// Milestone is how many other users at least added the templates of a
	// creator, or 0 if fewer than the lowest milestone.
	Milestone int
}

// UserBadgeGrant is a badge an admin gave a user.
type UserBadgeGrant struct {
	UserID          string
	Kind            UserBadgeKind
	GrantedByUserID null.String
	CreatedAt       time.Time
	// User is set when grants are shown to the admins.
	User *User
}

// CreatorStats is what the badges of a user with templates in the marketplace
// follow from.
type CreatorStats struct {
	UserID string
	// FirstListedAt is when the oldest of their listed templates was made.
	FirstListedAt time.Time
	// InstallerCount is how many other users added any of them.
	InstallerCount int
}
