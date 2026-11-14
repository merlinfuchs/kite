# JavaScript

Status: step 1, the Run JavaScript block, implemented 2026-11-08. Step 2, operations from code, implemented 2026-11-09 for reads, stored variables and requests, and 2026-11-14 for writes to Discord and integrations. Reads of members, roles and channels come from the gateway cache when it has them. Code calls the AI blocks with `ai.*` since 2026-11-14, except Search the Web, whose searches take longer than a run can. Handler operations for voice are missing. The runtime of modules in `pkg/js` is implemented 2026-11-14, and the engine doesn't run modules yet.

Kite runs user JavaScript in `kite-service/pkg/js`. The plan has three steps. The Run JavaScript block computes and returns a value, with no access to anything. Then scripts get the operations of Kite's integrations, like Discord's API, stored variables and AI. Maybe later, whole bot features are written as JavaScript modules that register commands and event handlers. The last step isn't decided.

Flows, the Run JavaScript block and modules are three ways to use the same operations. Flows stay a graph that the Go engine runs, see "Flows stay a graph" below.

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

Step 2 gives scripts the operations of Kite's integrations through `kite.call`. Go validates every call as if an attacker wrote it, and the JS wrappers are only for convenience.

Scripts don't call blocks. A block is shaped for the editor. It has a curated subset of an operation's options, text fields with placeholders and results in Kite's shapes. The Create channel block leaves out options that are hard to show in the editor, and code should have all of them. So blocks and scripts both use the operations underneath, which `design/integrations.md` defines:

- The operations of an integration's `api.json`, like the 208 operations of Discord's API that Kite allows, generated from its OpenAPI spec, and the hand-written ones of Roblox, ER:LC and Cookie API.
- Operations that aren't a request to a service, like stored variables, AI and voice. They have a Go handler and a JSON Schema for their parameters and result.
- Requests to any URL, like the HTTP Request block sends, which 432 active free apps used in October 2026. They're the one operation of a built-in `http` integration and go through the egress proxy like today.

An integration's `api.json` is the list of what scripts may call, not a copy of the service's spec. Discord's leaves out the writes that would break how Kite manages the app, like registering application commands, which would replace the commands Kite syncs, and editing the application or its interactions endpoint URL. Regenerating it from Discord's spec has to keep them out, and a test should check that it does.

The SDK is generated from the same files, so a new operation is in the SDK without writing a wrapper. What's generated is types, a `.d.ts` for editors and the AI, and the Go table. The JavaScript side is one `Proxy` that turns `discord.createMessage(args)` into a call of `discord.create_message`, so 200 generated wrappers don't end up in every snapshot. The argument is one object with the path and query parameters by name, the body under `body` and the audit log reason of a write under `reason`, like `discord.createMessage({ channel_id, body: { content: "Hi" } })`. That's how discord.js's REST client takes them too. A body can be a list, and no operation can have a parameter named `body` or `reason`, which a test checks. An operation's name in the SDK is fixed once it ships. Discord can rename an `operationId` in its spec, and a regenerated `api.json` would then break the scripts that use it, so a renamed operation keeps its old name as an alias. One Go executor runs every operation for blocks, scripts and the MCP server. It adds the bot token or the integration's credential, sends requests through the egress proxy, checks the route policy, charges credits and writes logs. A script never sees an API key.

The route policy is generated per operation, like `block_definitions.json` for blocks. It's the `policy` of each operation in `kite-service/pkg/flow/operations.json`, with the operation's credits, what its target is, and whether it's destructive. Nothing checks the target yet, as every app has its own bot. A shared bot would have to prove the target is in the app's server. Code can call what the Discord API Request block can, except the operations of the interaction, which need its token. Limits on destructive operations, if any, should apply to blocks and code alike, per app over time. It replaces Tessera's hand-written table, and the scope checks of Discord API Request and the MCP Discord API tool should move to it.

Most of it can be derived. The target of a Discord operation follows from its path, like `{channel.id}` or the server, and credits default to free for reads and 1 for writes, like the get blocks and other actions today. Fields in the body can't be derived, so a short list written by hand covers them: a message reference that forwards a message, emoji IDs and `sticker_ids`, which can name things in another server the bot is in. The list matters most on a shared bot, where every server is one the bot is in.

