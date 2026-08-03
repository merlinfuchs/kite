package server

import (
	"log/slog"
	"net/http"
	"net/url"
	"time"

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

	// Defaults live in default.toml, which is always the base config layer.
	// Restating them here would just be a second source of truth, so a
	// non-positive value means "leave Go's default alone" instead.
	if cfg.HTTP.MaxIdleConns > 0 {
		transport.MaxIdleConns = cfg.HTTP.MaxIdleConns
	}
	if cfg.HTTP.MaxIdleConnsPerHost > 0 {
		transport.MaxIdleConnsPerHost = cfg.HTTP.MaxIdleConnsPerHost
	}

	// Go leaves MaxConnsPerHost unlimited, which means a burst of concurrent
	// requests opens one socket each. On a cold start that produced thousands
	// of simultaneous connections to a single host, which upstream served at a
	// few requests per second -- almost certainly throttling what looks like a
	// connection flood. Bounding it makes the requests queue locally and go
	// out at a sane rate instead, which is both faster end to end and far less
	// hostile to whatever is on the other side.
	if cfg.HTTP.MaxConnsPerHost > 0 {
		transport.MaxConnsPerHost = cfg.HTTP.MaxConnsPerHost
	}

	slog.Info(
		"Configured HTTP connection pool",
		slog.Int("max_idle_conns", transport.MaxIdleConns),
		slog.Int("max_idle_conns_per_host", transport.MaxIdleConnsPerHost),
		slog.Int("max_conns_per_host", transport.MaxConnsPerHost),
	)
}

// warnUnverifiedBillingWebhook reports that /v1/billing/webhook is accepting
// unauthenticated writes.
//
// The signature check HMACs the body under billing.lemonsqueezy_signing_secret
// and compares hex strings. Neither the LemonSqueezy client nor the handler
// rejects an empty secret, so with the secret unset the key is public knowledge
// and anyone can produce a valid signature for a payload of their choosing. The
// endpoint grants entitlements, so a forged subscription_created is a free
// upgrade on any app ID the caller names.
func warnUnverifiedBillingWebhook(cfg *config.Config) {
	if cfg.Billing.LemonSqueezySigningSecret != "" {
		return
	}

	slog.Warn(
		"No billing.lemonsqueezy_signing_secret configured: the webhook " +
			"signature is an HMAC under an empty key, so /v1/billing/webhook " +
			"accepts forged events from anyone and will grant entitlements " +
			"from them. Set the secret, or keep the webhook unreachable.",
	)
}

// engineHTTPTimeout bounds an outbound flow request end to end.
//
// The request is already tied to the flow context, whose deadline is shorter
// than this, so in normal operation the context wins. This is the backstop for
// any caller that reaches the client without a deadline.
const engineHTTPTimeout = 30 * time.Second

func engineHTTPClient(cfg *config.Config) *http.Client {
	client := &http.Client{Timeout: engineHTTPTimeout}

	if cfg.Engine.HTTPProxyURL == "" {
		warnNoEgressProxy()
		return client
	}

	proxyURL, err := url.Parse(cfg.Engine.HTTPProxyURL)
	if err != nil {
		slog.With("error", err).Error("Failed to parse proxy URL")
		return nil
	}

	slog.Info("Using HTTP proxy for Engine", "url", cfg.Engine.HTTPProxyURL)
	client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	return client
}

// warnNoEgressProxy reports that flow requests leave the host unfiltered.
//
// The URL of an action_http_request node is authored by the tenant and is not
// validated anywhere, so without an egress proxy the engine will resolve and
// fetch whatever it is given -- including loopback, RFC1918, and cloud metadata
// endpoints -- and hand the response body back into the flow where it can be
// printed into a Discord channel.
func warnNoEgressProxy() {
	slog.Warn(
		"No HTTP proxy configured for the engine: outbound flow requests go " +
			"direct, so tenant-authored URLs can reach the host's internal " +
			"network. Set engine.http_proxy_url to an egress proxy that " +
			"rejects internal address ranges.",
	)
}
