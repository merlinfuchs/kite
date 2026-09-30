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
  // "always" for integrations every app can use, like Discord, "default" for
  // ones apps can turn off and "opt_in" for ones they turn on. Integrations
  // that need a credential are opt-in, and turned on by entering it.
  availability: "always" | "default" | "opt_in";
  // A cheap GET endpoint, relative to base_url, that checks the credential
  // when the app enters it.
  test_path?: string;
}

// How requests prove who they are. Integrations with the "header" or "query"
// type need a credential of the app, the others don't.
export type IntegrationAuth =
  | { type: "discord_bot" }
  | { type: "none" }
  | CredentialAuth;

export interface CredentialAuth {
  type: "header" | "query";
  // Name of the header or query parameter.
  name: string;
  // Put before the credential in the header, e.g. "Bearer ".
  prefix?: string;
  // What the credential is called, like "API key", and where to get it.
  label: string;
  help_url?: string;
}

export type CredentialIntegration = Integration & { auth: CredentialAuth };

export function needsCredential(
  integration: Integration
): integration is CredentialIntegration {
  return (
    integration.auth.type === "header" || integration.auth.type === "query"
  );
}
