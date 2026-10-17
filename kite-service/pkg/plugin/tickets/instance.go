package tickets

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

const (
	customIDOpen         = "kite-tickets:open"
	customIDClose        = "kite-tickets:close"
	customIDCloseConfirm = "kite-tickets:close-confirm"

	defaultPanelMessage = "Need help? Click the button below to open a private ticket with the staff team."

	ticketColor = 0x5865f2
	openColor   = 0x57f287
	closeColor  = 0xed4245
)

// memberPerms are what the opener, added members and the support role get in
// a ticket channel.
const memberPerms = discord.PermissionViewChannel |
	discord.PermissionSendMessages |
	discord.PermissionReadMessageHistory |
	discord.PermissionAttachFiles |
	discord.PermissionEmbedLinks

type TicketsPluginInstance struct {
	appID  string
	config plugin.ConfigValues

	// openMu serializes opening tickets, so a member double clicking the
	// button doesn't get two channels.
	openMu sync.Mutex

	botUserMu sync.Mutex
	botUserID discord.UserID
}

type TicketsConfig struct {
	PanelChannelID discord.ChannelID `json:"panel_channel_id"`
	CategoryID     discord.ChannelID `json:"category_id,omitempty"`
	SupportRoleID  discord.RoleID    `json:"support_role_id,omitempty"`
	LogChannelID   discord.ChannelID `json:"log_channel_id,omitempty"`
}

// Ticket is stored for every open ticket channel. Only channels with a ticket
// can be closed, so the plugin never deletes a channel it didn't create.
type Ticket struct {
	GuildID discord.GuildID `json:"guild_id"`
	OwnerID discord.UserID  `json:"owner_id"`
	Number  int64           `json:"number"`
}

func (p *TicketsPluginInstance) Update(ctx context.Context, config plugin.ConfigValues) error {
	p.config = config
	return nil
}

func (p *TicketsPluginInstance) HandleEvent(c plugin.Context, event gateway.Event) error {
	return nil
}

func (p *TicketsPluginInstance) HandleCommand(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	data, ok := event.Data.(*discord.CommandInteraction)
	if !ok {
		return nil
	}

	if event.GuildID == 0 {
		return reply(c, event, "Tickets only work inside a server.")
	}

	switch data.Name {
	case "tickets":
		for _, opt := range data.Options {
			switch opt.Name {
			case "setup":
				return p.handleSetup(c, event, opt.Options)
			case "disable":
				return p.handleDisable(c, event)
			}
		}
	case "ticket":
		for _, opt := range data.Options {
			switch opt.Name {
			case "close":
				return p.handleCloseRequest(c, event)
			case "add":
				return p.handleMemberChange(c, event, data, opt.Options, true)
			case "remove":
				return p.handleMemberChange(c, event, data, opt.Options, false)
			}
		}
	}

	return nil
}

func (p *TicketsPluginInstance) HandleComponent(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	data, ok := event.Data.(discord.ComponentInteraction)
	if !ok {
		return nil
	}

	switch string(data.ID()) {
	case customIDOpen:
		return p.handleOpen(c, event)
	case customIDClose:
		return p.handleCloseRequest(c, event)
	case customIDCloseConfirm:
		return p.handleClose(c, event)
	}

	return nil
}

func (p *TicketsPluginInstance) HandleModal(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	return nil
}

func (p *TicketsPluginInstance) Close() error {
	return nil
}

