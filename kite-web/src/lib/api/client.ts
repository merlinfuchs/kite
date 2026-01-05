import { QueryCache, QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { APIResponse } from "./response";
import env from "@/lib/env/client";

const queryClient = new QueryClient({
  queryCache: new QueryCache({
    onError: (err) => {
      toast.error(`Error: ${err}`);
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

export function apiRequest<T>(path: string, options?: RequestInit) {
  return fetch(getApiUrl(path), {
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
