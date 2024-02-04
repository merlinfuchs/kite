package log

import (
	"github.com/merlinfuchs/kite/kite-sdk-go/internal"
	"github.com/merlinfuchs/kite/kite-types/logmodel"
)

func Log(level logmodel.LogLevel, msg string) {
	internal.Log(int(level), msg)
}

func Debug(msg string) {
	internal.Log(int(logmodel.LogLevelDebug), msg)
}

func Info(msg string) {
	internal.Log(int(logmodel.LogLevelInfo), msg)
}

func Warn(msg string) {
	internal.Log(int(logmodel.LogLevelWarn), msg)
}

func Error(msg string) {
	internal.Log(int(logmodel.LogLevelError), msg)
}