Scripts get Discord's API JSON, like `global_name` and an `avatar` hash. Placeholders and the inputs of the Run JavaScript block use Kite's shapes, like `display_name` and `avatar_url`. That's on purpose. Modules are written against the real API, which docs and AI know. The block stays in the flow world.

With the app's own bot token, a script that burns the rate limit only hurts its own app. On a shared bot, the route policy and per-app budgets in Go have to stop it before Discord does.

Operations return promises, which Go settles while the script waits. Today Go calls `kite_run` once and the whole run happens in that call. It runs the code until its first `await`, then QuickJS's job queue in a loop in `kite.c` until nothing is left. Calls into Go, like the time zones of `formatDate`, return right away.

Step 2 makes that a back and forth. A call of an operation starts the request in a goroutine and returns an ID at once, and the prelude keeps a pending promise for it. When the job queue is empty and promises are still pending, `kite_run` returns that the run is waiting instead of a result. Go waits for the first request to finish and calls a `kite_resolve(id, result)` export, which settles its promise and runs the queue again. That repeats until the promise the code returned settles. Blocking in the host call would be simpler, but the calls of a `Promise.all` would run one after another.

The time limit then counts only the time inside `kite_run` and `kite_resolve`, when the script runs. Waiting on Discord doesn't count, and a cap on the wall time of the whole run, like 10 seconds, ends scripts that wait too long. A waiting run doesn't hold one of the slots for running scripts, and takes one again for `kite_resolve`. A run waiting between two calls has nothing on its stack, which is the state that saving a waiting run under Modules starts from.

## Modules

A module's top level only registers handlers, with `discord.command("level show", options, async (ctx) => ...)`, `discord.on("message_create", fn)` and `discord.component("claim", fn)`. The options of a command are Discord's. `js.LoadModule` runs the top level and returns the manifest. Calls fail while it runs, also the ones the code catches or doesn't wait for, so loading fails then. The top level isn't async, so it can't use `await`.

Loading keeps a snapshot of the instance, and every run of a handler starts from it through `__kiteDispatch` in the prelude. Parsing a module of 256 KB, the limit, takes 165 ms on kite.onl's i7-6700. A run from its 7 MB snapshot takes 3 ms, and one of a small module 0.5 ms. Loading gets 500 ms of CPU and a run 100 ms. The snapshots are cached by the hash of the code, up to 512 MB in each process, and a module that isn't cached loads again on its next event. Modules are often bundled, so 256 KB counts bytes, and bundlers have to output a script without `import`, like esbuild's IIFE format.

Globals start with what the top level set and don't carry over from one run to the next. Modules can't rely on that, as events may share instances later. Durable state goes in storage.

A handler of a command or component gets `ctx` with Discord's interaction, and `ctx.reply`, `defer`, `deferUpdate`, `update`, `showModal` and `followUp`. They call `interaction.respond` with the callback type and `interaction.follow_up`, which the host answers for the interaction of the run. A command's handler also gets `ctx.options`, the values of its options by name, and a component's the data saved with it in `ctx.data`. A handler of an event gets Discord's event in `ctx.event`. An event can have several handlers, which run one after another, and a failing one doesn't stop the others.

Durable state is stored variables, which code reaches by names it chooses, like `variables.get({ name: "xp", scope })`. The host turns each name into a variable ID, so code never holds IDs, and a variable can be renamed in the dashboard or get another ID when a template adds it. The Run JavaScript block declares its names under Stored variables, each with the variable picked in the editor. A module declares them at its top level instead, like `variables.define("xp", { scoped: true })`, which loading puts in the manifest. Installing the module creates the variables or links them to ones the app has, like a template install, and the runtime gets the same map of names to IDs. So `variables.*` and its handlers stay the same for both, and only where the map comes from differs. Looking variables up by the app's own names would be simpler, but breaks on a rename or when an install has to call a variable `xp-2`. Storage only a module sees, like the values of the old plugins, would need no declaration, but flows couldn't read it, like the XP a leveling module counts.

