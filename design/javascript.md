# JavaScript

Status: step 1, the Run JavaScript block, implemented 2026-11-08. Later steps are proposals.

Kite runs user JavaScript in `kite-service/pkg/js`. The plan has three steps. The Run JavaScript block computes and returns a value, with no access to anything. Then scripts get host functions to call Discord, HTTP and stored variables. Maybe later, whole bot features are written as JavaScript modules that register commands and event handlers. The last step isn't decided.

## Runtime

Scripts run in QuickJS-NG compiled to Wasm, in the service process through wazero. The module is built from `pkg/js/wasm/kite.c` and a pinned QuickJS-NG release by `pkg/js/wasm/build.sh`, and the built `kite.wasm` is committed.

goja was rejected because it has no memory limit. kite-service runs every bot in one process, and `"x".repeat(2**30)` or doubling a string 30 times allocates gigabytes in a few instructions, before an interrupt can stop it. Go can't recover from running out of memory, so one script would take every bot down. In Wasm the memory has a fixed maximum, and an allocation past it fails as a JS error.

Supabase edge-runtime, which the Tessera spike used, was rejected for the block. It's a separate service, a network hop per run, and about 10 MB per V8 isolate. It's good at what a module mode needs, like JIT speed and isolates that stay warm, which a block doesn't.

`github.com/fastschema/qjs` was rejected after testing v0.0.6. It mounts the working directory as `/` and gives scripts quickjs-libc's `std` and `os`, so `std.loadFile("/kite.toml")` read the file. Its `MaxExecutionTime` does nothing, as the interrupt handler is commented out. Our glue is about 200 lines without quickjs-libc. A script gets the language, `console` and what `prelude.js` defines. The module imports the clock and four WASI stubs, and no filesystem, environment or network.

## What has to stay this way

`WithCloseOnContextDone` must be on. QuickJS only calls the interrupt handler from its interpreter loop and some builtins, so `new Array(2**28).join("x")` ran for 8.7 s past a 100 ms limit without it. With it, wazero kills the module wherever it is. Without it the same script also held the Go GC to 6 cycles in 8.8 s, because the goroutine couldn't be stopped inside the native loop. With it, a 1 ms ticker in another goroutine was never more than 0.5 ms late. It makes scripts 1.4 to 3.4 times slower.

Scripts that fill every core make other goroutines 25 to 60 ms late. `pkg/js` caps scripts running at once to half the cores, which kept that under 5 ms in a test with 48 spinning scripts on 12 cores. It also caps the memory scripts hold at once.

Compiling the module and setting up the snapshot takes 0.23 s on an M4 Pro and about 0.9 s on kite.onl's i7-6700, in each of its 4 processes. `js.Warm` does it a minute after the server starts, once the gateways have reconnected. A script that runs before then compiles it itself. On the i7-6700 a small script runs in 0.5 to 1.2 ms, and creating 2,000 small objects takes 10 ms.

Each run starts from a copy of the memory of an instance where QuickJS and the prelude are set up. That's valid because the module's only mutable global is the stack pointer, which is back at its start between calls. Starting from a copy takes about 0.15 ms, and a 2 KB string-processing script about 1.8 ms with the kill on, on an M4 Pro. Every run gets a fresh instance, so nothing a script changes carries over to the next.

Each instance's memory is allocated with the capacity of the snapshot through wazero's experimental `MemoryAllocator`. By default wazero sizes it for the module's initial 18 pages, so growing it to the snapshot's 21 pages copied it again. On Linux that took a trivial run from about 265 to about 200 µs and from 2.8 to 1.5 MB of allocations.

Copy-on-write memory was measured against that and isn't worth it yet. It maps the snapshot from a file privately for each instance through the same allocator, so the OS only copies pages a script writes. Creating the instance still writes the module's data segments, so those 113 KB have to be copied back from the snapshot. In Docker on Linux it ran a trivial script in 175 µs against 190 µs, and the same in parallel, while allocating 133 KB instead of 1.5 MB. Most of the remaining time is wazero creating the instance, not memory.

Reusing an instance and copying the snapshot back into it took 65 µs, the biggest win left for modules. An instance whose script grew its memory has to be thrown away, as Wasm memory can't shrink.

QuickJS seeds `Math.random` from the clock when a context is created, so instances copied from one snapshot would share the seed. `Math.random` calls Go instead.

There's no `Intl`. `prelude.js` adds `formatDate` with time zones from Go, and makes `Number.prototype.toLocaleString` group digits like en-US.

## Host functions

The C glue has one way out of the sandbox, `kite.call(name, json)`, which runs a Go function and blocks until it returns. `prelude.js` keeps a reference and deletes the global, so scripts only get the wrappers it defines. Today the only host function is `date_parts`, for `formatDate`.

Step 2 adds functions like `discord.request`, `http.fetch` and `variables.get` to `hostFuncs`. Each needs an argument schema, a credit cost, a rate limit, and whether a block can call it or only a module. Go validates every call as if an attacker wrote it, and the JS wrappers are only for convenience. Discord calls go through a route policy table like Tessera's, which says per route how its target is proven to be in the right server. The Discord API Request block and the MCP Discord API already check scopes, and should share that table.

The public API should return Promises from the start, so scripts `await` host calls. Go can resolve them right away at first, and run them in parallel later without breaking scripts. That needs the QuickJS job queue, which also lifts the block's ban on returning a Promise.

## Modules

A module's top level would only register handlers, like `command("level", options, async (ctx) => ...)` and `on("message_create", filters, fn)`. On deploy, Go runs the top level once in collect mode to get the manifest, with host calls failing. Then it takes a snapshot per app version, so each event starts with the module loaded. Go calls a second export, `kite_dispatch(handler, payload)`, which calls the handler. State at the top level only lasts within one instance, so durable state goes in storage.

MEE6's Bot Maker, launched in November 2026, works the same way and is worth copying in a few places:

- Time spent waiting on host calls shouldn't count against a script's time limit. Modules need about 1 s of CPU per run and a per-server budget per minute.
- One event at a time per module, so code needs no locks. That needs a much shorter wall clock limit than the 30 s of flows.
- Storage with atomic `incr` and `compareAndSet`, keys per member, and a top-N query for leaderboards.
- Caps on kicks, bans and deletes per module over 10 minutes, 1 hour and 24 hours, counted by distinct members, that pause the action until an admin resumes it.
- Event filters declared next to each handler and checked in Go before any JS runs.
- Settings declared in code that become flow parameters, so templates work for modules.

Modules should run in a worker process before they launch, with a lower CPU priority, so a wazero or QuickJS crash can't take the gateways down. Keeping `pkg/js` behind `Run` makes that a change of transport.

QuickJS is 20 to 50 times slower than V8 on heavy compute. Bot handlers mostly wait on Discord, but a game module like chess should be measured before modules are built on it.
