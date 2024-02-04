package dismodel

import (
	"time"

	"github.com/merlinfuchs/kite/kite-types/null"
)

// Channel represents a Discord channel.
// https://discord.com/developers/docs/resources/channel#channel-object-channel-structure
type Channel struct {
	ID                            string                       `json:"id"`
	Type                          ChannelType                  `json:"type"`
	GuildID                       string                       `json:"guild_id,omitempty"`
	Position                      int                          `json:"position,omitempty"`
	PermissionOverwrites          []ChannelPermissionOverwrite `json:"permission_overwrites,omitempty"`
	Name                          string                       `json:"name,omitempty"`
	Topic                         string                       `json:"topic,omitempty"`
	NSFW                          bool                         `json:"nsfw,omitempty"`
	LastMessageID                 string                       `json:"last_message_id,omitempty"`
	Bitrate                       int                          `json:"bitrate,omitempty"`
	UserLimit                     int                          `json:"user_limit,omitempty"`
	RateLimitPerUser              int                          `json:"rate_limit_per_user,omitempty"`
	ParentID                      string                       `json:"parent_id,omitempty"`
	LastPinTimestamp              time.Time                    `json:"last_pin_timestamp,omitempty"`
	RTCRegion                     string                       `json:"rtc_region,omitempty"`
	VideoQualityMode              ChannelVideoQualityMode      `json:"video_quality_mode,omitempty"`
	ThreadMetadata                *ChannelThreadMetadata       `json:"thread_metadata,omitempty"`
	Member                        *ChannelThreadMember         `json:"member,omitempty"`
	DefaultAutoArchiveDuration    int                          `json:"default_auto_archive_duration,omitempty"`
	Permissions                   string                       `json:"permissions,omitempty"`
	Flags                         int                          `json:"flags,omitempty"`
	AvailableTags                 []ChannelTag                 `json:"available_tags,omitempty"`
	AppliedTags                   []string                     `json:"applied_tags,omitempty"`
	DefaultReactionEmoji          *ChannelDefaultReactionEmoji `json:"default_reaction_emoji,omitempty"`
	DefaultThreadRateLimitPerUser int                          `json:"default_thread_rate_limit_per_user,omitempty"`
	DefaultSortOrder              ChannelDefaultSortOrder      `json:"default_sort_order,omitempty"`
	DefaultForumLayout            ChannelDefaultForumLayout    `json:"default_forum_layout,omitempty"`
}

type ChannelType int

const (
	ChannelTypeGuildText          ChannelType = 0
	ChannelTypeDM                 ChannelType = 1
	ChannelTypeGuildVoice         ChannelType = 2
	ChannelTypeGroupDM            ChannelType = 3
	ChannelTypeGuildCategory      ChannelType = 4
	ChannelTypeGuildAnnouncement  ChannelType = 5
	ChannelTypeAnnouncementThread ChannelType = 10
	ChannelTypePublicThread       ChannelType = 11
	ChannelTypePrivateThread      ChannelType = 12
	ChannelTypeStageVoice         ChannelType = 13
	ChannelTypeGuildDirectory     ChannelType = 14
	ChannelTypeGuildForum         ChannelType = 15
	ChannelTypeMedia              ChannelType = 16
)

type ChannelPermissionOverwrite struct {
	ID    string                         `json:"id"`
	Type  ChannelpermissionoverwriteType `json:"type"`
	Allow string                         `json:"allow"`
	Deny  string                         `json:"deny"`
}

type ChannelpermissionoverwriteType int

const (
	ChannelpermissionoverwriteTypeRole   ChannelpermissionoverwriteType = 0
	ChannelpermissionoverwriteTypeMember ChannelpermissionoverwriteType = 1
)

type ChannelVideoQualityMode int

const (
	ChannelVideoQualityModeAuto ChannelVideoQualityMode = 1
	ChannelVideoQualityModeFull ChannelVideoQualityMode = 2
)

type ChannelThreadMetadata struct {
	Archived            bool      `json:"archived"`
	AutoArchiveDuration int       `json:"auto_archive_duration"`
	ArchiveTimestamp    time.Time `json:"archive_timestamp"`
	Locked              bool      `json:"locked"`
	Invitable           bool      `json:"invitable,omitempty"`
	CreateTimestamp     time.Time `json:"create_timestamp,omitempty"`
}

type ChannelThreadMember struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	JoinTimestamp time.Time `json:"join_timestamp"`
	Flags         int       `json:"flags"`
	Member        *Member   `json:"member,omitempty"`
}

type ChannelTag struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Moderated bool   `json:"moderated"`
	EmojiID   string `json:"emoji_id,omitempty"`
	EmojiName string `json:"emoji_name,omitempty"`
}

type ChannelDefaultReactionEmoji struct {
	EmojiID   string `json:"emoji_id,omitempty"`
	EmojiName string `json:"emoji_name,omitempty"`
}

type ChannelDefaultSortOrder int

const (
	ChannelDefaultSortOrderLatestActivity ChannelDefaultSortOrder = 0
	ChannelDefaultSortOrderCreationDate   ChannelDefaultSortOrder = 1
)

