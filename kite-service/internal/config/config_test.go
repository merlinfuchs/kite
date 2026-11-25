package config

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"

	tomlparser "github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/rawbytes"
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

// Identifying with an intent Discord doesn't know fails, so a typo stops the
// start instead.
func TestSharedBotIntentsValidate(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	for intents, ok := range map[string]bool{
		"":                              true,
		"message_content":               true,
		"message_content,guild_members": true,
		"presence":                      false,
		"message-content":               false,
	} {
		cfg := SharedBotConfig{}
		if intents != "" {
			cfg.Intents = strings.Split(intents, ",")
		}
		if err := validate.Struct(cfg); (err == nil) != ok {
			t.Errorf("intents %q: err = %v, want ok %v", intents, err, ok)
		}
	}
}

// The times of a sale can be TOML datetimes or quoted strings, and the sale
// only runs from its start until its end.
func TestBillingSaleParse(t *testing.T) {
	for name, toml := range map[string]string{
		"datetime": "[billing.sale]\nname = \"Black Week\"\ncode = \"BW\"\npercent = 30\nends_at = 2026-12-01T00:00:00+01:00\n",
		"string":   "[billing.sale]\nname = \"Black Week\"\ncode = \"BW\"\npercent = 30\nends_at = \"2026-12-01T00:00:00+01:00\"\n",
	} {
		t.Run(name, func(t *testing.T) {
			k, err := defaultBase()
			if err != nil {
				t.Fatal(err)
			}
			if err := k.Load(rawbytes.Provider([]byte(toml)), tomlparser.Parser()); err != nil {
				t.Fatal(err)
			}
			var cfg Config
			if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
				t.Fatal(err)
			}

			sale := cfg.Billing.Sale
			end := time.Date(2026, 11, 30, 23, 0, 0, 0, time.UTC)
			if !sale.EndsAt.Equal(end) {
				t.Fatalf("ends_at = %v, want %v", sale.EndsAt, end)
			}
			if !sale.Active(end.Add(-time.Second)) {
				t.Error("sale should run until its end")
			}
			if sale.Active(end) {
				t.Error("sale should be over at its end")
			}
			if err := validator.New(validator.WithRequiredStructEnabled()).Struct(sale); err != nil {
				t.Errorf("valid sale fails validation: %v", err)
			}
		})
	}
}

func TestBillingPaddleCheckoutValidate(t *testing.T) {
	now := time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)
	paddle := BillingPaddleConfig{APIKey: "key", ClientToken: "token", WebhookSecret: "secret"}
	plans := []BillingPlanConfig{
		{ID: "basic"},
		{ID: "plus", LemonSqueezyVariantID: "10", LemonSqueezyYearlyVariantID: "11", PaddlePriceID: "pri_1", PaddleYearlyPriceID: "pri_2"},
	}
	sale := BillingSaleConfig{Name: "Black Week", Code: "BW", Percent: 30, EndsAt: now.Add(time.Hour)}

	valid := BillingConfig{CheckoutProvider: "paddle", Paddle: paddle, Plans: plans}
	for name, tt := range map[string]struct {
		change func(c *BillingConfig)
		ok     bool
	}{
		"complete": {func(c *BillingConfig) {}, true},
		"on Lemon Squeezy": {func(c *BillingConfig) {
			c.CheckoutProvider = "lemonsqueezy"
			c.Paddle = BillingPaddleConfig{}
			c.Sale = sale
		}, true},
		"without a client token":   {func(c *BillingConfig) { c.Paddle.ClientToken = "" }, false},
		"without a webhook secret": {func(c *BillingConfig) { c.Paddle.WebhookSecret = "" }, false},
		"without a yearly price": {func(c *BillingConfig) {
			c.Plans = []BillingPlanConfig{{ID: "plus", LemonSqueezyVariantID: "10", LemonSqueezyYearlyVariantID: "11", PaddlePriceID: "pri_1"}}
		}, false},
		"with a sale without a discount": {func(c *BillingConfig) { c.Sale = sale }, false},
		"with a sale and a discount": {func(c *BillingConfig) {
			c.Sale = sale
			c.Sale.PaddleDiscountID = "dsc_1"
		}, true},
		"after the sale": {func(c *BillingConfig) {
			c.Sale = sale
			c.Sale.EndsAt = now.Add(-time.Hour)
		}, true},
	} {
		cfg := valid
		cfg.Plans = slices.Clone(plans)
		tt.change(&cfg)
		if err := cfg.validate(now); (err == nil) != tt.ok {
			t.Errorf("%s: err = %v, want ok %v", name, err, tt.ok)
		}
	}
}

func TestBillingSaleWithoutCode(t *testing.T) {
	cfg := unmarshalDefaults(t)
	if cfg.Billing.Sale.Active(time.Now()) {
		t.Error("the default config has no sale")
	}
	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(BillingSaleConfig{Code: "BW"}); err == nil {
		t.Error("a sale with a code but no percent or end should fail validation")
	}
}
