package config

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kitecloud/kite/kite-service/internal/model"
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
	OpenRouter OpenRouterConfig `toml:"openrouter"`
	Assistant  AssistantConfig  `toml:"assistant"`
	Prescreen  PrescreenConfig  `toml:"prescreen"`
	Billing    BillingConfig    `toml:"billing"`
	Encryption EncryptionConfig `toml:"encryption"`
	HTTP       HTTPConfig       `toml:"http"`
	Debug      DebugConfig      `toml:"debug"`
	// Kite's own registration at the services of integrations, by
	// integration ID, like [integrations.erlc].
	Integrations map[string]IntegrationConfig `toml:"integrations"`
	// SharedBot is the Kite bot, which shared apps run on. See
	// design/kite-lite.md.
	SharedBot SharedBotConfig `toml:"shared_bot"`

	ClusterCount int `toml:"cluster_count"`
	ClusterIndex int `toml:"cluster_index"`
	// RunSharedBot is set in the process of the Kite bot, which runs the
	// shared apps instead of a cluster's share of the others.
	RunSharedBot bool `toml:"-"`
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
	// AdminUserIDs are the Kite user IDs of who can review bundles for the
	// marketplace.
	AdminUserIDs []string `toml:"admin_user_ids"`
}

type EncryptionConfig struct {
	TokenEncryptionKey string `toml:"token_encryption_key" validate:"required"`
}

type SharedBotConfig struct {
	// Token is the Kite bot's token, from the environment like other secrets.
	// Without it, shared apps don't run.
	Token string `toml:"token"`
	// ClientSecret is the OAuth2 secret of the Kite bot's application, for
	// logging in and adding the bot on one page. Without it, it's
	// discord.client_secret, as on kite.onl users log in with the Kite bot's
	// application.
	ClientSecret string `toml:"client_secret"`
	// Intents are the privileged intents Discord granted the Kite bot, like
	// "message_content" and "guild_members". It asks for them, and templates
	// that read them can be added to its apps. Identifying with an intent
	// Discord didn't grant fails, so only list granted ones.
	Intents []string `toml:"intents" validate:"dive,oneof=message_content guild_members"`
}

// GrantedIntents are Intents as flags.
func (c SharedBotConfig) GrantedIntents() model.Intents {
	// The names are validated, so parsing can't fail.
	intents, _ := model.ParseIntents(c.Intents)
	return intents
}

type DiscordConfig struct {
	ClientID     string `toml:"client_id" validate:"required"`
	ClientSecret string `toml:"client_secret" validate:"required"`
	// BotToken is the bot of the Kite server, which hands out roles to users
	// and sends Kite's notifications
	BotToken string `toml:"bot_token"`
	// GuildID is the ID of the guild to hand out roles to users
	GuildID string `toml:"guild_id"`
	// CreatorRoleID is the role users with a template in the marketplace that
	// enough others installed get in the guild, see creatorMinInstallers.
	CreatorRoleID string `toml:"creator_role_id"`
	// ReviewChannelID is the channel the bot posts templates in that wait
	// for review for the marketplace.
	ReviewChannelID string `toml:"review_channel_id"`
	ProxyURL        string `toml:"proxy_url"`
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
	// Port of cluster 0, cluster N listens on Port + N.
	Port int `toml:"port"`
	// HeapDumpDir gets a heap and a goroutine profile when the Go heap of a
	// process passes HeapDumpThreshold bytes, for finding out afterwards what
	// filled it. Empty turns it off.
	HeapDumpDir       string `toml:"heap_dump_dir"`
	HeapDumpThreshold uint64 `toml:"heap_dump_threshold" validate:"gt=0"`
}

type UserLimitsConfig struct {
	MaxAppsPerUser        int `toml:"max_apps_per_user"`
	MaxAssetSize          int `toml:"max_asset_size"`
	MaxAssetStoragePerApp int `toml:"max_asset_storage_per_app"`
	MaxBundlesPerUser     int `toml:"max_bundles_per_user"`
	MaxAPITokensPerUser   int `toml:"max_api_tokens_per_user"`
}

type IntegrationConfig struct {
	// Kite's own key, sent next to the app's credential.
	APIKey string `toml:"api_key"`
	// Kite's ID at the service, used in authorization links.
	AppID string `toml:"app_id"`
}

type OpenAIConfig struct {
	APIKey string `toml:"api_key"`
}

// OpenRouterConfig is for the AI blocks that decide, like AI Choice, which
// run on a decision model only OpenRouter serves.
type OpenRouterConfig struct {
	APIKey string `toml:"api_key"`
}

// AssistantConfig configures the AI that helps users in the dashboard, like
// the one that edits flows. It uses the OpenAI API key.
type AssistantConfig struct {
	Model           string `toml:"model"`
	ReasoningEffort string `toml:"reasoning_effort"`
	// Caps each model call, reasoning included, so a prompt's cost is bounded.
	MaxOutputTokens int `toml:"max_output_tokens"`
	// Repairs of a flow prompt's edits that don't count as new prompts.
	MaxRepairs int `toml:"max_repairs"`
	// Format is how the flow AI changes flows: "edits", or "code" to write
	// the whole flow as code.
	Format string `toml:"format" validate:"omitempty,oneof=edits code"`
}

