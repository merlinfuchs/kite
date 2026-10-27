import { createHash } from "crypto";
import { apiRequest } from "../api/client";
import { APIResponse } from "../api/response";
import env from "../env/server";
import { APITokenCurrentGetResponse } from "../types/wire.gen";

// KiteAPI calls the Kite API as the owner of an API token. The MCP server has
// no access of its own, so the API checks every call like any other client's.
export type KiteAPI = <T>(
  path: string,
  init?: RequestInit
) => Promise<APIResponse<T>>;

// clientIp is the IP of the MCP client. nginx trusts CF-Connecting-IP from the
// same machine, so it limits each client by its own IP instead of putting all
// of them in the bucket of this server.
export function kiteAPI(token: string, clientIp?: string): KiteAPI {
  return <T>(path: string, init?: RequestInit) =>
    apiRequest<T>(
      path,
      {
        ...init,
        signal: AbortSignal.timeout(15_000),
        headers: {
          ...init?.headers,
          Authorization: `Bearer ${token}`,
          ...(clientIp ? { "CF-Connecting-IP": clientIp } : {}),
        },
      },
      env.API_INTERNAL_BASE_URL
    );
}

const scopesTTL = 60_000;
const maxCachedScopes = 1000;

// The scopes of tokens by their hash. Every MCP request needs them to list
// its tools, and this saves an API call for most. Scopes of a token never
// change, and the API checks the token on every call anyway, so a deleted
// token only fails on its next tool call.
const cachedScopes = new Map<string, { scopes: string[]; expires: number }>();

// tokenScopes returns the scopes of the token, from the API if fresh is set
// or they aren't cached.
export async function tokenScopes(
  api: KiteAPI,
  token: string,
  fresh: boolean
): Promise<APIResponse<string[]>> {
  const key = createHash("sha256").update(token).digest("base64");
  const cached = cachedScopes.get(key);
  if (!fresh && cached && cached.expires > Date.now()) {
    return { success: true, data: cached.scopes };
  }

  const res = await api<APITokenCurrentGetResponse>(
    "/v1/users/@me/api-tokens/current"
  );
  if (!res.success) {
    cachedScopes.delete(key);
    return res;
  }

  if (cachedScopes.size >= maxCachedScopes) cachedScopes.clear();
  cachedScopes.set(key, {
    scopes: res.data.scopes,
    expires: Date.now() + scopesTTL,
  });
  return { success: true, data: res.data.scopes };
}
