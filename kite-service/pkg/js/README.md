# pkg/js

`pkg/js` runs the code of the Run JavaScript block. Scripts run in QuickJS-NG compiled to WebAssembly, inside the service through [wazero](https://wazero.io). See `design/javascript.md` for why it's built this way, the measurements behind the limits, and the plans for host functions and modules.

`Run` takes the code and its input, and returns the JSON of what the code returns or an `*Error` with a message for the user. Every run gets a fresh Wasm instance, so nothing a script does carries over to the next one.

The package is made of these files:

- `js.go` has `Run`, the limits and the conversion of errors.
- `runtime.go` compiles the module, takes the snapshot every run starts from, and has the host functions.
- `prelude.js` runs once before the snapshot. It defines the globals scripts get, like `formatDate`, and removes `kite.call` so scripts only reach what it wraps.
- `wasm/kite.c` is the C glue between QuickJS-NG and Go: the imports from Go, `kite_init` and `kite_run`.
- `kite.wasm` is QuickJS-NG and `kite.c` built by `wasm/build.sh`. It's committed, because building it needs wasi-sdk.

Changes to `prelude.js` and the Go files don't need `kite.wasm` rebuilt, as the prelude is embedded in the service and runs when it starts. Changes to `wasm/kite.c` and upgrades of QuickJS-NG do:

1. Install [wasi-sdk](https://github.com/WebAssembly/wasi-sdk/releases). The current build used wasi-sdk 22 (clang 18). Put it in `/opt/wasi-sdk`, or set `WASI_SDK` to where it is.
2. Run `wasm/build.sh`. It clones the QuickJS-NG release in `QUICKJS_VERSION` and writes `kite.wasm`.
3. Run `go test ./pkg/js ./pkg/flow`, and commit `kite.wasm` with the change.

To upgrade QuickJS-NG, change `QUICKJS_VERSION` in `wasm/build.sh` and do the same.

`TestOneMutableGlobal` has to pass. Each run copies the memory of the snapshot into a new instance. That's only a full copy of its state if the module's only mutable global is the C stack pointer, which is back at its start between calls. A build that adds one, for example for thread-local storage, would make runs start from a wrong state.

The module must also not import anything `runtime.go` doesn't provide. wazero refuses to start it then, so a new import shows up as a failing test, not as a gap in the sandbox. Don't add quickjs-libc or WASI functions for files, the environment or the network.

Helpers that need Go, like the time zones of `formatDate`, are host functions:

1. Add a function to `hostFuncs` in `runtime.go`. It takes its argument as JSON and returns a value that's encoded as JSON.
2. Wrap it in `prelude.js` with `call("name", arg)`, and give the wrapper the name scripts use.
3. Add a test to `js_test.go`.

Every helper can be called by any app's code with any argument. It must not have side effects, read anything outside its argument, or take long, as the time it takes counts against the script's 25 ms.

Calls that act on Discord or other services aren't helpers. Each one needs a credit cost and a check of what the app may do, which the caller of `Run` is responsible for. It passes them in `Options.Host`, by names like `discord.create_message`, and code calls them as `discord.createMessage(arg)`, which returns a promise. Each call runs in a goroutine while the code waits, and `kite_run` returns that the code is waiting when nothing else is left to run. `Run` then gives up its slot, waits for a call to finish, and passes its result to `kite_resolve`, which continues the code. Only the time the code runs counts against the 25 ms, and a run takes 10 seconds at most in all. Calls the code doesn't wait for still finish before `Run` returns. A failed call rejects its promise with an `Error` with the message of its error, and a `CallError` adds the status and code of the response as properties. A call has to return soon once its context is canceled, as `Run` gives up on it 5 seconds later.
