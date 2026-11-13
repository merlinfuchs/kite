// Generates src/lib/integrations/discord/api.json, the endpoints the Discord
// API Request block can call, and the TypeScript types of their bodies and
// results for code, from Discord's OpenAPI spec: operationTypes.json has them
// for each operation, and types.json declares the types they name. Bump specCommit
// and run `node scripts/discord-api.mjs`, then `pnpm test -u` to update the
// copies the service and the docs embed.

import { writeFile } from "node:fs/promises";

const specCommit = "bb8eb1ec745a3dfc1e06c40b6e641c9738393df6";
const specUrl = `https://raw.githubusercontent.com/discord/discord-api-spec/${specCommit}/specs/openapi.json`;
const outFile = new URL(
  "../src/lib/integrations/discord/api.json",
  import.meta.url
);
const operationTypesFile = new URL(
  "../src/lib/integrations/discord/operationTypes.json",
  import.meta.url
);
const typesFile = new URL(
  "../src/lib/integrations/discord/types.json",
  import.meta.url
);

const methods = ["get", "post", "put", "patch", "delete"];

// Endpoints the bot could call but a flow shouldn't.
function isDenied(method, path) {
  // Kite responds to interactions itself, and webhook tokens are credentials
  // of their own that don't need the bot token. Followups and edits of the
  // flow's own interaction are added below.
  if (
    path.startsWith("/interactions") ||
    path.includes("{webhook_token}") ||
    path.includes("{interaction_token}")
  ) {
    return true;
  }
  // Commands, emojis and the app's settings are managed by Kite.
  if (path.startsWith("/applications") && method !== "get") return true;
  // The bot's name and avatar are managed by Kite.
  if (path === "/users/@me" && method === "patch") return true;
  // Server profiles are a premium feature, set with the Edit server profile
  // block or in the app's servers.
  return path === "/guilds/{guild_id}/members/@me" && method === "patch";
}

function paramType(schema) {
  if (schema.$ref?.endsWith("/SnowflakeType")) return "snowflake";
  if (schema.$ref) return paramType(resolveSchema(schema));
  const type = [schema.type].flat().find((t) => t && t !== "null");
  if (type === "array") return "array";
  if (["integer", "number", "boolean"].includes(type)) return type;
  if (type) return "string";
  // Nullable properties are a union with null.
  const variant = (schema.oneOf ?? schema.anyOf)?.find(
    (s) => s.type !== "null"
  );
  return variant ? paramType(variant) : "string";
}

const spec = await (await fetch(specUrl)).json();

function resolve(obj) {
  if (!obj.$ref) return obj;
  const name = obj.$ref.split("/").pop();
  return spec.components.parameters[name];
}

function resolveSchema(schema) {
  while (schema?.$ref) {
    schema = spec.components.schemas[schema.$ref.split("/").pop()];
  }
  return schema;
}

// The top-level properties of a JSON body. A union of objects, like the
// invites of servers and group DMs, has those of all of them, required if
// every one requires them. Bodies that are lists have none, and their fields
// aren't checked.
function bodyParams(content) {
  const schema = resolveSchema(content?.["application/json"]?.schema);
  const variants = (schema?.oneOf ?? schema?.anyOf ?? [schema]).map(
    resolveSchema
  );
  if (!variants.every((v) => v?.properties)) return undefined;

  const params = new Map();
  for (const variant of variants) {
    for (const [name, s] of Object.entries(variant.properties)) {
      if (!params.has(name)) params.set(name, paramType(s));
    }
  }
  return [...params].map(([name, type]) => ({
    name,
    type,
    required: variants.every((v) => v.required?.includes(name)),
  }));
}