MEE6's Bot Maker, launched in November 2026, works the same way and is worth copying in a few places:

- Time spent waiting on host calls shouldn't count against a script's time limit. Modules get 100 ms of CPU per run, and need a per-server budget per minute.
- One event at a time per module, so code needs no locks. That needs a much shorter wall clock limit than the 30 s of flows.
- Storage with atomic `incr` and `compareAndSet`, keys per member, and a top-N query for leaderboards.
- Caps on kicks, bans and deletes per module over 10 minutes, 1 hour and 24 hours, counted by distinct members, that pause the action until an admin resumes it.
- Event filters declared next to each handler and checked in Go before any JS runs.
- Settings declared in code that become flow parameters, so templates work for modules.

Modules wait like Discord does, with handlers and saved data instead of paused code:

```js
command("giveaway", opts, async (ctx) => {
  const msg = await ctx.reply({
    content: "Click to enter!",
    components: [
      button("enter", { label: "Enter" }, { prize: ctx.options.prize }),
    ],
  });
  schedule.in("24h", "draw", { messageId: msg.id, prize: ctx.options.prize });
});

on.button("enter", async (ctx, data) => {
  /* data.prize */
});
on.schedule("draw", async (ctx, data) => {
  /* pick a winner */
});
```

Kite stores the data of a button and puts only an ID in its `custom_id`, which Discord caps at 100 characters, like a resume point of a flow. `schedule.in` is a timer resume point with a handler name instead of a block ID. Nothing is paused, so deploys, restarts and edits to the code can't break waiting runs. It also fits buttons better than `await`. A message's buttons can be clicked by anyone, many times, days later. Short waits in one run, like `await sleep(2000)`, can keep the instance alive within the time limit.

Each event gets its own instance at first, like a run of the block. An instance takes 1.4 MB when it starts and up to its memory limit, and the event holds it for as long as it waits on calls. That's why `Run` caps live runs at 128. Events of a module mostly wait on Discord, so later several of them can share an instance, like requests share Node's event loop. Only events of the same app and version share one, as they share its globals. The Run JavaScript block never does, since every run has to start fresh.

`kite_dispatch(event_id, handler, payload)` then keeps the promise the handler returns under the event's ID. After every call into the instance, `kite.c` checks the promises of all events and returns the ones that settled, with their results. A call into Go already has an ID, and the resolve functions of its promise are stored under it. So `kite_resolve(id, result)` settles exactly that promise, whichever event made the call. Promises the code creates itself are never looked up and need no owner. wazero can't call into one instance from two goroutines, so one goroutine owns each instance and takes events and call results from a queue.

Go hands over one result per `kite_resolve`. The glue runs the job queue until it's empty before it returns, so every call into the instance starts with an empty queue. The jobs it runs are then caused by the one event that was dispatched or whose call finished, and the time inside the call counts against that event's CPU budget. Go counts the calls started during it toward that event too, for the 50 calls per run and the credits. The exception is a promise events share through globals, like a cached `discord.getGuild()`. All events waiting on it continue in one call, which counts against the event whose result arrived. Several results per call would save calls into Wasm, but those take microseconds, and the jobs of several events would run mixed together.

An event that goes over its budget is stopped by the interrupt handler, which throws an uncatchable error in the job that's running. QuickJS unwinds cleanly, and the other events keep running. Go cancels the event's calls that are still running and rejects their promises, so its code unwinds and its memory is freed. Native loops that never check the interrupt, like `join` on a sparse array of 2^28 items, can only be stopped by `CloseOnContextDone`, which ends the whole module. That kill stays as a backstop a few seconds after the interrupt, and it fails every event in the instance. Those events can't be retried, as they may have sent messages already.

Memory is shared too. The memory limit is per instance, so one event that uses most of it makes the others fail, and QuickJS can't tell which event allocated what. A cap of a few events per instance keeps a kill or a full heap from failing many events at once.

