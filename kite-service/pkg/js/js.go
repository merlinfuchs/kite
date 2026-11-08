// Package js runs untrusted JavaScript in QuickJS-NG compiled to Wasm, with
// limits on time and memory. A script can compute, log and call the pure
// helpers in prelude.js, and nothing else: it has no files, network, timers
// or access to the service.
package js

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// Limits of a run. MaxCodeLength and MaxInputs match the editor in
// kite-web/src/lib/blocks/javascriptRun.ts.
const (
	MaxCodeLength   = 4000
	MaxInputs       = 20
	Timeout         = 25 * time.Millisecond
	MemoryLimit     = 8 << 20
	MaxInputLength  = 1 << 20
	MaxResultLength = 16 << 10
	MaxLogLength    = 4 << 10
)

// killGrace is how long after Timeout wazero kills a script the interrupt
// handler didn't stop.
const killGrace = 25 * time.Millisecond

var defaultEngineConfig = engineConfig{
	memoryLimit: MemoryLimit,
	maxPages:    256, // 16 MiB
	stackSize:   256 << 10,
}

var (
	engineOnce sync.Once
	engineInst *engine
	engineErr  error
)

// slots caps the scripts running at once to half the cores, so a burst of
// scripts can't delay the gateways and everything else. Scripts past it wait.
var slots = make(chan struct{}, max(1, runtime.GOMAXPROCS(0)/2))

func getEngine() (*engine, error) {
	engineOnce.Do(func() {
		engineInst, engineErr = newEngine(context.Background(), defaultEngineConfig)
	})
	return engineInst, engineErr
}

type LogLevel int

const (
	LogLevelInfo LogLevel = iota
	LogLevelWarn
	LogLevelError
)

type Options struct {
	Code string
	// Input is passed to the script as input, after encoding it as JSON.
	Input any
	// Log gets what the script logs with console.log, warn and error, up to
	// MaxLogLength in total.
	Log func(level LogLevel, message string)
}

// Error is an error of the script, like one it threw or running out of
// time, with a message for the user.
type Error struct {
	Message string
	// Line in the code, 0 if unknown.
	Line int
}

func (e *Error) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s (line %d)", e.Message, e.Line)
	}
	return e.Message
}

var (
	errTimeout = &Error{Message: fmt.Sprintf("the code ran longer than %d ms", Timeout.Milliseconds())}
	errMemory  = &Error{Message: fmt.Sprintf("the code used more than %d MB of memory", MemoryLimit>>20)}
)

// Run runs opts.Code as the body of a function with the parameter input, and
// returns the JSON of what it returns.
func Run(ctx context.Context, opts Options) (json.RawMessage, error) {
	if n := utf8.RuneCountInString(opts.Code); n > MaxCodeLength {
		return nil, &Error{Message: fmt.Sprintf("the code is %d characters, more than the limit of %d", n, MaxCodeLength)}
	}
	input, err := json.Marshal(opts.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to encode input: %w", err)
	}
	if len(input) > MaxInputLength {
		return nil, &Error{Message: fmt.Sprintf("the inputs are more than %d MB", MaxInputLength>>20)}
	}

	e, err := getEngine()
	if err != nil {
		return nil, err
	}

	select {
	case slots <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-slots }()

	logged := 0
	st := &runState{
		deadline: time.Now().Add(Timeout),
		log: func(level LogLevel, message string) {
			if opts.Log == nil || logged+len(message) > MaxLogLength {
				return
			}
			logged += len(message)
			opts.Log(level, message)
		},
	}
	runCtx, cancel := context.WithTimeout(withState(ctx, st), Timeout+killGrace)
	defer cancel()

	inst, err := e.newInstance(runCtx)
	if err != nil {
		return nil, err
	}
	defer inst.close(context.Background())

	out, ok, err := inst.run(runCtx, []byte(opts.Code), input)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if runCtx.Err() != nil {
			return nil, errTimeout
		}
		return nil, fmt.Errorf("the code stopped unexpectedly: %w", err)
	}
	if !ok {
		return nil, scriptError(string(out))
	}
	if len(out) > MaxResultLength {
		return nil, &Error{Message: fmt.Sprintf("the result is %d KB, more than the limit of %d KB", len(out)>>10, MaxResultLength>>10)}
	}
	return out, nil
}

var codeLineRe = regexp.MustCompile(`code\.js:(\d+):(\d+)`)

// scriptError turns an error message from kite.c, the message and the stack
// on the following lines, into an Error.
func scriptError(out string) *Error {
	message, stack, _ := strings.Cut(out, "\n")
	switch message {
	case "InternalError: interrupted":
		return errTimeout
	case "InternalError: out of memory":
		return errMemory
	}
	e := &Error{Message: message}
	if m := codeLineRe.FindStringSubmatch(stack); m != nil {
		// QuickJS gives some errors, like reading a property of null, the
		// position of the function, which is the wrapper kite.c adds.
		if m[1] != "1" || m[2] != "1" {
			e.Line, _ = strconv.Atoi(m[1])
		}
	}
	return e
}

// IsError reports whether err is the script's fault, like an error it threw,
// rather than the service's.
func IsError(err error) bool {
	var e *Error
	return errors.As(err, &e)
}