function operation(id, method, path, content, params) {
  const paramsIn = (location) =>
    params
      .filter((p) => p.in === location)
      .map((p) => ({
        name: p.name,
        type: paramType(p.schema),
        required: !!p.required,
      }));

  return {
    id,
    method: method.toUpperCase(),
    path,
    // Reads are free like the get blocks, other calls cost 1 like most
    // actions.
    credits: method === "get" ? 0 : 1,
    path_params: paramsIn("path"),
    query_params: paramsIn("query"),
    has_body: !!content,
    body_params: bodyParams(content),
  };
}

// TypeScript types of the spec's schemas. Schemas of objects, unions and
// enums become named types in the discord namespace of kite.d.ts, which
// types.json has by name. Plain ones, like SnowflakeType, are inlined.
const namedTypes = new Map();

function isNamed(schema) {
  return !!(
    schema.properties ||
    schema.oneOf ||
    schema.anyOf ||
    schema.allOf ||
    schema.enum?.length
  );
}

function tsType(schema) {
  if (!schema) return "any";
  if (schema.$ref) {
    const name = schema.$ref.split("/").pop();
    const target = spec.components.schemas[name];
    if (!isNamed(target)) return tsType(target);
    if (!namedTypes.has(name)) {
      namedTypes.set(name, undefined);
      namedTypes.set(name, declaration(name, target));
    }
    return `discord.${name}`;
  }
  if (schema.const !== undefined) return JSON.stringify(schema.const);
  // Some enums of the spec are empty, with the type in allOf.
  if (schema.enum?.length) {
    return union(schema.enum.map((v) => JSON.stringify(v)));
  }
  const variants = schema.oneOf ?? schema.anyOf;
  if (variants) return union(variants.map(tsType));
  if (schema.allOf) return schema.allOf.map((s) => wrap(tsType(s))).join(" & ");
  const types = [schema.type].flat().filter(Boolean);
  if (!types.length) return "any";
  return union(types.map((t) => singleType(t, schema)));
}

function singleType(type, schema) {
  switch (type) {
    case "string":
      return "string";
    case "integer":
    case "number":
      return "number";
    case "boolean":
      return "boolean";
    case "null":
      return "null";
    case "array":
      return `${wrap(tsType(schema.items))}[]`;
    case "object":
      if (schema.properties) {
        return `{ ${properties(schema)
          .map((p) => p.field)
          .join("; ")} }`;
      }
      if (schema.additionalProperties && schema.additionalProperties !== true) {
        return `Record<string, ${tsType(schema.additionalProperties)}>`;
      }
      return "Record<string, any>";
  }
  return "any";
}

function properties(schema) {
  return Object.entries(schema.properties).map(([name, p]) => {
    const key = /^[A-Za-z_$][\w$]*$/.test(name) ? name : JSON.stringify(name);
    const optional = schema.required?.includes(name) ? "" : "?";
    return {
      field: `${key}${optional}: ${tsType(p)}`,
      description: p.description?.replaceAll("*/", "* /").replace(/\s+/g, " "),
    };
  });
}

function declaration(name, schema) {
  if (schema.properties && [schema.type].flat().includes("object")) {
    const fields = properties(schema).map(({ field, description }) =>
      description ? `  /** ${description} */\n  ${field};` : `  ${field};`
    );
    return `interface ${name} {\n${fields.join("\n")}\n}`;
  }
  return `type ${name} = ${tsType({ ...schema, title: undefined })};`;
}

function union(types) {
  const flat = [...new Set(types.flatMap((t) => splitUnion(t)))];
  if (!flat.length || flat.includes("any")) return "any";
  // null last, like string | null.
  return [
    ...flat.filter((t) => t !== "null"),
    ...flat.filter((t) => t === "null"),
  ].join(" | ");
}

// The members of a union at its top level.
function splitUnion(type) {
  const parts = [];
  let depth = 0;
  let start = 0;
  for (let i = 0; i < type.length; i++) {
    const c = type[i];
    if ("({[<".includes(c)) depth++;
    else if (")}]>".includes(c)) depth--;
    else if (depth === 0 && type.startsWith(" | ", i)) {
      parts.push(type.slice(start, i));
      start = i + 3;
    }
  }
  parts.push(type.slice(start));
  return parts;
}

