package log

import (
	"github.com/merlinfuchs/kite/kite-sdk-go/internal"
)

func Log(level LogLevel, msg string) {
	internal.Log(int(level), msg)
}

func Debug(msg string) {
	internal.Log(int(LogLevelDebug), msg)
}

func Info(msg string) {
	internal.Log(int(LogLevelInfo), msg)
}

func Warn(msg string) {
	internal.Log(int(LogLevelWarn), msg)
}

func Error(msg string) {
	internal.Log(int(LogLevelError), msg)
}
