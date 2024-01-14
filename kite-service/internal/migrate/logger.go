package migrate

import "log/slog"

type migrationLogger struct {
	logger  *slog.Logger
	verbose bool
}

// Printf is like fmt.Printf.
func (ml migrationLogger) Printf(format string, v ...interface{}) {
	ml.logger.Info(format, v...)
}

// Printf is like fmt.Printf.
func (ml migrationLogger) Verbose() bool {
	return ml.verbose
}
