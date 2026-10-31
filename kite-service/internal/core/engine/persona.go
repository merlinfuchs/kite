package engine

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"golang.org/x/sync/singleflight"
)

// personaWebhookName is the name of the webhook Kite creates in a channel to
// send persona messages. It doesn't show, as every message sets its own name.
const personaWebhookName = "Kite"

// guildMediaChannel is Discord's media channel type, which arikawa doesn't
// have.
const guildMediaChannel discord.ChannelType = 16

type channelWebhook struct {
	ID    discord.WebhookID
	Token string
}

// WebhookCache remembers the webhook of each app in a channel, so sending a
// persona message usually needs no request besides the message itself.
type WebhookCache struct {
	mu       sync.Mutex
	webhooks map[string]channelWebhook
	// creating makes runs that send to a channel at once share one lookup,
	// instead of each creating a webhook.
	creating singleflight.Group
}

func NewWebhookCache() *WebhookCache {
	return &WebhookCache{webhooks: make(map[string]channelWebhook)}
}

func webhookCacheKey(appID string, channelID discord.ChannelID) string {
	return appID + ":" + channelID.String()
}

func (c *WebhookCache) get(appID string, channelID discord.ChannelID) (channelWebhook, bool) {
	if c == nil {
		return channelWebhook{}, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	w, ok := c.webhooks[webhookCacheKey(appID, channelID)]
	return w, ok
}

func (c *WebhookCache) set(appID string, channelID discord.ChannelID, w channelWebhook) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.webhooks[webhookCacheKey(appID, channelID)] = w
}

func (c *WebhookCache) lookup(appID string, channelID discord.ChannelID, fn func() (channelWebhook, error)) (channelWebhook, error) {
	if c == nil {
		return fn()
	}
	w, err, _ := c.creating.Do(webhookCacheKey(appID, channelID), func() (any, error) {
		return fn()
	})
	if err != nil {
		return channelWebhook{}, err
	}
	return w.(channelWebhook), nil
}

func (c *WebhookCache) delete(appID string, channelID discord.ChannelID) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.webhooks, webhookCacheKey(appID, channelID))
}

// personaMessageData is the body of a webhook execution. arikawa's
// webhook.ExecuteData has no flags, which components v2 messages need.
type personaMessageData struct {
	Content         string                     `json:"content,omitempty"`
	Username        string                     `json:"username"`
	AvatarURL       string                     `json:"avatar_url,omitempty"`
	TTS             bool                       `json:"tts,omitempty"`
	Embeds          []discord.Embed            `json:"embeds,omitempty"`
	Components      discord.TopLevelComponents `json:"components,omitempty"`
	AllowedMentions *api.AllowedMentions       `json:"allowed_mentions,omitempty"`
	Flags           discord.MessageFlags       `json:"flags,omitempty"`
}

func (p *DiscordProvider) CreatePersonaMessage(ctx context.Context, channelID discord.ChannelID, persona provider.Persona, message api.SendMessageData) (*discord.Message, error) {
	features, err := p.featureProvider.AppFeatures(ctx, p.appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app features: %w", err)
	}
	if !features.Personas {
		return nil, fmt.Errorf("sending messages with another name and avatar requires premium")
	}

	channel, err := p.session.Channel(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	// Threads have no webhooks of their own, messages to them go through the
	// webhook of their channel.
	webhookChannelID := channelID
	var threadID discord.ChannelID
	switch channel.Type {
	case discord.GuildPublicThread, discord.GuildPrivateThread, discord.GuildAnnouncementThread:
		webhookChannelID = channel.ParentID
		threadID = channelID
	case discord.GuildForum, guildMediaChannel:
		return nil, fmt.Errorf("messages can't be sent to a forum directly, pick one of its posts")
	}

	data := personaMessageData{
		Content:         message.Content,
		Username:        persona.Name,
		AvatarURL:       persona.AvatarURL,
		TTS:             message.TTS,
		Embeds:          message.Embeds,
		Components:      message.Components,
		AllowedMentions: message.AllowedMentions,
		Flags:           message.Flags,
	}

	// A cached webhook can have been deleted since, then a new one is made.
	for attempt := 0; ; attempt++ {
		webhook, err := p.channelWebhook(webhookChannelID)
		if err != nil {
			return nil, err
		}

		msg, err := p.executeWebhook(ctx, webhook, threadID, data)
		if err == nil {
			return msg, nil
		}

		if attempt == 0 && util.IsDiscordRestStatusCode(err, http.StatusNotFound) {
			p.webhooks.delete(p.appID, webhookChannelID)
			continue
		}
		// The URL of the error has the webhook's token.
		return nil, fmt.Errorf("failed to send message: %w", util.WithoutRequestURL(err))
	}
}

// channelWebhook returns a webhook the bot made in the channel, and makes one
// if there's none.
func (p *DiscordProvider) channelWebhook(channelID discord.ChannelID) (channelWebhook, error) {
	if w, ok := p.webhooks.get(p.appID, channelID); ok {
		return w, nil
	}
	return p.webhooks.lookup(p.appID, channelID, func() (channelWebhook, error) {
		return p.findOrCreateWebhook(channelID)
	})
}

func (p *DiscordProvider) findOrCreateWebhook(channelID discord.ChannelID) (channelWebhook, error) {
	me, err := p.session.Me()
	if err != nil {
		return channelWebhook{}, fmt.Errorf("failed to get bot user: %w", err)
	}

	webhooks, err := p.session.ChannelWebhooks(channelID)
	if err != nil {
		return channelWebhook{}, fmt.Errorf("failed to get webhooks of the channel, does the bot have the Manage Webhooks permission? %w", err)
	}

	var webhook channelWebhook
	for _, w := range webhooks {
		if w.Type == discord.IncomingWebhook && w.Token != "" && w.User != nil && w.User.ID == me.ID {
			webhook = channelWebhook{ID: w.ID, Token: w.Token}
			break
		}
	}

	if webhook.ID == 0 {
		w, err := p.session.CreateWebhook(channelID, api.CreateWebhookData{Name: personaWebhookName})
		if err != nil {
			return channelWebhook{}, fmt.Errorf("failed to create webhook: %w", err)
		}
		webhook = channelWebhook{ID: w.ID, Token: w.Token}
	}

	p.webhooks.set(p.appID, channelID, webhook)
	return webhook, nil
}

func (p *DiscordProvider) executeWebhook(ctx context.Context, webhook channelWebhook, threadID discord.ChannelID, data personaMessageData) (*discord.Message, error) {
	query := url.Values{}
	query.Set("wait", "true")
	// Webhooks of an app can send interactive components.
	query.Set("with_components", "true")
	if threadID.IsValid() {
		query.Set("thread_id", threadID.String())
	}

	var msg discord.Message
	err := p.session.Client.WithContext(ctx).RequestJSON(
		&msg,
		"POST",
		api.EndpointWebhooks+webhook.ID.String()+"/"+url.PathEscape(webhook.Token)+"?"+query.Encode(),
		httputil.WithJSONBody(data),
	)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
