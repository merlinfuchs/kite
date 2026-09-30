package bot

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-support/internal/config"
	"github.com/kitecloud/kite/kite-support/internal/index"
	"github.com/kitecloud/kite/kite-support/internal/llm"
)

type Bot struct {
	state           *state.State
	llm             *llm.Client
	embed           *index.Embedder
	store           *index.Store
	cfg             *config.Config
	limiter         *rateLimiter
	feedbackLimiter *rateLimiter
	helpLimiter     *rateLimiter
	appID           discord.AppID
	feedback        *feedbackCache
	feedbackC       discord.ChannelID
	helpC           discord.ChannelID
	feedbackRoleID  discord.RoleID
	helpRoleID      discord.RoleID
}

func New(cfg *config.Config, store *index.Store, embedder *index.Embedder, llmClient *llm.Client) (*Bot, error) {
	if cfg.Discord.Token == "" {
		return nil, fmt.Errorf("discord token is empty")
	}
	if cfg.Discord.AppID == "" {
		return nil, fmt.Errorf("discord app_id is empty")
	}

	appIDSnow, err := discord.ParseSnowflake(cfg.Discord.AppID)
	if err != nil {
		return nil, fmt.Errorf("parse app_id: %w", err)
	}

	s := state.New("Bot " + cfg.Discord.Token)
	s.AddIntents(gateway.IntentGuilds)

	window, err := time.ParseDuration(cfg.RateLimit.PerUserWindow)
	if err != nil || window <= 0 {
		window = 10 * time.Second
	}
	max := cfg.RateLimit.PerUserMax
	if max <= 0 {
		max = 2
	}

	var feedbackChannel discord.ChannelID
	if cfg.Feedback.ChannelID != "" {
		fid, err := discord.ParseSnowflake(cfg.Feedback.ChannelID)
		if err != nil {
			return nil, fmt.Errorf("parse feedback channel_id: %w", err)
		}
		feedbackChannel = discord.ChannelID(fid)
	}

	var helpChannel discord.ChannelID
	if cfg.Help.ChannelID != "" {
		hid, err := discord.ParseSnowflake(cfg.Help.ChannelID)
		if err != nil {
			return nil, fmt.Errorf("parse help channel_id: %w", err)
		}
		helpChannel = discord.ChannelID(hid)
	}

	b := &Bot{
		state:           s,
		llm:             llmClient,
		embed:           embedder,
		store:           store,
		cfg:             cfg,
		limiter:         newRateLimiter(max, window),
		feedbackLimiter: newRateLimiter(1, 10*time.Minute),
		helpLimiter:     newRateLimiter(1, 10*time.Minute),
		appID:           discord.AppID(appIDSnow),
		feedback:        newFeedbackCache(),
		feedbackC:       feedbackChannel,
		helpC:           helpChannel,
		feedbackRoleID:  parseRoleID(cfg.Feedback.RoleID),
		helpRoleID:      parseRoleID(cfg.Help.RoleID),
	}

	s.AddHandler(b.onInteraction)
	return b, nil
}

func parseRoleID(s string) discord.RoleID {
	if s == "" {
		return 0
	}
	id, err := discord.ParseSnowflake(s)
	if err != nil {
		return 0
	}
	return discord.RoleID(id)
}

func (b *Bot) Run(ctx context.Context) error {
	if err := b.state.Open(ctx); err != nil {
		return fmt.Errorf("open gateway: %w", err)
	}
	defer b.state.Close()

	if err := b.registerCommands(); err != nil {
		return fmt.Errorf("register commands: %w", err)
	}

	slog.Info("kite-support bot online",
		"app_id", b.appID,
		"feedback_channel", b.feedbackC,
		"help_channel", b.helpC,
	)
	<-ctx.Done()
	return nil
}

func (b *Bot) registerCommands() error {
	manageGuild := discord.PermissionManageGuild
	cmds := []api.CreateCommandData{
		{
			Name:        "ask",
			Description: "Ask the Kite support bot a question.",
			Options: discord.CommandOptions{
				&discord.StringOption{
					OptionName:  "question",
					Description: "What do you want to know about Kite?",
					Required:    true,
				},
			},
		},
		{
			Name:                     "setup-menu",
			Description:              "Post a help menu in this channel.",
			DefaultMemberPermissions: &manageGuild,
		},
	}
	res, err := b.state.BulkOverwriteCommands(b.appID, cmds)
	if err != nil {
		return err
	}
	slog.Info("commands registered", "count", len(res))
	return nil
}
