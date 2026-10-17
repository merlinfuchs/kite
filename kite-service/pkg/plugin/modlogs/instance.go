package modlogs

import (
	"context"
	"encoding/json"
	"fmt"
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
	colorMessageEdit   = 0xfee75c
	colorMessageDelete = 0xed4245
	colorMemberJoin    = 0x57f287
	colorMemberLeave   = 0x99aab5
	colorModeration    = 0xe67e22

	// configCacheTTL is how long the log channel of a server is kept in
	// memory, so not every message reads it from the database.
	configCacheTTL = time.Minute

	// auditLogDelay gives Discord time to write the audit log entry of an
	// action before the plugin looks for it.
	auditLogDelay = time.Second
	// auditLogMaxAge is how old an audit log entry can be to still belong to
	// the event that is being logged.
	auditLogMaxAge = 20 * time.Second

	// maxTimeouts bounds the remembered timeouts. Expired ones are dropped
	// when it's reached.
	maxTimeouts = 10_000

	// Audit log change key of the timeout of a member.
	auditKeyTimeout = "communication_disabled_until"
)

type ModlogsPluginInstance struct {
	appID  string
	config plugin.ConfigValues

	messages *messageCache

	configMu    sync.Mutex
	configCache map[discord.GuildID]cachedConfig

	// timeouts remembers until when members are timed out, so a member update
	// can tell that a timeout was added or removed.
	timeoutsMu sync.Mutex
	timeouts   map[string]time.Time
}

type cachedConfig struct {
	channelID discord.ChannelID
	expiresAt time.Time
}

func (p *ModlogsPluginInstance) Update(ctx context.Context, config plugin.ConfigValues) error {
	p.config = config
	return nil
}

func (p *ModlogsPluginInstance) HandleEvent(c plugin.Context, event gateway.Event) error {
	switch e := event.(type) {
	case *gateway.MessageCreateEvent:
		return p.handleMessageCreate(c, e)
	case *gateway.MessageUpdateEvent:
		return p.handleMessageUpdate(c, e)
	case *gateway.MessageDeleteEvent:
		return p.handleMessageDelete(c, e)
	case *gateway.GuildMemberAddEvent:
		return p.handleMemberAdd(c, e)
	case *gateway.GuildMemberRemoveEvent:
		return p.handleMemberRemove(c, e)
	case *gateway.GuildMemberUpdateEvent:
		return p.handleMemberUpdate(c, e)
	case *gateway.GuildBanAddEvent:
		return p.handleBan(c, e.GuildID, e.User, true)
	case *gateway.GuildBanRemoveEvent:
		return p.handleBan(c, e.GuildID, e.User, false)
	}

	return nil
}

func (p *ModlogsPluginInstance) HandleCommand(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	data, ok := event.Data.(*discord.CommandInteraction)
	if !ok || data.Name != "modlogs" {
		return nil
	}

	if event.GuildID == 0 {
		return reply(c, event, "Mod logs only work inside a server.")
	}

	for _, opt := range data.Options {
		switch opt.Name {
		case "enable":
			var channelID discord.ChannelID
			for _, subOpt := range opt.Options {
				if subOpt.Name == "channel" {
					_ = subOpt.Value.UnmarshalTo(&channelID)
				}
			}
			if !channelID.IsValid() {
				return reply(c, event, "Pick a channel to send the logs to.")
			}

			_, err := c.Discord().CreateMessage(c, channelID, api.SendMessageData{
				Embeds: []discord.Embed{
					{
						Title:       "Mod Logs Enabled",
						Description: "Deleted and edited messages, joins, leaves, kicks, bans and timeouts will be logged in this channel.",
						Color:       colorMemberJoin,
					},
				},
			})
			if err != nil {
				replyErr := reply(c, event, fmt.Sprintf("Failed to send a message in <#%d>. Make sure the bot can see and send messages in that channel.", channelID))
				if replyErr != nil {
					return replyErr
				}
				return fmt.Errorf("failed to send modlogs test message: %w", err)
			}

			_, err = c.UpdateValue(c, modlogsConfigKey(event.GuildID.String()), provider.VariableOperationOverwrite, thing.NewInt(int64(channelID)))
			if err != nil {
				return err
			}
			p.setCachedChannel(event.GuildID, channelID)

			return reply(c, event, fmt.Sprintf(
				"Mod logs are enabled in <#%d>. The bot needs the `View Audit Log` permission to show who kicked, banned or timed out a member.",
				channelID,
			))
		case "disable":
			err := c.DeleteValue(c, modlogsConfigKey(event.GuildID.String()))
			if err != nil {
				return err
			}
			p.setCachedChannel(event.GuildID, 0)

			return reply(c, event, "Mod logs are disabled for this server.")
		}
	}

	return nil
}

