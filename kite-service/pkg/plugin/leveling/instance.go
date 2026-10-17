package leveling

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"slices"
	"strings"
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
	defaultLevelUpMessage = "🎉 {user} reached level **{level}**!"

	levelingColor = 0xfee75c

	// configCacheTTL is how long the settings of a server are kept in memory,
	// so not every message reads them from the database.
	configCacheTTL = time.Minute

	// maxCooldowns bounds the memory of the cooldowns. Expired ones are
	// dropped when it's reached.
	maxCooldowns = 50_000
)

type LevelingPluginInstance struct {
	appID  string
	config plugin.ConfigValues

	// mu guards the cooldowns and the leaderboard updates, so two messages of
	// the same server don't overwrite each other's leaderboard.
	mu        sync.Mutex
	cooldowns map[string]int64

	configMu    sync.Mutex
	configCache map[discord.GuildID]cachedConfig
}

type LevelingConfig struct {
	Enabled   bool                   `json:"enabled"`
	ChannelID discord.ChannelID      `json:"channel_id,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Rewards   map[int]discord.RoleID `json:"rewards,omitempty"`
}

type cachedConfig struct {
	config    *LevelingConfig
	expiresAt time.Time
}

func (p *LevelingPluginInstance) Update(ctx context.Context, config plugin.ConfigValues) error {
	p.config = config
	return nil
}

func (p *LevelingPluginInstance) HandleEvent(c plugin.Context, event gateway.Event) error {
	e, ok := event.(*gateway.MessageCreateEvent)
	if !ok {
		return nil
	}

	if e.GuildID == 0 || e.Author.Bot || e.WebhookID.IsValid() {
		return nil
	}

	config, err := p.getConfig(c, e.GuildID)
	if err != nil {
		return err
	}
	if config == nil || !config.Enabled {
		return nil
	}

	if !p.takeCooldown(e.GuildID, e.Author.ID) {
		return nil
	}

	gained := int64(minMessageXP + rand.IntN(maxMessageXP-minMessageXP+1))

	newXP, err := c.UpdateValue(
		c,
		levelingXPKey(e.GuildID.String(), e.Author.ID.String()),
		provider.VariableOperationIncrement,
		thing.NewInt(gained),
	)
	if err != nil {
		return err
	}

	if err := p.updateLeaderboard(c, e.GuildID, e.Author.ID, newXP.Int()); err != nil {
		return err
	}

	oldLevel := levelForXP(newXP.Int() - gained)
	newLevel := levelForXP(newXP.Int())
	if newLevel <= oldLevel {
		return nil
	}

	for level := oldLevel + 1; level <= newLevel; level++ {
		roleID, ok := config.Rewards[level]
		if !ok {
			continue
		}

		_, err := c.Discord().APIRequest(c, provider.DiscordAPIRequest{
			Method: http.MethodPut,
			Path:   fmt.Sprintf("guilds/%s/members/%s/roles/%s", e.GuildID, e.Author.ID, roleID),
			Reason: api.AuditLogReason(fmt.Sprintf("Reached level %d", level)),
		})
		if err != nil {
			// A missing permission for one reward shouldn't stop the
			// announcement or the other rewards.
			slog.Debug("Failed to give leveling reward", slog.String("error", err.Error()))
		}
	}

	channelID := e.ChannelID
	if config.ChannelID.IsValid() {
		channelID = config.ChannelID
	}

	message := config.Message
	if message == "" {
		message = defaultLevelUpMessage
	}
	message = strings.NewReplacer(
		"{user}", e.Author.Mention(),
		"{level}", fmt.Sprint(newLevel),
	).Replace(message)

	_, err = c.Discord().CreateMessage(c, channelID, api.SendMessageData{
		Content: message,
		AllowedMentions: &api.AllowedMentions{
			Users: []discord.UserID{e.Author.ID},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send level up message: %w", err)
	}

	return nil
}

func (p *LevelingPluginInstance) HandleCommand(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	data, ok := event.Data.(*discord.CommandInteraction)
	if !ok {
		return nil
	}

	if event.GuildID == 0 {
		return reply(c, event, "Leveling only works inside a server.", true)
	}

	switch data.Name {
	case "rank":
		return p.handleRank(c, event, data)
	case "leaderboard":
		return p.handleLeaderboard(c, event)
	case "leveling":
		for _, opt := range data.Options {
			switch opt.Name {
			case "enable":
				return p.handleEnable(c, event, opt.Options)
			case "disable":
				return p.handleDisable(c, event)
			case "reward-add":
				return p.handleRewardAdd(c, event, opt.Options)
			case "reward-remove":
				return p.handleRewardRemove(c, event, opt.Options)
			case "reset":
				return p.handleReset(c, event, opt.Options)
			}
		}
	}

	return nil
}

func (p *LevelingPluginInstance) HandleComponent(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	return nil
}

func (p *LevelingPluginInstance) HandleModal(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	return nil
}

func (p *LevelingPluginInstance) Close() error {
	return nil
}

func (p *LevelingPluginInstance) handleRank(c plugin.Context, event *gateway.InteractionCreateEvent, data *discord.CommandInteraction) error {
	sender := event.Sender()
	if sender == nil {
		return nil
	}

	user := *sender
	for _, opt := range data.Options {
		if opt.Name != "user" {
			continue
		}

		var userID discord.UserID
		_ = opt.Value.UnmarshalTo(&userID)
		if resolved, ok := data.Resolved.Users[userID]; ok {
			user = resolved
		}
	}

	if user.Bot {
		return reply(c, event, "Bots don't collect XP.", true)
	}

	xpValue, err := c.GetValue(c, levelingXPKey(event.GuildID.String(), user.ID.String()))
	if err != nil {
		return err
	}

	xp := xpValue.Int()
	level := levelForXP(xp)
	current := xp - xpForLevel(level)
	needed := xpForNextLevel(level)

	fields := []discord.EmbedField{
		{Name: "Level", Value: fmt.Sprint(level), Inline: true},
		{Name: "XP", Value: fmt.Sprint(xp), Inline: true},
	}

	leaderboard, err := getLeaderboard(c, event.GuildID)
	if err != nil {
		return err
	}
	for i, entry := range leaderboard {
		if entry.UserID == user.ID.String() {
			fields = append(fields, discord.EmbedField{Name: "Rank", Value: fmt.Sprintf("#%d", i+1), Inline: true})
			break
		}
	}

	description := fmt.Sprintf("%s\n`%d / %d XP` to level %d", progressBar(current, needed, 12), current, needed, level+1)
	if level >= maxLevel {
		description = "Reached the highest level!"
	}

	return replyEmbed(c, event, discord.Embed{
		Author: &discord.EmbedAuthor{
			Name: user.DisplayOrUsername(),
			Icon: user.AvatarURL(),
		},
		Description: description,
		Fields:      fields,
		Color:       levelingColor,
	})
}

func (p *LevelingPluginInstance) handleLeaderboard(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	leaderboard, err := getLeaderboard(c, event.GuildID)
	if err != nil {
		return err
	}

	if len(leaderboard) == 0 {
		return reply(c, event, "Nobody has collected any XP yet.", false)
	}

	medals := []string{"🥇", "🥈", "🥉"}

	var b strings.Builder
	for i, entry := range leaderboard {
		position := fmt.Sprintf("`#%d`", i+1)
		if i < len(medals) {
			position = medals[i]
		}

		fmt.Fprintf(&b, "%s <@%s> • Level %d • %d XP\n", position, entry.UserID, levelForXP(entry.XP), entry.XP)
	}

	return replyEmbed(c, event, discord.Embed{
		Title:       "🏆 Leaderboard",
		Description: b.String(),
		Color:       levelingColor,
	})
}

