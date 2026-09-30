package config

import (
	"github.com/go-playground/validator/v10"
)

type Config struct {
	Discord   DiscordConfig   `toml:"discord"`
	OpenAI    OpenAIConfig    `toml:"openai"`
	Index     IndexConfig     `toml:"index"`
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
	APIKey         string `toml:"api_key"`
	ChatModel      string `toml:"chat_model"`
	EmbeddingModel string `toml:"embedding_model"`
	SummaryModel   string `toml:"summary_model"`
}

type IndexConfig struct {
	DBPath       string  `toml:"db_path"`
	DocsPath     string  `toml:"docs_path"`
	ConceptsPath string  `toml:"concepts_path"`
	ServicePath  string  `toml:"service_path"`
	TopK         int     `toml:"top_k"`
	ChunkSize    int     `toml:"chunk_size"`
	ChunkOverlap int     `toml:"chunk_overlap"`
	ScoreFloor   float32 `toml:"score_floor"`
	DocsBaseURL  string  `toml:"docs_base_url"`
}

type RateLimitConfig struct {
	PerUserMax    int    `toml:"per_user_max"`
	PerUserWindow string `toml:"per_user_window"`
}

func LoadConfig(basePath string) (*Config, error) {
	return loadConfig[*Config](basePath)
}
