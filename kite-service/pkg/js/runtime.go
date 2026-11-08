package js

import (
	"context"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
	// formatDate takes time zones by name, and the service image has no
	// zoneinfo. Here, not in main, so the package works on its own.
	_ "time/tzdata"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
)

// kite.wasm is QuickJS-NG with the glue in wasm/kite.c, built by
// wasm/build.sh.
//
//go:embed kite.wasm
var wasmBytes []byte

//go:embed prelude.js
var prelude []byte

// engine holds the compiled module and the memory of an instance right after
// QuickJS and the prelude are set up, which every run starts from.
type engine struct {
	rt       wazero.Runtime
	compiled wazero.CompiledModule
	snapshot []byte
}

func newEngine(ctx context.Context) (*engine, error) {
	// The context kill has to be on: QuickJS only calls the interrupt handler
	// from its interpreter loop and some builtins, so a native loop like
	// new Array(2**28).join() runs on until wazero closes the module. It also
	// lets the Go GC stop the goroutine.
	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfigCompiler().
		WithMemoryLimitPages(maxPages).
		WithCloseOnContextDone(true))

	if err := instantiateHost(ctx, rt); err != nil {
		rt.Close(ctx)
		return nil, err
	}
	compiled, err := rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		rt.Close(ctx)
		return nil, fmt.Errorf("failed to compile QuickJS: %w", err)
	}
	e := &engine{rt: rt, compiled: compiled}

	mod, err := rt.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().WithName("").WithStartFunctions("_initialize"))
	if err != nil {
		rt.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate QuickJS: %w", err)
	}
	defer mod.Close(ctx)
	inst := &instance{mod: mod}
	ptr, err := inst.write(ctx, prelude)
	if err != nil {
		rt.Close(ctx)
		return nil, err
	}
	res, err := mod.ExportedFunction("kite_init").Call(ctx, MemoryLimit, stackSize, uint64(ptr), uint64(len(prelude)))
	if err != nil {
		rt.Close(ctx)
		return nil, fmt.Errorf("failed to set up QuickJS: %w", err)
	}
	if int32(res[0]) != 0 {
		out, _ := inst.output(ctx)
		rt.Close(ctx)
		return nil, fmt.Errorf("failed to set up QuickJS: %d %s", int32(res[0]), out)
	}
	inst.free(ctx, ptr)

	// Copying the memory is a full snapshot because the only mutable global
	// is the stack pointer, which is back at its start between calls.
	mem, _ := mod.Memory().Read(0, mod.Memory().Size())
	e.snapshot = append([]byte(nil), mem...)
	return e, nil
}

func (e *engine) newInstance(ctx context.Context) (*instance, error) {
	// The memory starts with the capacity of the snapshot, so growing it to
	// the snapshot's size doesn't copy it. Without that a run took 3 times as
	// long on Linux, see design/javascript.md.
	allocCtx := experimental.WithMemoryAllocator(ctx, snapshotAllocator{capacity: uint64(len(e.snapshot))})
	mod, err := e.rt.InstantiateModule(allocCtx, e.compiled, wazero.NewModuleConfig().WithName("").WithStartFunctions())
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate QuickJS: %w", err)
	}
	mem := mod.Memory()
	if pages := uint32(len(e.snapshot)) / 65536; pages > mem.Size()/65536 {
		if _, ok := mem.Grow(pages - mem.Size()/65536); !ok {
			mod.Close(ctx)
			return nil, errors.New("failed to grow memory")
		}
	}
	mem.Write(0, e.snapshot)
	return &instance{mod: mod}, nil
}

// snapshotAllocator allocates the memory of an instance with at least
// capacity bytes of room.
type snapshotAllocator struct {
	capacity uint64
}

func (a snapshotAllocator) Allocate(capacity, maxBytes uint64) experimental.LinearMemory {
	return &sliceMemory{buf: make([]byte, 0, min(maxBytes, max(capacity, a.capacity)))}
}

// sliceMemory is the memory of an instance, like wazero's own but with the
// capacity of the allocator.
type sliceMemory struct {
	buf []byte
}

