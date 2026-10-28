import env from "../env/server";

const mcpPath = "/api/mcp";

// The MCP server's URL, which OAuth clients get tokens for. The API only
// issues tokens for it if it's the API's app.public_base_url with /api/mcp.
const mcpResource = `${env.NEXT_PUBLIC_SITE_URL}${mcpPath}`;

// Where clients find protectedResourceMetadata, with the path of the
// resource appended as RFC 9728 says.
export const protectedResourceMetadataUrl = `${env.NEXT_PUBLIC_SITE_URL}/.well-known/oauth-protected-resource${mcpPath}`;

// protectedResourceMetadata tells OAuth clients like Claude to get tokens
// from the API, see RFC 9728.
export const protectedResourceMetadata = {
  resource: mcpResource,
  authorization_servers: [env.NEXT_PUBLIC_API_PUBLIC_BASE_URL],
  // The Discord API is only for API tokens, like in the API's metadata.
  scopes_supported: ["kite:read", "kite:write"],
  bearer_methods_supported: ["header"],
  resource_name: "Kite",
  resource_documentation: `${env.NEXT_PUBLIC_DOCS_LINK}/guides/mcp`,
};
