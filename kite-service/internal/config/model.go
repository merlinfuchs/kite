package config

import (
	"time"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	Logging    LoggingConfig    `toml:"logging"`
	Database   DatabaseConfig   `toml:"database"`
	API        APIConfig        `toml:"api"`
	App        AppConfig        `toml:"app"`
	UserLimits UserLimitsConfig `toml:"user_limits"`
	Discord    DiscordConfig    `toml:"discord"`
	Engine     EngineConfig     `toml:"engine"`
	Gateway    GatewayConfig    `toml:"gateway"`
	OpenAI     OpenAIConfig     `toml:"openai"`
	Billing    BillingConfig    `toml:"billing"`
	Encryption EncryptionConfig `toml:"encryption"`
	HTTP       HTTPConfig       `toml:"http"`
	Debug      DebugConfig      `toml:"debug"`

	ClusterCount int `toml:"cluster_count"`
	ClusterIndex int `toml:"cluster_index"`
}

func (cfg *Config) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(cfg)
}

func (cfg *Config) IsPrimaryCluster() bool {
	return cfg.ClusterIndex == 0
}

func LoadConfig(basePath string) (*Config, error) {
	return loadConfig[*Config](basePath)
}

type DatabaseConfig struct {
	Postgres PostgresConfig `toml:"postgres"`
	S3       S3Config       `toml:"s3"`
}

type LoggingConfig struct {
	Filename   string `toml:"filename"`
	MaxSize    int    `toml:"max_size"`
	MaxAge     int    `toml:"max_age"`
	MaxBackups int    `toml:"max_backups"`
}

type PostgresConfig struct {
	Host     string `toml:"host" validate:"required"`
	Port     int    `toml:"port" validate:"required"`
	DBName   string `toml:"db_name" validate:"required"`
	User     string `toml:"user" validate:"required"`
	Password string `toml:"password"`
}

type S3Config struct {
	Endpoint        string `toml:"endpoint" validate:"required"`
	AccessKeyID     string `toml:"access_key_id" validate:"required"`
	SecretAccessKey string `toml:"secret_access_key" validate:"required"`
	Secure          bool   `toml:"secure"`
	SSECKey         string `toml:"ssec_key"`
}

type APIConfig struct {
	Host          string `toml:"host" validate:"required"`
	Port          int    `toml:"port" validate:"required"`
	PublicBaseURL string `toml:"public_base_url" validate:"required"`
	SecureCookies bool   `toml:"secure_cookies"`
	StrictCookies bool   `toml:"strict_cookies"`
}

type AppConfig struct {
	PublicBaseURL string `toml:"public_base_url" validate:"required"`
}

type EncryptionConfig struct {
	TokenEncryptionKey string `toml:"token_encryption_key" validate:"required"`
}

type DiscordConfig struct {
	ClientID     string `toml:"client_id" validate:"required"`
	ClientSecret string `toml:"client_secret" validate:"required"`
	// BotToken is used to hand out roles to users
	BotToken string `toml:"bot_token"`
	// GuildID is the ID of the guild to hand out roles to users
	GuildID  string `toml:"guild_id"`
	ProxyURL string `toml:"proxy_url"`
}

type EngineConfig struct {
	MaxStackDepth int    `toml:"max_stack_depth"`
	MaxOperations int    `toml:"max_operations"`
	MaxCredits    int    `toml:"max_credits"`
	HTTPProxyURL  string `toml:"http_proxy_url"`

	// PopulateInterval controls how often the engine polls for changed
	// commands, event listeners, and plugin instances. This is the delay
	// between saving an entity in the dashboard and it taking effect.
	PopulateInterval time.Duration `toml:"populate_interval"`
	// RemoveDanglingInterval controls how often the engine scans for deleted
	// entities. Each run issues an unbounded query per entity type, so this
	// should stay well above PopulateInterval.
	RemoveDanglingInterval time.Duration `toml:"remove_dangling_interval"`
	// PopulateOverlap is subtracted from the poll cursor after each
	// successful poll, so entities written while a query was in flight are
	// not missed. Populating is idempotent, so re-reading a few rows is
	// harmless.
	PopulateOverlap time.Duration `toml:"populate_overlap"`
}

