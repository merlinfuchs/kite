package plugin

import (
	"time"

	"github.com/tetratelabs/wazero"
)

type PluginConfig struct {
	MemoryPagesLimit   int
	ExecutionTimeLimit time.Duration
	TotalTimeLimit     time.Duration
	CompilationCache   wazero.CompilationCache

	UserConfig map[string]string
}