func (p *LevelingPluginInstance) handleEnable(c plugin.Context, event *gateway.InteractionCreateEvent, options discord.CommandInteractionOptions) error {
	return p.updateConfig(c, event, func(config *LevelingConfig) string {
		config.Enabled = true
		config.ChannelID = 0
		config.Message = ""

		for _, opt := range options {
			switch opt.Name {
			case "channel":
				_ = opt.Value.UnmarshalTo(&config.ChannelID)
			case "message":
				_ = opt.Value.UnmarshalTo(&config.Message)
			}
		}

		where := "in the channel of the message"
		if config.ChannelID.IsValid() {
			where = fmt.Sprintf("in <#%d>", config.ChannelID)
		}

		return fmt.Sprintf("Leveling is enabled. Members get XP for chatting and new levels are announced %s.", where)
	})
}

func (p *LevelingPluginInstance) handleDisable(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	return p.updateConfig(c, event, func(config *LevelingConfig) string {
		config.Enabled = false
		return "Leveling is disabled. Members keep their XP and get no new XP until you enable it again."
	})
}

func (p *LevelingPluginInstance) handleRewardAdd(c plugin.Context, event *gateway.InteractionCreateEvent, options discord.CommandInteractionOptions) error {
	var level int
	var roleID discord.RoleID
	for _, opt := range options {
		switch opt.Name {
		case "level":
			_ = opt.Value.UnmarshalTo(&level)
		case "role":
			_ = opt.Value.UnmarshalTo(&roleID)
		}
	}

	if level < 1 || level > maxLevel || !roleID.IsValid() {
		return reply(c, event, fmt.Sprintf("Pick a level between 1 and %d and a role.", maxLevel), true)
	}

	return p.updateConfig(c, event, func(config *LevelingConfig) string {
		if config.Rewards == nil {
			config.Rewards = make(map[int]discord.RoleID)
		}
		config.Rewards[level] = roleID

		return fmt.Sprintf(
			"Members now get %s when they reach level %d. The bot needs the `Manage Roles` permission and its role has to be above that role.",
			roleID.Mention(), level,
		)
	})
}

