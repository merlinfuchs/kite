package logattr

import "log/slog"

const ErrorKey = "error"

func Error(err error) slog.Attr {
	return slog.String("error", err.Error())
}
