package config

import (
	"github.com/go-playground/validator/v10"
)

type Config struct {
	Discord   DiscordConfig   `toml:"discord"`
	OpenAI    OpenAIConfig    `toml:"openai"`
	Knowledge KnowledgeConfig `toml:"knowledge"`
	RateLimit RateLimitConfig `toml:"ratelimit"`
	Feedback  FeedbackConfig  `toml:"feedback"`
	Help      HelpConfig      `toml:"help"`
}

type FeedbackConfig struct {
	ChannelID string `toml:"channel_id"`
	RoleID    string `toml:"role_id"`
}

type HelpConfig struct {
	ChannelID string `toml:"channel_id"`
	RoleID    string `toml:"role_id"`
}

func (cfg *Config) Validate() error {
	v := validator.New(validator.WithRequiredStructEnabled())
	return v.Struct(cfg)
}

type DiscordConfig struct {
	Token string `toml:"token"`
	AppID string `toml:"app_id"`
}

type OpenAIConfig struct {
	APIKey          string `toml:"api_key"`
	Model           string `toml:"model"`
	ReasoningEffort string `toml:"reasoning_effort"`
	// Caps each answer, reasoning included.
	MaxOutputTokens int `toml:"max_output_tokens"`
}

type KnowledgeConfig struct {
	// Where `index` writes the knowledge, embedded into the binary.
	Path        string `toml:"path"`
	DocsPath    string `toml:"docs_path"`
	CatalogPath string `toml:"catalog_path"`
	DocsBaseURL string `toml:"docs_base_url"`
	PlansURL    string `toml:"plans_url"`
}

type RateLimitConfig struct {
	PerUserMax    int    `toml:"per_user_max"`
	PerUserWindow string `toml:"per_user_window"`
}

func LoadConfig(basePath string) (*Config, error) {
	return loadConfig[*Config](basePath)
}
