import { LookupResult } from "@/tools/common/hooks/useLookup";
import { decodeSnowflake } from "@/tools/snowflake-decoder/snowflake";

const idPatterns = [
  /^(\d{17,20})$/,
  // Mentions like <@123>
  /^<@!?(\d{17,20})>$/,
  // Profile links like discord.com/users/123
  /^(?:https?:\/\/)?(?:[a-z]+\.)?discord(?:app)?\.com\/users\/(\d{17,20})\/?(?:[?#].*)?$/,
  // Invite links of bots, with client_id=123
  /^(?:https?:\/\/)?(?:[a-z]+\.)?discord(?:app)?\.com\/(?:api\/)?oauth2\/authorize\?(?:.*&)?client_id=(\d{17,20})(?:[&#].*)?$/,
];

// Finds the ID in what was pasted: the ID, a mention, or a link to a profile
// or a bot invite. Anything else, like a message link, isn't taken apart.
export function parseLookupId(value: string) {
  for (const pattern of idPatterns) {
    const match = value.match(pattern);
    if (match && decodeSnowflake(match[1])) return match[1];
  }
  return null;
}

// Lists the names of the bits that are set, in the order of the list.
export function flagNames(flags: number, names: [number, string][]) {
  return names.filter(([bit]) => (flags & bit) === bit).map(([, n]) => n);
}

// fetchDiscord calls an endpoint of Discord that needs no token, from the
// browser.
export async function fetchDiscord<T>(
  path: string,
  notFound: string
): Promise<LookupResult<T>> {
  const res = await fetch(`https://discord.com/api/v10${path}`);
  if (res.status === 404) {
    return { success: false, error: notFound };
  }
  if (res.status === 429) {
    return {
      success: false,
      error: "Discord is limiting lookups right now. Try again in a minute.",
    };
  }
  if (!res.ok) {
    return {
      success: false,
      error: "Discord didn't answer. Try again in a moment.",
    };
  }
  return { success: true, data: await res.json() };
}

export const discordUnreachable =
  "Couldn't reach Discord. Check your connection or turn off ad blockers for this page, then try again.";
