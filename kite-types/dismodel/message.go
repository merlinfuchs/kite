package dismodel

import "time"

type Message struct {
	ID              string              `json:"id"`
	ChannelID       string              `json:"channel_id"`
	Author          User                `json:"author"`
	Content         string              `json:"content"`
	Timestamp       time.Time           `json:"timestamp"`
	EditedTimestamp time.Time           `json:"edited_timestamp"`
	TTS             bool                `json:"tts"`
	MentionEveryone bool                `json:"mention_everyone"`
	Mentions        []User              `json:"mentions,omitempty"`
	MentionRoles    []Role              `json:"mention_roles,omitempty"`
	MentionChannels []Channel           `json:"mention_channels,omitempty"`
	Attachments     []MessageAttachment `json:"attachments,omitempty"`
	Embeds          []MessageEmbed      `json:"embeds,omitempty"`
	Reactions       []MessageReaction   `json:"reactions,omitempty"`
	Nonce           interface{}         `json:"nonce,omitempty"`
	Pinned          bool                `json:"pinned"`
	WebhookID       string              `json:"webhook_id,omitempty"`
	Type            MessageType         `json:"type,omitempty"`
	// Activity     interface{} `json:"activity,omitempty"`
	Application       *Application        `json:"application,omitempty"`
	ApplicationID     string              `json:"application_id,omitempty"`
	MessageReference  *MessageReference   `json:"message_reference,omitempty"`
	Flags             int                 `json:"flags,omitempty"`
	ReferencedMessage *Message            `json:"referenced_message,omitempty"`
	Interaction       *MessageInteraction `json:"interaction,omitempty"`
	Thread            *Channel            `json:"thread,omitempty"`
	Components        []MessageComponent  `json:"components,omitempty"`
	// StickerItems 	[]StickerItem       `json:"sticker_items,omitempty"`
	// Stickers     	[]Sticker           `json:"stickers,omitempty"`
	Position int `json:"position,omitempty"`
	// RoleSubscriptionData
	Resolved *ResolvedData `json:"resolved,omitempty"`
}

type MessageAttachment struct {
	ID           string  `json:"id"`
	Filename     string  `json:"filename"`
	Description  string  `json:"description,omitempty"`
	ContentType  string  `json:"content_type,omitempty"`
	Size         int     `json:"size"`
	URL          string  `json:"url"`
	ProxyURL     string  `json:"proxy_url"`
	Height       int     `json:"height,omitempty"`
	Width        int     `json:"width,omitempty"`
	Ephemeral    bool    `json:"ephemeral,omitempty"`
	DurationSecs float32 `json:"duration_secs,omitempty"`
	WaveForm     string  `json:"waveform,omitempty"`
	Flags        int     `json:"flags,omitempty"`
}

type MessageEmbed struct{}

type MessageReaction struct {
	Count        int                         `json:"count"`
	CountDetails MessageReactionCountDetails `json:"count_details,omitempty"`
	Me           bool                        `json:"me"`
	MeBurst      bool                        `json:"me_burst"`
	Emoji        Emoji
	// BurstColors []int `json:"burst_colors"`
}

type MessageReactionCountDetails struct {
	Burst  int `json:"burst"`
	Normal int `json:"normal"`
}

type MessageType int

const (
	MessageTypeDefault                                 MessageType = 0
	MessageTypeRecipientAdd                            MessageType = 1
	MessageTypeRecipientRemove                         MessageType = 2
	MessageTypeCall                                    MessageType = 3
	MessageTypeChannelNameChange                       MessageType = 4
	MessageTypeChannelIconChange                       MessageType = 5
	MessageTypeChannelPinnedMessage                    MessageType = 6
	MessageTypeGuildUserJoin                           MessageType = 7
	MessageTypeGuildBoost                              MessageType = 8
	MessageTypeGuildBoostTier1                         MessageType = 9
	MessageTypeGuildBoostTier2                         MessageType = 10
	MessageTypeGuildBoostTier3                         MessageType = 11
	MessageTypeChannelFollowAdd                        MessageType = 12
	MessageTypeGuildDiscoveryDisqualified              MessageType = 14
	MessageTypeGuildDiscoveryRequalified               MessageType = 15
	MessageTypeGuildDiscoveryGracePeriodInitialWarning MessageType = 16
	MessageTypeGuildDiscoveryGracePeriodFinalWarning   MessageType = 17
	MessageTypeThreadCreated                           MessageType = 18
	MessageTypeReply                                   MessageType = 19
	MessageTypeChatInputCommand                        MessageType = 20
	MessageTypeThreadStarterMessage                    MessageType = 21
	MessageTypeGuildInviteReminder                     MessageType = 22
	MessageTypeContextMenuCommand                      MessageType = 23
	MessageTypeAutoModerationAction                    MessageType = 24
	MessageTypeRoleSubscriptionPurchase                MessageType = 25
	MessageTypeInteractionPremiumUpsell                MessageType = 26
	MessageTypeStageStart                              MessageType = 27
	MessageTypeStageEnd                                MessageType = 28
	MessageTypeStageSpeaker                            MessageType = 29
	MessageTypeStageTopic                              MessageType = 31
	MessageTypeGuildApplicationPremiumSubscription     MessageType = 32
)

