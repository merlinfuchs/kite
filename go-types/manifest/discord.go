package manifest

type DiscordCommand struct {
	Type                     string                  `json:"type"`
	Name                     string                  `json:"name"`
	Description              string                  `json:"description"`
	DefaultMemberPermissions []string                `json:"default_member_permissions"`
	DMPermission             bool                    `json:"dm_permission"`
	NSFW                     bool                    `json:"nsfw"`
	Options                  []DiscordCommandOptions `json:"options"`
}

type DiscordCommandOptions struct {
	Type        string                       `json:"type"`
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Required    bool                         `json:"required"`
	MinValue    int                          `json:"min_value"`
	MaxValue    int                          `json:"max_value"`
	MinLength   int                          `json:"min_length"`
	MaxLength   int                          `json:"max_length"`
	Choices     []DiscordCommandOptionChoice `json:"choices"`
	Options     []DiscordCommandOptions      `json:"options"`
}

type DiscordCommandOptionChoice struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
