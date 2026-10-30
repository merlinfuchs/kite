package database

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/db/s3"
	"github.com/kitecloud/kite/kite-service/internal/logging"
)

type DatabaseBackupOpts struct {
	Operation string
	Name      string
}

func RunBackup(ctx context.Context, db string, opts DatabaseBackupOpts) error {
	if db != "postgres" {
		return fmt.Errorf("database %s is not supported for backup", db)
	}

	if opts.Operation != "create" {
		return fmt.Errorf("operation %s is not supported for backup", opts.Operation)
	}

	cfg, err := config.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("Failed to load server config: %v", err)
	}

	logging.SetupLogger(cfg.Logging)

	s3.RetryLonger()
	s3, err := s3.New(cfg.Database.S3)
	if err != nil {
		slog.With("error", err).Error("Failed to create s3 client")
		return fmt.Errorf("failed to create s3 client: %w", err)
	}

	slog.Info("Creating database backup", "database", db, "operation", opts.Operation)

	// A failed upload leaves the previous backup of the same name in place,
	// as S3 only replaces it when the upload completes. Parts are retried on
	// their own, this is for when the storage is down for longer.
	for attempt := 1; ; attempt++ {
		err = dumpAndStore(ctx, s3, cfg.Database.Postgres, db, opts.Name)
		if err == nil {
			return nil
		}
		if attempt == backupAttempts {
			return err
		}

		slog.Error("Failed to create database backup, retrying", "attempt", attempt, "error", err)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backupRetryDelay):
		}
	}
}

const (
	backupAttempts   = 3
	backupRetryDelay = 5 * time.Minute
)

// dumpAndStore streams the dump through gzip into S3. Nothing is written to
// local disk, as writing the dump there slowed down Postgres on the same disk
// enough for queries to time out.
func dumpAndStore(ctx context.Context, s3 *s3.Client, pgConfig config.PostgresConfig, db string, name string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pg_dump",
		"--host="+pgConfig.Host,
		"--username="+pgConfig.User,
		"--port="+fmt.Sprintf("%d", pgConfig.Port),
		"--dbname="+pgConfig.DBName,
		"--format=tar",
	)

	if pgConfig.Password != "" {
		cmd.Env = append(os.Environ(), "PGPASSWORD="+pgConfig.Password)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get dump output: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start database dump: %w", err)
	}

	pr, pw := io.Pipe()
	var size int64

	done := make(chan struct{})
	go func() {
		defer close(done)

		gzipWriter := gzip.NewWriter(pw)
		n, err := io.Copy(gzipWriter, stdout)
		size = n
		if err == nil {
			err = gzipWriter.Close()
		}
		if waitErr := cmd.Wait(); err == nil && waitErr != nil {
			err = fmt.Errorf("failed to create database dump: %w\noutput: %s", waitErr, stderr.String())
		}

		// An error here fails the upload instead of storing a cut off dump.
		pw.CloseWithError(err)
	}()

	slog.Info("Storing compressed database backup", "name", name)

	compressedSize, err := s3.StoreDBBackup(ctx, db, name, pr)
	if err != nil {
		// Stops pg_dump, and unblocks the goroutine if it's still writing.
		cancel()
		pr.CloseWithError(err)
	}
	<-done

	if err != nil {
		return err
	}

	slog.Info("Successfully stored compressed db backup", "name", name, "original_size", size, "compressed_size", compressedSize)

	return nil
}
