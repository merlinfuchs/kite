// Generates src/lib/integrations/discord/api.json, the endpoints the Discord
// API Request block can call, from Discord's OpenAPI spec. Bump specCommit and run
// `node scripts/discord-api.mjs`, then `pnpm test -u` to update the copy the
// service embeds.

import { writeFile } from "node:fs/promises";

const specCommit = "bb8eb1ec745a3dfc1e06c40b6e641c9738393df6";
const specUrl = `https://raw.githubusercontent.com/discord/discord-api-spec/${specCommit}/specs/openapi.json`;
const outFile = new URL(
  "../src/lib/integrations/discord/api.json",
  import.meta.url
);

const methods = ["get", "post", "put", "patch", "delete"];

// Endpoints the bot could call but a flow shouldn't.
function isDenied(method, path) {
  // Kite responds to interactions itself, and webhook tokens are credentials
  // of their own that don't need the bot token.
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
  return path === "/users/@me" && method === "patch";
}

function paramType(schema) {
  if (schema.$ref?.endsWith("/SnowflakeType")) return "snowflake";
  if (schema.$ref) return paramType(resolveSchema(schema));
  const type = [schema.type].flat().find((t) => t && t !== "null");
  if (type === "array") return "array";
  if (["integer", "number", "boolean"].includes(type)) return type;
  if (type) return "string";
  // Nullable properties are a union with null.
  const variant = (schema.oneOf ?? schema.anyOf)?.find((s) => s.type !== "null");
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

const operations = [];
for (const [path, item] of Object.entries(spec.paths)) {
  const pathParams = (item.parameters ?? []).map(resolve);

  for (const method of methods) {
    const op = item[method];
    if (!op) continue;
    if (!op.security?.some((s) => "BotToken" in s)) continue;
    if (isDenied(method, path)) continue;

    // Multipart bodies are for file uploads, which the block doesn't support.
    const content = op.requestBody?.content;
    if (content && !content["application/json"]) continue;

    const params = [...pathParams, ...(op.parameters ?? []).map(resolve)];
    const paramsIn = (location) =>
      params
        .filter((p) => p.in === location)
        .map((p) => ({
          name: p.name,
          type: paramType(p.schema),
          required: !!p.required,
        }));

    operations.push({
      id: op.operationId,
      method: method.toUpperCase(),
      path,
      path_params: paramsIn("path"),
      query_params: paramsIn("query"),
      has_body: !!content,
      body_params: bodyParams(content),
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
