package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/session"
	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres"
	"github.com/kitecloud/kite/kite-service/internal/db/s3"
	"github.com/kitecloud/kite/kite-service/internal/logging"
	"github.com/urfave/cli/v2"
)

var adminCMD = cli.Command{
	Name:  "admin",
	Usage: "Admin commands used for debugging and administration",
	Subcommands: []*cli.Command{
		{
			Name:   "impersonate",
			Usage:  "Impersonate a user by creating a session token that can be injected into the session cookie",
			Action: adminImpersonateCMD,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "user_id",
					Usage: "The ID of the user to impersonate.",
				},
			},
		},
		// Temporary: apps deleted before their image files were deleted with
		// them left those files behind. Remove once it ran in production.
		{
			Name:   "cleanup-assets",
			Usage:  "List image files that no asset uses anymore, and delete them with --delete",
			Action: adminCleanupAssetsCMD,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "delete",
					Usage: "Delete the files instead of only listing them.",
				},
			},
		},
	},
}

func adminImpersonateCMD(c *cli.Context) error {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logging.SetupLogger(cfg.Logging)

	pg, err := postgres.New(postgres.BuildConnectionDSN(cfg.Database.Postgres), 1)
	if err != nil {
		slog.Error(
			"Failed to create postgres client",
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to create postgres client: %w", err)
	}

	sessionManager := session.NewSessionManager(session.SessionManagerConfig{}, pg)

	key, _, err := sessionManager.CreateSession(context.Background(), c.String("user_id"))
	if err != nil {
		slog.Error(
			"Failed to create session",
			slog.String("error", err.Error()),
			slog.String("user_id", c.String("user_id")),
		)
		return fmt.Errorf("failed to create session: %w", err)
	}

	fmt.Println(key)
	return nil
}

// Files are uploaded before their asset is inserted, so recent ones may belong
// to an upload that is still in progress.
const cleanupAssetsMinAge = time.Hour

func adminCleanupAssetsCMD(c *cli.Context) error {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logging.SetupLogger(cfg.Logging)

	pg, err := postgres.New(postgres.BuildConnectionDSN(cfg.Database.Postgres), 1)
	if err != nil {
		return fmt.Errorf("failed to create postgres client: %w", err)
	}

	objectStore, err := s3.New(cfg.Database.S3)
	if err != nil {
		return fmt.Errorf("failed to create s3 client: %w", err)
	}

	ctx := c.Context
	objects, err := objectStore.ListObjects(ctx, postgres.AssetBucketName)
	if err != nil {
		return err
	}

	contentHashes, err := pg.Q.GetAllAssetContentHashes(ctx)
	if err != nil {
		return fmt.Errorf("failed to get asset content hashes: %w", err)
	}
	used := make(map[string]bool, len(contentHashes))
	for _, contentHash := range contentHashes {
		used[contentHash] = true
	}

	var unused []s3.ObjectInfo
	var unusedSize int64
	for _, object := range objects {
		if used[object.Name] || time.Since(object.LastModified) < cleanupAssetsMinAge {
			continue
		}
		unused = append(unused, object)
		unusedSize += object.Size
	}

	fmt.Printf("%d files, %d used by assets, %d unused (%.1f MB)\n", len(objects), len(contentHashes), len(unused), float64(unusedSize)/1e6)
	if !c.Bool("delete") {
		fmt.Println("Run with --delete to delete the unused files.")
		return nil
	}

	deleted := 0
	for _, object := range unused {
		if err := objectStore.DeleteObject(ctx, postgres.AssetBucketName, object.Name); err != nil {
			slog.Error("Failed to delete file", slog.String("name", object.Name), slog.String("error", err.Error()))
			continue
		}
		deleted++
	}
	fmt.Printf("Deleted %d files\n", deleted)
	return nil
}
