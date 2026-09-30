# Kite

Kite is a no-code Discord bot builder. Users build commands and event listeners as flows of blocks in the web editor, and the service runs them.

- `kite-service`: Go backend (API, Discord gateway, flow engine). Go 1.25.
- `kite-web`: Next.js 14 app (pages router, pnpm). The flow editor lives here.
- `kite-docs`: Docusaurus docs site (pnpm).

## Checks

CI runs all of these on every PR. Run them before pushing.

```shell
# Service
cd kite-service
gofmt -l .        # must print nothing
go vet ./...
go test ./...

# Generated files, must leave no diff (see below)
go run github.com/gzuidhof/tygo@v0.2.19 generate
sqlc generate     # sqlc v1.31.1

# Web
cd kite-web
pnpm install
pnpm run format   # prettier
pnpm test -u      # also regenerates catalog.json
pnpm run build    # typechecks and lints

# Docs
cd kite-docs
pnpm install
pnpm run build
```

## Generated files

Never edit these by hand. Change the source and regenerate.

| File                                          | Source                                                   | Command                              |
| --------------------------------------------- | -------------------------------------------------------- | ------------------------------------ |
| `kite-web/src/lib/types/*.gen.ts`             | Go types in `kite-service` (see `tygo.yaml`)             | `tygo generate` in `kite-service`    |
| `kite-service/internal/db/postgres/pgmodel/*` | `internal/db/postgres/queries` and `migrations`          | `sqlc generate` in `kite-service`    |
| `kite-service/pkg/flow/catalog.json`          | block definitions in `kite-web/src/lib/blocks`           | `pnpm test -u` in `kite-web`         |
| `kite-service/pkg/flow/block_definitions.json` | block definitions in `kite-web/src/lib/blocks`         | `pnpm test -u` in `kite-web`         |

`catalog.json` describes every block to the flow AI, and `block_definitions.json` tells the service how each block runs. `TestCatalogHasEveryNodeType` and `TestEveryBlockRuns` fail if a block is missing from them.

## Adding a block

Look at an existing block that does something similar and copy its shape. Before adding a new block, check that no existing block already covers the use case, or could with one extra option. See `design/integrations.md` for the design.

Every block has a definition in `kite-web/src/lib/blocks`, named after the integration it mainly acts on and its type, like `discordInviteCreate.ts` for `action_invite_create`, and listed in `blocks/index.ts` in the order of the block explorer. Blocks for a service other than Discord include the service in their type, e.g. `action_roblox_user_get`, so they can't collide with Discord blocks or blocks of other services.

If the block is a single API request, give it `fields` and a request `run`, following `discordInviteCreate.ts`. The service runs it from the generated `block_definitions.json`, a test checks it against its integration's `api.json`, and it needs no Go. Otherwise give it a custom `run` and `fields` whose `schema` describes each setting, following `discordChannelGet.ts`, and write it by hand. A field is edited by the input registered under its name in `FlowNodeEditor.tsx`, or the one it names in `input`. Blocks built around a widget like the message builder keep a block `schema` and `inputs`, like `discordMessageCreate.ts`.

Service:

1. `pkg/flow/data.go`: add the `FlowNodeType` constant and any new fields on `FlowNodeData`. Reuse existing fields (`ChannelTarget`, `MessageTarget`, `AuditLogReason`, ...) where they fit.
2. `pkg/flow/handlers_*.go`: add a handler function to the file of its group, like `handlers_message.go`, and register it in the file's `init`. Evaluate inputs with `ctx.EvalTemplate`, return errors with `traceError(n, err)`, and store a result if the block returns data.
3. `pkg/provider/discord.go`: add the method to the `DiscordProvider` interface and to `MockDiscordProvider`.
4. `internal/core/engine/providers.go`: implement the method.
5. `CreditsCost()` in `pkg/flow/execute.go`: `action_*` blocks cost 1 by default. Blocks that call external services or do a lot of work cost more. A test checks it matches `credits` of the definition.

Web:

1. `src/lib/flow/dataSchema.ts`: zod schemas for settings other blocks share, like `channelTargetSchema`. A setting's schema needs `.describe(...)`.
2. `src/lib/flow/resultSchema.ts`: schema for the result, if the block returns data. Without it the placeholder picker and flow AI can't see the output.
3. `src/lib/blocks`: the definition, with title, description, icon, category, fields, result and credits.
4. `src/components/flow/FlowNodeEditor.tsx`: only if you added a new field name that needs an input.
5. Run `pnpm test -u` to regenerate `catalog.json` and `block_definitions.json`, and `tygo generate` if you changed Go types.

Docs:

1. Add `kite-docs/docs/reference/blocks/actions/<type>.md`, following `action_message_pin.md`. CI fails if a block has no page.
2. Add a line to `kite-docs/docs/reference/blocks/index.md`.

## Adding an event listener type

1. `internal/model/event_listener.go`: add the `EventListenerType`. The value must be the lowercased Discord gateway event name.
2. `internal/model/app.go`: add it to the matching `Needs*` method if it needs a privileged or non-default intent. Only request intents when an app actually uses them.
3. `internal/core/engine/event_listener.go`: allow it in `shouldHandleEvent`, including any filtering (e.g. ignore the bot's own actions).
4. `internal/core/engine/data.go`: user, guild and channel IDs for the event.
5. `pkg/eval/ctx.go`: fill the placeholder env (`user`, `member`, `channel`, `guild`, ...) for the event.
6. `internal/core/gateway/events_test.go` and `helpers_test.go`: add the type to the existing test tables.
7. Web: `event_type` enum in `src/lib/flow/dataSchema.ts`, the options in `FlowNodeEditor.tsx` and `EventListenerCreateDialog.tsx`. Regenerate `catalog.json`.
8. Docs: `kite-docs/docs/reference/event.md`.

## Database

Migrations are in `kite-service/internal/db/postgres/migrations` as `NNN_description.up.sql` and `.down.sql`. Take the next free number and check `main` hasn't used it since you branched. After changing migrations or queries, run `sqlc generate`.

## Rules

- Match the surrounding code: naming, structure, comment density. Run gofmt and prettier.
- Keep PRs to one feature. No drive-by refactors, reformatting, dependency bumps or Go version changes.
- Never vendor dependencies. Add Go dependencies with `go get` so `go.mod` and `go.sum` are updated, and don't paper over build errors with `go mod tidy` in the Dockerfile.
- Don't delete or loosen `.gitignore`, validation, tests or existing comments unless that's the point of the change.
- Treat templated input as untrusted. An empty or invalid number should be an error, not silently become 0 or "unlimited".
- A flow execution times out after 30 seconds. Don't build blocks that need to run longer.
- Destructive blocks (deleting things, leaving servers, bulk actions) need sensible limits and must not be easy to aim at the wrong target.
