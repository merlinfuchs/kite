package js

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testTimeout is the time limit of tests that don't test it, as a busy CI
// machine can take longer than Timeout for any script.
const testTimeout = time.Second

func run(t *testing.T, code string, input any) (string, error) {
	t.Helper()
	res, err := Run(context.Background(), Options{Code: code, Input: input, timeout: testTimeout})
	return string(res), err
}

func TestRun(t *testing.T) {
	tests := []struct {
		name  string
		code  string
		input any
		want  string
	}{
		{"number", `return 1 + 2`, nil, `3`},
		{"input", `return input.items.map(x => x * input.factor)`, map[string]any{"items": []int{1, 2, 3}, "factor": 2}, `[2,4,6]`},
		{"statements", "let total = 0\nfor (const n of input.n) {\n  if (n % 2) total += n\n}\nreturn {total}", map[string]any{"n": []int{1, 2, 3, 4, 5}}, `{"total":9}`},
		{"no return", `const x = 1`, nil, `null`},
		{"undefined", `return undefined`, nil, `null`},
		{"string", `return "a" + "b"`, nil, `"ab"`},
		{"es2023", `return [3, 1, 2].toSorted().findLast(x => x > 1)`, nil, `3`},
		{"regex", `return "<@123> and <@456>".match(/\d+/g)`, nil, `["123","456"]`},
		{"random differs", `return Math.random() !== Math.random()`, nil, `true`},
		{"date", `return new Date(0).toISOString()`, nil, `"1970-01-01T00:00:00.000Z"`},
		{"await", "const x = await 1\nreturn x + await Promise.resolve(2)", nil, `3`},
		{"promise all", `return await Promise.all([1, Promise.resolve(2), (async () => 3)()])`, nil, `[1,2,3]`},
		{"async helper", "async function double(n) {\n  await null\n  return n * 2\n}\nreturn [await double(1), await double(2)]", nil, `[2,4]`},
		{"microtask", `const seen = []; queueMicrotask(() => seen.push(1)); await null; return seen`, nil, `[1]`},
		{"escaped wrapper", `}), (function () { return 1`, nil, `1`},
		{"unhandled rejection", `Promise.reject(new Error("ignored")); (async () => { throw 1 })(); return 1`, nil, `1`},
		{"no globals", `return [typeof std, typeof os, typeof kite, typeof require, typeof fetch, typeof setTimeout, typeof process, typeof Deno]`, nil,
			`["undefined","undefined","undefined","undefined","undefined","undefined","undefined","undefined"]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := run(t, tt.code, tt.input)
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, res)
		})
	}
}

func TestRunErrors(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{"throw", "const a = 1\nthrow new Error('nope')", "Error: nope (line 2)"},
		{"reference error", "\n\nreturn missing()", "ReferenceError: missing is not defined (line 3)"},
		{"null property", "\n\nreturn null.x", "TypeError: cannot read property 'x' of null"},
		{"syntax", "return {", "SyntaxError: expecting ';' (line 2)"},
		{"throw string", `throw "nope"`, "nope"},
		{"rejected", "\nreturn Promise.reject(new Error('nope'))", "Error: nope (line 2)"},
		{"rejected helper", "async function f() {\n  await null\n  throw new TypeError('bad')\n}\nreturn await f()", "TypeError: bad (line 3)"},
		{"null property after await", "const x = 1\nawait null\n\nreturn null.x", "TypeError: cannot read property 'x' of null"},
		{"error after await", "await null\n\nreturn missing()", "ReferenceError: missing is not defined (line 3)"},
		{"slow rejection", `return Promise.reject({toString() { while (true) {} }})`, "the code ran longer than 25 ms"},
		{"never finishes", `await new Promise(() => {})`, "the code waited for something that never finished"},
		{"spin in microtask", `queueMicrotask(() => { while (true) {} }); return 1`, "the code ran longer than 25 ms"},
		{"endless microtasks", `const f = () => queueMicrotask(f); f(); return 1`, "the code ran longer than 25 ms"},
		{"caught spin", `const spin = async () => { await null; while (true) {} }; try { await spin() } catch {} return 1`, "the code ran longer than 25 ms"},
		{"bigint", `return 1n`, "TypeError: BigInt are forbidden in JSON.stringify"},
		{"cyclic", `const a = {}; a.a = a; return a`, "TypeError: circular reference"},
		{"spin", `while (true) {}`, "the code ran longer than 25 ms"},
		{"spin catch", `while (true) { try { while (true) {} } catch {} }`, "the code ran longer than 25 ms"},
		{"native loop", `return new Array(2**28).join("x").length`, "the code ran longer than 25 ms"},
		{"regex backtracking", `return /^(a+)+$/.test("a".repeat(40) + "b")`, "the code ran longer than 25 ms"},
		{"memory", `const a = []; while (true) a.push(new Uint8Array(1e6))`, "the code used more than 8 MB of memory"},
		{"array buffer", `return new ArrayBuffer(2**30).byteLength`, "the code used more than 8 MB of memory"},
		{"string", `return "x".repeat(2**30).length`, "RangeError: string too long (line 1)"},
		{"recursion", `function f() { return f() + 1 }; return f()`, "RangeError: Maximum call stack size exceeded (line 1)"},
		{"result too large", `return "x".repeat(20000)`, "the result is 19 KB, more than the limit of 16 KB"},
		{"multiline message", "throw new Error('Invalid amount\\nUse a number from 1 to 10')", "Error: Invalid amount\nUse a number from 1 to 10 (line 1)"},
		{"long message", `throw "x".repeat(1e4)`, strings.Repeat("x", MaxErrorLength-len(cutOff)) + cutOff},
		{"unprintable", `throw {toString() { throw 1 }}`, "Error: the code threw something that can't be shown as text"},
		{"code too long", strings.Repeat(" ", MaxCodeLength+1), "the code is 4001 characters, more than the limit of 4000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{Code: tt.code, timeout: testTimeout}
			if tt.want == errTimeout.Message {
				opts.timeout = 0
			}
			start := time.Now()
			_, err := Run(context.Background(), opts)
			require.Error(t, err)
			var scriptErr *Error
			assert.ErrorAs(t, err, &scriptErr)
			assert.Equal(t, tt.want, err.Error())
			// Far below the 8.7 s a native loop ran without the kill, with
			// room for a busy CI machine.
			assert.Less(t, time.Since(start), 500*time.Millisecond)
		})
	}
}

func TestRunInputTooLarge(t *testing.T) {
	_, err := run(t, `return 1`, strings.Repeat("x", MaxInputLength))
	assert.EqualError(t, err, "the inputs are more than 1 MB")
}

func TestRunState(t *testing.T) {
	_, err := run(t, `Object.prototype.polluted = 1; globalThis.leak = 1; return 1`, nil)
	require.NoError(t, err)
	res, err := run(t, `return [typeof ({}).polluted, typeof leak]`, nil)
	require.NoError(t, err)
	assert.JSONEq(t, `["undefined","undefined"]`, res)
}

func TestRunLog(t *testing.T) {
	type entry struct {
		level   LogLevel
		message string
	}
	var logs []entry
	_, err := Run(context.Background(), Options{
		Code: `console.log("a", 1, {b: [2]}); console.warn("w"); console.error(new Error("e")); console.log(undefined, null)
for (let i = 0; i < 100; i++) console.log("x".repeat(100))`,
		Log:     func(level LogLevel, message string) { logs = append(logs, entry{level, message}) },
		timeout: testTimeout,
	})
	require.NoError(t, err)
	require.Greater(t, len(logs), 4)
	assert.Equal(t, []entry{
		{LogLevelInfo, `a 1 {"b":[2]}`},
		{LogLevelWarn, "w"},
		{LogLevelError, "Error: e"},
		{LogLevelInfo, "undefined null"},
	}, logs[:4])
	total := 0
	for _, l := range logs {
		total += len(l.message)
	}
	assert.LessOrEqual(t, total, MaxLogLength)

	assert.Contains(t, logs[len(logs)-1].message, "(cut off)")

	logs = nil
	_, err = Run(context.Background(), Options{
		Code:    `for (let i = 0; i < 100; i++) console.log("")`,
		Log:     func(level LogLevel, message string) { logs = append(logs, entry{level, message}) },
		timeout: testTimeout,
	})
	require.NoError(t, err)
	assert.Len(t, logs, MaxLogEntries)
}

func TestFormatDate(t *testing.T) {
	// 2026-03-29 01:30 UTC, half an hour before Berlin switches to summer time.
	const ts = 1774747800000
	tests := []struct {
		code string
		want string
	}{
		{`return formatDate(input.ts)`, `"2026-03-29 01:30"`},
		{`return formatDate(input.ts, "dddd, MMMM D YYYY [at] h:mm A Z", "Europe/Berlin")`, `"Sunday, March 29 2026 at 3:30 AM +02:00"`},
		{`return formatDate(input.ts - 3600000, "HH:mm Z", "Europe/Berlin")`, `"01:30 +01:00"`},
		{`return formatDate(new Date(input.ts), "ddd DD.MM.YY hh:mm:ss a", "America/New_York")`, `"Sat 28.03.26 09:30:00 pm"`},
		{`return formatDate("2026-01-01T12:00:00Z", "YYYY-MM-DD HH:mm", "Asia/Kolkata")`, `"2026-01-01 17:30"`},
	}
	for _, tt := range tests {
		res, err := run(t, tt.code, map[string]any{"ts": ts})
		require.NoError(t, err, tt.code)
		assert.Equal(t, tt.want, res, tt.code)
	}

	_, err := run(t, `return formatDate(0, "HH", "Mars/Olympus")`, nil)
	assert.EqualError(t, err, `TypeError: formatDate: unknown time zone "Mars/Olympus", use a name like Europe/Berlin (line 1)`)
	_, err = run(t, `return formatDate(0, "HH", "Local")`, nil)
	assert.Error(t, err)
	_, err = run(t, `return formatDate("soon")`, nil)
	assert.EqualError(t, err, `TypeError: formatDate: invalid date soon (line 1)`)
}

func TestNumberToLocaleString(t *testing.T) {
	res, err := run(t, `return [(1234567).toLocaleString(), (-1234.5678).toLocaleString("de-DE"), (0.5).toLocaleString(), (1234.5).toLocaleString(undefined, {minimumFractionDigits: 2}), (12).toLocaleString(undefined, {maximumFractionDigits: 0}), (NaN).toLocaleString()]`, nil)
	require.NoError(t, err)
	assert.JSONEq(t, `["1,234,567","-1,234.568","0.5","1,234.50","12","NaN"]`, res)
}

func TestRunConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	errs := make(chan error, 50)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			opts := Options{Code: `return input * 2`, Input: i, timeout: testTimeout}
			if i%5 == 0 {
				opts = Options{Code: `while (true) {}`}
			}
			res, err := Run(context.Background(), opts)
			if i%5 == 0 {
				if err != errTimeout {
					errs <- fmt.Errorf("spin: %v", err)
				}
			} else if err != nil || string(res) != strconv.Itoa(i*2) {
				errs <- fmt.Errorf("%d: %s %v", i, res, err)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

func TestRunCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Run(ctx, Options{Code: `return 1`})
	assert.ErrorIs(t, err, context.Canceled)
}

// Compiles the module up front, so the first test's timing doesn't include it.
func TestMain(m *testing.M) {
	if _, err := getEngine(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

// Copying the memory of an instance is only a full snapshot if its only
// mutable global is the stack pointer. A new QuickJS-NG version or build flag
// could add one, like for thread-local storage.
func TestOneMutableGlobal(t *testing.T) {
	b := wasmBytes[8:]
	mutable := 0
	for len(b) > 0 {
		id := b[0]
		size, n := binary.Uvarint(b[1:])
		section := b[1+n : 1+n+int(size)]
		b = b[1+n+int(size):]
		if id != 6 {
			continue
		}
		count, n := binary.Uvarint(section)
		section = section[n:]
		for range count {
			// Value type, mutability, then an init expression: i32.const,
			// i64.const or global.get with a LEB128 operand, and end.
			if section[1] == 1 {
				mutable++
			}
			require.Contains(t, []byte{0x41, 0x42, 0x23}, section[2], "unexpected init expression")
			i := 3
			for section[i]&0x80 != 0 {
				i++
			}
			require.Equal(t, byte(0x0b), section[i+1])
			section = section[i+2:]
		}
	}
	assert.Equal(t, 1, mutable)
}

func TestRunMicrotaskError(t *testing.T) {
	var logs []string
	res, err := Run(context.Background(), Options{
		Code: `queueMicrotask(() => { throw new Error("x") }); await null; return 1`,
		Log:  func(level LogLevel, message string) { logs = append(logs, string(level)+" "+message) },
	})
	require.NoError(t, err)
	assert.Equal(t, "1", string(res))
	assert.Equal(t, []string{"error Error: x"}, logs)
}

var testHost = map[string]HostFunc{
	"discord.echo": func(ctx context.Context, arg json.RawMessage) (json.RawMessage, error) { return arg, nil },
	"discord.get_thing": func(ctx context.Context, arg json.RawMessage) (json.RawMessage, error) {
		return json.RawMessage(`{"thing":true}`), nil
	},
	"discord.sleep": func(ctx context.Context, arg json.RawMessage) (json.RawMessage, error) {
		var ms int
		json.Unmarshal(arg, &ms)
		select {
		case <-time.After(time.Duration(ms) * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return arg, nil
	},
	"discord.fail": func(ctx context.Context, arg json.RawMessage) (json.RawMessage, error) {
		return nil, errors.New("Missing Permissions")
	},
	"discord.panic": func(ctx context.Context, arg json.RawMessage) (json.RawMessage, error) { panic("boom") },
	"cookie_api.make_card": func(ctx context.Context, arg json.RawMessage) (json.RawMessage, error) {
		return json.RawMessage(`"card"`), nil
	},
}

func hostFrom(funcs map[string]HostFunc) func(string) HostFunc {
	return func(name string) HostFunc { return funcs[name] }
}

func runHost(t *testing.T, code string, opts ...func(*Options)) (string, error) {
	t.Helper()
	o := Options{Code: code, Host: hostFrom(testHost)}
	for _, f := range opts {
		f(&o)
	}
	res, err := Run(context.Background(), o)
	return string(res), err
}

func TestHostCalls(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{"echo", `return await discord.echo({a: 1})`, `{"a":1}`},
		{"camel case", `return [await discord.getThing(), await cookieApi.makeCard()]`, `[{"thing":true},"card"]`},
		{"then", `return (await discord.echo(1)) + await discord.echo(2).then((n) => n * 10)`, `21`},
		{"caught", `try { await discord.fail() } catch (e) { return e.message }`, `"discord.fail: Missing Permissions"`},
		{"unknown", `try { await discord.nope() } catch (e) { return e.message }`, `"discord.nope isn't available"`},
		{"panic", `try { await discord.panic() } catch (e) { return e.message }`, `"discord.panic: the call failed"`},
		{"not calls", `return [typeof discord.then, typeof discord.toJSON, typeof discord.toString, JSON.stringify(discord), String(await discord === discord)]`, `["undefined","undefined","undefined","{}","true"]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := runHost(t, tt.code)
			require.NoError(t, err)
			assert.JSONEq(t, tt.want, res)
		})
	}

	_, err := runHost(t, "await discord.fail()")
	assert.EqualError(t, err, "Error: discord.fail: Missing Permissions")
	_, err = runHost(t, `for (let i = 0; i < 21; i++) await discord.echo(i)`)
	assert.EqualError(t, err, "Error: the code made more than 20 calls")
}

