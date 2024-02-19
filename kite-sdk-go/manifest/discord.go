package manifest

import (
	"github.com/merlinfuchs/dismod/distype"
)

type DiscordCommand struct {
	Type                     distype.ApplicationCommandType `json:"type,omitempty"`
	Name                     string                         `json:"name"`
	NameLocalizations        map[string]string              `json:"name_localizations,omitempty"`
	Description              string                         `json:"description"`
	DescriptionLocalizations map[string]string              `json:"description_localizations,omitempty"`
	Options                  []DiscordCommandOption         `json:"options,omitempty"`
	DefaultMemberPermissions string                         `json:"default_member_permissions"`
	DMPermission             *bool                          `json:"dm_permission,omitempty"`
	NSFW                     *bool                          `json:"nsfw,omitempty"`
}

type DiscordCommandOption struct {
	Type                 distype.ApplicationCommandOptionType `json:"type"`
	Name                 string                               `json:"name"`
	NameLocations        map[string]string                    `json:"name_locations,omitempty"`
	Description          string                               `json:"description"`
	DescriptionLocations map[string]string                    `json:"description_locations,omitempty"`
	Required             *bool                                `json:"required,omitempty"`
	Choices              []DiscordCommandOptionChoice         `json:"choices,omitempty"`
	Options              []DiscordCommandOption               `json:"options,omitempty"`
	ChannelTypes         []distype.ChannelType                `json:"channel_types,omitempty"`
	MinValue             *int                                 `json:"min_value,omitempty"`
	MaxValue             *int                                 `json:"max_value,omitempty"`
	MinLength            *int                                 `json:"min_length,omitempty"`
	MaxLength            *int                                 `json:"max_length,omitempty"`
	Autocomplete         *bool                                `json:"autocomplete,omitempty"`
}

type DiscordCommandOptionChoice struct {
	Name              string            `json:"name"`
	NameLocalizations map[string]string `json:"name_localizations,omitempty"`
	Value             interface{}       `json:"value"`
}

func (m *Manifest) DiscordApplicationCommands() []distype.ApplicationCommandCreateRequest {
	res := make([]distype.ApplicationCommandCreateRequest, len(m.DiscordCommands))
	for i, cmd := range m.DiscordCommands {
		var cmdType *distype.ApplicationCommandType
		if cmd.Type != 0 {
			temp := distype.ApplicationCommandType(cmd.Type)
			cmdType = &temp
		}

		options := make([]distype.ApplicationCommandOption, len(cmd.Options))
		for i, opt := range cmd.Options {
			options[i] = distype.ApplicationCommandOption{
				Type:                 distype.ApplicationCommandOptionType(opt.Type),
				Name:                 opt.Name,
				NameLocations:        opt.NameLocations,
				Description:          opt.Description,
				DescriptionLocations: opt.DescriptionLocations,
				Required:             opt.Required,

				// TODO: support all types
			}
		}

		res[i] = distype.ApplicationCommandCreateRequest{
			Type:                     cmdType,
			Name:                     cmd.Name,
			NameLocalizations:        cmd.NameLocalizations,
			Description:              cmd.Description,
			DescriptionLocalizations: cmd.DescriptionLocalizations,
			DefaultMemberPermissions: distype.NullString(cmd.DefaultMemberPermissions, cmd.DefaultMemberPermissions != ""),
			DMPermission:             cmd.DMPermission,
			NSFW:                     cmd.NSFW,
			Options:                  options,
		}
	}

	return res
}
