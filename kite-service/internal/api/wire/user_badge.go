package wire

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

// UserBadge is a badge on a user's creator page and in the dashboard.
type UserBadge struct {
	// Kind is staff, contributor, early_supporter, subscriber,
	// early_creator or creator.
	Kind string `json:"kind"`
	// Plan is the title of the plan of a subscriber.
	Plan string `json:"plan,omitempty"`
	// Milestone is how many other users at least added the templates of a
	// creator, missing if fewer than the lowest milestone.
	Milestone int `json:"milestone,omitempty"`
}

func UserBadgesToWire(badges []model.UserBadge) []UserBadge {
	res := make([]UserBadge, len(badges))
	for i, b := range badges {
		res[i] = UserBadge{Kind: string(b.Kind), Plan: b.Plan, Milestone: b.Milestone}
	}
	return res
}

// UserBadgeGrant is a badge an admin gave a user, for the admins.
type UserBadgeGrant struct {
	UserID    string    `json:"user_id"`
	Kind      string    `json:"kind"`
	User      *User     `json:"user"`
	CreatedAt time.Time `json:"created_at"`
}

type UserBadgeGrantListResponse = []*UserBadgeGrant

// UserBadgeGrantCreateRequest gives the user with the Discord ID a badge.
type UserBadgeGrantCreateRequest struct {
	DiscordID string `json:"discord_id"`
	Kind      string `json:"kind"`
}

func (req UserBadgeGrantCreateRequest) Validate() error {
	kinds := make([]any, len(model.GrantableUserBadges))
	for i, k := range model.GrantableUserBadges {
		kinds[i] = string(k)
	}
	return validation.ValidateStruct(&req,
		validation.Field(&req.DiscordID, validation.Required),
		validation.Field(&req.Kind, validation.Required, validation.In(kinds...)),
	)
}

type UserBadgeGrantCreateResponse = UserBadgeGrant

type UserBadgeGrantDeleteResponse struct{}

func UserBadgeGrantToWire(grant *model.UserBadgeGrant) *UserBadgeGrant {
	return &UserBadgeGrant{
		UserID:    grant.UserID,
		Kind:      string(grant.Kind),
		User:      UserToWire(grant.User, false),
		CreatedAt: grant.CreatedAt,
	}
}
