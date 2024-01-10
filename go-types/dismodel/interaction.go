package dismodel

import "encoding/json"

type Interaction struct {
	ID             string          `json:"id"`
	ApplicationID  string          `json:"application_id"`
	Type           InteractionType `json:"type"`
	Data           interface{}     `json:"data,omitempty"`
	GuildID        string          `json:"guild_id,omitempty"`
	Channel        *Channel        `json:"channel,omitempty"`
	ChannelID      string          `json:"channel_id,omitempty"`
	Member         *Member         `json:"member,omitempty"`
	User           *User           `json:"user,omitempty"`
	Token          string          `json:"token"`
	Version        int             `json:"version"`
	Message        *Message        `json:"message,omitempty"`
	AppPermissions string          `json:"application_permissions,omitempty"`
	Locale         string          `json:"locale,omitempty"`
	GuildLocale    string          `json:"guild_locale,omitempty"`
}

func (i *Interaction) UnmarshalJSON(b []byte) error {
	var raw struct {
		ID             string          `json:"id"`
		ApplicationID  string          `json:"application_id"`
		Type           InteractionType `json:"type"`
		Data           json.RawMessage `json:"data,omitempty"`
		GuildID        string          `json:"guild_id,omitempty"`
		Channel        *Channel        `json:"channel,omitempty"`
		ChannelID      string          `json:"channel_id,omitempty"`
		Member         *Member         `json:"member,omitempty"`
		User           *User           `json:"user,omitempty"`
		Token          string          `json:"token"`
		Version        int             `json:"version"`
		Message        *Message        `json:"message,omitempty"`
		AppPermissions string          `json:"application_permissions,omitempty"`
		Locale         string          `json:"locale,omitempty"`
		GuildLocale    string          `json:"guild_locale,omitempty"`
	}

	json.Unmarshal(b, &raw)

	*i = Interaction{
		ID:             raw.ID,
		ApplicationID:  raw.ApplicationID,
		Type:           raw.Type,
		GuildID:        raw.GuildID,
		Channel:        raw.Channel,
		ChannelID:      raw.ChannelID,
		Member:         raw.Member,
		User:           raw.User,
		Token:          raw.Token,
		Version:        raw.Version,
		Message:        raw.Message,
		AppPermissions: raw.AppPermissions,
		Locale:         raw.Locale,
		GuildLocale:    raw.GuildLocale,
	}

	switch raw.Type {
	case InteractionTypeApplicationCommand:
		var data ApplicationCommandInteractionData
		err := json.Unmarshal(raw.Data, &data)
		if err != nil {
			return err
		}
		i.Data = data
		break
	case InteractionTypeMessageComponent:
		var data MessageComponentInteractionData
		err := json.Unmarshal(raw.Data, &data)
		if err != nil {
			return err
		}
		i.Data = data
		break
	}

	return nil
}

type InteractionType int

const (
	InteractionTypePing                           InteractionType = 1
	InteractionTypeApplicationCommand             InteractionType = 2
	InteractionTypeMessageComponent               InteractionType = 3
	InteractionTypeApplicationCommandAutocomplete InteractionType = 4
	InteractionTypeModalSubmit                    InteractionType = 5
)

type ApplicationCommandInteractionData struct {
	ID       string                         `json:"id"`
	Name     string                         `json:"name"`
	Type     ApplicationCommandType         `json:"type"`
	Resolved *ResolvedData                  `json:"resolved,omitempty"`
	Options  []ApplicationCommandOptionData `json:"options,omitempty"`
	GuildID  string                         `json:"guild_id,omitempty"`
	TargetID string                         `json:"target_id,omitempty"`
}

type ApplicationCommandType int

const (
	ApplicationCommandTypeChatInput ApplicationCommandType = 1
	ApplicationCommandTypeUser      ApplicationCommandType = 2
	ApplicationCommandTypeMessage   ApplicationCommandType = 3
)

type ApplicationCommandOptionData struct {
	Name    string                         `json:"name"`
	Type    ApplicationCommandOptionType   `json:"type"`
	Value   interface{}                    `json:"value,omitempty"`
	Options []ApplicationCommandOptionData `json:"options,omitempty"`
	Focused bool                           `json:"focused,omitempty"`
}

