package config

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	gotoml "github.com/pelletier/go-toml/v2"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

const ConfigFile = "kite.toml"

//go:embed default.toml
var defaultConfig []byte

type Validate interface {
	Validate() error
}

func loadConfig[T Validate](basePath string) (T, error) {
	var res T

	k, err := loadBase(basePath)
	if err != nil {
		return res, fmt.Errorf("Failed to load base config: %v", err)
	}

	if err := k.UnmarshalWithConf("", &res, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return res, fmt.Errorf("Failed to unmarshal full config: %v", err)
	}
	if cfg, ok := any(res).(*Config); ok {
		applySelfHostedDefaults(cfg)
	}

	if err := res.Validate(); err != nil {
		return res, fmt.Errorf("Failed to validate plugin config: %v", err)
	}

	return res, nil
}

// applySelfHostedDefaults turns on the product features and generous quotas
// intended for an instance operated by one owner. Infrastructure and Discord
// still impose their own resource and API limits.
func applySelfHostedDefaults(cfg *Config) {
	if !cfg.SelfHosted.Enabled {
		return
	}

	const highLimit = 1_000_000
	cfg.UserLimits.MaxAppsPerUser = highLimit
	cfg.UserLimits.MaxAssetStoragePerApp = 10_000_000_000
	cfg.UserLimits.MaxBundlesPerUser = highLimit
	cfg.UserLimits.MaxAPITokensPerUser = highLimit

	for i := range cfg.Billing.Plans {
		plan := &cfg.Billing.Plans[i]
		if !plan.Default {
			continue
		}

		plan.Title = "Self-hosted"
		plan.Description = "Personal instance with features enabled and configurable infrastructure limits."
		plan.FeatureMaxCollaborators = 10
		plan.FeatureUsageCreditsPerMonth = highLimit
		plan.FeatureMaxGuilds = highLimit
		plan.FeatureMaxCommands = highLimit
		plan.FeatureMaxVariables = highLimit
		plan.FeatureMaxSecrets = highLimit
		plan.FeatureMaxMessages = highLimit
		plan.FeatureMaxEventListeners = highLimit
		plan.FeaturePrioritySupport = true
		plan.FeatureRotatingStatus = true
		plan.FeatureServerProfiles = true
		plan.FeatureAnimatedProfile = true
		plan.FeaturePersonas = true
		plan.FeatureVoiceAudio = true
		plan.FeatureMaxScheduledEventListeners = highLimit
		plan.FeatureMinScheduleIntervalSeconds = 1
		plan.FeatureMaxWebhookEventListeners = highLimit
		plan.FeatureMaxAIPromptsPerMonth = highLimit
	}
}

func defaultBase() (*koanf.Koanf, error) {
	k := koanf.New(".")
	parser := toml.Parser()

	if err := k.Load(rawbytes.Provider(defaultConfig), parser); err != nil {
		return nil, fmt.Errorf("Failed to load default config: %v", err)
	}

	return k, nil
}

func loadBase(basePath string) (*koanf.Koanf, error) {
	k := koanf.New(".")
	parser := toml.Parser()

	if err := k.Load(rawbytes.Provider(defaultConfig), parser); err != nil {
		return nil, fmt.Errorf("Failed to load default config: %v", err)
	}

	configPath := filepath.Join(basePath, ConfigFile)
	if err := k.Load(file.Provider(configPath), parser); err != nil {
		var pathError *fs.PathError
		if !errors.As(err, &pathError) {
			return nil, fmt.Errorf("Failed to load config file: %v", err)
		}
	}

	envProvider := env.Provider("KITE_", ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "KITE_")), "__", ".", -1)
	})
	if err := k.Load(envProvider, nil); err != nil {
		return nil, fmt.Errorf("Failed to load env config: %v", err)
	}
	// Environment variables cannot represent a TOML string slice directly.
	// Accept a comma-separated value for Railway and other hosts that only
	// provide environment-based configuration.
	if allowedIDs, ok := os.LookupEnv("KITE_SELF_HOSTED__ALLOWED_DISCORD_IDS"); ok {
		ids := make([]string, 0)
		for _, id := range strings.Split(allowedIDs, ",") {
			if id = strings.TrimSpace(id); id != "" {
				ids = append(ids, id)
			}
		}
		k.Set("self_hosted.allowed_discord_ids", ids)
	}

	return k, nil
}

func ConfigExists(basePath string) bool {
	configPath := filepath.Join(basePath, ConfigFile)
	_, err := os.Stat(configPath)
	return err == nil
}

func WriteConfig(basePath string, conf interface{}) error {
	configPath := filepath.Join(basePath, ConfigFile)

	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("Failed to create config file: %v", err)
	}

	defer f.Close()

	encoder := gotoml.NewEncoder(f)
	if err := encoder.Encode(conf); err != nil {
		return fmt.Errorf("Failed to encode config: %v", err)
	}

	return nil
}