func (p *ModlogsPluginInstance) HandleComponent(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	return nil
}

func (p *ModlogsPluginInstance) HandleModal(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	return nil
}

func (p *ModlogsPluginInstance) Close() error {
	return nil
}

func (p *ModlogsPluginInstance) handleMessageCreate(c plugin.Context, e *gateway.MessageCreateEvent) error {
	if e.GuildID == 0 {
		return nil
	}

	logChannelID, err := p.getLogChannel(c, e.GuildID)
	if err != nil {
		return err
	}
	if !logChannelID.IsValid() || logChannelID == e.ChannelID {
		return nil
	}

	p.messages.Set(e.ID, cachedMessageFrom(e.Message))
	return nil
}

func (p *ModlogsPluginInstance) handleMessageUpdate(c plugin.Context, e *gateway.MessageUpdateEvent) error {
	if e.GuildID == 0 || e.Author.Bot || e.WebhookID.IsValid() {
		return nil
	}

	logChannelID, err := p.getLogChannel(c, e.GuildID)
	if err != nil {
		return err
	}
	if !logChannelID.IsValid() || logChannelID == e.ChannelID {
		return nil
	}

	before, cached := p.messages.Get(e.ID)
	p.messages.Set(e.ID, cachedMessageFrom(e.Message))

	// Discord also sends updates when it adds link previews, those don't
	// change the content.
	if cached && before.Content == e.Content {
		return nil
	}
	if !cached && !e.EditedTimestamp.IsValid() {
		return nil
	}

	beforeContent := "*Not cached, the message is older than the bot's memory.*"
	if cached {
		beforeContent = orEmpty(before.Content)
	}

	return sendLog(c, logChannelID, discord.Embed{
		Author:      authorOf(e.Author),
		Title:       "Message Edited",
		Description: fmt.Sprintf("In <#%d> • [Jump to message](%s)", e.ChannelID, e.URL()),
		Fields: []discord.EmbedField{
			{Name: "Before", Value: truncate(beforeContent, 1024)},
			{Name: "After", Value: truncate(orEmpty(e.Content), 1024)},
		},
		Footer:    &discord.EmbedFooter{Text: "User ID: " + e.Author.ID.String()},
		Color:     colorMessageEdit,
		Timestamp: discord.NewTimestamp(time.Now()),
	})
}

func (p *ModlogsPluginInstance) handleMessageDelete(c plugin.Context, e *gateway.MessageDeleteEvent) error {
	if e.GuildID == 0 {
		return nil
	}

	msg, cached := p.messages.Delete(e.ID)

	logChannelID, err := p.getLogChannel(c, e.GuildID)
	if err != nil {
		return err
	}
	if !logChannelID.IsValid() || logChannelID == e.ChannelID {
		return nil
	}

	if !cached {
		return sendLog(c, logChannelID, discord.Embed{
			Title:       "Message Deleted",
			Description: fmt.Sprintf("A message was deleted in <#%d>. It's older than the bot's memory, so its content is unknown.", e.ChannelID),
			Footer:      &discord.EmbedFooter{Text: "Message ID: " + e.ID.String()},
			Color:       colorMessageDelete,
			Timestamp:   discord.NewTimestamp(time.Now()),
		})
	}

	if msg.Author.Bot {
		return nil
	}

	description := fmt.Sprintf("Sent by %s in <#%d>\n\n%s", msg.Author.Mention(), e.ChannelID, truncate(orEmpty(msg.Content), 3500))
	if msg.Attachments > 0 {
		description += fmt.Sprintf("\n\n📎 %d attachment(s)", msg.Attachments)
	}

	return sendLog(c, logChannelID, discord.Embed{
		Author:      authorOf(msg.Author),
		Title:       "Message Deleted",
		Description: description,
		Footer:      &discord.EmbedFooter{Text: "User ID: " + msg.Author.ID.String()},
		Color:       colorMessageDelete,
		Timestamp:   discord.NewTimestamp(time.Now()),
	})
}

func (p *ModlogsPluginInstance) handleMemberAdd(c plugin.Context, e *gateway.GuildMemberAddEvent) error {
	logChannelID, err := p.getLogChannel(c, e.GuildID)
	if err != nil || !logChannelID.IsValid() {
		return err
	}

	user := e.Member.User
	return sendLog(c, logChannelID, discord.Embed{
		Author:      authorOf(user),
		Title:       "Member Joined",
		Description: fmt.Sprintf("%s joined the server.\nAccount created <t:%d:R>", user.Mention(), user.ID.Time().Unix()),
		Thumbnail:   &discord.EmbedThumbnail{URL: user.AvatarURL()},
		Footer:      &discord.EmbedFooter{Text: "User ID: " + user.ID.String()},
		Color:       colorMemberJoin,
		Timestamp:   discord.NewTimestamp(time.Now()),
	})
}

