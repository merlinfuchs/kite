# Kite

Kite is a no-code Discord bot builder. Users build commands and event listeners as flows of blocks in the web editor, and the service runs them.

- `kite-service`: Go backend (API, Discord gateway, flow engine). Go 1.25.
- `kite-web`: Next.js 14 app (pages router, pnpm). The flow editor lives here.
- `kite-docs`: Docusaurus docs site (pnpm).
- `deploy`: scripts the maintainer uses to deploy kite.onl. Don't change them unless asked.

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
| `kite-service/pkg/flow/catalog.json`          | zod schemas and `nodeTypes` in `kite-web/src/lib/flow`   | `pnpm test -u` in `kite-web`         |

`catalog.json` describes every block to the flow AI. `TestCatalogHasEveryNodeType` fails if a block is missing from it.

## Adding a block

Look at an existing block that does something similar and copy its shape. `action_message_pin` is a good minimal example. Before adding a new block, check that no existing block already covers the use case, or could with one extra option.

Service:

1. `pkg/flow/data.go`: add the `FlowNodeType` constant and any new fields on `FlowNodeData`. Reuse existing fields (`ChannelTarget`, `MessageTarget`, `AuditLogReason`, ...) where they fit.
2. `pkg/flow/execute.go`: add a `case` to the switch in `Execute`. Evaluate inputs with `ctx.EvalTemplate`, return errors with `traceError(n, err)`, and store a result if the block returns data.
3. `pkg/provider/discord.go`: add the method to the `DiscordProvider` interface and to `MockDiscordProvider`.
4. `internal/core/engine/providers.go`: implement the method. The engine's provider embeds the mock, so if you skip this it compiles and silently does nothing.
5. `CreditsCost()` in `pkg/flow/execute.go`: `action_*` blocks cost 1 by default. Blocks that call external services or do a lot of work cost more.

Web:

1. `src/lib/flow/dataSchema.ts`: zod schema for the block's data. Every field needs `.describe(...)`.
2. `src/lib/flow/resultSchema.ts`: schema for the result, if the block returns data. Without it the placeholder picker and flow AI can't see the output.
3. `src/lib/flow/nodes.ts`: entry in `nodeTypes` with title, description, icon, `dataFields`, schemas and `creditsCost`.
4. `src/lib/flow/components.ts`: map the type to a component (usually `FlowNodeActionBase`).
5. `src/lib/flow/categories.ts`: add the type to a category.
6. `src/components/flow/FlowNodeEditor.tsx`: only if you added a new field name that needs an input.
7. Run `pnpm test -u` to regenerate `catalog.json`, and `tygo generate` if you changed Go types.

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