func (p *TicketsPluginInstance) handleSetup(c plugin.Context, event *gateway.InteractionCreateEvent, options discord.CommandInteractionOptions) error {
	config := TicketsConfig{}
	message := defaultPanelMessage

	for _, opt := range options {
		switch opt.Name {
		case "channel":
			_ = opt.Value.UnmarshalTo(&config.PanelChannelID)
		case "category":
			_ = opt.Value.UnmarshalTo(&config.CategoryID)
		case "support-role":
			_ = opt.Value.UnmarshalTo(&config.SupportRoleID)
		case "log-channel":
			_ = opt.Value.UnmarshalTo(&config.LogChannelID)
		case "message":
			_ = opt.Value.UnmarshalTo(&message)
		}
	}

	if !config.PanelChannelID.IsValid() {
		return reply(c, event, "Pick a channel to post the ticket panel in.")
	}

	_, err := c.Discord().CreateMessage(c, config.PanelChannelID, api.SendMessageData{
		Embeds: []discord.Embed{
			{
				Title:       "🎫 Support Tickets",
				Description: message,
				Color:       ticketColor,
			},
		},
		Components: discord.TopLevelComponents{
			&discord.ActionRowComponent{
				&discord.ButtonComponent{
					Label:    "Open Ticket",
					CustomID: customIDOpen,
					Style:    discord.PrimaryButtonStyle(),
					Emoji:    &discord.ComponentEmoji{Name: "🎫"},
				},
			},
		},
	})
	if err != nil {
		replyErr := reply(c, event, fmt.Sprintf("Failed to post the ticket panel in <#%d>. Make sure the bot can see and send messages in that channel.", config.PanelChannelID))
		if replyErr != nil {
			return replyErr
		}
		return fmt.Errorf("failed to post ticket panel: %w", err)
	}

	rawConfig, err := json.Marshal(config)
	if err != nil {
		return err
	}

	_, err = c.UpdateValue(
		c,
		ticketsConfigKey(event.GuildID.String()),
		provider.VariableOperationOverwrite,
		thing.NewString(string(rawConfig)),
	)
	if err != nil {
		return err
	}

	return reply(c, event, fmt.Sprintf(
		"Tickets are set up. The panel was posted in <#%d>. The bot needs the `Manage Channels` and `Manage Roles` permissions to open tickets.",
		config.PanelChannelID,
	))
}

func (p *TicketsPluginInstance) handleDisable(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	err := c.DeleteValue(c, ticketsConfigKey(event.GuildID.String()))
	if err != nil {
		return err
	}

	return reply(c, event, "Tickets have been disabled for this server. Open tickets can still be closed.")
}

