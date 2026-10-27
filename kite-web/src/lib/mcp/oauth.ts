import env from "../env/server";

// The MCP server's URL, which OAuth clients get tokens for. The API only
// issues tokens for it if it's the API's app.public_base_url with /api/mcp.
export const mcpResource = `${env.NEXT_PUBLIC_SITE_URL}/api/mcp`;

// Where clients find protectedResourceMetadata, with the path of the
// resource appended as RFC 9728 says.
export const protectedResourceMetadataUrl = `${env.NEXT_PUBLIC_SITE_URL}/.well-known/oauth-protected-resource/api/mcp`;

// protectedResourceMetadata tells OAuth clients like Claude to get tokens
// from the API, see RFC 9728.
export const protectedResourceMetadata = {
  resource: mcpResource,
  authorization_servers: [env.NEXT_PUBLIC_API_PUBLIC_BASE_URL],
  scopes_supported: [
    "kite:read",
    "kite:write",
    "discord:read",
    "discord:write",
  ],
  bearer_methods_supported: ["header"],
  resource_name: "Kite",
  resource_documentation: `${env.NEXT_PUBLIC_DOCS_LINK}/guides/mcp`,
};
