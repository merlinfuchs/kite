package js

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testModule = `
variables.define("xp", { scoped: true })
variables.define("total")

discord.command("ping", { description: "Pong" }, async (ctx) => {
  await ctx.reply({ content: "Pong" })
})

discord.command("level show", { description: "Shows a level", options: [{ type: 6, name: "user", description: "Whose" }] }, async (ctx) => {
  await ctx.reply({ content: "level of " + ctx.options.user })
  await ctx.followUp({ content: "done" })
})

discord.component("claim", async (ctx) => {
  await ctx.update({ content: "claimed " + ctx.data.ticket + " by " + ctx.interaction.member.user.id })
})

discord.on("message_create", async (ctx) => {
  await discord.createMessage({ channel_id: ctx.event.channel_id, content: "first" })
})
discord.on("message_create", async (ctx) => {
  await discord.createMessage({ channel_id: ctx.event.channel_id, content: "second" })
})
`

func loadModule(t *testing.T, code string) (*Manifest, error) {
	t.Helper()
	return LoadModule(context.Background(), ModuleOptions{Code: code, timeout: testTimeout})
}

// recorder is a host that records the calls of a module.
type recorder struct {
	mu    sync.Mutex
	calls []string
}

func (r *recorder) host(name string) HostFunc {
	switch name {
	case "interaction.respond", "interaction.follow_up", "discord.create_message":
	default:
		return nil
	}
	return func(ctx context.Context, arg json.RawMessage) (json.RawMessage, error) {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.calls = append(r.calls, name+" "+string(arg))
		return nil, nil
	}
}

func runModuleCall(t *testing.T, code string, call ModuleCall) ([]string, error) {
	t.Helper()
	r := &recorder{}
	err := RunModule(context.Background(), ModuleOptions{Code: code, Host: r.host, timeout: testTimeout}, call)
	return r.calls, err
}

func TestLoadModule(t *testing.T) {
	m, err := loadModule(t, testModule)
	require.NoError(t, err)
	b, err := json.Marshal(m)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"commands": [
			{"name": "ping", "data": {"description": "Pong"}},
			{"name": "level show", "data": {"description": "Shows a level", "options": [{"type": 6, "name": "user", "description": "Whose"}]}}
		],
		"events": ["message_create"],
		"components": ["claim"],
		"variables": [{"name": "xp", "scoped": true}, {"name": "total", "scoped": false}]
	}`, string(b))

	m, err = loadModule(t, `discord.command("ping", () => {})`)
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(m.Commands[0].Data))
}

func TestLoadModuleErrors(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{"call", "\ndiscord.createMessage({})", "Error: discord.createMessage can't be used at the top level of a module. Call it inside a handler instead. (line 2)"},
		// Kept even though nothing waits for it.
		{"call in async function", "\n(async () => { discord.createMessage({}) })()", "Error: discord.createMessage can't be used at the top level of a module. Call it inside a handler instead. (line 2)"},
		{"caught call", "try { discord.createMessage({}) } catch {}", "Error: discord.createMessage can't be used at the top level of a module. Call it inside a handler instead. (line 1)"},
		{"fetch", `fetch("https://example.com")`, "Error: fetch can't be used at the top level of a module. Call it inside a handler instead. (line 1)"},
		// The top level isn't async.
		{"await", "await discord.createMessage({})", "SyntaxError: expecting ';'"},
		{"duplicate command", "discord.command(\"a\", () => {})\ndiscord.command(\"a\", () => {})", `TypeError: discord.command: there's already a command called "a" (line 2)`},
		{"no handler", `discord.command("a", {})`, "TypeError: discord.command: the handler must be a function, like async (ctx) => { ... } (line 1)"},
		{"options", `discord.command("a", [], () => {})`, `TypeError: discord.command: the options must be an object, like { description: "Shows your level" } (line 1)`},
		{"event", `discord.on("MESSAGE_CREATE", () => {})`, `TypeError: discord.on: the event must be a name like "message_create" (line 1)`},
		{"component", `discord.component("a:b", () => {})`, "TypeError: discord.component: the name must be up to 32 letters, digits, - and _ (line 1)"},
		{"scoped", `variables.define("xp", { scoped: "yes" })`, "TypeError: variables.define: scoped must be true or false (line 1)"},
		{"duplicate variable", `variables.define("xp"); variables.define("xp")`, `TypeError: variables.define: there's already a variable called "xp" (line 1)`},
		{"thrown", "\n\nthrow new Error(\"nope\")", "Error: nope (line 3)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadModule(t, tt.code)
			require.Error(t, err)
			assert.Equal(t, tt.want, err.Error())
		})
	}

	_, err := loadModule(t, strings.Repeat(" ", MaxModuleLength+1))
	assert.EqualError(t, err, "the module is 256 KB, more than the limit of 256 KB")
}