func (m *sliceMemory) Reallocate(size uint64) []byte {
	if size > uint64(cap(m.buf)) {
		m.buf = append(m.buf[:cap(m.buf)], make([]byte, size-uint64(cap(m.buf)))...)
	}
	m.buf = m.buf[:size]
	return m.buf
}

func (m *sliceMemory) Free() {}

type instance struct {
	mod api.Module
}

func (i *instance) close(ctx context.Context) { i.mod.Close(ctx) }

// write copies b into the instance with a null after it, which QuickJS needs
// after code and JSON.
func (i *instance) write(ctx context.Context, b []byte) (uint32, error) {
	res, err := i.mod.ExportedFunction("kite_alloc").Call(ctx, uint64(len(b)+1))
	if err != nil {
		return 0, err
	}
	ptr := uint32(res[0])
	if ptr == 0 {
		return 0, errors.New("out of memory")
	}
	i.mod.Memory().Write(ptr, append(b, 0))
	return ptr, nil
}

func (i *instance) free(ctx context.Context, ptr uint32) {
	i.mod.ExportedFunction("kite_free").Call(ctx, uint64(ptr))
}

func (i *instance) output(ctx context.Context) ([]byte, error) {
	ptr, err := i.mod.ExportedFunction("kite_out_ptr").Call(ctx)
	if err != nil {
		return nil, err
	}
	n, err := i.mod.ExportedFunction("kite_out_len").Call(ctx)
	if err != nil {
		return nil, err
	}
	out, ok := i.mod.Memory().Read(uint32(ptr[0]), uint32(n[0]))
	if !ok {
		return nil, errors.New("invalid output")
	}
	return append([]byte(nil), out...), nil
}

// run returns the JSON of the script's return value, or the error message the
// script threw for runError. An error means the instance is dead. The
// time limit starts when the script does, so a slow start on a busy machine
// doesn't count against it.
func (i *instance) run(ctx context.Context, code []byte, input []byte, timeout time.Duration) (out []byte, status runStatus, err error) {
	codePtr, err := i.write(ctx, code)
	if err != nil {
		return nil, 0, err
	}
	inputPtr, err := i.write(ctx, input)
	if err != nil {
		return nil, 0, err
	}

	stateFrom(ctx).deadline = time.Now().Add(timeout)
	runCtx, cancel := context.WithTimeout(ctx, timeout+killGrace)
	defer cancel()
	res, err := i.mod.ExportedFunction("kite_run").Call(runCtx, uint64(codePtr), uint64(len(code)), uint64(inputPtr), uint64(len(input)))
	if err != nil {
		if ctx.Err() == nil && runCtx.Err() != nil {
			return nil, 0, errTimeout
		}
		return nil, 0, err
	}
	out, err = i.output(ctx)
	if err != nil {
		return nil, 0, err
	}
	return out, runStatus(res[0]), nil
}

// runStatus is what kite_run returns.
type runStatus int32

const (
	runResult runStatus = iota
	runError
	// The promise the code returned is still pending, with nothing left to
	// run.
	runPending
)

// runState is what the host functions of one run see, passed in the context
// of the calls into the instance.
type runState struct {
	deadline time.Time
	log      func(level LogLevel, message string)
	pending  []byte
}

// logLevels are the levels kite.c logs with, by its number for them.
var logLevels = []LogLevel{LogLevelInfo, LogLevelWarn, LogLevelError}

type stateKey struct{}

func withState(ctx context.Context, st *runState) context.Context {
	return context.WithValue(ctx, stateKey{}, st)
}

func stateFrom(ctx context.Context) *runState {
	st, _ := ctx.Value(stateKey{}).(*runState)
	return st
}

