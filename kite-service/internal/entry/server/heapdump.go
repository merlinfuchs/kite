package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/config"
)

// watchHeap writes heap and goroutine profiles to cfg.HeapDumpDir when the
// heap passes cfg.HeapDumpThreshold, and again each time it grows by another
// quarter of it. A process that runs out of memory gets killed before anyone
// can take a profile, so these are all that's left of what filled it.
//
// It starts after delay, as the heap passes the threshold for a moment while
// the apps load after a start.
func watchHeap(ctx context.Context, cfg config.DebugConfig, name string, delay time.Duration, interval time.Duration) {
	if cfg.HeapDumpDir == "" {
		return
	}

	select {
	case <-ctx.Done():
		return
	case <-time.After(delay):
	}

	step := max(cfg.HeapDumpThreshold/4, 1)
	next := cfg.HeapDumpThreshold
	var lastDump time.Time

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)

		// Start over once the heap is back to normal, but not more than once
		// an hour, so a heap around the threshold doesn't fill the disk.
		if stats.HeapAlloc < cfg.HeapDumpThreshold && time.Since(lastDump) > time.Hour {
			next = cfg.HeapDumpThreshold
		}
		if stats.HeapAlloc < next {
			continue
		}

		next = stats.HeapAlloc + step
		lastDump = time.Now()

		if err := writeHeapDump(cfg.HeapDumpDir, name, lastDump); err != nil {
			slog.Error("Failed to write heap dump", slog.String("error", err.Error()))
			continue
		}
		slog.Warn(
			"Heap is large, wrote heap dump",
			slog.Uint64("heap_alloc", stats.HeapAlloc),
			slog.String("dir", cfg.HeapDumpDir),
		)
	}
}

func writeHeapDump(dir string, name string, at time.Time) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	stamp := at.UTC().Format("20060102-150405")
	for _, profile := range []string{"heap", "goroutine"} {
		path := filepath.Join(dir, fmt.Sprintf("%s-%s-%s.pb.gz", name, stamp, profile))
		if err := writeProfile(path, profile); err != nil {
			return fmt.Errorf("failed to write %s: %w", path, err)
		}
	}
	return nil
}

func writeProfile(path string, profile string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return err
	}
	if err := pprof.Lookup(profile).WriteTo(f, 0); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
