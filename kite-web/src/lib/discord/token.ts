// The other values on the Developer Portal that people paste instead of the
// bot token.
export type WrongTokenKind = "application_id" | "public_key" | "client_secret";

export const wrongTokenMessages: Record<WrongTokenKind, string> = {
  application_id: "This is the application ID, not the bot token.",
  public_key: "This is the public key, not the bot token.",
  // Other values can have the same format, but none that people paste here.
  client_secret: "This looks like the client secret, not the bot token.",
};

// Recognizes the values that aren't the bot token. Anything else is left for
// Discord to check.
export function checkTokenFormat(token: string): WrongTokenKind | null {
  if (/^\d{17,20}$/.test(token)) return "application_id";
  if (/^[a-f0-9]{64}$/.test(token)) return "public_key";
  if (/^[\w-]{32}$/.test(token)) return "client_secret";
  return null;
}