// PrescreenConfig configures the AI that screens templates submitted to the
// marketplace before the admins review them. It runs when there's an OpenAI
// API key, set the model to empty to turn it off.
type PrescreenConfig struct {
	Model string `toml:"model"`
}

type BillingConfig struct {
	LemonSqueezyAPIKey        string              `toml:"lemonsqueezy_api_key"`
	LemonSqueezySigningSecret string              `toml:"lemonsqueezy_signing_secret"`
	LemonSqueezyStoreID       string              `toml:"lemonsqueezy_store_id"`
	TestMode                  bool                `toml:"test_mode"`
	Paddle                    BillingPaddleConfig `toml:"paddle"`
	Plans                     []BillingPlanConfig `toml:"plans"`
	Sale                      BillingSaleConfig   `toml:"sale"`

	// CheckoutProvider is where new subscriptions are bought, lemonsqueezy
	// or paddle. Existing ones stay with theirs.
	CheckoutProvider string `toml:"checkout_provider" validate:"omitempty,oneof=lemonsqueezy paddle"`
}

// BillingPaddleConfig connects Paddle, which replaces Lemon Squeezy. Without
// an API key, Paddle subscriptions aren't synced.
type BillingPaddleConfig struct {
	APIKey string `toml:"api_key"`
	// WebhookSecret is the secret key of the notification destination, which
	// signs the webhooks Paddle sends.
	WebhookSecret string `toml:"webhook_secret"`
	// ClientToken lets Paddle.js open the checkout in the browser.
	ClientToken string `toml:"client_token"`
	// Sandbox uses the sandbox environment, whose API keys don't work in the
	// live one.
	Sandbox bool `toml:"sandbox"`
}

// BillingSaleConfig is a discount on the first payment of yearly plans, with
// a discount code that has to exist in Lemon Squeezy. The site and dashboard
// show it between StartsAt and EndsAt, and checkouts of yearly plans apply the
// code. Without a code there's no sale. Paddle checkouts apply the discount
// PaddleDiscountID instead, and without it the sale isn't shown while
// checkouts go to Paddle.
type BillingSaleConfig struct {
	// Name is what the site calls the sale, like "Black Week".
	Name     string    `toml:"name" validate:"required_with=Code"`
	Code     string    `toml:"code"`
	Percent  int       `toml:"percent" validate:"required_with=Code,omitempty,min=1,max=100"`
	StartsAt time.Time `toml:"starts_at"`
	EndsAt   time.Time `toml:"ends_at" validate:"required_with=Code"`

	PaddleDiscountID string `toml:"paddle_discount_id"`
}

// Active is whether the sale runs at now.
func (c BillingSaleConfig) Active(now time.Time) bool {
	return c.Code != "" && !now.Before(c.StartsAt) && now.Before(c.EndsAt)
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

	// The yearly variant of the same product. Without it the plan is only
	// offered monthly.
	YearlyPrice                 float32 `toml:"yearly_price"`
	LemonSqueezyYearlyVariantID string  `toml:"lemonsqueezy_yearly_variant_id"`

	// The Paddle product of the plan and its monthly and yearly prices.
	PaddleProductID     string `toml:"paddle_product_id"`
	PaddlePriceID       string `toml:"paddle_price_id"`
	PaddleYearlyPriceID string `toml:"paddle_yearly_price_id"`

	DiscordRoleID string `toml:"discord_role_id"`

	FeatureMaxCollaborators     int  `toml:"feature_max_collaborators"`
	FeatureUsageCreditsPerMonth int  `toml:"feature_usage_credits_per_month"`
	FeatureMaxGuilds            int  `toml:"feature_max_guilds"`
	FeatureMaxCommands          int  `toml:"feature_max_commands"`
	FeatureMaxVariables         int  `toml:"feature_max_variables"`
	FeatureMaxSecrets           int  `toml:"feature_max_secrets"`
	FeatureMaxMessages          int  `toml:"feature_max_messages"`
	FeatureMaxEventListeners    int  `toml:"feature_max_event_listeners"`
	FeaturePrioritySupport      bool `toml:"feature_priority_support"`
	FeatureRotatingStatus       bool `toml:"feature_rotating_status"`
	FeatureServerProfiles       bool `toml:"feature_server_profiles"`
	FeatureAnimatedProfile      bool `toml:"feature_animated_profile"`
	FeaturePersonas             bool `toml:"feature_personas"`
	FeatureVoiceAudio           bool `toml:"feature_voice_audio"`

	FeatureMaxScheduledEventListeners int `toml:"feature_max_scheduled_event_listeners"`
	FeatureMinScheduleIntervalSeconds int `toml:"feature_min_schedule_interval_seconds"`

	FeatureMaxWebhookEventListeners int `toml:"feature_max_webhook_event_listeners"`

	// Prompts to the flow AI per month. Unlike other limits, 0 means none.
	FeatureMaxAIPromptsPerMonth int `toml:"feature_max_ai_prompts_per_month"`
}
