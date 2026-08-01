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

// configureHTTPTransport raises the idle connection limits on the default
// transport.
//
// Every arikawa api.Client is built from a zero-value http.Client, so they all
// share http.DefaultTransport and its connection pool. That part is fine, but
// DefaultTransport leaves MaxIdleConnsPerHost at the package default of 2:
// under any sustained parallel load to discord.com, all but two connections
// are closed after each request and have to redo the TLS handshake next time.
//
// MaxIdleConns has to move with it. Leaving the total at its default of 100
// while allowing more per host just means one host can starve the pool.
//
// Note this transport is shared with the engine's outbound HTTP client when no
// proxy is configured, so flow HTTP requests draw from the same pool.
func configureHTTPTransport(cfg *config.Config) {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		slog.Warn("Default HTTP transport is not *http.Transport, skipping connection pool tuning")
		return
	}

	maxIdleConns := cfg.HTTP.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 512
	}

	maxIdleConnsPerHost := cfg.HTTP.MaxIdleConnsPerHost
	if maxIdleConnsPerHost <= 0 {
		maxIdleConnsPerHost = 128
	}

	transport.MaxIdleConns = maxIdleConns
	transport.MaxIdleConnsPerHost = maxIdleConnsPerHost

	slog.Info(
		"Configured HTTP connection pool",
		slog.Int("max_idle_conns", maxIdleConns),
		slog.Int("max_idle_conns_per_host", maxIdleConnsPerHost),
	)
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