func (p *LevelingPluginInstance) handleRewardRemove(c plugin.Context, event *gateway.InteractionCreateEvent, options discord.CommandInteractionOptions) error {
	var level int
	for _, opt := range options {
		if opt.Name == "level" {
			_ = opt.Value.UnmarshalTo(&level)
		}
	}

	return p.updateConfig(c, event, func(config *LevelingConfig) string {
		if _, ok := config.Rewards[level]; !ok {
			return fmt.Sprintf("There is no reward for level %d.", level)
		}

		delete(config.Rewards, level)
		return fmt.Sprintf("Removed the reward for level %d. Members who already got the role keep it.", level)
	})
}

func (p *LevelingPluginInstance) handleReset(c plugin.Context, event *gateway.InteractionCreateEvent, options discord.CommandInteractionOptions) error {
	var userID discord.UserID
	for _, opt := range options {
		if opt.Name == "user" {
			_ = opt.Value.UnmarshalTo(&userID)
		}
	}
	if !userID.IsValid() {
		return reply(c, event, "Pick a member.", true)
	}

	err := c.DeleteValue(c, levelingXPKey(event.GuildID.String(), userID.String()))
	if err != nil {
		return err
	}

	if err := p.updateLeaderboard(c, event.GuildID, userID, 0); err != nil {
		return err
	}

	return reply(c, event, fmt.Sprintf("Reset the XP of <@%d>.", userID), true)
}

// updateConfig loads the settings of the server, lets change modify them and
// saves them. change returns the response for the user.
func (p *LevelingPluginInstance) updateConfig(
	c plugin.Context,
	event *gateway.InteractionCreateEvent,
	change func(config *LevelingConfig) string,
) error {
	p.configMu.Lock()
	defer p.configMu.Unlock()

	config, err := loadConfig(c, event.GuildID)
	if err != nil {
		return err
	}
	if config == nil {
		config = &LevelingConfig{}
	}

	response := change(config)

	raw, err := json.Marshal(config)
	if err != nil {
		return err
	}

	_, err = c.UpdateValue(c, levelingConfigKey(event.GuildID.String()), provider.VariableOperationOverwrite, thing.NewString(string(raw)))
	if err != nil {
		return err
	}

	if p.configCache == nil {
		p.configCache = make(map[discord.GuildID]cachedConfig)
	}
	p.configCache[event.GuildID] = cachedConfig{config: config, expiresAt: time.Now().Add(configCacheTTL)}

	return reply(c, event, response, true)
}

