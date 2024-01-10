package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/merlinfuchs/kite/go-types/dismodel"
	"github.com/merlinfuchs/kite/go-types/event"
	"github.com/merlinfuchs/kite/kite-service/config"
	"github.com/merlinfuchs/kite/kite-service/internal/bot"
	"github.com/merlinfuchs/kite/kite-service/internal/host"
	"github.com/merlinfuchs/kite/kite-service/pkg/engine"
	"github.com/merlinfuchs/kite/kite-service/pkg/plugin"
)

func RunServer(cfg *config.ServerConfig) error {
	bot, err := bot.New(cfg.Discord.Token)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	engine := engine.New()
	env := host.NewEnv(bot)

	commands := []*discordgo.ApplicationCommand{}

	for _, pl := range cfg.Plugins {
		pluginCFG, err := config.LoadPluginConfig(pl.Path)
		if err != nil {
			return fmt.Errorf("failed to load plugin config: %w", err)
		}

		for _, cmd := range pluginCFG.Commands {
			options := []*discordgo.ApplicationCommandOption{}
			for _, subCMD := range cmd.SubCommands {
				options = append(options, &discordgo.ApplicationCommandOption{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        subCMD.Name,
					Description: subCMD.Description,
				})
			}

			commands = append(commands, &discordgo.ApplicationCommand{
				Name:        cmd.Name,
				Description: cmd.Description,
				Options:     options,
			})
		}

		wasm, err := os.ReadFile(filepath.Join(pl.Path, pluginCFG.Build.Out))
		if err != nil {
			return fmt.Errorf("failed to read plugin wasm: %w", err)
		}

		userConfig := make(map[string]string, len(pluginCFG.DefaultConfig)+len(pl.Config))
		for k, v := range pluginCFG.DefaultConfig {
			userConfig[k] = v
		}
		for k, v := range pl.Config {
			userConfig[k] = v
		}

		manifest := plugin.PluginManifest{
			ID:          "abc",
			Events:      pluginCFG.Events,
			Permissions: pluginCFG.Permissions,
		}
		config := plugin.PluginConfig{
			MemoryPagesLimit:   32,
			UserConfig:         userConfig,
			TotalTimeLimit:     time.Second * 10,
			ExecutionTimeLimit: time.Millisecond * 100,
		}

		plugin, err := plugin.New(context.Background(), wasm, manifest, config, &env)
		if err != nil {
			return fmt.Errorf("failed to create plugin: %w", err)
		}

		err = engine.LoadPlugin(plugin, pl.GuildIDs)
		if err != nil {
			return fmt.Errorf("failed to load plugin: %w", err)
		}
	}

	_, err = bot.Session.ApplicationCommandBulkOverwrite(cfg.Discord.ClientID, "", commands)
	if err != nil {
		return fmt.Errorf("failed to overwrite commands: %w", err)
	}

	bot.Session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		err := engine.HandleEvent(context.Background(), &event.Event{
			Type:    event.DiscordMessageCreate,
			GuildID: m.GuildID,
			Data: dismodel.MessageCreateEvent{
				ID:        m.ID,
				ChannelID: m.ChannelID,
				Content:   m.Content,
			},
		})
		if err != nil {
			fmt.Printf("failed to handle event: %v\n", err)
		}
	})

	bot.Session.AddHandler(func(s *discordgo.Session, m *discordgo.InteractionCreate) {
		err := engine.HandleEvent(context.Background(), &event.Event{
			Type:    event.DiscordInteractionCreate,
			GuildID: m.GuildID,
			Data: dismodel.InteractionCreateEvent{
				ID:        m.ID,
				Type:      dismodel.InteractionType(m.Type),
				Token:     m.Token,
				ChannelID: m.ChannelID,
				Data:      m.Data,
			},
		})
		if err != nil {
			fmt.Printf("failed to handle event: %v\n", err)
		}
	})

	err = bot.Session.Open()
	if err != nil {
		return fmt.Errorf("failed to open discord session: %w", err)
	}

	<-make(chan struct{})
	return nil
}
