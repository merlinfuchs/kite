// A service blocks can talk to. Blocks reference integrations, see
// ../blocks and design/integrations.md.
export interface Integration {
  id: string;
  name: string;
  description: string;
  // Where the requests of the integration's blocks go. Discord's go through
  // the bot's session instead.
  base_url?: string;
  auth: IntegrationAuth;
  // A cheap GET endpoint, relative to base_url, that checks the credential
  // when the app connects the integration.
  test_path?: string;
}

// How requests prove who they are. Integrations with the "discord_bot" or
// "none" type are always connected, the others need a credential of the app.
export type IntegrationAuth =
  | { type: "discord_bot" }
  | { type: "none" }
  | {
      type: "header" | "query";
      // Name of the header or query parameter.
      name: string;
      // Put before the credential in the header, e.g. "Bearer ".
      prefix?: string;
      // What the credential is called, like "API key", and where to get it.
      label: string;
      help_url?: string;
    };

export function needsCredential(integration: Integration) {
  return (
    integration.auth.type === "header" || integration.auth.type === "query"
  );
}