// getConfig returns the settings of a server from memory, or from the
// database if they aren't cached or are too old.
func (p *LevelingPluginInstance) getConfig(c plugin.Context, guildID discord.GuildID) (*LevelingConfig, error) {
	p.configMu.Lock()
	defer p.configMu.Unlock()

	if cached, ok := p.configCache[guildID]; ok && time.Now().Before(cached.expiresAt) {
		return cached.config, nil
	}

	config, err := loadConfig(c, guildID)
	if err != nil {
		return nil, err
	}

	if p.configCache == nil {
		p.configCache = make(map[discord.GuildID]cachedConfig)
	}
	p.configCache[guildID] = cachedConfig{config: config, expiresAt: time.Now().Add(configCacheTTL)}

	return config, nil
}

// takeCooldown reports whether the member can get XP again and starts their
// cooldown if so.
func (p *LevelingPluginInstance) takeCooldown(guildID discord.GuildID, userID discord.UserID) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now().Unix()
	key := guildID.String() + ":" + userID.String()

	if last, ok := p.cooldowns[key]; ok && now-last < cooldownSeconds {
		return false
	}

	if len(p.cooldowns) >= maxCooldowns {
		for k, last := range p.cooldowns {
			if now-last >= cooldownSeconds {
				delete(p.cooldowns, k)
			}
		}
	}

	p.cooldowns[key] = now
	return true
}

func (p *LevelingPluginInstance) updateLeaderboard(c plugin.Context, guildID discord.GuildID, userID discord.UserID, xp int64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	entries, err := getLeaderboard(c, guildID)
	if err != nil {
		return err
	}

	// Skip the write if the member isn't on the leaderboard and doesn't make
	// it on, which is the case for most messages in a busy server.
	onBoard := slices.ContainsFunc(entries, func(e leaderboardEntry) bool { return e.UserID == userID.String() })
	if !onBoard && len(entries) >= leaderboardSize && xp <= entries[len(entries)-1].XP {
		return nil
	}

	raw, err := json.Marshal(updateLeaderboard(entries, userID.String(), xp))
	if err != nil {
		return err
	}

	_, err = c.UpdateValue(c, levelingLeaderboardKey(guildID.String()), provider.VariableOperationOverwrite, thing.NewString(string(raw)))
	return err
}

func getLeaderboard(c plugin.Context, guildID discord.GuildID) ([]leaderboardEntry, error) {
	raw, err := c.GetValue(c, levelingLeaderboardKey(guildID.String()))
	if err != nil {
		return nil, err
	}
	if raw == thing.Null {
		return nil, nil
	}

	var entries []leaderboardEntry
	if err := json.Unmarshal([]byte(raw.String()), &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func loadConfig(c plugin.Context, guildID discord.GuildID) (*LevelingConfig, error) {
	raw, err := c.GetValue(c, levelingConfigKey(guildID.String()))
	if err != nil {
		return nil, err
	}
	if raw == thing.Null {
		return nil, nil
	}

	config := LevelingConfig{}
	if err := json.Unmarshal([]byte(raw.String()), &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func reply(c plugin.Context, event *gateway.InteractionCreateEvent, content string, ephemeral bool) error {
	var flags discord.MessageFlags
	if ephemeral {
		flags = discord.EphemeralMessage
	}

	_, err := c.Discord().CreateInteractionResponse(c, event.ID, event.Token, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content:         option.NewNullableString(content),
			Flags:           flags,
			AllowedMentions: &api.AllowedMentions{},
		},
	})
	return err
}

func replyEmbed(c plugin.Context, event *gateway.InteractionCreateEvent, embed discord.Embed) error {
	_, err := c.Discord().CreateInteractionResponse(c, event.ID, event.Token, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Embeds:          &[]discord.Embed{embed},
			AllowedMentions: &api.AllowedMentions{},
		},
	})
	return err
}

func levelingConfigKey(guildID string) string {
	return "leveling-config:" + guildID
}

func levelingXPKey(guildID string, userID string) string {
	return "leveling-xp:" + guildID + ":" + userID
}

func levelingLeaderboardKey(guildID string) string {
	return "leveling-leaderboard:" + guildID
}
