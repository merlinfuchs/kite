# Integrations

Status: phase 1 implemented for Discord, 2026-09-29. Later phases are proposals.

Integrations let Kite call other services through blocks that are defined as data instead of code. An integration is an API with a spec, a host, an auth scheme and a set of blocks. An app connects it once by entering a credential, and its blocks work in every flow of the app. Discord is an integration too, always connected with the app's bot token.

## Why

Most action blocks are the same thing: a few inputs, one HTTP request, a result. Adding one still touches about a dozen files across Go, TypeScript and docs (see "Adding a block" in AGENTS.md), and forgetting the provider implementation compiles and silently does nothing.

Meanwhile users work around the missing blocks with the HTTP request block, and pay for it with their credentials. From production flows (2026-09-29, commands and event listeners only):

- 207 apps call `discord.com` from HTTP blocks, and on the most used endpoints 85 to 90% paste their bot token into a header.
- 178 apps call cookie-api, and 95 of them paste a Discord bot token, almost all for `/api/transcript`. None use cookie-api's own stored-token option.
- 89 apps call cookie-api's Roblox user lookup, although Kite has had `action_roblox_user_get` since 2025-07. cookie-api also returns an avatar URL, which is probably why.

Pasted credentials end up in exports and share codes, which any logged-in user can read.

