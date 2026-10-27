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
  return async <T>(path: string, init?: RequestInit) => {
    const res = await apiRequest<T>(
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
    // The token was deleted or replaced, so the client's next request gets a
    // 401 instead of its cached access, and connects again.
    if (!res.success && res.error.code === "invalid_api_token") {
      cachedAccess.delete(tokenKey(token));
    }
    return res;
  };
}

const accessTTL = 60_000;
const maxCachedTokens = 1000;

// What a token can do, which decides the tools of the MCP server.
export interface TokenAccess {
  scopes: string[];
  // Whether it's the token of an OAuth client, like Claude or ChatGPT.
  oauth: boolean;
}

// The access of tokens by their hash. Every MCP request needs it to list its
// tools, and this saves an API call for most. It never changes, and the API
// checks the token on every call anyway, so a deleted token only fails on its
// next tool call. Tokens of OAuth clients aren't cached past their expiry, so
// the client gets a 401 and refreshes, instead of tool calls that fail.
const cachedAccess = new Map<
  string,
  { access: TokenAccess; expires: number }
>();

// tokenAccess returns what the token can do, from the API if fresh is set or
// it isn't cached.
export async function tokenAccess(
  api: KiteAPI,
  token: string,
  fresh: boolean
): Promise<APIResponse<TokenAccess>> {
  const key = tokenKey(token);
  const cached = cachedAccess.get(key);
  if (!fresh && cached && cached.expires > Date.now()) {
    return { success: true, data: cached.access };
  }

  const res = await api<APITokenCurrentGetResponse>(
    "/v1/users/@me/api-tokens/current"
  );
  if (!res.success) {
    cachedAccess.delete(key);
    return res;
  }

  if (cachedAccess.size >= maxCachedTokens) cachedAccess.clear();
  const accessExpires = res.data.access_expires_at
    ? new Date(res.data.access_expires_at).getTime()
    : Infinity;
  const access = {
    scopes: res.data.scopes,
    oauth: !!res.data.oauth_client_id,
  };
  cachedAccess.set(key, {
    access,
    expires: Math.min(Date.now() + accessTTL, accessExpires),
  });
  return { success: true, data: access };
}

function tokenKey(token: string) {
  return createHash("sha256").update(token).digest("base64");
}
