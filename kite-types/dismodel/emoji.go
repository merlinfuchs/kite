package dismodel

type Emoji struct{}

type GuildEmojisUpdateEvent struct {
	Emojis []Emoji `json:"emojis"`
}

type EmojiListCall struct {
	GuildID string `json:"guild_id"`
}

type EmojiListResponse = []Emoji

type EmojiGetCall struct {
	EmojiID string `json:"emoji_id"`
}

type EmojiGetResponse = Emoji

type EmojiCreateCall struct {
	Name  string   `json:"name"`
	Image string   `json:"image"`
	Roles []string `json:"roles"`
}

type EmojiCreateResponse = Emoji

type EmojiUpdateCall struct {
	EmojiID string   `json:"emoji_id"`
	Name    string   `json:"name"`
	Roles   []string `json:"roles"`
}

type EmojiUpdateResponse = Emoji

type EmojiDeleteCall struct {
	EmojiID string `json:"emoji_id"`
}

type EmojiDeleteResponse struct{}
