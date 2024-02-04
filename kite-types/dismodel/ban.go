package dismodel

type Ban struct {
	Reason string `json:"reason"`
	User   User   `json:"user"`
}

type GuildBanAddEvent struct {
	GuildID string `json:"guild_id"`
	User    User   `json:"user"`
}

type GuildBanRemoveEvent struct {
	GuildID string `json:"guild_id"`
	User    User   `json:"user"`
}

type BanListCall struct {
	Limit  int    `json:"limit"`
	Before string `json:"before"`
	After  string `json:"after"`
}

type BanListResponse = []Ban

type BanGetCall struct {
	UserID string `json:"user_id"`
}

type BanGetResponse = Ban

type BanCreateCall struct {
	UserID               string `json:"user_id"`
	DeleteMessageSeconds int    `json:"delete_message_seconds,omitempty"`
}

type BanCreateResponse struct{}

type BanRemoveCall struct {
	UserID string `json:"user_id"`
}

type BanRemoveResponse struct{}