The Discord API Request block (#469) already has most of the machinery: endpoints generated from Discord's OpenAPI spec, path building that can't be pointed at another endpoint, placeholder evaluation inside JSON bodies (`EvalJSONTemplate`), and results that later blocks can read field by field (`thing.NewFromJSONValue`). This design generalizes it.

## Concepts

**Integration.** A service Kite can talk to, defined in the repo, one directory each. It has an id, name, icon, an OpenAPI spec, a base URL and an auth scheme. Discord, cookie-api and Roblox would be the first ones. Integrations don't list blocks: blocks reference the integrations they need.

**Credential.** What an app enters to connect an integration, usually an API key. It's stored per app, encrypted, write-only, and bound to the integration's host. Discord's credential is the bot token the app already has.

**Block definition.** One file per block, and self-contained: the fields the user fills in with their types, how the block runs, e.g. a request to an integration and how the fields map onto it, and what the block returns. Both the service and the editor read the same file, and nothing at runtime reads a spec. A block needs the integrations its requests go to, plus any it lists in `requires`, and only works when the app has them connected.

**Spec.** Every integration has an `openapi.json`, the spec Kite builds from. For a service with an official spec it's a trimmed copy at a pinned commit, written by a script. For one without, like cookie-api, it's written by hand from their docs. Nothing at runtime reads it: it's the input for generating block definitions, for the raw request block's operation list, and for checking definitions when the service changes its API.

**Raw request block.** An integration can turn on a raw block like Discord API Request, where any operation of its spec can be picked, as the escape hatch for endpoints nobody has curated yet. The block is generated for the integration, and its operation list is a build output generated from `openapi.json`, like `discordApi.json` today.

## Definition files

Definitions are TypeScript objects, so the compiler checks them and result schemas can be zod schemas like those of other blocks. A test writes what the service needs to `kite-service/pkg/flow/block_definitions.json`, which the service embeds, the same way `catalog.json` is generated. `pnpm test -u` updates it and CI fails when it's stale. JSON or YAML files read by both sides would need a JSON Schema and a validator on each side for the checks TypeScript gives for free.

Blocks and integrations live in separate folders and reference each other by ID:

```
kite-web/src/lib/
  integrations/
    types.ts
    index.ts             all integrations
    discord/
      index.ts
    cookie_api/          later
      index.ts
      openapi.json       hand-written from their docs, they publish none
  blocks/
    types.ts
    index.ts             all blocks, and the editor schema for them
    discordMessageBulkDelete.ts
    discordInviteCreate.ts
    discordRoleCreate.ts
    discordMessageList.ts
    controlConditionCompare.ts  a block of Kite itself
```

The blocks folder is flat. File names are the integration a block mainly acts on, if any, followed by its type, like `discordInviteCreate.ts` for `action_invite_create`, so sorting groups them. Blocks of Kite itself, like conditions and variables, have no prefix. Category folders would repeat the `category` field and drift from it when a block moves to another section of the block explorer. A block that needs several integrations, like a transcript that reads Discord messages and renders them with cookie-api, lists them all.

Discord's spec is trimmed to `src/lib/flow/discordApi.json` by `scripts/discord-api.mjs`, which the raw block uses already. Other integrations would keep theirs in their folder as `openapi.json`.

Proposed for later integrations, the integration's own settings in its `index.ts` (shown as JSON):

```json
{
  "id": "cookie_api",
  "name": "Cookie API",
  "icon": "cookie",
  "base_url": "https://api.cookie-api.com",
  "spec_source": null,
  "auth": {
    "type": "header",
    "name": "Authorization",
    "label": "API key",
    "help_url": "https://docs.cookie-api.com/en/docs/getting-started/faq/api-key/"
  },
  "raw_block": false
}
```

`spec_source` is the URL and commit an `openapi.json` was trimmed from, and `null` for a hand-written one. `auth.type` is `header`, `query` or, for Discord only, `discord_bot`. OAuth for a user's own login to another service is out of scope.

A block definition, as implemented:

```ts
export const discordRoleCreate: BlockDefinition = {
  type: "action_role_create",
  title: "Create role",
  description: "Create a new role in the server",
  icon: "shield-plus",
  category: "Roles",
  credits: 1,
  audit_log_reason: true,
  run: {
    kind: "request",
    integration: "discord",
    operation: "create_guild_role",
    method: "POST",
    path: "/guilds/{guild_id}/roles",
  },
  fields: [
    {
      name: "guild_target",
      in: "path",
      target: "guild_id",
      type: "snowflake",
      label: "Server",
      description: "ID of the server. Leave empty to use the server the flow runs in.",
      fallback: "guild",
    },
    { name: "name", in: "body", type: "string", max_length: 100, label: "Name", description: "..." },
    { name: "permissions", in: "body", type: "string", widget: "permissions", label: "Permissions", description: "..." },
    { name: "color", in: "body", type: "integer", min: 0, max: 16777215, label: "Color", description: "..." },
    { name: "hoist", in: "body", type: "boolean", label: "Show Separately", description: "..." },
  ],
  result: { thing: "discord_role", schema: roleResultSchema },
};
```

`run.integration` is the integration the request goes to, which the block needs. Blocks that need more, or that run custom code, list integrations in `requires`:

```ts
export const cookieApiTranscriptCreate: BlockDefinition = {
  type: "action_cookie_api_transcript_create",
  // ...
  requires: ["discord", "cookie_api"],
  run: { kind: "custom" },
};
```

A block's integrations are those of its requests and its `requires` together. Discord blocks name Discord too, although it's always connected, so they keep working if Discord ever becomes one platform among several.

A field's `name` is the setting in the node's data, and `target` its name in the request if it differs. Field types are `snowflake`, `snowflake_list`, `integer`, `boolean` and `string`, checked in the editor and again when the flow runs. `fallback` fills an empty field with the server or channel the flow runs in, like `guild_target` of other blocks.

The generator copies what the spec knows into the definition: method, path, types, required fields, allowed values, limits and the response schema. People then add what the spec lacks: which fields to show and in what order, labels and descriptions (Discord's spec describes about 10% of its properties), input widgets, fixed or hidden values, and how the result is exposed. `run.operation` stays as a reference back to the spec, for the drift check below.

`result.thing` wraps the response as an existing thing type, so a migrated block keeps `{{result('x').mention}}` working, and `result.list` a list of them. Without it the result is the plain JSON, like Discord API Request.

Block types include the id of the integration they mainly act on, e.g. `action_cookie_api_transcript_create`, so blocks of different integrations can't collide. Folders can move, types can't. Discord's existing blocks, like `action_message_create`, keep their names. New Discord blocks may use the plain form, since Discord owns it. Nothing in the definition format is specific to Discord: Discord is just the integration whose auth type is `discord_bot` and that is always connected.

Destructive operations need limits, like a maximum count, which AGENTS.md asks for and review should check.

## Flow data

Every block has its own node type, e.g. `action_role_create`, like built-in blocks today. The type decides the integration, the request and the host. The node's data only holds what the user filled in, as flat settings like every other block:

```json
{
  "id": "k3j2",
  "type": "action_role_create",
  "data": {
    "name": "{{arg('name')}}",
    "color": "16711680",
    "hoist": "true",
    "audit_log_reason": "Requested by {{user.username}}",
    "temporary_name": "role"
  }
}
```

One shared type with the block named in the data would need a second lookup in everything already keyed by node type: the block explorer, docs pages, the flow AI's catalog, validation and allowed blocks per flow type (#393). It would also let edited flow data turn one block into another.

`FlowNodeData` keeps settings it has no field for in `Fields` when it's decoded and writes them back next to the others, so adding a block needs no Go struct field. A field named like an existing setting, e.g. `channel_target` or `name`, reads that setting instead, which a test limits to text settings. Values are usually templates, but numbers, booleans and lists are accepted, since the flow AI writes them that way.

An earlier version kept the values in a `fields` object. The flow AI kept writing them flat anyway, copying other blocks, so the layout now matches them. Converted built-in blocks keep their current data for the same reason.

Node types are permanent. A breaking change gets a new type, e.g. `..._v2`, and the old one keeps working. A node whose type no longer exists, e.g. from a removed integration, renders as the same error node as a disconnected integration.

## Execution

The service embeds all definitions and builds a table of node type to block definition at startup. A node of an integration block runs like this:

1. Look up the definition by node type (`blockDefinitions` in `pkg/flow/block_definitions.go`). The node's data holds only field values, never the operation or the host, so an imported flow can't point a block at another endpoint.
2. Check the app has the block's integrations connected, otherwise fail with "Cookie API isn't connected".
3. Evaluate each field and place it: path parameters through the same validation as Discord API Request (IDs must be IDs, no `/`, `\`, `.` or `..`), query parameters encoded, body fields set at their JSON pointer with `EvalJSONTemplate` semantics. Fields not in the definition are ignored.
4. Send the request. Discord goes through the session client, which adds the token and shares the rate limiter. Everything else goes through the HTTP provider and the egress proxy, with the credential added by the executor and redirects not followed.
5. Parse the result and store it, typed if `result.type` is set.

The credential is only ever added to requests to the integration's own base URL, which comes from the repo, never from a flow. It never appears in templates, results, errors or logs. This is the same property Discord API Request has for the bot token, extended to every integration.

Credits come from the definition. Partner APIs that charge Kite can cost more.

## Editor

The block explorer groups integration blocks by integration. Blocks of integrations the app hasn't connected still show, with a prompt to connect it, the same way premium blocks show that they need Premium.

Blocks of a disconnected integration that are already in a flow, or arrive through an import or the flow AI, render as an error node: the block keeps its settings, shows "Cookie API isn't connected" and links to the integration settings. Validation reports it as an issue so the flow AI sees it too, and at runtime the block fails with the same message instead of a 401 from the service. Importing a flow lists the integrations it needs.

The settings form is rendered from the definition, reusing `BaseInput` so every field accepts placeholders. Nested or `oneOf` bodies that a definition doesn't cover fall back to the JSON editor from #469. The result schema goes into the catalog like the one of any other block.

## Connecting an integration

A new settings page lists the integrations. Connecting one asks for the credential described by `auth`, and optionally tests it.

Storage, if #419 doesn't provide it (see below):

```sql
CREATE TABLE app_integrations (
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    integration_id TEXT NOT NULL,
    credential_encrypted TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    PRIMARY KEY (app_id, integration_id)
);
```

Encrypted like `apps.discord_token`, write-only over the API, one credential per integration per app. As with #419, every collaborator of the app can use it, so this protects against leaks through exports and share codes, not against a malicious collaborator.

## Where #419 fits

#419 (app-level secrets) is the general version of the same problem: users reference a secret by name with `{{secrets.NAME}}`, and it's only resolved in HTTP block fields. It's still needed for APIs without an integration.

The two should share one implementation: the encrypted storage, the write-only API, the redaction of resolved values from logs and errors, and the import warning for missing names. An integration credential is then a secret with an `integration_id`, which the executor attaches to requests to that integration's host instead of a template resolving it. That's stricter than a named secret, which a flow can send to any URL.

So #419 comes first. Its `app_secrets` table has a nullable `integration_id` next to the nullable `name`, and exactly one of them is set: named secrets are what `{{secrets.NAME}}` resolves, integration credentials are what the executor attaches.

Once integrations exist, most users won't need named secrets at all. The ones who still paste a Discord token into an HTTP block get a warning in the editor whenever a value looks like a bot token.

## Flow AI

The catalog is about 190 KB and sent with every request, so hundreds of integration blocks can't all go in it. The catalog would carry the blocks of integrations the app has connected. Everything else is found with a search tool the AI can call, which also fits running the flow AI through MCP.

What phase 1 showed with gpt-5-mini:

- It copies the settings layout of other blocks rather than reading a block's schema, which is why settings are flat.
- It prefers Discord API Request for anything Discord-shaped. The catalog is now ordered like the block explorer, the raw block's description lists the endpoints with their own block, and the instructions say to use raw request blocks only when no other block fits. It picks Create invite reliably now, but still often uses the raw block to list messages or create roles. Those flows work, just less readable.
- Warning it to switch blocks made it swap back and forth until it ran out of repairs, leaving broken flows. The editor still suggests the dedicated block to users, but the warning isn't sent to the AI.

## Adding an integration

1. Add `openapi.json`: trimmed from an official spec by the script, or written from the service's docs, as for cookie-api.
2. A script drafts `integration.json` and block definitions from the spec with an LLM, and opens a PR.
3. CI checks every definition against the definition JSON Schema and against `openapi.json`: the operation exists, every field exists at its location with the same type, required fields are covered or fixed, labels and descriptions are present.
4. Review decides which endpoints deserve a block, the wording, and whether something is destructive.

Adding a block no longer needs Go or TypeScript changes, so reviews are small and uniform.

## Compatibility

Saved flows reference block types and field names forever, so both are permanent once released. Renaming means adding a new field and deprecating the old one.

When a service changes its API, a script fetches the new spec, writes the trimmed `openapi.json` and compares it with every definition by `operation`: removed operations, removed or renamed fields, changed types and new required fields. It reports what needs a decision. CI runs the same comparison against the committed `openapi.json`, so it never has to fetch anything.

## One format for all blocks

Some blocks will always need code: conditions, loops, sleep, variables, AI, responses, voice and status. Two ways of defining blocks side by side is the complexity to avoid, so every block has a definition, and only how a block runs differs.

Before, a hand-written block was spread over `nodes.ts`, `dataSchema.ts`, `resultSchema.ts`, `categories.ts`, `components.ts` and a `case` in `Execute`. Now every block has one definition with its title, icon, category, credits, settings, result, structure and a `run`, and `nodes.ts`, the block explorer's sections and the canvas components are built from them:

```ts
export const discordMessagePin: BlockDefinition = {
  type: "action_message_pin",
  title: "Pin channel message",
  description: "Bot pins a message in a channel",
  icon: "pin",
  category: "Messages",
  requires: ["discord"],
  credits: 1,
  schema: nodeActionMessagePinDataSchema,
  inputs: ["channel_target", "message_target", "audit_log_reason", "custom_label"],
  run: { kind: "custom" },
};
```

A custom block runs the Go handler registered under its node type in `nodeHandlers`, so no separate ID is needed. `block_definitions.json` lists every block, and tests check that every custom block has a handler and every handler a definition, and that the credits of a definition match what `CreditsCost` charges.

All definitions live in the flat `blocks` folder, whether they're requests or custom. Blocks of Kite itself, like conditions, loops, sleep, variables, AI, calculate value and log, need no integration. Custom blocks that call a service, like pin or ban, name it in `requires`, so connect prompts and error nodes work the same for every block. Roblox is an integration too, without a credential.

The definitions are listed in `blocks/index.ts` in the order of the block explorer. The explorer's sections take their blocks from the definitions' `category`, and the flow AI's catalog follows the same order, which is how the move could be checked: the only change in the catalog is the order of the raw block's list of dedicated endpoints.

### Widgets

Custom blocks name the editor inputs that edit their settings in `inputs`, which are the widgets: `message_data` (the message builder, with templates), `emoji_data`, `modal_data`, `channel_data`, `command_permissions` and the rest of the inputs registered in `FlowNodeEditor.tsx`. Their settings keep their zod schema. Blocks defined with `fields` get a generated schema and form instead, and a field can use a widget too, like `permissions`.

Moving a custom block from `schema` and `inputs` to `fields` is optional and can happen block by block, whenever a block's settings fit plain fields. Outputs that depend on settings, like one per button of a message, still come from `getNodeOutputs`.

### Structure

Conditions and loops own other blocks, have several outputs and their own canvas components. That structure is in the definition now:

```ts
export const controlConditionCompare: BlockDefinition = {
  type: "control_condition_compare",
  // ...
  outputs: [],
  owns: ["control_condition_item_else", "control_condition_item_compare"],
  component: "condition_compare",
  run: { kind: "custom" },
};

export const controlLoop: BlockDefinition = {
  type: "control_loop",
  // ...
  owns: ["control_loop_end", "control_loop_each"],
  component: "control_loop",
  run: { kind: "custom" },
};
```

Condition items and the loop's each and end blocks are definitions of their own, without a category. `createNode` creates the owned blocks, the first to the right and the second to the left, and `getOwnedChildTypes` reads `owns`. What these blocks do stays in code: branching, looping and resuming in Go, adding condition items and the canvas components in the editor. The definition only names them.

## Existing blocks

About 18 of Kite's actions are a single Discord REST call: reactions, pin and unpin, message delete, ban, unban, kick, timeout, member edit, role add and remove, the channel blocks and the thread blocks. After the step above they're definitions with a custom `run`, and converting one means replacing its Go handler with a request `run`, until custom handlers are left only where there's real logic: responses, message blocks with the builder, voice and status (gateway), AI, variables and control flow.

A converted block keeps its node type, its field names (`channel_target`, `message_target`, `emoji_data`, ...) and its result type, so saved flows keep working unchanged. Definitions therefore map existing field names onto the request, and a few fields need a named conversion, done by the field types `emoji` (`emoji_data` becomes `name:id` in the reaction path) and `seconds_until` (a timeout's duration in seconds becomes the `communication_disabled_until` timestamp). A block that needs more than a named conversion stays in code. Converted blocks keep their zod schema and editor inputs, so the editor and the flow AI see no change, and their fields only describe the request. The ban's `seconds` field drops fractions like the Go code did. A timeout is marked `partial`, as it's one of several things `update_guild_member` does, so the raw request block isn't pointed at it.

The get blocks (`action_message_get`, `action_channel_get` and the others) read the gateway cache before calling Discord. A converted get block would always call Discord and use more of the rate limit on busy event flows, so they convert last, once definitions can say "check the cache first".

Each converted block has a test for the exact request it sends, matching the arikawa call it replaced. Then the Go handler and the provider method are deleted.

## Future: platform integrations

Integrations as described here only add actions. What makes Discord the primary platform isn't its API calls but everything flows are built around: triggers (gateway events, commands, buttons, select menus, modals), the message builder, the placeholder environment (`user`, `member`, `channel`, `guild`) and the app model of one bot account per app with a gateway connection.

Discord could at some point become one platform among others, like Revolt, Matrix or Telegram. That would mean a second kind of integration, a platform integration, which also declares triggers, an identity, messages and a placeholder environment. Flows would target one platform, with that platform's blocks, instead of shared blocks like a single "send message" for every platform. Shared blocks end up with only what every platform supports: no components, no embeds, no threads.

That's far off. The two rules above keep it open without extra work now: block types are namespaced by integration, and the definition format doesn't assume Discord.

## Phases

1. Done: definition format, executor and editor rendering, with create invite (#212), bulk delete (#210), role create (#208) and the message list from #468.
2. Done: one format for all blocks. Every block has a definition with a custom or request `run`, custom blocks name their widgets, conditions and loops their structure, and Go runs custom blocks from a handler map instead of the switch in `Execute`.
3. Done: 14 blocks that are a single request run as requests: message delete, reactions, pin and unpin, ban, unban, kick, timeout, member roles, channel delete and thread members. They keep their schema and editor inputs, and their fields only describe the request, with the field types `emoji` and `seconds_until` for the two conversions they need. Still custom: member edit (nested settings), channel and thread create and edit and forum posts (settings that are whole objects), and the get blocks, which read Kite's cache.
4. #419 with room for integration credentials, then the integrations settings page, connect prompts and error nodes. Cookie API as the first non-Discord integration, pending the partnership. For transcripts, ask cookie-api for a mode where Kite sends the messages instead of a bot token.
5. The LLM draft script and contributor docs.
6. Triggers from other services (the webhook listener, #181), if Kite goes beyond Discord. Integrations only add actions.

## Open questions

1. Should the flow AI's catalog include blocks of disconnected integrations, so it can suggest connecting them?
2. Test the credential when connecting, which needs a test operation per integration?
3. Partner pricing: do cookie-api calls cost more credits, and who pays cookie-api?