func instantiateHost(ctx context.Context, rt wazero.Runtime) error {
	_, err := rt.NewHostModuleBuilder("kite").
		NewFunctionBuilder().WithFunc(func(ctx context.Context) int32 {
		if st := stateFrom(ctx); st != nil && time.Now().After(st.deadline) {
			return 1
		}
		return 0
	}).Export("interrupt").
		NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, level int32, p, n uint32) {
		st := stateFrom(ctx)
		if st == nil || level < 0 || int(level) >= len(logLevels) {
			return
		}
		b, _ := m.Memory().Read(p, n)
		st.log(logLevels[level], string(b))
	}).Export("log").
		NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, namePtr, nameLen, argPtr, argLen uint32) int32 {
		st := stateFrom(ctx)
		if st == nil {
			return 0
		}
		name, _ := m.Memory().Read(namePtr, nameLen)
		arg, _ := m.Memory().Read(argPtr, argLen)
		res, err := callHost(string(name), arg)
		if err != nil {
			st.pending = []byte(err.Error())
			return -int32(len(st.pending))
		}
		st.pending = res
		return int32(len(res))
	}).Export("call").
		NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, dst uint32) {
		if st := stateFrom(ctx); st != nil {
			m.Memory().Write(dst, st.pending)
			st.pending = nil
		}
	}).Export("take").
		NewFunctionBuilder().WithFunc(func() float64 { return rand.Float64() }).Export("random").
		Instantiate(ctx)
	if err != nil {
		return fmt.Errorf("failed to instantiate host module: %w", err)
	}

	// The only WASI calls QuickJS makes without quickjs-libc are the clock,
	// and stdio when printing a fatal error, which goes nowhere.
	const errBadFile = 8
	_, err = rt.NewHostModuleBuilder("wasi_snapshot_preview1").
		NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, id uint32, precision uint64, resultPtr uint32) uint32 {
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], uint64(time.Now().UnixNano()))
		m.Memory().Write(resultPtr, buf[:])
		return 0
	}).Export("clock_time_get").
		NewFunctionBuilder().WithFunc(func(fd uint32) uint32 { return errBadFile }).Export("fd_close").
		NewFunctionBuilder().WithFunc(func(fd, ptr uint32) uint32 { return errBadFile }).Export("fd_fdstat_get").
		NewFunctionBuilder().WithFunc(func(fd uint32, offset uint64, whence, ptr uint32) uint32 { return errBadFile }).Export("fd_seek").
		NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, fd, iovs, iovsLen, nwritten uint32) uint32 {
		var total uint32
		for k := uint32(0); k < iovsLen; k++ {
			l, _ := m.Memory().ReadUint32Le(iovs + k*8 + 4)
			total += l
		}
		m.Memory().WriteUint32Le(nwritten, total)
		return 0
	}).Export("fd_write").
		Instantiate(ctx)
	if err != nil {
		return fmt.Errorf("failed to instantiate WASI module: %w", err)
	}
	return nil
}

// hostFuncs are what the prelude can call. They must not have side effects,
// as scripts in the Run JavaScript block can't do anything but compute.
var hostFuncs = map[string]func(arg json.RawMessage) (any, error){
	"date_parts": dateParts,
}

func callHost(name string, arg json.RawMessage) (res []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s failed", name)
		}
	}()
	fn, ok := hostFuncs[name]
	if !ok {
		return nil, fmt.Errorf("unknown host function %s", name)
	}
	v, err := fn(arg)
	if err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

// locations caches time zones, as LoadLocation reads the zone each time.
var locations sync.Map

func loadLocation(name string) (*time.Location, error) {
	if loc, ok := locations.Load(name); ok {
		return loc.(*time.Location), nil
	}
	loc, err := time.LoadLocation(name)
	// Local is the zone of the server.
	if err != nil || name == "Local" {
		return nil, fmt.Errorf("formatDate: unknown time zone %q, use a name like Europe/Berlin", name)
	}
	locations.Store(name, loc)
	return loc, nil
}

func dateParts(arg json.RawMessage) (any, error) {
	var a struct {
		Time     float64 `json:"time"`
		TimeZone string  `json:"time_zone"`
	}
	if err := json.Unmarshal(arg, &a); err != nil {
		return nil, err
	}
	loc, err := loadLocation(a.TimeZone)
	if err != nil {
		return nil, err
	}
	t := time.UnixMilli(int64(a.Time)).In(loc)
	_, offset := t.Zone()
	return map[string]int{
		"year":    t.Year(),
		"month":   int(t.Month()),
		"day":     t.Day(),
		"hour":    t.Hour(),
		"minute":  t.Minute(),
		"second":  t.Second(),
		"weekday": int(t.Weekday()),
		"offset":  offset / 60,
	}, nil
}