function wrap(type) {
  return splitUnion(type).length > 1 || type.includes(" & ")
    ? `(${type})`
    : type;
}

// The types of an operation's JSON body and of what it returns. An empty
// response, like 204, is null for code.
function operationTypes(op) {
  const json = op.requestBody?.content?.["application/json"]?.schema;
  const results = Object.entries(op.responses ?? {})
    .filter(([status]) => status.startsWith("2"))
    .map(([, r]) => {
      const schema = r.content?.["application/json"]?.schema;
      if (!schema) return "null";
      // The spec makes some results nullable, like lists, but a response
      // with a body always has one.
      const types = splitUnion(tsType(schema)).filter((t) => t !== "null");
      return types.length ? union(types) : "null";
    });
  return {
    ...(json ? { body: tsType(json) } : {}),
    result: results.length ? union(results) : "any",
  };
}

// Endpoints only flows can call, not MCP clients: ones where one wrong call
// does a lot of damage that can't be undone, and ones that return webhook
// tokens, which would let a token that can only read send messages.
const flowOnlyOperations = new Set([
  "bulk_ban_users_from_guild",
  "bulk_delete_messages",
  "get_guild_webhooks",
  "get_webhook",
  "leave_guild",
  "list_channel_webhooks",
  "prune_guild",
]);

// Followups and edits of the interaction the flow runs with, which the spec
// only has as webhook endpoints. The service fills in the app's ID and the
// interaction's token, so a flow never sees the token, and their query
// parameters are only for other webhooks.
const interactionOperations = {
  execute_webhook: "create_followup_message",
  update_original_webhook_message: "update_original_interaction_response",
  update_webhook_message: "update_followup_message",
};

const operations = [];
const types = {};
for (const [path, item] of Object.entries(spec.paths)) {
  const pathParams = (item.parameters ?? []).map(resolve);

  for (const method of methods) {
    const op = item[method];
    if (!op) continue;

    const interactionId = interactionOperations[op.operationId];
    if (interactionId) {
      types[interactionId] = operationTypes(op);
      operations.push({
        ...operation(
          interactionId,
          method,
          path
            .replace("{webhook_id}", "{application_id}")
            .replace("{webhook_token}", "{interaction_token}"),
          op.requestBody?.content,
          pathParams.filter((p) => !p.name.startsWith("webhook_"))
        ),
        interaction: true,
      });
      continue;
    }

    if (!op.security?.some((s) => "BotToken" in s)) continue;
    if (isDenied(method, path)) continue;

    // Multipart bodies are for file uploads, which the block doesn't support.
    const content = op.requestBody?.content;
    if (content && !content["application/json"]) continue;

    const params = [...pathParams, ...(op.parameters ?? []).map(resolve)];
    types[op.operationId] = operationTypes(op);
    operations.push({
      ...operation(op.operationId, method, path, content, params),
      ...(flowOnlyOperations.has(op.operationId) ? { flow_only: true } : {}),
    });
  }
}

operations.sort((a, b) => a.id.localeCompare(b.id));

await writeFile(
  outFile,
  JSON.stringify(
    {
      source: `https://github.com/discord/discord-api-spec/blob/${specCommit}/specs/openapi.json`,
      operations,
    },
    null,
    2
  ) + "\n"
);
console.log(`Wrote ${operations.length} operations`);

// Apart, so the editor only loads the small one for its completions.
const sorted = (entries) =>
  Object.fromEntries([...entries].sort(([a], [b]) => a.localeCompare(b)));
await writeFile(
  operationTypesFile,
  JSON.stringify(sorted(Object.entries(types)), null, 2) + "\n"
);
await writeFile(typesFile, JSON.stringify(sorted(namedTypes), null, 2) + "\n");
console.log(`Wrote the types of ${namedTypes.size} schemas`);
