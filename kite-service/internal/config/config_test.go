package config

import (
	"testing"
	"time"

	"github.com/knadh/koanf/v2"
)

// unmarshalDefaults loads the embedded default config the same way LoadConfig
// does, but skips Validate: the defaults intentionally omit required secrets
// like the encryption key, which callers are expected to supply.
func unmarshalDefaults(t *testing.T) Config {
	t.Helper()

	k, err := defaultBase()
	if err != nil {
		t.Fatalf("failed to load default config: %v", err)
	}

	var cfg Config
	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		t.Fatalf("failed to unmarshal default config: %v", err)
	}

	return cfg
}

// The polling intervals are written as duration strings ("5s") in
// default.toml. If the decode hook that turns those into time.Duration ever
// stops being applied they would silently decode to zero, which turns every
// ticker into a busy loop. Assert the parsed values directly.
func TestDefaultDurationsParse(t *testing.T) {
	cfg := unmarshalDefaults(t)

	tests := []struct {
		name string
		got  time.Duration
		want time.Duration
	}{
		{"engine.populate_interval", cfg.Engine.PopulateInterval, 5 * time.Second},
		{"engine.remove_dangling_interval", cfg.Engine.RemoveDanglingInterval, 5 * time.Minute},
		{"engine.populate_overlap", cfg.Engine.PopulateOverlap, 5 * time.Second},
		{"gateway.populate_interval", cfg.Gateway.PopulateInterval, 10 * time.Second},
		{"gateway.remove_dangling_interval", cfg.Gateway.RemoveDanglingInterval, 60 * time.Second},
		{"gateway.populate_overlap", cfg.Gateway.PopulateOverlap, 5 * time.Second},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

func TestDefaultGatewayStartInterval(t *testing.T) {
	cfg := unmarshalDefaults(t)

	if cfg.Gateway.StartInterval <= 0 {
		t.Errorf("gateway.start_interval = %v, want > 0", cfg.Gateway.StartInterval)
	}
}

// Go's DefaultTransport leaves MaxIdleConnsPerHost at 2, which forces a TLS
// handshake for all but two concurrent requests to discord.com. Raising the
// per-host limit only helps if the total moves with it, otherwise a single
// host can still only keep its share of the default 100.
func TestDefaultHTTPPoolLimits(t *testing.T) {
	cfg := unmarshalDefaults(t)

	if cfg.HTTP.MaxIdleConnsPerHost <= 2 {
		t.Errorf("http.max_idle_conns_per_host = %d, want more than the Go default of 2",
			cfg.HTTP.MaxIdleConnsPerHost)
	}
	if cfg.HTTP.MaxIdleConns < cfg.HTTP.MaxIdleConnsPerHost {
		t.Errorf("http.max_idle_conns = %d is below max_idle_conns_per_host = %d, which caps the per-host limit",
			cfg.HTTP.MaxIdleConns, cfg.HTTP.MaxIdleConnsPerHost)
	}
}

// The debug server exposes uncontrolled pprof profiles, so it must stay off
// unless explicitly enabled and must not default to a public bind address.
func TestDebugServerDefaultsToDisabledAndLocal(t *testing.T) {
	cfg := unmarshalDefaults(t)

	if cfg.Debug.Enabled {
		t.Error("debug.enabled = true, want false by default")
	}
	if cfg.Debug.Host != "127.0.0.1" {
		t.Errorf("debug.host = %q, want %q", cfg.Debug.Host, "127.0.0.1")
	}
}

func TestSelfHostedProfileEnablesFeaturesAndRaisesQuotas(t *testing.T) {
	cfg := unmarshalDefaults(t)
	if len(cfg.Billing.Plans) != 1 {
		t.Fatalf("default plans = %d, want 1", len(cfg.Billing.Plans))
	}

	defaultCredits := cfg.Billing.Plans[0].FeatureUsageCreditsPerMonth
	applySelfHostedDefaults(&cfg)
	if cfg.Billing.Plans[0].FeatureUsageCreditsPerMonth != defaultCredits {
		t.Fatal("self-hosted profile changed quotas while disabled")
	}

	cfg.SelfHosted.Enabled = true
	applySelfHostedDefaults(&cfg)
	plan := cfg.Billing.Plans[0]
	if plan.FeatureUsageCreditsPerMonth <= defaultCredits {
		t.Errorf("monthly credits = %d, want more than default %d", plan.FeatureUsageCreditsPerMonth, defaultCredits)
	}
	if plan.FeatureMaxCollaborators < 3 {
		t.Errorf("collaborators = %d, want at least 3", plan.FeatureMaxCollaborators)
	}
	if !plan.FeatureVoiceAudio || !plan.FeaturePersonas || !plan.FeatureAnimatedProfile {
		t.Error("self-hosted profile should enable all boolean plan features")
	}
	if cfg.UserLimits.MaxAppsPerUser < 100 {
		t.Errorf("max apps per user = %d, want a generous self-hosted limit", cfg.UserLimits.MaxAppsPerUser)
	}
}

func TestSelfHostedAllowedDiscordIDsEnvironmentValue(t *testing.T) {
	t.Setenv("KITE_SELF_HOSTED__ALLOWED_DISCORD_IDS", "123, 456")
	k, err := loadBase(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ids := k.Strings("self_hosted.allowed_discord_ids")
	if len(ids) != 2 || ids[0] != "123" || ids[1] != "456" {
		t.Fatalf("allowed Discord IDs = %#v, want [123 456]", ids)
	}
}
