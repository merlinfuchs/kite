import { ApiOperation } from "../integrations/api";
import api from "../integrations/discord/api.json";

// The Discord API endpoints the Discord API Request block can call, generated
// from Discord's OpenAPI spec by scripts/discord-api.mjs.

export const discordApiOperations: ApiOperation[] = api.operations;

const operationsById = new Map(discordApiOperations.map((o) => [o.id, o]));

export function getDiscordApiOperation(id: string | undefined) {
  return id ? operationsById.get(id) : undefined;
}

// create_message -> Create message
export function discordApiOperationLabel(id: string) {
  const words = id.replaceAll("_", " ");
  return words.charAt(0).toUpperCase() + words.slice(1);
}

// The spec's names differ from the ones in Discord's docs, e.g. Modify Guild is
// update_guild.
const synonyms: Record<string, string> = {
  modify: "update",
  edit: "update",
  fetch: "get",
  remove: "delete",
  add: "create",
};

function words(id: string) {
  const res = id
    .toLowerCase()
    .split(/[^a-z0-9]+/)
    .filter(Boolean);
  // The docs' "Get Channel Messages" is list_messages.
  if (res[0] === "get" && res[res.length - 1]?.endsWith("s")) res[0] = "list";
  return res.map((w) => synonyms[w] ?? w.replace(/s$/, ""));
}

// The endpoints most similar to an unknown one, so a typo or a name from the
// docs can be corrected. The last word is what the endpoint acts on, so it
// counts the most.
export function similarDiscordApiOperations(id: string, count = 5) {
  const target = words(id);
  return discordApiOperations
    .map((o) => {
      const opWords = words(o.id);
      const shared = target.filter((w) => opWords.includes(w)).length;
      const sameObject = target.at(-1) === opWords.at(-1) ? 10 : 0;
      const extra = opWords.length - shared;
      return { id: o.id, score: shared * 10 + sameObject - extra };
    })
    .sort((a, b) => b.score - a.score || a.id.localeCompare(b.id))
    .slice(0, count)
    .map((o) => o.id);
}
