import { decodeSnowflake } from "@/tools/snowflake-decoder/snowflake";

// Finds the ID in what was pasted, like a mention, a link or the ID itself.
export function parseLookupId(value: string) {
  const match = value.match(/\d{17,20}/);
  if (!match || !decodeSnowflake(match[0])) return null;
  return match[0];
}

// Lists the names of the bits that are set, in the order of the list.
export function flagNames(flags: number, names: [number, string][]) {
  return names.filter(([bit]) => (flags & bit) === bit).map(([, n]) => n);
}