func (p *TicketsPluginInstance) handleOpen(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	if event.GuildID == 0 || event.Member == nil {
		return nil
	}

	// Creating the channel can take longer than Discord waits for a response.
	_, err := c.Discord().CreateInteractionResponse(c, event.ID, event.Token, api.InteractionResponse{
		Type: api.DeferredMessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Flags: discord.EphemeralMessage,
		},
	})
	if err != nil {
		return err
	}

	p.openMu.Lock()
	defer p.openMu.Unlock()

	config, err := getTicketsConfig(c, event.GuildID)
	if err != nil {
		return err
	}
	if config == nil {
		return editReply(c, event, "Tickets are disabled in this server.")
	}

	user := event.Member.User

	existing, err := c.GetValue(c, ticketsOpenKey(event.GuildID.String(), user.ID.String()))
	if err != nil {
		return err
	}
	if existing != thing.Null {
		channelID := discord.ChannelID(existing.Int())
		if _, err := c.Discord().Channel(c, channelID); err == nil {
			return editReply(c, event, fmt.Sprintf("You already have an open ticket: <#%d>", channelID))
		}

		// The channel was deleted by hand, so the ticket isn't open anymore.
		if err := cleanupTicket(c, event.GuildID, user.ID, channelID); err != nil {
			return err
		}
	}

	botUserID, err := p.getBotUserID(c)
	if err != nil {
		return err
	}

	number, err := c.UpdateValue(
		c,
		ticketsCounterKey(event.GuildID.String()),
		provider.VariableOperationIncrement,
		thing.NewInt(1),
	)
	if err != nil {
		return err
	}

	overwrites := []discord.Overwrite{
		{
			ID:   discord.Snowflake(event.GuildID),
			Type: discord.OverwriteRole,
			Deny: discord.PermissionViewChannel,
		},
		{
			ID:    discord.Snowflake(user.ID),
			Type:  discord.OverwriteMember,
			Allow: memberPerms,
		},
		{
			ID:    discord.Snowflake(botUserID),
			Type:  discord.OverwriteMember,
			Allow: memberPerms,
		},
	}
	if config.SupportRoleID.IsValid() {
		overwrites = append(overwrites, discord.Overwrite{
			ID:    discord.Snowflake(config.SupportRoleID),
			Type:  discord.OverwriteRole,
			Allow: memberPerms,
		})
	}

	channel, err := c.Discord().CreateChannel(c, event.GuildID, api.CreateChannelData{
		Name:           fmt.Sprintf("ticket-%04d", number.Int()),
		Type:           discord.GuildText,
		Topic:          fmt.Sprintf("Ticket #%04d opened by %s (%s)", number.Int(), user.Tag(), user.ID),
		CategoryID:     config.CategoryID,
		Overwrites:     overwrites,
		AuditLogReason: api.AuditLogReason(fmt.Sprintf("Ticket opened by %s", user.Tag())),
	})
	if err != nil {
		replyErr := editReply(c, event, "Failed to create your ticket. Ask an admin to make sure the bot has the `Manage Channels` and `Manage Roles` permissions.")
		if replyErr != nil {
			return replyErr
		}
		return fmt.Errorf("failed to create ticket channel: %w", err)
	}

	rawTicket, err := json.Marshal(Ticket{
		GuildID: event.GuildID,
		OwnerID: user.ID,
		Number:  number.Int(),
	})
	if err != nil {
		return err
	}

	_, err = c.UpdateValue(c, ticketsChannelKey(channel.ID.String()), provider.VariableOperationOverwrite, thing.NewString(string(rawTicket)))
	if err != nil {
		return err
	}

	_, err = c.UpdateValue(c, ticketsOpenKey(event.GuildID.String(), user.ID.String()), provider.VariableOperationOverwrite, thing.NewInt(int64(channel.ID)))
	if err != nil {
		return err
	}

	content := user.Mention()
	allowedMentions := &api.AllowedMentions{
		Users: []discord.UserID{user.ID},
	}
	if config.SupportRoleID.IsValid() {
		content += " " + config.SupportRoleID.Mention()
		allowedMentions.Roles = []discord.RoleID{config.SupportRoleID}
	}

	_, err = c.Discord().CreateMessage(c, channel.ID, api.SendMessageData{
		Content: content,
		Embeds: []discord.Embed{
			{
				Title:       fmt.Sprintf("Ticket #%04d", number.Int()),
				Description: fmt.Sprintf("Thanks for opening a ticket, %s. Describe your issue and the staff team will be with you soon.\n\nClick the button below or use `/ticket close` when you're done.", user.Mention()),
				Color:       ticketColor,
			},
		},
		Components: discord.TopLevelComponents{
			&discord.ActionRowComponent{
				&discord.ButtonComponent{
					Label:    "Close Ticket",
					CustomID: customIDClose,
					Style:    discord.DangerButtonStyle(),
					Emoji:    &discord.ComponentEmoji{Name: "🔒"},
				},
			},
		},
		AllowedMentions: allowedMentions,
	})
	if err != nil {
		return fmt.Errorf("failed to send ticket welcome message: %w", err)
	}

	logTicket(c, config, discord.Embed{
		Title:       "Ticket Opened",
		Description: fmt.Sprintf("%s opened ticket #%04d: %s", user.Mention(), number.Int(), channel.Mention()),
		Color:       openColor,
		Timestamp:   discord.NewTimestamp(time.Now()),
	})

	return editReply(c, event, fmt.Sprintf("Your ticket has been opened: %s", channel.Mention()))
}

func (p *TicketsPluginInstance) handleCloseRequest(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	ticket, err := getTicket(c, event.ChannelID)
	if err != nil {
		return err
	}
	if ticket == nil || ticket.GuildID != event.GuildID {
		return reply(c, event, "This command only works inside a ticket.")
	}

	_, err = c.Discord().CreateInteractionResponse(c, event.ID, event.Token, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content: option.NewNullableString("Are you sure you want to close this ticket? The channel will be deleted."),
			Flags:   discord.EphemeralMessage,
			Components: &discord.TopLevelComponents{
				&discord.ActionRowComponent{
					&discord.ButtonComponent{
						Label:    "Close Ticket",
						CustomID: customIDCloseConfirm,
						Style:    discord.DangerButtonStyle(),
						Emoji:    &discord.ComponentEmoji{Name: "🔒"},
					},
				},
			},
		},
	})
	return err
}

