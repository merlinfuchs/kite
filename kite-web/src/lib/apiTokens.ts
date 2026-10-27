// Scopes of API tokens are "<resource>:<access>", like in the service's
// model/api_token.go. Write access includes read.
export type APITokenResource = "kite" | "discord";
export type APITokenAccess = "read" | "write";

export const apiTokenScopes = [
  "kite:read",
  "kite:write",
  "discord:read",
  "discord:write",
];

export function hasAccess(
  scopes: string[],
  resource: APITokenResource,
  access: APITokenAccess
) {
  return (
    scopes.includes(`${resource}:write`) ||
    (access === "read" && scopes.includes(`${resource}:read`))
  );
}
