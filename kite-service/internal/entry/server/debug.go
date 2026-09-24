package server

import (
	"context"
	"expvar"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/config"
)

// startDebugServer serves pprof profiles and expvar counters on a dedicated
// listener.
//
// This is intentionally kept off the API server's router: the profiles are not
// access-controlled and would expose heap contents and goroutine stacks to
// anyone who can reach them. Bind it to a private interface only.
func startDebugServer(ctx context.Context, cfg config.DebugConfig) {
	if !cfg.Enabled {
		return
	}

	mux := http.NewServeMux()

	mux.Handle("GET /debug/vars", expvar.Handler())

	mux.HandleFunc("GET /debug/pprof/", pprof.Index)
	mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)

	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:    address,
		Handler: mux,
		// Profiles are collected over a caller-supplied duration, so the read
		// timeout has to be generous enough for a 30s CPU profile.
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Minute,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("Failed to shut down debug server", slog.String("error", err.Error()))
		}
	}()

	go func() {
		slog.Info("Starting debug server", slog.String("address", address))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error(
				"Debug server stopped",
				slog.String("address", address),
				slog.String("error", err.Error()),
			)
		}
	}()
}