func (p *TicketsPluginInstance) handleClose(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	ticket, err := getTicket(c, event.ChannelID)
	if err != nil {
		return err
	}
	if ticket == nil || ticket.GuildID != event.GuildID {
		return reply(c, event, "This ticket has already been closed.")
	}

	sender := event.Sender()
	if sender == nil {
		return nil
	}

	_, err = c.Discord().CreateInteractionResponse(c, event.ID, event.Token, api.InteractionResponse{
		Type: api.UpdateMessage,
		Data: &api.InteractionResponseData{
			Content:    option.NewNullableString("Closing the ticket..."),
			Components: &discord.TopLevelComponents{},
		},
	})
	if err != nil {
		return err
	}

	_, err = c.Discord().APIRequest(c, provider.DiscordAPIRequest{
		Method: http.MethodDelete,
		Path:   "channels/" + event.ChannelID.String(),
		Reason: api.AuditLogReason(fmt.Sprintf("Ticket closed by %s", sender.Tag())),
	})
	if err != nil {
		_, editErr := c.Discord().EditInteractionResponse(c, event.AppID, event.Token, api.EditInteractionResponseData{
			Content: option.NewNullableString("Failed to delete the ticket channel. Make sure the bot has the `Manage Channels` permission."),
		})
		if editErr != nil {
			return editErr
		}
		return fmt.Errorf("failed to delete ticket channel: %w", err)
	}

	if err := cleanupTicket(c, ticket.GuildID, ticket.OwnerID, event.ChannelID); err != nil {
		return err
	}

	config, err := getTicketsConfig(c, event.GuildID)
	if err != nil {
		return err
	}
	if config != nil {
		logTicket(c, config, discord.Embed{
			Title:       "Ticket Closed",
			Description: fmt.Sprintf("%s closed ticket #%04d of <@%d>", sender.Mention(), ticket.Number, ticket.OwnerID),
			Color:       closeColor,
			Timestamp:   discord.NewTimestamp(time.Now()),
		})
	}

	return nil
}

func (p *TicketsPluginInstance) handleMemberChange(
	c plugin.Context,
	event *gateway.InteractionCreateEvent,
	data *discord.CommandInteraction,
	options discord.CommandInteractionOptions,
	add bool,
) error {
	ticket, err := getTicket(c, event.ChannelID)
	if err != nil {
		return err
	}
	if ticket == nil || ticket.GuildID != event.GuildID {
		return reply(c, event, "This command only works inside a ticket.")
	}

	var userID discord.UserID
	for _, opt := range options {
		if opt.Name == "user" {
			_ = opt.Value.UnmarshalTo(&userID)
		}
	}
	if !userID.IsValid() {
		return reply(c, event, "Pick a member.")
	}

	if user, ok := data.Resolved.Users[userID]; ok && user.Bot {
		return reply(c, event, "Bots can't be added to or removed from tickets.")
	}

	path := fmt.Sprintf("channels/%s/permissions/%s", event.ChannelID, userID)

	if add {
		body, err := json.Marshal(map[string]any{
			"type":  discord.OverwriteMember,
			"allow": fmt.Sprint(uint64(memberPerms)),
			"deny":  "0",
		})
		if err != nil {
			return err
		}

		_, err = c.Discord().APIRequest(c, provider.DiscordAPIRequest{
			Method: http.MethodPut,
			Path:   path,
			Body:   body,
			Reason: "Member added to ticket",
		})
		if err != nil {
			replyErr := reply(c, event, "Failed to add the member. Make sure the bot has the `Manage Roles` permission.")
			if replyErr != nil {
				return replyErr
			}
			return fmt.Errorf("failed to add member to ticket: %w", err)
		}

		return replyPublic(c, event, fmt.Sprintf("Added <@%d> to the ticket.", userID))
	}

	if userID == ticket.OwnerID {
		return reply(c, event, "The member who opened the ticket can't be removed from it.")
	}

	_, err = c.Discord().APIRequest(c, provider.DiscordAPIRequest{
		Method: http.MethodDelete,
		Path:   path,
		Reason: "Member removed from ticket",
	})
	if err != nil {
		replyErr := reply(c, event, "Failed to remove the member. Make sure the bot has the `Manage Roles` permission.")
		if replyErr != nil {
			return replyErr
		}
		return fmt.Errorf("failed to remove member from ticket: %w", err)
	}

	return replyPublic(c, event, fmt.Sprintf("Removed <@%d> from the ticket.", userID))
}

