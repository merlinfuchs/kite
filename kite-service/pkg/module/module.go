package module

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-sdk-go/event"
	"github.com/merlinfuchs/kite/kite-sdk-go/log"
	"github.com/merlinfuchs/kite/kite-sdk-go/manifest"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type ModuleState string

const (
	ModuleStateInit  ModuleState = "init"
	ModuleStateReady ModuleState = "ready"
	ModuleStateEvent ModuleState = "event"
)

type Module struct {
	r wazero.Runtime
	m api.Module

	config      ModuleConfig
	env         HostEnvironment
	apiVersion  manifest.APIVersion
	apiEncoding manifest.APIEncoding

	cancel           context.CancelFunc
	ticker           *time.Ticker
	handleStartAt    time.Time
	hostCallStartAt  time.Time
	hostCallDuration time.Duration

	state                ModuleState
	currentManifest      manifest.Manifest
	currentCallResponse  []byte
	currentEvent         []byte
	currentEventResponse *event.EventResponse
	currentGuildID       distype.Snowflake
}

func New(
	ctx context.Context,
	wasm []byte,
	config ModuleConfig,
	env HostEnvironment,
) (*Module, error) {
	r := wazero.NewRuntimeWithConfig(ctx,
		// At some point we probably want to switch to the optimization compiler
		// right not it doesn't seem to be faster for short living functions
		// https://github.com/tetratelabs/wazero/issues/1496
		// opt.NewRuntimeConfigOptimizingCompiler().
		wazero.NewRuntimeConfigCompiler().
			WithCloseOnContextDone(true).
			WithMemoryLimitPages(uint32(config.MemoryPagesLimit)).
			WithCompilationCache(config.CompilationCache),
	)

	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	p := &Module{
		r:      r,
		config: config,
		env:    env,
		state:  ModuleStateInit,
	}

	_, err := r.NewHostModuleBuilder("env").
		NewFunctionBuilder().WithFunc(p.kiteSetManifest).Export("kite_set_manifest").
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
			WithStdout(&moduleLogForwarder{
				env:   env,
				level: log.LogLevelDebug,
			}).
			WithStderr(&moduleLogForwarder{
				env:   env,
				level: log.LogLevelError,
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
	if p.apiVersion != manifest.APIVersionAlpha {
		return nil, fmt.Errorf("unsupported api version: %d", p.apiVersion)
	}

	p.apiEncoding, err = p.getAPIEncoding(ctx)
	if err != nil {
		return nil, err
	}
	if p.apiEncoding != manifest.APIEncodingJSON {
		return nil, fmt.Errorf("unsupported api encoding: %d", p.apiEncoding)
	}

	p.state = ModuleStateReady
	return p, nil
}

func (p *Module) Config() ModuleConfig {
	return p.config
}

func (p *Module) startHostCall() {
	p.hostCallStartAt = time.Now()
}

func (p *Module) endHostCall() {
	p.hostCallDuration += time.Since(p.hostCallStartAt)
	p.hostCallStartAt = time.Time{}
}
