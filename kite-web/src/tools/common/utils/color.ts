// A number, or a hex color like #ff0000, #f00 or 0xff0000, which the service
// reads with message.ParseColor.
export const colorRegex =
  /^\s*(?:([0-9]+)|(?:#|0[xX])([0-9a-fA-F]{6}|[0-9a-fA-F]{3}))\s*$/;

// A color as a number, if it's one or a hex code.
export function parseColorValue(value: string) {
  const match = colorRegex.exec(value);
  if (!match) return;
  if (match[1]) return Number(match[1]);
  const hex = match[2].length === 3 ? match[2].replace(/./g, "$&$&") : match[2];
  return parseInt(hex, 16);
}

export function colorIntToHex(color: number) {
  return "#" + color.toString(16).padStart(6, "0");
}

export function colorHexToInt(color: string) {
  return parseInt(color.replace("#", ""), 16);
}

// Placeholders are only filled in when the message is sent, so they're shown
// like no color.
export function colorToHex(
  color: number | string | undefined,
  fallback: string
) {
  return typeof color === "number" ? colorIntToHex(color) : fallback;
}
