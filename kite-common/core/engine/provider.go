package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-common/core/flow"
	"github.com/kitecloud/kite/kite-common/model"
	"github.com/kitecloud/kite/kite-common/store"
)

type DiscordProvider struct {
	appID    string
	appStore store.AppStore
}

func NewDiscordProvider(appID string, appStore store.AppStore) *DiscordProvider {
	return &DiscordProvider{
		appID:    appID,
		appStore: appStore,
	}
}

func (p *DiscordProvider) appCredentials(ctx context.Context) (*model.AppCredentials, error) {
	cred, err := p.appStore.AppCredentials(ctx, p.appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app credentials: %w", err)
	}
	return cred, nil
}

func (p *DiscordProvider) clientWithCredentials(ctx context.Context) (*api.Client, error) {
	cred, err := p.appCredentials(ctx)
	if err != nil {
		return nil, err
	}

	return api.NewClient("Bot " + cred.DiscordToken), nil
}

func (p *DiscordProvider) CreateInteractionResponse(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse) error {
	client := api.NewClient("").WithContext(ctx)

	err := client.RespondInteraction(interactionID, interactionToken, response)
	if err != nil {
		return fmt.Errorf("failed to respond to interaction: %w", err)
	}

	return nil
}

func (p *DiscordProvider) CreateMessage(ctx context.Context, channelID discord.ChannelID, message api.SendMessageData) (*discord.Message, error) {
	client, err := p.clientWithCredentials(ctx)
	if err != nil {
		return nil, err
	}

	msg, err := client.SendMessageComplex(channelID, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return msg, nil
}

type LogProvider struct {
	appID    string
	logStore store.LogStore
}

func NewLogProvider(appID string, logStore store.LogStore) *LogProvider {
	return &LogProvider{
		appID:    appID,
		logStore: logStore,
	}
}

func (p *LogProvider) CreateLogEntry(ctx context.Context, level flow.LogLevel, message string) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err := p.logStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:     p.appID,
		Level:     model.LogLevel(level),
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		slog.With("error", err).With("app_id", p.appID).Error("Failed to create log entry")
	}
}
