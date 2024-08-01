package cmd

import (
	"fmt"

	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres"
	"github.com/kitecloud/kite/kite-service/internal/logging"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/urfave/cli/v2"
)

var databases = []string{"postgres"}

var databaseCMD cli.Command

func init() {
	migrateCommands := []*cli.Command{}
	for _, db := range databases {
		migrateCommands = append(migrateCommands, &cli.Command{
			Name:  db,
			Usage: fmt.Sprintf("Run migrations against the %s database.", db),
			Args:  true,
			Subcommands: []*cli.Command{
				{
					Name:  "up",
					Usage: "Migrate the database to the latest version.",
					Action: func(c *cli.Context) error {
						return runDatabaseMigrations(db, "up", databaseMigrationOpts{})
					},
				},
				{
					Name:  "down",
					Usage: "Rollback the database to the earliest version.",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  "danger",
							Usage: "Confirm that you want to run this command.",
						},
					},
					Action: func(c *cli.Context) error {
						if !c.Bool("danger") {
							return fmt.Errorf("this command is dangerous, use --danger flag to confirm")
						}

						return runDatabaseMigrations(db, "down", databaseMigrationOpts{})
					},
				},
				{
					Name:  "version",
					Usage: "Print the current database version.",
					Action: func(c *cli.Context) error {
						return runDatabaseMigrations(db, "version", databaseMigrationOpts{})
					},
				},
				{
					Name:  "list",
					Usage: "List all available database migrations.",
					Action: func(c *cli.Context) error {
						return runDatabaseMigrations(db, "list", databaseMigrationOpts{})
					},
				},
				{
					Name:  "force",
					Usage: "Force a specific migration version.",
					Flags: []cli.Flag{
						&cli.IntFlag{
							Name:  "version",
							Usage: "The target version to force to.",
						},
						&cli.BoolFlag{
							Name:  "danger",
							Usage: "Confirm that you want to run this command.",
						},
					},
					Action: func(c *cli.Context) error {
						if !c.Bool("danger") {
							return fmt.Errorf("this command is dangerous, use --danger flag to confirm")
						}

						return runDatabaseMigrations(db, "force", databaseMigrationOpts{
							TargetVersion: c.Int("version"),
						})
					},
				},
				{
					Name:  "to",
					Usage: "Migrate the database to a specific version.",
					Flags: []cli.Flag{
						&cli.IntFlag{
							Name:  "version",
							Usage: "The target version to migrate to.",
						},
						&cli.BoolFlag{
							Name:  "danger",
							Usage: "Confirm that you want to run this command.",
						},
					},
					Action: func(c *cli.Context) error {
						if !c.Bool("danger") {
							return fmt.Errorf("this command is dangerous, use --danger flag to confirm")
						}

						return runDatabaseMigrations(db, "to", databaseMigrationOpts{
							TargetVersion: c.Int("version"),
						})
					},
				},
			},
		})
	}

	databaseCMD = cli.Command{
		Name:  "database",
		Usage: "Manage and migrate databases used by Kite.",
		Subcommands: []*cli.Command{
			{
				Name:        "migrate",
				Description: "Run database migrations.",
				Subcommands: migrateCommands,
			},
		},
	}
}

type databaseMigrationOpts struct {
	TargetVersion int
}

func runDatabaseMigrations(database string, operation string, opts databaseMigrationOpts) error {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("Failed to load server config: %v", err)
	}

	logging.SetupLogger(cfg.Logging)

	l := slog.With("database", database).With("operation", operation)

	var migrater store.MigrationStore

	switch database {
	case "postgres":
		pg, err := postgres.New(postgres.BuildConnectionDSN(cfg.Database.Postgres))
		if err != nil {
			l.With("error", err).Error("Failed to create postgres client")
			os.Exit(1)
		}

		pgMigrater, err := pg.GetMigrater()
		if err != nil {
			l.With("error", err).Error("Failed to get migrater")
			os.Exit(1)
		}
		migrater = pgMigrater
		defer migrater.Close()
	default:
		return fmt.Errorf("unsupported database: %s", database)
	}

	migrater.SetLogger(databaseMigrationLogger{
		logger: l,
	})

	switch operation {
	case "up":
		err = migrater.Up()
	case "down":
		err = migrater.Down()
	case "list":
		var migrations []string
		migrations, err = migrater.List()
		if err != nil {
			break
		}
		l.With("migrations", migrations).Info("")
	case "version":
		var version uint
		var dirty bool
		version, dirty, err = migrater.Version()
		if err != nil {
			break
		}
		l.With("version", version).With("dirty", dirty).Info("")

	case "force":
		l = l.With("target_version", opts.TargetVersion)
		err = migrater.Force(opts.TargetVersion)
		if err != nil {
			break
		}
	case "to":
		l = l.With("target_version", opts.TargetVersion)
		if opts.TargetVersion < 0 {
			l.With("error", err).Error("Invalid target version for migrate")
			return err
		}
		err = migrater.To(uint(opts.TargetVersion))
		if err != nil {
			break
		}
	}

	if err == migrate.ErrNoChange {
		l.Warn("Already at the correct version, migration was skipped")
	} else if err == migrate.ErrNilVersion {
		l.Warn("Migration is at nil version (no migrations have been performed)")
	} else if err != nil {
		l.With("error", err).Error("Migration operation failed")
		return err
	}

	l.Info("Migration completed")

	return nil
}

type databaseMigrationLogger struct {
	logger  *slog.Logger
	verbose bool
}

// Printf is like fmt.Printf
func (ml databaseMigrationLogger) Printf(format string, v ...interface{}) {
	ml.logger.Info(fmt.Sprintf(format, v...))
}

// Printf is like fmt.Printf
func (ml databaseMigrationLogger) Verbose() bool {
	return ml.verbose
}
