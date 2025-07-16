package server

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/kitecloud/kite/kite-service/internal/config"
)

func patchDiscordProxyURL(cfg *config.Config) {
	if cfg.Discord.ProxyURL == "" {
		return
	}

	slog.Info("Using Proxy for Discord API", "url", cfg.Discord.ProxyURL)

	httputil.Retries = 10

	api.BaseEndpoint = cfg.Discord.ProxyURL
	api.Endpoint = api.BaseEndpoint + api.Path + "/"
	api.EndpointGateway = api.Endpoint + "gateway"
	api.EndpointGatewayBot = api.EndpointGateway + "/bot"
	api.EndpointApplications = api.Endpoint + "applications/"
	api.EndpointChannels = api.Endpoint + "channels/"
	api.EndpointGuilds = api.Endpoint + "guilds/"
	api.EndpointUsers = api.Endpoint + "users/"
	api.EndpointWebhooks = api.Endpoint + "webhooks/"
	api.EndpointInvites = api.Endpoint + "invites/"
	api.EndpointInteractions = api.Endpoint + "interactions/"
	api.EndpointStageInstances = api.Endpoint + "stage-instances/"
	api.EndpointMe = api.Endpoint + "users/@me"
	api.EndpointAuth = api.Endpoint + "auth/"
	api.EndpointLogin = api.EndpointAuth + "login"
	api.EndpointTOTP = api.EndpointAuth + "mfa/totp"
}

func engineHTTPClient(cfg *config.Config) *http.Client {
	if cfg.Engine.HTTPProxyURL != "" {
		proxyURL, err := url.Parse(cfg.Engine.HTTPProxyURL)
		if err != nil {
			slog.With("error", err).Error("Failed to parse proxy URL")
			return nil
		}

		slog.Info("Using HTTP proxy for Engine", "url", cfg.Engine.HTTPProxyURL)

		return &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}
	}

	return &http.Client{}
}
