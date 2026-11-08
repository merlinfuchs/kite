// Package js runs untrusted JavaScript in QuickJS-NG compiled to Wasm, with
// limits on time and memory. A script can compute, log and call the pure
// helpers in prelude.js, and nothing else: it has no files, network, timers
// or access to the service.
package js

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
	MaxLogEntries   = 50
)

const (
	// killGrace is how long after Timeout wazero kills a script the
	// interrupt handler didn't stop.
	killGrace = 25 * time.Millisecond
	// maxPages caps the Wasm memory in pages of 64 KiB, as a backstop for
	// memory QuickJS doesn't count against MemoryLimit.
	maxPages = 256
	// stackSize is the QuickJS stack limit, below the 1 MiB Wasm stack.
	stackSize = 256 << 10
)

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
		engineInst, engineErr = newEngine(context.Background())
	})
	return engineInst, engineErr
}

// Warm compiles QuickJS in the background after delay, which takes about a
// second on kite.onl's server, so the first script doesn't wait for it. A
// script that runs before then compiles it itself.
func Warm(ctx context.Context, delay time.Duration) {
	go func() {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return
		}
		if _, err := getEngine(); err != nil {
			slog.With("error", err).Error("Failed to compile QuickJS")
		}
	}()
}

// LogLevel is one of the levels of provider.LogLevel.
type LogLevel string

const (
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type Options struct {
	Code string
	// Input is passed to the script as input, after encoding it as JSON.
	Input any
	// Log gets what the script logs with console.log, warn and error, up to
	// MaxLogEntries entries and MaxLogLength in total.
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

	logged, entries := 0, 0
	st := &runState{
		deadline: time.Now().Add(Timeout),
		log: func(level LogLevel, message string) {
			if opts.Log == nil || entries >= MaxLogEntries || logged+len(message) > MaxLogLength {
				return
			}
			logged += len(message)
			entries++
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

	out, ok, err := inst.run(runCtx, []byte(codeHead+opts.Code+codeTail), input)
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

// The code runs as the body of a function, which starts on its first line so
// line numbers match.
const (
	codeHead = "(function (input) {"
	codeTail = "\n})"
)

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
		// position of the function, which is codeHead.
		if m[1] != "1" || m[2] != "1" {
			e.Line, _ = strconv.Atoi(m[1])
		}
	}
	return e
}
