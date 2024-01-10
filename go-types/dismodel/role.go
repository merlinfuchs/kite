package dismodel

type Role struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Color        int       `json:"color"`
	Hoist        bool      `json:"hoist"`
	Icon         string    `json:"icon,omitempty"`
	UnicodeEmoji string    `json:"unicode_emoji,omitempty"`
	Position     int       `json:"position"`
	Permissions  string    `json:"permissions"`
	Managed      bool      `json:"managed"`
	Mentionable  bool      `json:"mentionable"`
	Tags         []RoleTag `json:"tags,omitempty"`
	Flags        int       `json:"flags"`
}

type RoleTag struct {
	BotID                 string `json:"bot_id,omitempty"`
	IntegrationID         string `json:"integration_id,omitempty"`
	PremiumSubscriber     bool   `json:"premium_subscriber,omitempty"`
	SubscriptionListingID string `json:"subscription_listing_id,omitempty"`
	AvailableForPurchase  bool   `json:"available_for_purchase,omitempty"`
	GuildConnections      bool   `json:"guild_connections,omitempty"`
}

type GuildRoleCreateEvent struct {
	Role    Role   `json:"role"`
	GuildID string `json:"guild_id"`
}

type GuildRoleUpdateEvent struct {
	Role    Role   `json:"role"`
	GuildID string `json:"guild_id"`
}

type GuildRoleDeleteEvent struct {
	RoleID  string `json:"role_id"`
	GuildID string `json:"guild_id"`
}

type RoleListCall struct{}

type RoleListResponse = []Role

type RoleCreateCall struct {
	Name         string `json:"name"`
	Permissions  string `json:"permissions"`
	Color        int    `json:"color,omitempty"`
	Hoist        bool   `json:"hoist,omitempty"`
	Icon         string `json:"icon,omitempty"`
	UnicodeEmoji string `json:"unicode_emoji,omitempty"`
	Mentionable  bool   `json:"mentionable,omitempty"`
}

type RoleCreateResponse = Role

type RoleUpdateCall struct {
	RoleID       string `json:"role_id"`
	Name         string `json:"name"`
	Permissions  string `json:"permissions"`
	Color        int    `json:"color,omitempty"`
	Hoist        bool   `json:"hoist,omitempty"`
	Icon         string `json:"icon,omitempty"`
	UnicodeEmoji string `json:"unicode_emoji,omitempty"`
	Mentionable  bool   `json:"mentionable,omitempty"`
}

type RoleUpdateResponse = Role

type RoleUpdatePositionsCall = []RoleUpdatePositionsEntry

type RoleUpdatePositionsEntry struct {
	RoleID   string `json:"role_id"`
	Position int    `json:"position"`
}

type RoleUpdatePositionsResponse = []Role

type RoleDeleteCall struct {
	RoleID string `json:"role_id"`
}

type RoleDeleteResponse struct{}