type ChannelDefaultForumLayout int

const (
	ChannelDefaultForumLayoutNotSet      ChannelDefaultForumLayout = 0
	ChannelDefaultForumLayoutListView    ChannelDefaultForumLayout = 1
	ChannelDefaultForumLayoutGalleryView ChannelDefaultForumLayout = 2
)

type ChannelCreateEvent = Channel

type ChannelUpdateEvent = Channel

type ChannelDeleteEvent = Channel

type ChannelPinsUpdateEvent struct {
	GuildID          string    `json:"guild_id,omitempty"`
	ChannelID        string    `json:"channel_id"`
	LastPinTimestamp time.Time `json:"last_pin_timestamp,omitempty"`
}

type ThreadCreateEvent = Channel

type ThreadUpdateEvent = Channel

type ThreadDeleteEvent struct {
	ID       string      `json:"id"`
	GuildID  string      `json:"guild_id"`
	ParentID string      `json:"parent_id"`
	Type     ChannelType `json:"type"`
}

type ThreadListSyncEvent struct {
	GuildID    string                `json:"guild_id"`
	ChannelIDs []string              `json:"channel_ids"`
	Threads    []Channel             `json:"threads"`
	Members    []ChannelThreadMember `json:"members"`
}

type ThreadMemberUpdateEvent struct {
	ChannelThreadMember
	GuildID string `json:"guild_id"`
}

type ThreadMembersUpdateEvent struct {
	ID               string                `json:"id"`
	GuildID          string                `json:"guild_id"`
	MemberCount      int                   `json:"member_count"`
	AddedMembers     []ChannelThreadMember `json:"added_members"`
	RemovedMemberIDs []string              `json:"removed_member_ids"`
}

type ChannelGetCall struct {
	ID string `json:"id"`
}

type ChannelGetResponse = Channel

type ChannelListCall struct{}

type ChannelListResponse []Channel

type ChannelCreateCall struct {
	Name                          string                       `json:"name,omitempty"`
	Type                          ChannelType                  `json:"type,omitempty"`
	Topic                         string                       `json:"topic,omitempty"`
	Bitrate                       int                          `json:"bitrate,omitempty"`
	UserLimit                     int                          `json:"user_limit,omitempty"`
	RateLimitPerUser              int                          `json:"rate_limit_per_user,omitempty"`
	Position                      int                          `json:"position,omitempty"`
	PermissionOverwrites          []ChannelPermissionOverwrite `json:"permission_overwrites,omitempty"`
	ParentID                      string                       `json:"parent_id,omitempty"`
	NSFW                          bool                         `json:"nsfw,omitempty"`
	RTCRegion                     string                       `json:"rtc_region,omitempty"`
	VideoQualityMode              ChannelVideoQualityMode      `json:"video_quality_mode,omitempty"`
	DefaultAutoArchiveDuration    int                          `json:"default_auto_archive_duration,omitempty"`
	DefaultReactionEmoji          *ChannelDefaultReactionEmoji `json:"default_reaction_emoji,omitempty"`
	AvailableTags                 []ChannelTag                 `json:"available_tags,omitempty"`
	DefaultSortOrder              ChannelDefaultSortOrder      `json:"default_sort_order,omitempty"`
	DefaultForumLayout            ChannelDefaultForumLayout    `json:"default_forum_layout,omitempty"`
	DefaultThreadRateLimitPerUser int                          `json:"default_thread_rate_limit_per_user,omitempty"`
}

type ChannelCreateResponse = Channel

type ChannelUpdateCall struct {
	ID                            string                                  `json:"id"`
	Name                          null.Null[string]                       `json:"name,omitempty"`
	Type                          null.Null[ChannelType]                  `json:"type,omitempty"`
	Position                      null.Null[int]                          `json:"position,omitempty"`
	Topic                         null.Null[string]                       `json:"topic,omitempty"`
	NSFW                          null.Null[bool]                         `json:"nsfw,omitempty"`
	RateLimitPerUser              null.Null[int]                          `json:"rate_limit_per_user,omitempty"`
	Bitrate                       null.Null[int]                          `json:"bitrate,omitempty"`
	UserLimit                     null.Null[int]                          `json:"user_limit,omitempty"`
	PermissionOverwrites          null.Null[[]ChannelPermissionOverwrite] `json:"permission_overwrites,omitempty"`
	ParentID                      null.Null[string]                       `json:"parent_id,omitempty"`
	RTCRegion                     null.Null[string]                       `json:"rtc_region,omitempty"`
	VideoQualityMode              null.Null[ChannelVideoQualityMode]      `json:"video_quality_mode,omitempty"`
	DefaultAutoArchiveDuration    null.Null[int]                          `json:"default_auto_archive_duration,omitempty"`
	Flags                         null.Null[int]                          `json:"flags,omitempty"`
	AvailableTags                 null.Null[[]ChannelTag]                 `json:"available_tags,omitempty"`
	DefaultReactionEmoji          null.Null[ChannelDefaultReactionEmoji]  `json:"default_reaction_emoji,omitempty"`
	DefaultThreadRateLimitPerUser null.Null[int]                          `json:"default_thread_rate_limit_per_user,omitempty"`
	DefaultSortOrder              null.Null[ChannelDefaultSortOrder]      `json:"default_sort_order,omitempty"`
	DefaultForumLayout            null.Null[ChannelDefaultForumLayout]    `json:"default_forum_layout,omitempty"`
}

