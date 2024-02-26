package module

import (
	"time"

	"github.com/tetratelabs/wazero"
)

type ModuleConfig struct {
	MemoryPagesLimit   int
	ExecutionTimeLimit time.Duration
	TotalTimeLimit     time.Duration
	HostCallLimit      int
	CompilationCache   wazero.CompilationCache
}
