package starboard

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

var customEmojiRegex = regexp.MustCompile(`<a?:(\w+):(\d+)>`)

type StarboardPluginInstance struct {
	appID  string
	config plugin.ConfigValues
}

func (p *StarboardPluginInstance) Update(ctx context.Context, config plugin.ConfigValues) error {
	p.config = config
	return nil
}

func (p *StarboardPluginInstance) HandleEvent(c plugin.Context, event gateway.Event) error {
	e, ok := event.(*gateway.MessageReactionAddEvent)
	if !ok {
		return nil
	}

	config, err := getPluginConfig(c, e.GuildID)
	if err != nil {
		return fmt.Errorf("failed to get starboard config: %w", err)
	}

	if config == nil {
		return nil
	}

	if e.Emoji.ID != config.Emoji.ID || e.Emoji.Name != config.Emoji.Name {
		return nil
	}

	message, err := c.Discord().Message(c, e.ChannelID, e.MessageID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	reactionCount, err := getMessageReactionCount(message, config.Emoji)
	if err != nil {
		return fmt.Errorf("failed to get message reaction count: %w", err)
	}

	if reactionCount >= config.Threshold {
		messageData, err := getStarboardMessageData(message, config.Emoji, reactionCount)
		if err != nil {
			return fmt.Errorf("failed to get starboard message data: %w", err)
		}

		existingMessageID, err := c.GetValue(c, starboardMessageKey(config.ChannelID.String(), e.MessageID.String()))
		if err != nil {
			return fmt.Errorf("failed to get existing message ID: %w", err)
		}

		if existingMessageID != thing.Null {
			messageID := discord.MessageID(existingMessageID.Int())
			_, err := c.Discord().EditMessage(c, config.ChannelID, messageID, api.EditMessageData{
				Content:    option.NewNullableString(messageData.Content),
				Embeds:     &messageData.Embeds,
				Components: &messageData.Components,
			})
			if err != nil {
				return fmt.Errorf("failed to edit message: %w", err)
			}

			return nil
		}

		newMessage, err := c.Discord().CreateMessage(c, config.ChannelID, api.SendMessageData{
			Content:    messageData.Content,
			Embeds:     messageData.Embeds,
			Components: messageData.Components,
		})
		if err != nil {
			return fmt.Errorf("failed to create new message: %w", err)
		}

		_, err = c.UpdateValue(
			c,
			starboardMessageKey(config.ChannelID.String(), e.MessageID.String()),
			provider.VariableOperationOverwrite,
			thing.NewInt(int64(newMessage.ID)),
		)
		if err != nil {
			return fmt.Errorf("failed to update existing message ID: %w", err)
		}

		return nil
	}

	return nil
}

func (p *StarboardPluginInstance) HandleCommand(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	data, ok := event.Data.(*discord.CommandInteraction)
	if !ok {
		return nil
	}

	if data.Name != "starboard" {
		return nil
	}

	for _, opt := range data.Options {
		switch opt.Name {
		case "enable":
			var channelID discord.ChannelID
			var threshold int
			emoji := discord.Emoji{
				Name: "⭐",
			}
			for _, subOpt := range opt.Options {
				switch subOpt.Name {
				case "channel":
					_ = subOpt.Value.UnmarshalTo(&channelID)
				case "threshold":
					_ = subOpt.Value.UnmarshalTo(&threshold)
				case "emoji":
					var rawEmoji string
					_ = subOpt.Value.UnmarshalTo(&rawEmoji)
					if customEmojiRegex.MatchString(rawEmoji) {
						matches := customEmojiRegex.FindStringSubmatch(rawEmoji)
						emojiID, err := discord.ParseSnowflake(matches[2])
						if err != nil {
							return err
						}
						emoji = discord.Emoji{
							Name: matches[1],
							ID:   discord.EmojiID(emojiID),
						}
					} else {
						emoji = discord.Emoji{
							Name: rawEmoji,
						}
					}
				}
			}

			rawConfig, err := json.Marshal(StarboardConfig{
				ChannelID: channelID,
				Threshold: threshold,
				Emoji:     emoji,
			})
			if err != nil {
				return err
			}

			_, err = c.UpdateValue(
				c,
				starboardConfigKey(event.GuildID.String()),
				provider.VariableOperationOverwrite,
				thing.NewString(string(rawConfig)),
			)
			if err != nil {
				return err
			}

			_, err = c.Discord().CreateInteractionResponse(c, event.ID, event.Token, api.InteractionResponse{
				Type: api.MessageInteractionWithSource,
				Data: &api.InteractionResponseData{
					Content: option.NewNullableString(fmt.Sprintf(
						"Starboard has been enabled for this server. Messages will be copied to <#%d> when they have %d %s reactions.",
						channelID,
						threshold,
						emoji,
					)),
					Flags: discord.EphemeralMessage,
				},
			})
			if err != nil {
				return err
			}

			return nil
		case "disable":
			err := c.DeleteValue(c, starboardConfigKey(event.GuildID.String()))
			if err != nil {
				return err
			}

			_, err = c.Discord().CreateInteractionResponse(c, event.ID, event.Token, api.InteractionResponse{
				Type: api.MessageInteractionWithSource,
				Data: &api.InteractionResponseData{
					Content: option.NewNullableString("Starboard has been disabled for this server."),
					Flags:   discord.EphemeralMessage,
				},
			})
			if err != nil {
				return err
			}

			return nil
		}
	}

	return nil
}

func (p *StarboardPluginInstance) HandleComponent(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	return nil
}

func (p *StarboardPluginInstance) HandleModal(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	return nil
}

func (p *StarboardPluginInstance) Close() error {
	return nil
}

func getPluginConfig(c plugin.Context, guildID discord.GuildID) (*StarboardConfig, error) {
	rawConfig, err := c.GetValue(c, starboardConfigKey(guildID.String()))
	if err != nil {
		return nil, err
	}

	if rawConfig == thing.Null {
		return nil, nil
	}

	config := StarboardConfig{}
	err = json.Unmarshal([]byte(rawConfig.String()), &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func getMessageReactionCount(message *discord.Message, emoji discord.Emoji) (int, error) {
	for _, reaction := range message.Reactions {
		if reaction.Emoji.ID != emoji.ID || reaction.Emoji.Name != emoji.Name {
			continue
		}

		return reaction.Count, nil
	}

	return 0, nil
}

func getStarboardMessageData(message *discord.Message, emoji discord.Emoji, reactionCount int) (*discord.Message, error) {
	emojiText := emoji.Name
	if emoji.ID.IsValid() {
		emojiText = fmt.Sprintf("<%s:%d>", emoji.Name, emoji.ID)
	}

	return &discord.Message{
		Content: fmt.Sprintf("%s **%d**", emojiText, reactionCount),
		Embeds: []discord.Embed{
			{
				Author: &discord.EmbedAuthor{
					Name: message.Author.Username,
					Icon: message.Author.AvatarURL(),
					URL:  message.URL(),
				},
				Description: message.Content,
				Color:       0xffac33,
				Timestamp:   message.Timestamp,
			},
		},
		Components: discord.ContainerComponents{
			&discord.ActionRowComponent{
				&discord.ButtonComponent{
					Label: "Go to Message",
					Style: discord.LinkButtonStyle(discord.URL(message.URL())),
				},
			},
		},
	}, nil
}

func starboardConfigKey(guildID string) string {
	return "starboard-config:" + guildID
}

func starboardMessageKey(channelID string, messageID string) string {
	return "starboard-message:" + channelID + ":" + messageID
}

type StarboardConfig struct {
	ChannelID discord.ChannelID `json:"channel_id"`
	Threshold int               `json:"threshold"`
	Emoji     discord.Emoji     `json:"emoji"`
}
