package server

import (
	"fmt"

	"github.com/merlinfuchs/kite/kite-service/config"
	"github.com/merlinfuchs/kite/kite-service/internal/logging"
	"github.com/merlinfuchs/kite/kite-service/internal/migrate"
	"github.com/urfave/cli/v2"
)

var migratableStores = []string{"postgres"}

func migrateCMD() *cli.Command {
	subCommands := []*cli.Command{}
	for _, store := range migratableStores {
		subCommands = append(subCommands, migrateStoreCMD(store))
	}

	return &cli.Command{
		Name:        "migrate",
		Usage:       "Run and manage database migrations",
		Subcommands: subCommands,
	}
}

func migrateStoreCMD(store string) *cli.Command {
	return &cli.Command{
		Name:  store,
		Usage: "Run and manage database migrations for " + store,
		Subcommands: []*cli.Command{
			{
				Name:  "up",
				Usage: "Run all pending migrations",
				Action: func(c *cli.Context) error {
					return migrateRun(store, "up", migrate.MigrateOpts{})
				},
			},
			{
				Name:  "down",
				Usage: "Rollback all applied migrations",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:     "danger",
						Usage:    "Confirm that you know what you're doing",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					return migrateRun(store, "down", migrate.MigrateOpts{})
				},
			},
			{
				Name:  "version",
				Usage: "Print the current migration version",
				Action: func(c *cli.Context) error {
					return migrateRun(store, "version", migrate.MigrateOpts{})
				},
			},
			{
				Name:  "list",
				Usage: "List all migrations",
				Action: func(c *cli.Context) error {
					return migrateRun(store, "list", migrate.MigrateOpts{})
				},
			},
			{
				Name:  "force",
				Usage: "Force set the migration version",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:     "version",
						Usage:    "Version to set",
						Required: true,
					},
					&cli.BoolFlag{
						Name:     "danger",
						Usage:    "Confirm that you know what you're doing",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					return migrateRun(store, "force", migrate.MigrateOpts{
						TargetVersion: c.Int("version"),
					})
				},
			},
			{
				Name:  "to",
				Usage: "Migrate to a specific version",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:     "version",
						Usage:    "Version to set",
						Required: true,
					},
					&cli.BoolFlag{
						Name:     "danger",
						Usage:    "Confirm that you know what you're doing",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					return migrateRun(store, "to", migrate.MigrateOpts{
						TargetVersion: c.Int("version"),
					})
				},
			},
		},
	}
}

func migrateRun(store string, operation string, opts migrate.MigrateOpts) error {
	cfg, err := config.LoadServerConfig(".")
	if err != nil {
		return fmt.Errorf("Failed to load server config: %v", err)
	}

	logging.SetupLogger(cfg.Log)

	migrate.Migrate(cfg, store, operation, migrate.MigrateOpts{})
	return nil
}
