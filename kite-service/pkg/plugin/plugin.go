package plugin

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/merlinfuchs/kite/go-types/event"
	"github.com/merlinfuchs/kite/go-types/logmodel"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type PluginState string

const PluginStateInit PluginState = "init"
const PluginStateReady PluginState = "ready"
const PluginStateEvent PluginState = "event"

type Plugin struct {
	r wazero.Runtime
	m api.Module

	manifest   Manifest
	config     PluginConfig
	env        HostEnvironment
	apiVersion int

	cancel           context.CancelFunc
	ticker           *time.Ticker
	handleStartAt    time.Time
	hostCallStartAt  time.Time
	hostCallDuration time.Duration

	state                PluginState
	currentCallResponse  []byte
	currentEvent         []byte
	currentEventResponse *event.EventResponse
	currentGuildID       string
}

func New(
	ctx context.Context,
	wasm []byte,
	manifest Manifest,
	config PluginConfig,
	env HostEnvironment,
	compilationCache wazero.CompilationCache,
) (*Plugin, error) {
	r := wazero.NewRuntimeWithConfig(ctx,
		wazero.NewRuntimeConfigCompiler().
			WithCloseOnContextDone(true).
			WithMemoryLimitPages(uint32(config.MemoryPagesLimit)).
			WithCompilationCache(config.CompilationCache).
			WithCompilationCache(compilationCache),
	)

	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	p := &Plugin{
		r:        r,
		manifest: manifest,
		config:   config,
		env:      env,
		state:    PluginStateInit,
	}

	_, err := r.NewHostModuleBuilder("env").
		NewFunctionBuilder().WithFunc(p.kiteGetConfigSize).Export("kite_get_config_size").
		NewFunctionBuilder().WithFunc(p.kiteGetConfig).Export("kite_get_config").
		NewFunctionBuilder().WithFunc(p.kiteLog).Export("kite_log").
		NewFunctionBuilder().WithFunc(p.kiteCall).Export("kite_call").
		NewFunctionBuilder().WithFunc(p.kiteGetEvent).Export("kite_get_event").
		NewFunctionBuilder().WithFunc(p.kiteSetEventResponse).Export("kite_set_event_response").
		NewFunctionBuilder().WithFunc(p.kiteGetCallResponse).Export("kite_get_call_response").
		Instantiate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate host module: %w", err)
	}

	m, err := r.InstantiateWithConfig(ctx, wasm,
		wazero.NewModuleConfig().
			WithStartFunctions().
			WithRandSource(rand.Reader).
			WithSysWalltime().
			WithSysNanosleep().
			WithSysNanotime().
			WithStdout(&pluginLogForwader{
				env:   env,
				level: logmodel.LogLevelDebug,
			}).
			WithStderr(&pluginLogForwader{
				env:   env,
				level: logmodel.LogLevelError,
			}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %w", err)
	}

	p.m = m

	p.apiVersion, err = p.getAPIVersion(ctx)
	if err != nil {
		return nil, err
	}

	if p.apiVersion != 0 {
		return nil, fmt.Errorf("unsupported api version: %d", p.apiVersion)
	}

	// We call _start function manually because we need to call it after the module is instantiated.
	err = p.callStart(ctx)
	if err != nil {
		return nil, err
	}

	p.state = PluginStateReady
	return p, nil
}

func (p *Plugin) Manifest() Manifest {
	return p.manifest
}

func (p *Plugin) Config() PluginConfig {
	return p.config
}

func (p *Plugin) getAPIVersion(ctx context.Context) (int, error) {
	fn := p.m.ExportedFunction("kite_get_api_version")
	if fn == nil {
		return 0, fmt.Errorf("kite_api_version not defined")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Millisecond)
	defer cancel()

	res, err := fn.Call(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to call kite_api_version: %w", err)
	}

	if len(res) != 1 {
		return 0, fmt.Errorf("kite_api_version returned invalid number of results")
	}

	return int(res[0]), nil
}

func (p *Plugin) callStart(ctx context.Context) error {
	fn := p.m.ExportedFunction("_start")
	if fn != nil {
		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		_, err := fn.Call(ctx)
		if err != nil {
			return fmt.Errorf("failed to call _start: %w", err)
		}
	}
	return nil
}

func (p *Plugin) startHostCall() {
	p.hostCallStartAt = time.Now()
}

func (p *Plugin) endHostCall() {
	p.hostCallDuration += time.Since(p.hostCallStartAt)
	p.hostCallStartAt = time.Time{}
}