type GatewayConfig struct {
	// PopulateInterval controls how often the gateway manager polls for new
	// and changed apps.
	PopulateInterval time.Duration `toml:"populate_interval"`
	// RemoveDanglingInterval controls how often the gateway manager runs a
	// full scan of enabled app IDs to find gateways whose app was deleted.
	// Disabled apps are caught by the cheaper per-poll query instead.
	RemoveDanglingInterval time.Duration `toml:"remove_dangling_interval"`
	// PopulateOverlap works the same way as EngineConfig.PopulateOverlap.
	PopulateOverlap time.Duration `toml:"populate_overlap"`
	// StartInterval paces how quickly gateway connections are initiated. Each
	// app has its own bot token and therefore its own identify budget, so
	// this exists to keep a cold start from opening tens of thousands of TLS
	// handshakes at once rather than to satisfy a Discord rate limit.
	StartInterval time.Duration `toml:"start_interval"`
}

// HTTPConfig tunes the shared default HTTP transport, which every Discord API
// client draws connections from.
type HTTPConfig struct {
	MaxIdleConns        int `toml:"max_idle_conns"`
	MaxIdleConnsPerHost int `toml:"max_idle_conns_per_host"`
	// MaxConnsPerHost bounds concurrent connections to a single host. Go
	// leaves this unlimited, so a burst opens one socket per in-flight
	// request; bounding it makes requests queue locally instead of arriving
	// as a connection flood.
	MaxConnsPerHost int `toml:"max_conns_per_host"`
}

// DebugConfig controls the debug server, which exposes pprof profiles and
// expvar counters. It must only ever be bound to a private interface: the
// profiles it serves are not access-controlled.
type DebugConfig struct {
	Enabled bool   `toml:"enabled"`
	Host    string `toml:"host"`
	Port    int    `toml:"port"`
}

type UserLimitsConfig struct {
	MaxAppsPerUser int `toml:"max_apps_per_user"`
	MaxAssetSize   int `toml:"max_asset_size"`
}

type OpenAIConfig struct {
	APIKey string `toml:"api_key"`
}

type BillingConfig struct {
	LemonSqueezyAPIKey        string              `toml:"lemonsqueezy_api_key"`
	LemonSqueezySigningSecret string              `toml:"lemonsqueezy_signing_secret"`
	LemonSqueezyStoreID       string              `toml:"lemonsqueezy_store_id"`
	TestMode                  bool                `toml:"test_mode"`
	Plans                     []BillingPlanConfig `toml:"plans"`
}

type BillingPlanConfig struct {
	ID          string  `toml:"id" validate:"required"`
	Title       string  `toml:"title" validate:"required"`
	Description string  `toml:"description" validate:"required"`
	Price       float32 `toml:"price" validate:"required"`
	Default     bool    `toml:"default"`
	Popular     bool    `toml:"popular"`
	Hidden      bool    `toml:"hidden"`

	LemonSqueezyProductID string `toml:"lemonsqueezy_product_id"`
	LemonSqueezyVariantID string `toml:"lemonsqueezy_variant_id"`

	DiscordRoleID string `toml:"discord_role_id"`

	FeatureMaxCollaborators     int  `toml:"feature_max_collaborators"`
	FeatureUsageCreditsPerMonth int  `toml:"feature_usage_credits_per_month"`
	FeatureMaxGuilds            int  `toml:"feature_max_guilds"`
	FeatureMaxCommands          int  `toml:"feature_max_commands"`
	FeatureMaxVariables         int  `toml:"feature_max_variables"`
	FeatureMaxMessages          int  `toml:"feature_max_messages"`
	FeatureMaxEventListeners    int  `toml:"feature_max_event_listeners"`
	FeaturePrioritySupport      bool `toml:"feature_priority_support"`
}