func TestRunModule(t *testing.T) {
	calls, err := runModuleCall(t, testModule, ModuleCall{
		Kind: ModuleCallCommand,
		Name: "level show",
		Interaction: json.RawMessage(`{"data": {"name": "level", "options": [
			{"type": 1, "name": "show", "options": [{"type": 6, "name": "user", "value": "123"}]}
		]}}`),
	})
	require.NoError(t, err)
	assert.Equal(t, []string{
		`interaction.respond {"type":4,"data":{"content":"level of 123"}}`,
		`interaction.follow_up {"content":"done"}`,
	}, calls)

	calls, err = runModuleCall(t, testModule, ModuleCall{
		Kind:        ModuleCallComponent,
		Name:        "claim",
		Interaction: json.RawMessage(`{"member": {"user": {"id": "1"}}}`),
		Data:        json.RawMessage(`{"ticket": 7}`),
	})
	require.NoError(t, err)
	assert.Equal(t, []string{`interaction.respond {"type":7,"data":{"content":"claimed 7 by 1"}}`}, calls)

	calls, err = runModuleCall(t, testModule, ModuleCall{
		Kind:  ModuleCallEvent,
		Name:  "message_create",
		Event: json.RawMessage(`{"channel_id": "5"}`),
	})
	require.NoError(t, err)
	assert.Equal(t, []string{
		`discord.create_message {"channel_id":"5","content":"first"}`,
		`discord.create_message {"channel_id":"5","content":"second"}`,
	}, calls)

	_, err = runModuleCall(t, testModule, ModuleCall{Kind: ModuleCallCommand, Name: "nope"})
	assert.EqualError(t, err, `Error: the module has no command called "nope"`)
}

func TestRunModuleEventErrors(t *testing.T) {
	// A failing handler doesn't stop the next one.
	code := `
discord.on("message_create", () => { throw new Error("first") })
discord.on("message_create", () => { throw new Error("second") })
discord.on("message_create", async (ctx) => { await discord.createMessage({ content: "third" }) })`
	r := &recorder{}
	var logs []string
	err := RunModule(context.Background(), ModuleOptions{
		Code:    code,
		Host:    r.host,
		Log:     func(level LogLevel, message string) { logs = append(logs, string(level)+" "+message) },
		timeout: testTimeout,
	}, ModuleCall{Kind: ModuleCallEvent, Name: "message_create", Event: json.RawMessage(`{}`)})
	assert.EqualError(t, err, "Error: first (line 2)")
	assert.Equal(t, []string{`discord.create_message {"content":"third"}`}, r.calls)
	assert.Equal(t, []string{"error Error: second"}, logs)
}

func TestModuleRegistersOnlyAtTopLevel(t *testing.T) {
	_, err := runModuleCall(t, `discord.command("a", () => { discord.command("b", () => {}) })`, ModuleCall{Kind: ModuleCallCommand, Name: "a"})
	assert.EqualError(t, err, "Error: discord.command can only be used at the top level of a module (line 1)")

	// The Run JavaScript block isn't a module.
	_, err = run(t, `variables.define("xp")`, nil)
	assert.EqualError(t, err, "Error: variables.define can only be used at the top level of a module (line 1)")
}

