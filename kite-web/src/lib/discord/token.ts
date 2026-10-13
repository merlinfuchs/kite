// Catches the other values on the Developer Portal that people paste instead
// of the bot token. Anything else is left for Discord to check.
export function checkTokenFormat(token: string): string | null {
  if (/^\d{17,20}$/.test(token)) {
    return "This is the application ID. Copy the token under Bot instead.";
  }
  if (/^[a-f0-9]{64}$/.test(token)) {
    return "This is the public key. Copy the token under Bot instead.";
  }
  if (/^[\w-]{32}$/.test(token)) {
    return "This looks like the client secret. Copy the token under Bot instead.";
  }
  return null;
}
