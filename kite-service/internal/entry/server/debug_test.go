package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/metrics"
)

// freePort asks the OS for an unused port and immediately gives it back. There
// is a small race before the server binds it, which is acceptable in a test.
func freePort(t *testing.T) int {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find a free port: %v", err)
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port
}

func waitForServer(t *testing.T, url string) {
	t.Helper()

	for range 100 {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("debug server never became reachable at %s", url)
}

// The whole point of the debug server is to be there when the service is
// misbehaving in production, so verify it actually serves both surfaces rather
// than just that it compiles.
func TestDebugServerServesVarsAndPprof(t *testing.T) {
	port := freePort(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startDebugServer(ctx, config.DebugConfig{
		Enabled: true,
		Host:    "127.0.0.1",
		Port:    port,
	})

	base := fmt.Sprintf("http://127.0.0.1:%d", port)
	waitForServer(t, base+"/debug/vars")

	t.Run("expvar reports our counters", func(t *testing.T) {
		metrics.GatewayEvents.Add("MESSAGE_CREATE", 7)

		resp, err := http.Get(base + "/debug/vars")
		if err != nil {
			t.Fatalf("GET /debug/vars: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET /debug/vars = %d, want 200", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		var vars map[string]any
		if err := json.Unmarshal(body, &vars); err != nil {
			t.Fatalf("body is not valid JSON: %v", err)
		}

		// The counters the deploy is meant to be observed through.
		for _, key := range []string{
			"gateway_events_total",
			"gateway_connections_active",
			"db_poll_count",
			"lock_wait_slow_total",
			"go_goroutines",
			"memstats",
		} {
			if _, ok := vars[key]; !ok {
				t.Errorf("expvar output is missing %q", key)
			}
		}

		events, _ := vars["gateway_events_total"].(map[string]any)
		if events["MESSAGE_CREATE"] == nil {
			t.Error("gateway_events_total did not report the event type we incremented")
		}
	})

	t.Run("pprof profiles are reachable", func(t *testing.T) {
		for _, path := range []string{"/debug/pprof/", "/debug/pprof/cmdline"} {
			resp, err := http.Get(base + path)
			if err != nil {
				t.Fatalf("GET %s: %v", path, err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("GET %s = %d, want 200", path, resp.StatusCode)
			}
		}
	})

	t.Run("heap profile downloads", func(t *testing.T) {
		resp, err := http.Get(base + "/debug/pprof/heap?debug=1")
		if err != nil {
			t.Fatalf("GET heap profile: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read heap profile: %v", err)
		}
		if len(body) == 0 {
			t.Error("heap profile was empty")
		}
	})
}

// Disabled is the default, and it must mean no listener at all rather than a
// listener that refuses requests.
func TestDebugServerDisabledDoesNotListen(t *testing.T) {
	port := freePort(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startDebugServer(ctx, config.DebugConfig{
		Enabled: false,
		Host:    "127.0.0.1",
		Port:    port,
	})

	time.Sleep(100 * time.Millisecond)

	if _, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 200*time.Millisecond); err == nil {
		t.Error("debug server is listening despite being disabled")
	}
}

func TestDebugServerShutsDownWithContext(t *testing.T) {
	port := freePort(t)
	ctx, cancel := context.WithCancel(context.Background())

	startDebugServer(ctx, config.DebugConfig{
		Enabled: true,
		Host:    "127.0.0.1",
		Port:    port,
	})

	base := fmt.Sprintf("http://127.0.0.1:%d", port)
	waitForServer(t, base+"/debug/vars")

	cancel()

	for range 100 {
		if _, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 100*time.Millisecond); err != nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Error("debug server still listening after context cancellation")
}
