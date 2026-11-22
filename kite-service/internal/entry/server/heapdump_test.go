package server

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/config"
)

func TestWatchHeapWritesProfiles(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go watchHeap(ctx, config.DebugConfig{
		HeapDumpDir: dir,
		// Every process has more heap than this.
		HeapDumpThreshold: 1,
	}, "cluster-0", 0, 10*time.Millisecond)

	for range 200 {
		heap, _ := filepath.Glob(filepath.Join(dir, "cluster-0-*-heap.pb.gz"))
		goroutine, _ := filepath.Glob(filepath.Join(dir, "cluster-0-*-goroutine.pb.gz"))
		if len(heap) > 0 && len(goroutine) > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("no heap and goroutine profiles were written")
}

func TestWatchHeapOffWithoutDir(t *testing.T) {
	done := make(chan struct{})
	go func() {
		watchHeap(context.Background(), config.DebugConfig{HeapDumpThreshold: 1}, "cluster-0", 0, time.Millisecond)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watchHeap kept running without heap_dump_dir")
	}
}

func TestWatchHeapWaitsForDelay(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go watchHeap(ctx, config.DebugConfig{
		HeapDumpDir:       dir,
		HeapDumpThreshold: 1,
	}, "cluster-0", time.Hour, time.Millisecond)

	time.Sleep(100 * time.Millisecond)
	if dumps, _ := filepath.Glob(filepath.Join(dir, "*")); len(dumps) > 0 {
		t.Fatalf("wrote %v before the delay", dumps)
	}
}