func (p *ModlogsPluginInstance) handleMemberRemove(c plugin.Context, e *gateway.GuildMemberRemoveEvent) error {
	logChannelID, err := p.getLogChannel(c, e.GuildID)
	if err != nil || !logChannelID.IsValid() {
		return err
	}

	time.Sleep(auditLogDelay)

	// Leaving, being kicked and being banned all look the same here. Bans have
	// their own event, so they're skipped, and kicks are found in the audit log.
	entry := findAuditLogEntry(c, e.GuildID, discord.Snowflake(e.User.ID), 0, func(entry discord.AuditLogEntry) bool {
		return entry.ActionType == discord.MemberKick || entry.ActionType == discord.MemberBanAdd
	})
	if entry != nil && entry.ActionType == discord.MemberBanAdd {
		return nil
	}

	if entry != nil {
		return sendLog(c, logChannelID, moderationEmbed("Member Kicked", e.User, entry))
	}

	return sendLog(c, logChannelID, discord.Embed{
		Author:      authorOf(e.User),
		Title:       "Member Left",
		Description: fmt.Sprintf("%s left the server.", e.User.Mention()),
		Footer:      &discord.EmbedFooter{Text: "User ID: " + e.User.ID.String()},
		Color:       colorMemberLeave,
		Timestamp:   discord.NewTimestamp(time.Now()),
	})
}

func (p *ModlogsPluginInstance) handleMemberUpdate(c plugin.Context, e *gateway.GuildMemberUpdateEvent) error {
	now := time.Now()
	key := e.GuildID.String() + ":" + e.User.ID.String()

	until := e.CommunicationDisabledUntil.Time()
	timedOut := e.CommunicationDisabledUntil.IsValid() && until.After(now)

	p.timeoutsMu.Lock()
	if p.timeouts == nil {
		p.timeouts = make(map[string]time.Time)
	}
	prev, hadTimeout := p.timeouts[key]
	if timedOut {
		if len(p.timeouts) >= maxTimeouts {
			for k, t := range p.timeouts {
				if t.Before(now) {
					delete(p.timeouts, k)
				}
			}
		}
		p.timeouts[key] = until
	} else {
		delete(p.timeouts, key)
	}
	p.timeoutsMu.Unlock()

	added := timedOut && (!hadTimeout || !prev.Equal(until))
	// A timeout that ran out on its own isn't a moderation action.
	removed := !timedOut && hadTimeout && prev.After(now)
	if !added && !removed {
		return nil
	}

	logChannelID, err := p.getLogChannel(c, e.GuildID)
	if err != nil || !logChannelID.IsValid() {
		return err
	}

	time.Sleep(auditLogDelay)

	entry := findAuditLogEntry(c, e.GuildID, discord.Snowflake(e.User.ID), discord.MemberUpdate, func(entry discord.AuditLogEntry) bool {
		for _, change := range entry.Changes {
			if string(change.Key) == auditKeyTimeout {
				return true
			}
		}
		return false
	})

	if added {
		// Without the audit log entry this update could also be another
		// change of a member that was already timed out before the bot knew.
		if entry == nil {
			return nil
		}

		embed := moderationEmbed("Member Timed Out", e.User, entry)
		embed.Fields = append(embed.Fields, discord.EmbedField{
			Name:   "Until",
			Value:  fmt.Sprintf("<t:%d:f> (<t:%d:R>)", until.Unix(), until.Unix()),
			Inline: true,
		})
		return sendLog(c, logChannelID, embed)
	}

	if entry != nil {
		return sendLog(c, logChannelID, moderationEmbed("Timeout Removed", e.User, entry))
	}

	return sendLog(c, logChannelID, discord.Embed{
		Author:      authorOf(e.User),
		Title:       "Timeout Removed",
		Description: fmt.Sprintf("The timeout of %s was removed.", e.User.Mention()),
		Footer:      &discord.EmbedFooter{Text: "User ID: " + e.User.ID.String()},
		Color:       colorModeration,
		Timestamp:   discord.NewTimestamp(time.Now()),
	})
}

func (p *ModlogsPluginInstance) handleBan(c plugin.Context, guildID discord.GuildID, user discord.User, banned bool) error {
	logChannelID, err := p.getLogChannel(c, guildID)
	if err != nil || !logChannelID.IsValid() {
		return err
	}

	time.Sleep(auditLogDelay)

	actionType := discord.MemberBanAdd
	title := "Member Banned"
	if !banned {
		actionType = discord.MemberBanRemove
		title = "Member Unbanned"
	}

	entry := findAuditLogEntry(c, guildID, discord.Snowflake(user.ID), actionType, nil)
	return sendLog(c, logChannelID, moderationEmbed(title, user, entry))
}

