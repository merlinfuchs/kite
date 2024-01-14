package logging

import (
	"io"
	"os"

	"log/slog"

	"github.com/cyrusaf/ctxlog"
	"github.com/endobit/clog"
	"github.com/merlinfuchs/kite/kite-service/config"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"gopkg.in/natefinch/lumberjack.v2"
)

func getLogWriter(cfg config.ServerLogConfig) io.Writer {
	logWriters := make([]io.Writer, 0)
	logWriters = append(logWriters, os.Stdout)

	if cfg.Filename != "" {
		lj := lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
		}
		logWriters = append(logWriters, &lj)
	}
	writer := io.MultiWriter(logWriters...)
	return writer
}

func SetupLogger(cfg config.ServerLogConfig) *slog.Logger {
	writer := getLogWriter(cfg)

	handler := ctxlog.NewHandler(clog.NewHandler(writer))

	logger := slog.New(handler)
	hostname, err := os.Hostname()
	if err != nil {
		logger.Error("failed to get hostname", logattr.Error(err))
		hostname = ""
	}
	logger = logger.With(slog.String("host", hostname))

	slog.SetDefault(logger)
	return logger
}
