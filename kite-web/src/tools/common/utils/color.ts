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
