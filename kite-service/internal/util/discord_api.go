package util

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// DiscordAPIRequest sends a request to the Discord API with the client of an
// app's session, which adds the bot token and shares its rate limiter with
// every other request of the app.
func DiscordAPIRequest(ctx context.Context, client *api.Client, req provider.DiscordAPIRequest) ([]byte, error) {
	opts := []httputil.RequestOption{httputil.WithHeaders(req.Reason.Header())}
	if req.Body != nil {
		opts = append(opts, httputil.JSONRequest, httputil.WithBodyBytes(req.Body))
	}

	resp, err := client.WithContext(ctx).Request(
		req.Method,
		api.Endpoint+strings.TrimPrefix(req.Path, "/"),
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send Discord API request: %w", WithoutRequestURL(err))
	}

	body := resp.GetBody()
	defer body.Close()

	data, err := io.ReadAll(io.LimitReader(body, thing.MaxBodySize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read Discord API response: %w", err)
	}
	if len(data) > thing.MaxBodySize {
		return nil, fmt.Errorf("body size exceeds max body size of %d bytes", thing.MaxBodySize)
	}

	return data, nil
}

// EditServerProfile changes the bot's profile in a server.
func EditServerProfile(ctx context.Context, client *api.Client, guildID discord.GuildID, data provider.ServerProfileData) error {
	err := client.WithContext(ctx).FastRequest(
		"PATCH",
		api.EndpointGuilds+guildID.String()+"/members/@me",
		httputil.WithJSONBody(data),
	)
	if err != nil {
		return fmt.Errorf("failed to edit server profile: %w", err)
	}
	return nil
}

// EditBotProfile changes the avatar and banner of the app's bot everywhere,
// and returns the updated bot user.
func EditBotProfile(ctx context.Context, client *api.Client, data provider.BotProfileData) (*discord.User, error) {
	var user discord.User
	err := client.WithContext(ctx).RequestJSON(&user, "PATCH", api.EndpointMe, httputil.WithJSONBody(data))
	if err != nil {
		return nil, fmt.Errorf("failed to edit bot profile: %w", err)
	}
	return &user, nil
}

// WithoutRequestURL removes the path from the error of a request that never
// reached Discord, as it can contain an interaction token.
func WithoutRequestURL(err error) error {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if u, parseErr := url.Parse(urlErr.URL); parseErr == nil {
			urlErr.URL = u.Scheme + "://" + u.Host
		} else {
			urlErr.URL = ""
		}
	}
	return err
}