type ApplicationCommandOptionType int

const (
	ApplicationCommandOptionTypeSubCommand      ApplicationCommandOptionType = 1
	ApplicationCommandOptionTypeSubCommandGroup ApplicationCommandOptionType = 2
	ApplicationCommandOptionTypeString          ApplicationCommandOptionType = 3
	ApplicationCommandOptionTypeInteger         ApplicationCommandOptionType = 4
	ApplicationCommandOptionTypeBoolean         ApplicationCommandOptionType = 5
	ApplicationCommandOptionTypeUser            ApplicationCommandOptionType = 6
	ApplicationCommandOptionTypeChannel         ApplicationCommandOptionType = 7
	ApplicationCommandOptionTypeRole            ApplicationCommandOptionType = 8
	ApplicationCommandOptionTypeMentionable     ApplicationCommandOptionType = 9
	ApplicationCommandOptionTypeNumber          ApplicationCommandOptionType = 10
	ApplicationCommandOptionTypeAttachment      ApplicationCommandOptionType = 11
)

type MessageComponentInteractionData struct {
	CustomID      string               `json:"custom_id"`
	ComponentType MessageComponentType `json:"component_type"`
	Values        []string             `json:"values,omitempty"`
	Resolved      *ResolvedData        `json:"resolved,omitempty"`
}

type MessageComponentType int

const (
	MessageComponentTypeActionRow         MessageComponentType = 1
	MessageComponentTypeButton            MessageComponentType = 2
	MessageComponentTypeStringSelect      MessageComponentType = 3
	MessageComponentTypeTextInput         MessageComponentType = 4
	MessageComponentTypeUserSelect        MessageComponentType = 5
	MessageComponentTypeRoleSelect        MessageComponentType = 6
	MessageComponentTypeMentionableSelect MessageComponentType = 7
	MessageComponentTypeChannelSelect     MessageComponentType = 8
)

type ResolvedData struct {
	Users       map[string]User              `json:"users,omitempty"`
	Members     map[string]Member            `json:"members,omitempty"`
	Roles       map[string]Role              `json:"roles,omitempty"`
	Channels    map[string]Channel           `json:"channels,omitempty"`
	Messages    map[string]Message           `json:"messages,omitempty"`
	Attachments map[string]MessageAttachment `json:"attachments,omitempty"`
}

type InteractionResponseData struct {
	TTS     bool           `json:"tts,omitempty"`
	Content string         `json:"content,omitempty"`
	Embeds  []MessageEmbed `json:"embeds,omitempty"`
	//AllowedMentions *AllowedMentions `json:"allowed_mentions,omitempty"`
	Flags      int                `json:"flags,omitempty"`
	Components []MessageComponent `json:"components,omitempty"`
	//Attachments    []MessageAttachment `json:"attachments,omitempty"`
}

type InteractionCreateEvent = Interaction

type InteractionResponseCreateCall struct {
	ID    string                  `json:"id"`
	Token string                  `json:"token"`
	Data  InteractionResponseData `json:"data"`
}

type InteractionResponseCreateResponse = Message

type InteractionResponseUpdateCall struct {
	ID    string `json:"id"`
	Token string `json:"token"`
	Data  string `json:"data"`
}

type InteractionResponseUpdateResponse = Message

type InteractionResponseDeleteCall struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

type InteractionResponseDeleteResponse struct{}

type InteractionResponseGetCall struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

type InteractionResponseGetResponse = Message

type InteractionFollowupCreateCall struct {
	ID    string                  `json:"id"`
	Token string                  `json:"token"`
	Data  InteractionResponseData `json:"data"`
}

type InteractionFollowupCreateResponse = Message

type InteractionFollowupUpdateCall struct {
	ID        string                  `json:"id"`
	Token     string                  `json:"token"`
	MessageID string                  `json:"message_id"`
	Data      InteractionResponseData `json:"data"`
}

type InteractionFollowupUpdateResponse = Message

type InteractionFollowupDeleteCall struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	MessageID string `json:"message_id"`
}

type InteractionFollowupDeleteResponse struct{}

type InteractionFollowupGetCall struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	MessageID string `json:"message_id"`
}

type InteractionFollowupGetResponse = Message