// getBotUserID returns the user ID of the bot, which needs an overwrite to see
// the ticket channels it creates.
func (p *TicketsPluginInstance) getBotUserID(c plugin.Context) (discord.UserID, error) {
	p.botUserMu.Lock()
	defer p.botUserMu.Unlock()

	if p.botUserID.IsValid() {
		return p.botUserID, nil
	}

	raw, err := c.Discord().APIRequest(c, provider.DiscordAPIRequest{
		Method: http.MethodGet,
		Path:   "users/@me",
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get bot user: %w", err)
	}

	var user discord.User
	if err := json.Unmarshal(raw, &user); err != nil {
		return 0, fmt.Errorf("failed to decode bot user: %w", err)
	}

	p.botUserID = user.ID
	return user.ID, nil
}

func getTicketsConfig(c plugin.Context, guildID discord.GuildID) (*TicketsConfig, error) {
	raw, err := c.GetValue(c, ticketsConfigKey(guildID.String()))
	if err != nil {
		return nil, err
	}
	if raw == thing.Null {
		return nil, nil
	}

	config := TicketsConfig{}
	if err := json.Unmarshal([]byte(raw.String()), &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func getTicket(c plugin.Context, channelID discord.ChannelID) (*Ticket, error) {
	raw, err := c.GetValue(c, ticketsChannelKey(channelID.String()))
	if err != nil {
		return nil, err
	}
	if raw == thing.Null {
		return nil, nil
	}

	ticket := Ticket{}
	if err := json.Unmarshal([]byte(raw.String()), &ticket); err != nil {
		return nil, err
	}

	return &ticket, nil
}

func cleanupTicket(c plugin.Context, guildID discord.GuildID, ownerID discord.UserID, channelID discord.ChannelID) error {
	if err := c.DeleteValue(c, ticketsChannelKey(channelID.String())); err != nil {
		return err
	}

	open, err := c.GetValue(c, ticketsOpenKey(guildID.String(), ownerID.String()))
	if err != nil {
		return err
	}

	// Only forget the open ticket of the owner if it's this one.
	if open != thing.Null && discord.ChannelID(open.Int()) == channelID {
		return c.DeleteValue(c, ticketsOpenKey(guildID.String(), ownerID.String()))
	}

	return nil
}

// logTicket posts to the log channel if there is one. A failure doesn't fail
// opening or closing the ticket.
func logTicket(c plugin.Context, config *TicketsConfig, embed discord.Embed) {
	if !config.LogChannelID.IsValid() {
		return
	}

	_, err := c.Discord().CreateMessage(c, config.LogChannelID, api.SendMessageData{
		Embeds: []discord.Embed{embed},
	})
	if err != nil {
		slog.Debug("Failed to log ticket", slog.String("error", err.Error()))
	}
}

func reply(c plugin.Context, event *gateway.InteractionCreateEvent, content string) error {
	_, err := c.Discord().CreateInteractionResponse(c, event.ID, event.Token, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content:         option.NewNullableString(content),
			Flags:           discord.EphemeralMessage,
			AllowedMentions: &api.AllowedMentions{},
		},
	})
	return err
}

func replyPublic(c plugin.Context, event *gateway.InteractionCreateEvent, content string) error {
	_, err := c.Discord().CreateInteractionResponse(c, event.ID, event.Token, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content:         option.NewNullableString(content),
			AllowedMentions: &api.AllowedMentions{},
		},
	})
	return err
}

func editReply(c plugin.Context, event *gateway.InteractionCreateEvent, content string) error {
	_, err := c.Discord().EditInteractionResponse(c, event.AppID, event.Token, api.EditInteractionResponseData{
		Content:         option.NewNullableString(content),
		AllowedMentions: &api.AllowedMentions{},
	})
	return err
}

func ticketsConfigKey(guildID string) string {
	return "tickets-config:" + guildID
}

func ticketsCounterKey(guildID string) string {
	return "tickets-counter:" + guildID
}

func ticketsOpenKey(guildID string, userID string) string {
	return "tickets-open:" + guildID + ":" + userID
}

func ticketsChannelKey(channelID string) string {
	return "tickets-channel:" + channelID
}
