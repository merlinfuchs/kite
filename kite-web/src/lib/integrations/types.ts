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
  // What the integration's blocks send to the service, shown before the app
  // enables it, with the service's policies.
  data_shared?: string;
  privacy_url?: string;
  terms_url?: string;
  // Other hosts of the service, like an old domain, which HTTP blocks are
  // pointed to the integration's blocks for, like those to base_url.
  other_hosts?: string[];
  // Where the owner of the credential authorizes Kite for what the
  // credential alone can't do, like running ER:LC commands. {credential_id}
  // is the public part of the credential, for ER:LC's server keys the part
  // after the last "-", and {app_id} Kite's ID at the service, from the
  // service's config.
  authorize_url?: string;
  // The service reports its rate limits in X-RateLimit headers and
  // Retry-After the way ER:LC does, which the service follows, see
  // design/integrations.md.
  rate_limit_headers?: boolean;
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
  // Header for Kite's own key for the service, from the service's config,
  // sent next to the app's credential.
  kite_key_header?: string;
}

export type CredentialIntegration = Integration & { auth: CredentialAuth };

export function needsCredential(
  integration: Integration
): integration is CredentialIntegration {
  return (
    integration.auth.type === "header" || integration.auth.type === "query"
  );
}