Events in one instance take turns at each `await`. Code between two awaits runs without interruption, but state in globals can change across an await. That's the opposite of running one event at a time per module, as MEE6 does above. Durable state goes in storage with atomic operations either way. Which of the two modules get is undecided.

If people want `await ctx.waitForClick()` later, the instance's memory can be saved while it waits. A paused `async` function lives in QuickJS's heap, which is the Wasm memory. When every function waits on a promise that isn't resolved yet, `kite_run` returns, and the memory is the run's whole state. Kite would save the pages that differ from the snapshot, compressed, and later copy them into a new instance and call a `kite_resolve(id, result)` export. Nothing runs again, so code doesn't have to be deterministic like with replay in Temporal or Durable Functions. A new `kite.wasm` breaks saved runs, as its memory layout differs, so old builds have to be kept as long as runs can wait, or the runs fail after an upgrade. A saved run also keeps the code it started with. The size of saved runs needs a spike before anything depends on it.

Modules should run in a worker process before they launch, with a lower CPU priority, so a wazero or QuickJS crash can't take the gateways down. Keeping `pkg/js` behind `Run` makes that a change of transport.

QuickJS is 20 to 50 times slower than V8 on heavy compute. Bot handlers mostly wait on Discord, but a game module like chess should be measured before modules are built on it.

## Flows stay a graph

Flows could be compiled to JavaScript and run like modules, so there's one way to run things. They aren't, for these reasons:

- Flows pause. A modal, a button on a message or a sleep stores a resume point, the ID of a block and the results so far, and the flow continues from it minutes or days later, after deploys. A paused JavaScript call stack can't be stored. Compiling flows to steps that can be resumed gives back the graph.
- Placeholders are expr-lang in Go, with lazy fields, `origin.`, `previous.`, `arg()` and `result()`, and behavior apps depend on, like the IDs of modal selects. Rewriting that in JavaScript risks every flow, and calling Go for each placeholder leaves JavaScript only ordering calls to Go.
- Every block calls Go anyway. Compiling would replace the graph interpreter, about 1,700 lines. The 5,700 lines of block handlers stay either way.
- Error logs, field highlighting and test runs point at blocks.

Speed isn't the reason. 967k runs a day at 0.5 ms each would be about 8 minutes of CPU a day on kite.onl's server.

What's shared is the operations. A request block is a preset of the operation in its `run.operation`, so a flow can be turned into a module by turning each block into a call of its operation and each button output into a handler. Custom blocks need that mapping written once each. The other way only works for code that is a list of operations, and isn't a goal.

## One interface for flows, modules and plugins

Flows, modules and plugins can all sit behind one interface in the engine. Each of them declares what it handles and handles it:

- A manifest with the commands, event listeners and their filters, schedules and webhooks it handles, and the settings it declares.
- A method that gets a trigger, like a command, an event, a schedule or a click on a button with its saved data, and runs it within the app's limits, credits and logs.

A flow is a unit with one trigger, its entry block. A module has as many as it declares. Plugins already work this way: `pkg/plugin` has `Commands()`, `Events()` and `Config()`, and an instance has `HandleCommand`, `HandleEvent`, `HandleComponent` and `HandleModal`. Counting and starboard are Go behind it. The interface for modules would grow out of it, with flows, modules and the plugins that are left as its implementations.

The engine routes events by the manifests and doesn't care what's behind them. The commands Discord gets and the intents an app needs are computed from all manifests together, where commands, event listeners and plugins each have their own path today. A resume record holds either a block ID and the state of a flow, or the name of a handler and its saved data, so one table and one dispatcher serve both.

Things to get right:

- Plan limits count what's in the manifests, not units. A module with 5 commands counts as 5 commands, or modules get around the 15 commands of the free plan.
- The Commands and Event listeners pages show those of modules too, marked as coming from their module and only editable there. Otherwise nobody sees what their bot does, and names that clash with flows go unnoticed.
- A module is one more kind of item in a template, as an installed template already groups commands, event listeners and messages like a module does. That keeps the Marketplace working for code.
- Flows stay a row per command or event listener. A module is one row with its code and the manifest Go takes from it on deploy, and routing is built from both.