type ChannelUpdateResponse = Channel

type ChannelUpdatePositionsCall = []ChannelUpdatePositionsEntry

type ChannelUpdatePositionsEntry struct {
	ID              string `json:"id"`
	Position        int    `json:"position"`
	LockPermissions bool   `json:"lock_permissions"`
	ParentID        string `json:"parent_id"`
}

type ChannelUpdatePositionsResponse struct{}

type ChannelDeleteCall struct {
	ID string `json:"id"`
}

type ChannelDeleteResponse = Channel

type ChannelUpdatePermissionsCall struct {
	ChannelID   string                         `json:"channel_id"`
	OverwriteID string                         `json:"overwrite_id"`
	Allow       string                         `json:"allow"`
	Deny        string                         `json:"deny"`
	Type        ChannelpermissionoverwriteType `json:"type"`
}

type ChannelUpdatePermissionsResponse struct{}

type ChannelDeletePermissionsCall struct {
	ChannelID   string `json:"channel_id"`
	OverwriteID string `json:"overwrite_id"`
}

type ChannelDeletePermissionsResponse struct{}

type ThreadStartFromMessageCall struct {
	ChannelID           string `json:"channel_id"`
	MessageID           string `json:"message_id"`
	Name                string `json:"name"`
	AutoArchiveDuration int    `json:"auto_archive_duration"`
	RateLimitPerUser    int    `json:"rate_limit_per_user"`
}

type ThreadStartFromMessageResponse = Channel

type ThreadStartCall struct {
	ChannelID           string      `json:"channel_id"`
	Name                string      `json:"name"`
	Type                ChannelType `json:"type"`
	AutoArchiveDuration int         `json:"auto_archive_duration"`
	Invitable           bool        `json:"invitable"`
	RateLimitPerUser    int         `json:"rate_limit_per_user"`
}

type ThreadStartResponse = Channel

type ThreadStartInForumCall struct {
	ChannelID           string   `json:"channel_id"`
	Name                string   `json:"name"`
	AutoArchiveDuration int      `json:"auto_archive_duration"`
	RateLimitPerUser    int      `json:"rate_limit_per_user"`
	Message             Message  `json:"message"`
	AppliedTags         []string `json:"applied_tags"`
}

type ThreadStartInForumResponse = Channel

type ThreadJoinCall struct {
	ChannelID string `json:"channel_id"`
}

type ThreadJoinResponse struct{}

type ThreadMemberAddCall struct {
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
}

type ThreadMemberAddResponse struct{}

type ThreadLeaveCall struct {
	ChannelID string `json:"channel_id"`
}

type ThreadLeaveResponse struct{}

type ThreadMemberRemoveCall struct {
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
}

type ThreadMemberRemoveResponse struct{}

type ThreadMemberGetCall struct {
	ChannelID  string `json:"channel_id"`
	UserID     string `json:"user_id"`
	WithMember bool   `json:"with_member"`
}

type ThreadMemberGetResponse = ChannelThreadMember

type ThreadMemberListCall struct {
	ChannelID  string `json:"channel_id"`
	WithMember bool   `json:"with_member"`
	After      string `json:"after"`
	Limit      int    `json:"limit"`
}

type ThreadMemberListResponse []ChannelThreadMember

type ThreadListPublicArchivedCall struct {
	ChannelID string `json:"channel_id"`
	Before    string `json:"before"`
	Limit     int    `json:"limit"`
}

type ThreadListPublicArchivedResponse struct {
	Threads []Channel             `json:"threads"`
	Members []ChannelThreadMember `json:"members"`
	HasMore bool                  `json:"has_more"`
}

type ThreadListPrivateArchivedCall struct {
	ChannelID string `json:"channel_id"`
	Before    string `json:"before"`
	Limit     int    `json:"limit"`
}

type ThreadListPrivateArchivedResponse struct {
	Threads []Channel             `json:"threads"`
	Members []ChannelThreadMember `json:"members"`
	HasMore bool                  `json:"has_more"`
}

type ThreadListJoinedPrivateArchivedCall struct {
	ChannelID string `json:"channel_id"`
	Before    string `json:"before"`
	Limit     int    `json:"limit"`
}

type ThreadListJoinedPrivateArchivedResponse struct {
	Threads []Channel             `json:"threads"`
	Members []ChannelThreadMember `json:"members"`
	HasMore bool                  `json:"has_more"`
}

type ThreadListActiveCall struct{}

type ThreadListActiveResponse struct {
	Threads []Channel             `json:"threads"`
	Members []ChannelThreadMember `json:"members"`
}
