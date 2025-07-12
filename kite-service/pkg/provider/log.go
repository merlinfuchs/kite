package provider

import "context"

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// LogProvider provides access to logging.
type LogProvider interface {
	CreateLogEntry(ctx context.Context, level LogLevel, message string)
}

type MockLogProvider struct{}

func (p *MockLogProvider) CreateLogEntry(ctx context.Context, level LogLevel, message string) {}
