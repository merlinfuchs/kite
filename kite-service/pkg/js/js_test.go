package js

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func run(t *testing.T, code string, input any) (string, error) {
	t.Helper()
	res, err := Run(context.Background(), Options{Code: code, Input: input})
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
		{"promise", `return (async () => 1)()`, "TypeError: the code returned a Promise, async code isn't supported"},
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
		{"code too long", strings.Repeat(" ", MaxCodeLength+1), "the code is 4001 characters, more than the limit of 4000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			_, err := run(t, tt.code, nil)
			require.Error(t, err)
			assert.True(t, IsError(err), "%v is not a script error", err)
			assert.Equal(t, tt.want, err.Error())
			assert.Less(t, time.Since(start), Timeout+killGrace+50*time.Millisecond)
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
for (let i = 0; i < 1000; i++) console.log("x".repeat(100))`,
		Log: func(level LogLevel, message string) { logs = append(logs, entry{level, message}) },
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
			code := `return input * 2`
			if i%5 == 0 {
				code = `while (true) {}`
			}
			res, err := Run(context.Background(), Options{Code: code, Input: i})
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