// getLogChannel returns the log channel of a server, or 0 if it has none.
func (p *ModlogsPluginInstance) getLogChannel(c plugin.Context, guildID discord.GuildID) (discord.ChannelID, error) {
	p.configMu.Lock()
	defer p.configMu.Unlock()

	if cached, ok := p.configCache[guildID]; ok && time.Now().Before(cached.expiresAt) {
		return cached.channelID, nil
	}

	value, err := c.GetValue(c, modlogsConfigKey(guildID.String()))
	if err != nil {
		return 0, err
	}

	channelID := discord.ChannelID(value.Int())
	p.setCachedChannelLocked(guildID, channelID)
	return channelID, nil
}

func (p *ModlogsPluginInstance) setCachedChannel(guildID discord.GuildID, channelID discord.ChannelID) {
	p.configMu.Lock()
	defer p.configMu.Unlock()
	p.setCachedChannelLocked(guildID, channelID)
}

func (p *ModlogsPluginInstance) setCachedChannelLocked(guildID discord.GuildID, channelID discord.ChannelID) {
	if p.configCache == nil {
		p.configCache = make(map[discord.GuildID]cachedConfig)
	}
	p.configCache[guildID] = cachedConfig{channelID: channelID, expiresAt: time.Now().Add(configCacheTTL)}
}

// findAuditLogEntry returns the most recent audit log entry about the target
// that matches, or nil if there is none or the bot can't read the audit log.
// An actionType of 0 looks at entries of every type.
func findAuditLogEntry(
	c plugin.Context,
	guildID discord.GuildID,
	targetID discord.Snowflake,
	actionType discord.AuditLogEvent,
	match func(entry discord.AuditLogEntry) bool,
) *discord.AuditLogEntry {
	path := fmt.Sprintf("guilds/%s/audit-logs?limit=10", guildID)
	if actionType != 0 {
		path += fmt.Sprintf("&action_type=%d", actionType)
	}

	raw, err := c.Discord().APIRequest(c, provider.DiscordAPIRequest{
		Method: http.MethodGet,
		Path:   path,
	})
	if err != nil {
		return nil
	}

	var auditLog discord.AuditLog
	if err := json.Unmarshal(raw, &auditLog); err != nil {
		return nil
	}

	for _, entry := range auditLog.Entries {
		if entry.TargetID != targetID {
			continue
		}
		if time.Since(entry.CreatedAt()) > auditLogMaxAge {
			continue
		}
		if match != nil && !match(entry) {
			continue
		}
		return &entry
	}

	return nil
}

func moderationEmbed(title string, target discord.User, entry *discord.AuditLogEntry) discord.Embed {
	moderator := "Unknown"
	reason := "No reason given"
	if entry != nil {
		if entry.UserID.IsValid() {
			moderator = entry.UserID.Mention()
		}
		if entry.Reason != "" {
			reason = truncate(entry.Reason, 1024)
		}
	}

	return discord.Embed{
		Author:      authorOf(target),
		Title:       title,
		Description: fmt.Sprintf("%s (%s)", target.Mention(), target.Tag()),
		Fields: []discord.EmbedField{
			{Name: "Moderator", Value: moderator, Inline: true},
			{Name: "Reason", Value: reason, Inline: true},
		},
		Footer:    &discord.EmbedFooter{Text: "User ID: " + target.ID.String()},
		Color:     colorModeration,
		Timestamp: discord.NewTimestamp(time.Now()),
	}
}

func sendLog(c plugin.Context, channelID discord.ChannelID, embed discord.Embed) error {
	_, err := c.Discord().CreateMessage(c, channelID, api.SendMessageData{
		Embeds:          []discord.Embed{embed},
		AllowedMentions: &api.AllowedMentions{},
	})
	if err != nil {
		return fmt.Errorf("failed to send mod log: %w", err)
	}
	return nil
}

func cachedMessageFrom(msg discord.Message) cachedMessage {
	return cachedMessage{
		GuildID:     msg.GuildID,
		ChannelID:   msg.ChannelID,
		Author:      msg.Author,
		Content:     msg.Content,
		Attachments: len(msg.Attachments),
	}
}

func authorOf(user discord.User) *discord.EmbedAuthor {
	return &discord.EmbedAuthor{
		Name: user.Tag(),
		Icon: user.AvatarURL(),
	}
}

func orEmpty(content string) string {
	if content == "" {
		return "*No text*"
	}
	return content
}

// truncate cuts text to at most limit characters, embeds reject longer fields.
func truncate(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit-1]) + "…"
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

func modlogsConfigKey(guildID string) string {
	return "modlogs-config:" + guildID
}
