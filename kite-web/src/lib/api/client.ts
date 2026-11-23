import { QueryCache, QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { APIResponse } from "./response";
import env from "@/lib/env/client";
import serverEnv from "@/lib/env/server";
import { BundleGetResponse } from "../types/wire.gen";

const queryClient = new QueryClient({
  queryCache: new QueryCache({
    onError: (err) => {
      toast.error(errorMessage(err));
    },
  }),
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: (failureCount, err: any) => {
        if (failureCount >= 3) {
          return false;
        }
        return err.status >= 500;
      },
      staleTime: 1000 * 60 * 3,
    },
  },
});

export default queryClient;

export function getApiUrl(path?: string) {
  const baseUrl = env.NEXT_PUBLIC_API_PUBLIC_BASE_URL;
  if (!path) {
    return baseUrl;
  }

  return baseUrl + path;
}

// loginUrlTo logs in and comes back to the path.
export function loginUrlTo(path: string) {
  return getApiUrl(`/v1/auth/login?redirect=${encodeURIComponent(path)}`);
}

export const loginUrl = loginUrlTo("/apps");

// installUrl logs in and adds the Kite bot to a server on one page of
// Discord. The query can pick the server with guild_id, and the rest goes to
// the setup page, like templates=verification,giveaway.
export function installUrl(query?: string) {
  return getApiUrl(`/v1/auth/install${query ? `?${query}` : ""}`);
}

// errorMessage is the text of an error for a toast, without the "Error:" it
// starts with as a string.
export function errorMessage(err: unknown) {
  return err instanceof Error ? err.message : String(err);
}

// React Query holds requests while the browser is offline, so one that fails
// without an answer while offline was cut off by the connection dropping. The
// server has no navigator.onLine.
function fetchOrOffline(...args: Parameters<typeof fetch>) {
  return fetch(...args).catch((err) => {
    if (typeof window !== "undefined" && !navigator.onLine) {
      throw new Error("You're offline. Try again once you're back online.");
    }
    throw err;
  });
}

// requestBaseUrl is where requests go. Pages rendered on the server reach the
// API on the same machine when there is one, browsers use the public URL.
function requestBaseUrl() {
  if (typeof window === "undefined" && serverEnv.API_INTERNAL_BASE_URL) {
    return serverEnv.API_INTERNAL_BASE_URL;
  }
  return getApiUrl();
}

export function apiRequest<T>(
  path: string,
  options?: RequestInit,
  baseUrl = requestBaseUrl()
) {
  return fetchOrOffline(baseUrl + path, {
    ...options,
    credentials: "include",
  }).then(async (res) => {
    try {
      return await res.json();
    } catch (err) {
      if (res.status === 429) {
        throw new Error("Rate limit exceeded, try again later");
      }
      throw new Error("Unexpected API error: " + res.statusText);
    }
  }) as Promise<APIResponse<T>>;
}

// fetchBundle gets a bundle as anyone without a session sees it, for what the
// server renders. It's null for bundles that don't exist or were taken down,
// other errors are thrown.
export async function fetchBundle(id: string) {
  const res = await apiRequest<BundleGetResponse>(
    `/v1/bundles/${encodeURIComponent(id)}`,
    { signal: AbortSignal.timeout(5000) }
  );
  if (res.success) return res.data;
  if (
    res.error.code === "unknown_bundle" ||
    res.error.code === "bundle_removed"
  )
    return null;
  throw new Error(res.error.message);
}
