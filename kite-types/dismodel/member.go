package dismodel

import (
	"time"

	"github.com/merlinfuchs/kite/kite-types/null"
)

type Member struct {
	User                       *User     `json:"user"`
	Nick                       string    `json:"nick,omitempty"`
	Avatar                     string    `json:"avatar,omitempty"`
	Roles                      []string  `json:"roles"`
	JoinedAt                   time.Time `json:"joined_at"`
	PremiumSince               time.Time `json:"premium_since,omitempty"`
	Deaf                       bool      `json:"deaf"`
	Mute                       bool      `json:"mute"`
	Flags                      int       `json:"flags"`
	Pending                    bool      `json:"pending,omitempty"`
	Permissions                string    `json:"permissions,omitempty"`
	CommunicationDisabledUntil time.Time `json:"communication_disabled_until,omitempty"`
}

type GuildMemberAddEvent struct {
	Member
	GuildID string `json:"guild_id"`
}

type GuildMemberRemoveEvent struct {
	User    User   `json:"user"`
	GuildID string `json:"guild_id"`
}

type GuildMemberUpdateEvent struct {
	Member
	GuildID string `json:"guild_id"`
}

type MemberGetCall struct {
	UserID string `json:"user_id"`
}

type MemberGetResponse = Member

type MemberListCall struct {
	Limit int `json:"limit,omitempty"`
	After int `json:"after,omitempty"`
}

type MemberListResponse = []Member

type MemberSearchCall struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

type MemberSearchResponse = []Member

type MemberUpdateCall struct {
	UserID                     string               `json:"user_id"`
	Nick                       null.Null[string]    `json:"nick,omitempty"`
	Roles                      null.Null[[]string]  `json:"roles,omitempty"`
	Mute                       null.Null[bool]      `json:"mute,omitempty"`
	Deaf                       null.Null[bool]      `json:"deaf,omitempty"`
	ChannelID                  null.Null[string]    `json:"channel_id,omitempty"`
	CommunicationDisabledUntil null.Null[time.Time] `json:"communication_disabled_until,omitempty"`
	Flags                      null.Null[int]       `json:"flags,omitempty"`
}

type MemberUpdateResponse = Member

type MemberUpdateOwnCall struct {
	Nick string `json:"nick,omitempty"`
}

type MemberUpdateOwnResponse = Member

type MemberRoleAddCall struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
}

type MemberAddRoleResponse struct{}

type MemberRoleRemoveCall struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
}

type MemberRemoveRoleResponse struct{}

type MemberRemoveCall struct {
	UserID string `json:"user_id"`
}

type MemberRemoveResponse struct{}

type MemberPruneCountCall struct {
	Days         int      `json:"days"`
	IncludeRoles []string `json:"include_roles"`
}

type MemberPruneCountResponse struct {
	Pruned int `json:"pruned"`
}

type MemberPruneBeginCall struct {
	Days              int      `json:"days"`
	ComputePruneCount bool     `json:"compute_prune_count"`
	IncludeRoles      []string `json:"include_roles"`
	Reason            string   `json:"reason,omitempty"`
}

type MemberPruneBeginResponse struct {
	Pruned int `json:"pruned"`
}
