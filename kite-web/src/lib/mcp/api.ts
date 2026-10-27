import { apiRequest } from "../api/client";
import { APIResponse } from "../api/response";
import env from "../env/server";

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