func TestHostCallsDontUseTime(t *testing.T) {
	// Waiting 60 ms is more than the 25 ms the code can run.
	res, err := runHost(t, `return await discord.sleep(60)`)
	require.NoError(t, err)
	assert.Equal(t, "60", res)

	// The calls of Promise.all run at the same time.
	start := time.Now()
	res, err = runHost(t, `return await Promise.all([discord.sleep(80), discord.sleep(80), discord.sleep(80)])`)
	require.NoError(t, err)
	assert.Equal(t, "[80,80,80]", res)
	assert.Less(t, time.Since(start), 200*time.Millisecond)

	// The time the code runs adds up across the calls it waits for.
	_, err = runHost(t, `
const spin = (ms) => { const end = Date.now() + ms; while (Date.now() < end) {} };
for (let i = 0; i < 4; i++) { spin(10); await discord.echo(i) }`)
	assert.Equal(t, errTimeout, err)

	_, err = runHost(t, `await discord.sleep(1000)`, func(o *Options) { o.wallTime = 100 * time.Millisecond })
	assert.Equal(t, errWallTime, err)
}

func TestUnawaitedHostCallsFinish(t *testing.T) {
	var done atomic.Bool
	host := map[string]HostFunc{"discord.send": func(ctx context.Context, arg json.RawMessage) (json.RawMessage, error) {
		time.Sleep(50 * time.Millisecond)
		done.Store(true)
		return nil, nil
	}}
	res, err := runHost(t, `discord.send(); return 1`, func(o *Options) { o.Host = hostFrom(host) })
	require.NoError(t, err)
	assert.Equal(t, "1", res)
	assert.True(t, done.Load())
}

// Runs that wait for calls give up their slot, so more of them than there
// are slots wait at the same time.
func TestWaitingRunsGiveUpSlots(t *testing.T) {
	n := 3*cap(slots) + 1
	start := time.Now()
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := runHost(t, `return await discord.sleep(150)`)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		assert.NoError(t, err)
	}
	assert.Less(t, time.Since(start), 450*time.Millisecond)
}
