package plugin

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/merlinfuchs/kite/go-types/event"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type PluginState string

const PluginStateInit PluginState = "init"
const PluginStateReady PluginState = "ready"
const PluginStateEvent PluginState = "event"

type Plugin struct {
	sync.RWMutex

	r        wazero.Runtime
	m        api.Module
	manifest PluginManifest
	config   PluginConfig
	env      HostEnvironment

	ctx              context.Context
	cancel           context.CancelFunc
	ticker           *time.Ticker
	handleStartAt    time.Time
	hostCallStartAt  time.Time
	hostCallDuration time.Duration

	state                PluginState
	currentCallResponse  []byte
	currentEvent         []byte
	currentEventResponse *event.EventResponse
}

func New(ctx context.Context, wasm []byte, manifest PluginManifest, config PluginConfig, env HostEnvironment) (*Plugin, error) {
	r := wazero.NewRuntimeWithConfig(ctx,
		wazero.NewRuntimeConfigCompiler().
			WithCloseOnContextDone(true).
			WithMemoryLimitPages(uint32(config.MemoryPagesLimit)).
			WithCompilationCache(config.CompilationCache),
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
			WithRandSource(rand.Reader),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %w", err)
	}

	p.m = m

	// We call _start function manually because we need to call it after the module is instantiated.
	fn := p.m.ExportedFunction("_start")
	if fn != nil {
		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		_, err = fn.Call(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to call _start: %w", err)
		}
	}

	p.state = PluginStateReady
	return p, nil
}

func (p *Plugin) Manifest() PluginManifest {
	return p.manifest
}

func (p *Plugin) Config() PluginConfig {
	return p.config
}

func (p *Plugin) Handle(ctx context.Context, e *event.Event) error {
	// TODO: when execution has been aborted with the context the plugin instance is unusable and must be replaced
	if p.config.TotalTimeLimit != 0 {
		ctx, p.cancel = context.WithTimeout(ctx, p.config.TotalTimeLimit)
	} else {
		ctx, p.cancel = context.WithCancel(ctx)
	}
	p.ctx = ctx

	raw, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	p.currentEvent = raw

	p.startHandle()
	defer p.endHandle()

	fn := p.m.ExportedFunction("kite_handle")
	if fn == nil {
		return fmt.Errorf("kite_handle not defined")
	}

	_, err = fn.Call(ctx, uint64(len(raw)))
	if err != nil {
		return fmt.Errorf("failed to call kite_handle: %w", err)
	}

	if p.currentEventResponse == nil {
		return nil
	}

	if !p.currentEventResponse.Success {
		err := p.currentEventResponse.Error
		p.currentEventResponse = nil
		return err
	}

	p.currentEventResponse = nil
	return nil
}

func (p *Plugin) Close(ctx context.Context) error {
	if err := p.m.Close(ctx); err != nil {
		return err
	}
	return p.r.Close(ctx)
}

func (p *Plugin) startHandle() {
	p.state = PluginStateEvent
	p.ticker = time.NewTicker(time.Millisecond * 5)
	p.handleStartAt = time.Now()
	p.hostCallDuration = 0

	if p.config.ExecutionTimeLimit != 0 {
		go func() {
			for {
				_, ok := <-p.ticker.C
				if !ok {
					return
				}

				totalDuration := time.Since(p.handleStartAt)

				hostCallDuration := p.hostCallDuration
				if !p.hostCallStartAt.IsZero() {
					hostCallDuration += time.Since(p.hostCallStartAt)
				}

				executionDuration := totalDuration - hostCallDuration
				if executionDuration > p.config.ExecutionTimeLimit {
					p.cancel()
				}
			}
		}()
	}
}

func (p *Plugin) endHandle() {
	p.state = PluginStateReady
	p.ticker.Stop()
}

func (p *Plugin) startHostCall() {
	p.hostCallStartAt = time.Now()
}

func (p *Plugin) endHostCall() {
	p.hostCallDuration += time.Since(p.hostCallStartAt)
	p.hostCallStartAt = time.Time{}
}
