import { apiRequest } from "../api/client";
import { APIResponse } from "../api/response";

// KiteAPI calls the Kite API as the owner of an API token. The MCP server has
// no access of its own, so the API checks every call like any other client's.
export type KiteAPI = <T>(
  path: string,
  init?: RequestInit
) => Promise<APIResponse<T>>;

export function kiteAPI(token: string): KiteAPI {
  return <T>(path: string, init?: RequestInit) =>
    apiRequest<T>(path, {
      ...init,
      signal: AbortSignal.timeout(15_000),
      headers: { ...init?.headers, Authorization: `Bearer ${token}` },
    });
}
