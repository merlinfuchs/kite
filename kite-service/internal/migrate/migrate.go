package migrate

import (
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/merlinfuchs/kite/kite-service/config"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

type MigrateOpts struct {
	TargetVersion int
}

func Migrate(cfg *config.ServerConfig, storeName string, operation string, opts MigrateOpts) {
	// Contextual logger
	l := slog.With("entry", "migrate").With("store", storeName).With("operation", operation)
	l.Debug("Starting migration")

	var migrater store.Migrater

	switch storeName {
	case "postgres":
		pg, err := postgres.New(postgres.BuildConnectionDSN(cfg.Postgres))
		if err != nil {
			l.With(logattr.Error(err)).Error("Failed to create postgres client")
			os.Exit(1)
		}

		pgMigrater, err := pg.GetMigrater()
		if err != nil {
			l.With(logattr.Error(err)).Error("Failed to get migrater")
			os.Exit(1)
		}
		migrater = pgMigrater
		defer migrater.Close()
	default:
		l.Error("Unknown store, can't migrate")
		os.Exit(1)
	}

	migrater.SetLogger(migrationLogger{
		logger:  l,
		verbose: false,
	})

	var err error
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
			l.With(logattr.Error(err)).Error("Invalid target version for migrate")
			os.Exit(1)
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
		l.With(logattr.Error(err)).Error("Migration operation failed")
		os.Exit(1)
	}
}