type MessageReference struct {
	MessageID       string `json:"message_id,omitempty"`
	ChannelID       string `json:"channel_id,omitempty"`
	GuildID         string `json:"guild_id,omitempty"`
	FailIfNotExists bool   `json:"fail_if_not_exists,omitempty"`
}

type MessageInteraction struct {
	ID     string          `json:"id"`
	Type   InteractionType `json:"type"`
	Name   string          `json:"name"`
	User   User            `json:"user"`
	Member *Member         `json:"member,omitempty"`
}

type MessageComponent struct{}

type MessageCreateEvent = Message

type MessageUpdateEvent = MessageDeleteResponse
type MessageDeleteEvent struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	GuildID   string `json:"guild_id,omitempty"`
}

type MessageDeleteBulkEvent struct {
	IDs       []string `json:"ids"`
	ChannelID string   `json:"channel_id"`
	GuildID   string   `json:"guild_id,omitempty"`
}

type MessageReactionAddEvent struct {
	UserID          string  `json:"user_id"`
	ChannelID       string  `json:"channel_id"`
	MessageID       string  `json:"message_id"`
	GuildID         string  `json:"guild_id,omitempty"`
	Member          *Member `json:"member,omitempty"`
	Emoji           Emoji   `json:"emoji"`
	MessageAuthorID string  `json:"message_author_id"`
}

type MessageReactionRemoveEvent struct {
	UserID    string `json:"user_id"`
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	GuildID   string `json:"guild_id,omitempty"`
	Emoji     Emoji  `json:"emoji"`
}

type MessageReactionRemoveAllEvent struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	GuildID   string `json:"guild_id,omitempty"`
}

type MessageReactionRemoveEmojiEvent struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	GuildID   string `json:"guild_id,omitempty"`
	Emoji     Emoji  `json:"emoji"`
}

type MessageListCall struct {
	ChannelID string `json:"channel_id"`
	Limit     int    `json:"limit,omitempty"`
	Before    string `json:"before,omitempty"`
	After     string `json:"after,omitempty"`
	Around    string `json:"around,omitempty"`
}

type MessageListResponse []Message

type MessageGetCall struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
}

type MessageGetResponse = Message

type MessageCreateCall struct {
	ChannelID string `json:"channel_id"`
	Content   string `json:"content"`
	TTS       bool   `json:"tts"`
	// Attachments      []MessageAttachment `json:"attachments,omitempty"`
	Embeds           []MessageEmbed     `json:"embeds,omitempty"`
	MessageReference *MessageReference  `json:"message_reference,omitempty"`
	Components       []MessageComponent `json:"components,omitempty"`
}

type MessageCreateResponse = Message

type MessageUpdateCall struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	Content   string `json:"content"`
	// Attachments      []MessageAttachment `json:"attachments,omitempty"`
	Embeds           []MessageEmbed     `json:"embeds,omitempty"`
	MessageReference *MessageReference  `json:"message_reference,omitempty"`
	Components       []MessageComponent `json:"components,omitempty"`
}

type MessageUpdateResponse = Message

type MessageDeleteCall struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
}

type MessageDeleteResponse struct{}

type MessageDeleteBulkCall struct {
	ChannelID string   `json:"channel_id"`
	Messages  []string `json:"messages"`
}

type MessageDeleteBulkResponse struct{}

type MessageReactionCreateCall struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	// unicode emoji or name:id for custom emojis
	Emoji string `json:"emoji"`
}

type MessageReactionCreateResponse struct{}

type MessageReactionDeleteOwnCall struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	// unicode emoji or name:id for custom emojis
	Emoji string `json:"emoji"`
}

type MessageReactionDeleteOwnResponse struct{}

type MessageReactionDeleteUserCall struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	// unicode emoji or name:id for custom emojis
	Emoji  string `json:"emoji"`
	UserID string `json:"user_id"`
}

type MessageReactionDeleteUserResponse struct{}

type MessageReactionListCall struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	// unicode emoji or name:id for custom emojis
	Emoji string `json:"emoji"`
	After string `json:"after,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

type MessageReactionListResponse []User

type MessageReactionDeleteAllCall struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
}

type MessageReactionDeleteAllResponse struct{}

type MessageReactionDeleteEmojiCall struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
	// unicode emoji or name:id for custom emojis
	Emoji string `json:"emoji"`
}

type MessageReactionDeleteEmojiResponse struct{}

type MessageGetPinnedCall struct {
	ChannelID string `json:"channel_id"`
}

type MessageGetPinnedResponse []Message

type MessagePinCall struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
}

type MessagePinResponse struct{}

type MessageUnpinCall struct {
	ChannelID string `json:"channel_id"`
	MessageID string `json:"message_id"`
}

type MessageUnpinResponse struct{}
