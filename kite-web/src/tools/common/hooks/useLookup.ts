import { useQuery } from "@tanstack/react-query";
import { useState } from "react";

export type LookupResult<T> =
  | { success: true; data: T }
  | { success: false; error: string };

// useLookup runs the lookup of a tool for what was last submitted. Failures,
// like an unknown ID or a rate limit, are results, so they show under the
// input instead of as a toast.
export function useLookup<T>({
  name,
  fetch,
  unreachable,
  staleTime,
}: {
  name: string;
  fetch: (key: string) => Promise<LookupResult<T>>;
  // Shown when the request didn't get an answer at all.
  unreachable: string;
  staleTime: number;
}) {
  const [key, setKey] = useState<string | null>(null);

  const query = useQuery({
    queryKey: ["lookup", name, key],
    queryFn: async (): Promise<LookupResult<T>> => {
      try {
        return await fetch(key!);
      } catch {
        return { success: false, error: unreachable };
      }
    },
    enabled: !!key,
    staleTime,
  });

  function lookup(next: string) {
    if (next !== key) {
      setKey(next);
    } else if (!query.data?.success) {
      // Submitting the same thing after an error tries again.
      query.refetch();
    }
  }

  return { result: query.data, loading: query.isFetching, lookup };
}
