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
	// MaxErrorLength caps the message of an error, which ends up in the
	// logs of the app.
	MaxErrorLength = 1000
	// MaxCalls caps the calls of Options.Host a run can make.
	MaxCalls = 20
	// MaxWallTime caps how long a run takes in all, with the time it waits
	// for calls, which doesn't count against Timeout.
	MaxWallTime = 10 * time.Second
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

// liveRuns caps the runs that have an instance, also the ones that wait for
// calls without a slot. An instance holds 1.5 MB of memory, up to 16 MB.
var liveRuns = make(chan struct{}, 128)

// callGrace is how long Run waits for calls to return after they were
// canceled, before it gives up on them.
const callGrace = 5 * time.Second

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

	// Host returns the call the code makes by a name like
	// discord.create_message, or nil if there's none. The code calls it as
	// discord.createMessage(arg), which returns a promise, and it runs in the
	// background while the code waits. Every call has returned when Run does.
	Host func(name string) HostFunc

	// Random is Math.random, so test runs can pick the same each time.
	Random func() float64

	// timeout replaces Timeout in tests, which can't rely on getting a busy
	// CI machine's CPU for 25 ms. wallTime replaces MaxWallTime.
	timeout  time.Duration
	wallTime time.Duration
}

// HostFunc is a call the code can make. arg is the JSON of the code's
// argument. Its error message is what the code's promise is rejected with. It
// has to return soon once ctx is done.
type HostFunc func(ctx context.Context, arg json.RawMessage) (json.RawMessage, error)

// CallError is an error of a HostFunc from the service it called, whose
// status and code the code can check, like e.status === 404. They become
// properties of the Error, and are left out when they're 0.
type CallError struct {
	Err error
	// The HTTP status of the response.
	Status int
	// The service's own code for the error, like 10003 for Discord's Unknown
	// Channel.
	Code int
}

func (e *CallError) Error() string { return e.Err.Error() }
func (e *CallError) Unwrap() error { return e.Err }

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
	errTimeout  = &Error{Message: fmt.Sprintf("the code ran longer than %d ms", Timeout.Milliseconds())}
	errPending  = &Error{Message: "the code waited for something that never finished"}
	errWallTime = &Error{Message: fmt.Sprintf("the code took longer than %d seconds", int(MaxWallTime.Seconds()))}
	errMemory   = &Error{Message: fmt.Sprintf("the code used more than %d MB of memory", MemoryLimit>>20)}
)

// Run runs opts.Code as the body of an async function with the parameter
// input, and returns the JSON of what it returns once the promise settles.
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
	case liveRuns <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-liveRuns }()

	// The run gives up its slot while it waits for calls.
	held := false
	acquire := func(ctx context.Context) error {
		select {
		case slots <- struct{}{}:
			held = true
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	release := func() {
		if held {
			<-slots
			held = false
		}
	}
	defer release()
	if err := acquire(ctx); err != nil {
		return nil, err
	}

	wallTime := MaxWallTime
	if opts.wallTime > 0 {
		wallTime = opts.wallTime
	}
	hostCtx, cancel := context.WithTimeout(ctx, wallTime)
	defer cancel()

	logged, entries := 0, 0
	st := &runState{
		random:  opts.Random,
		host:    opts.Host,
		hostCtx: hostCtx,
		results: make(chan callResult, MaxCalls),
		log: func(level LogLevel, message string) {
			if opts.Log == nil || entries >= MaxLogEntries || logged >= MaxLogLength {
				return
			}
			// The message that goes over the limit is cut off, not left out.
			if logged+len(message) > MaxLogLength {
				message = truncate(message, max(0, MaxLogLength-logged-len(cutOff))) + cutOff
			}
			logged += len(message)
			entries++
			opts.Log(level, message)
		},
	}
	stateCtx := withState(ctx, st)
	inst, err := e.newInstance(stateCtx)
	if err != nil {
		return nil, err
	}
	defer inst.close(context.Background())

	budget := Timeout
	if opts.timeout > 0 {
		budget = opts.timeout
	}
	out, status, err := inst.run(stateCtx, []byte(codeHead+opts.Code+codeTail), input, &budget)
	for err == nil && status == runPending && st.inflight > 0 {
		release()
		select {
		case r := <-st.results:
			st.inflight--
			if acquire(hostCtx) != nil {
				err = errWallTime
			} else {
				out, status, err = inst.resolve(stateCtx, r, &budget)
			}
		case <-hostCtx.Done():
			err = errWallTime
		}
	}
	release()
	// Calls the code didn't wait for, like a message it sent without await,
	// still finish, within the wall time. After it they're canceled, and Run
	// gives up on them after callGrace.
	done := hostCtx.Done()
	var giveUp <-chan time.Time
	for st.inflight > 0 {
		select {
		case <-st.results:
			st.inflight--
		case <-done:
			done = nil
			giveUp = time.After(callGrace)
		case <-giveUp:
			slog.Error("Script calls didn't return after they were canceled", slog.Int("calls", st.inflight))
			st.inflight = 0
		}
	}

	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if err == errTimeout || err == errWallTime {
			return nil, err
		}
		return nil, fmt.Errorf("the code stopped unexpectedly: %w", err)
	}
	switch status {
	case runError:
		return nil, scriptError(string(out), opts.Code)
	case runPending:
		// No timers or host calls can settle it yet.
		return nil, errPending
	case runResult:
	default:
		return nil, fmt.Errorf("the code stopped unexpectedly: status %d", status)
	}
	if len(out) > MaxResultLength {
		return nil, &Error{Message: fmt.Sprintf("the result is %d KB, more than the limit of %d KB", len(out)>>10, MaxResultLength>>10)}
	}
	return out, nil
}

// The code runs as the body of an async function, so it can use await. The
// function starts on its first line so line numbers match.
const (
	codeHead = "(async function (input) {"
	codeTail = "\n})"
)

var codeLineRe = regexp.MustCompile(`code\.js:(\d+):(\d+)`)

// scriptError turns an error from kite.c, the message, a null byte and the
// stack, into an Error. code is the code the error is from.
func scriptError(out string, code string) *Error {
	message, stack, _ := strings.Cut(out, "\x00")
	if len(message) > MaxErrorLength {
		message = truncate(message, MaxErrorLength-len(cutOff)) + cutOff
	}
	switch message {
	case "InternalError: interrupted":
		return errTimeout
	case "InternalError: out of memory":
		return errMemory
	}
	e := &Error{Message: message}
	if m := codeLineRe.FindStringSubmatch(stack); m != nil {
		line, _ := strconv.Atoi(m[1])
		// QuickJS gives some errors, like reading a property of null, the
		// last position it knows of. That's the start of codeHead, or after
		// an await the statement it resumed at, which would be a wrong line.
		lines := strings.Split(codeHead+code+codeTail, "\n")
		if (m[1] != "1" || m[2] != "1") && line <= len(lines) && !strings.Contains(lines[line-1], "await") {
			e.Line = line
		}
	}
	return e
}

const cutOff = " (cut off)"

// truncate cuts s to at most n bytes without splitting a character.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}