func TestModuleLimits(t *testing.T) {
	// More memory than the block has.
	code := `discord.command("a", () => { globalThis.s = "x".repeat(16 << 20) + "y" })`
	_, err := runModuleCall(t, code, ModuleCall{Kind: ModuleCallCommand, Name: "a"})
	require.NoError(t, err)

	code = `discord.command("a", () => { globalThis.s = "x".repeat(40 << 20) + "y" })`
	_, err = runModuleCall(t, code, ModuleCall{Kind: ModuleCallCommand, Name: "a"})
	assert.Equal(t, moduleLimits.errMemory, err)

	code = `discord.on("ready", async () => { for (let i = 0; i < 51; i++) await discord.createMessage({}) })`
	_, err = runModuleCall(t, code, ModuleCall{Kind: ModuleCallEvent, Name: "ready"})
	assert.EqualError(t, err, "Error: the code made more than 50 calls")

	err = RunModule(context.Background(), ModuleOptions{Code: `discord.on("ready", () => { while (true) {} })`}, ModuleCall{Kind: ModuleCallEvent, Name: "ready"})
	assert.Equal(t, moduleLimits.errTimeout, err)

	// Loading has more time.
	err = RunModule(context.Background(), ModuleOptions{Code: `while (true) {}`}, ModuleCall{})
	assert.Equal(t, moduleLoadLimits.errTimeout, err)
}

func TestModuleLoadsOnce(t *testing.T) {
	// Unique, so the module isn't loaded by another test.
	code := fmt.Sprintf(`// %d
console.log("loaded")
let count = 0
const startedAt = Math.random()
discord.on("ready", async () => {
  count++
  console.log(count, startedAt)
})`, time.Now().UnixNano())

	var mu sync.Mutex
	var logs []string
	opts := ModuleOptions{Code: code, timeout: testTimeout, Log: func(level LogLevel, message string) {
		mu.Lock()
		defer mu.Unlock()
		logs = append(logs, message)
	}}
	var wg sync.WaitGroup
	for range 5 {
		wg.Go(func() {
			assert.NoError(t, RunModule(context.Background(), opts, ModuleCall{Kind: ModuleCallEvent, Name: "ready"}))
		})
	}
	wg.Wait()

	// Every run starts from the globals the top level set.
	require.Len(t, logs, 6)
	assert.Equal(t, "loaded", logs[0])
	for _, l := range logs[1:] {
		assert.Equal(t, logs[1], l)
		assert.True(t, strings.HasPrefix(l, "1 0."), l)
	}
}

func TestModuleCacheDropsOldest(t *testing.T) {
	c := &moduleCache{max: 10}
	c.add(&loadedModule{key: [32]byte{1}, snapshot: make([]byte, 4)})
	c.add(&loadedModule{key: [32]byte{2}, snapshot: make([]byte, 4)})
	c.get([32]byte{1})
	c.add(&loadedModule{key: [32]byte{3}, snapshot: make([]byte, 4)})
	assert.NotNil(t, c.get([32]byte{1}))
	assert.Nil(t, c.get([32]byte{2}))
	assert.NotNil(t, c.get([32]byte{3}))
	assert.Equal(t, 8, c.size)
}

// BenchmarkRunModule measures a run of a loaded module, a small one and one
// of about MaxModuleLength.
func BenchmarkRunModule(b *testing.B) {
	var big strings.Builder
	for i := 0; big.Len() < MaxModuleLength-1000; i++ {
		fmt.Fprintf(&big, "function f%d(a) { return a.map((x) => x * %d).filter(Boolean).reduce((s, x) => s + x, 0) }\n", i, i)
	}
	big.WriteString(`discord.command("ping", () => {})`)
	for _, bm := range []struct {
		name string
		code string
	}{
		{"small", `discord.command("ping", () => {})`},
		{"256KB", big.String()},
	} {
		b.Run(bm.name, func(b *testing.B) {
			for b.Loop() {
				err := RunModule(context.Background(), ModuleOptions{Code: bm.code, timeout: testTimeout}, ModuleCall{Kind: ModuleCallCommand, Name: "ping"})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
