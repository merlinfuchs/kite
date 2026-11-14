import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/router";
import { useEffect, useRef, useState } from "react";

export type LookupResult<T> =
  | { success: true; data: T }
  | { success: false; error: string };

// useLookup runs the lookup of a tool for what was last submitted. Failures,
// like an unknown ID or a rate limit, are results, so they show under the
// input instead of as a toast.
//
// What was looked up goes into the URL, like ?id=123, so a lookup can be
// shared, and a link with it runs the lookup when the page opens.
export function useLookup<T>({
  name,
  param,
  parse,
  invalid,
  fetch,
  unreachable,
  staleTime,
  ready = true,
}: {
  name: string;
  // The query parameter of the URL, like id.
  param: string;
  // Gets what to look up from the input, null if it isn't valid.
  parse: (value: string) => string | null;
  // Shown when parse returns null.
  invalid: string;
  fetch: (key: string) => Promise<LookupResult<T>>;
  // Shown when the request didn't get an answer at all.
  unreachable: string;
  staleTime: number;
  // Whether a lookup from the URL can run yet, like once the user is known
  // to be logged in.
  ready?: boolean;
}) {
  const router = useRouter();
  const [key, setKey] = useState<string | null>(null);
  const [inputError, setInputError] = useState<string>();
  // What the URL had when the page opened, for the input.
  const [initialValue, setInitialValue] = useState("");
  const ranFromUrl = useRef(false);

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

  function lookup(value: string) {
    const next = parse(value);
    if (!next) {
      setInputError(invalid);
      return;
    }
    setInputError(undefined);

    if (router.query[param] !== next) {
      router.replace(
        { pathname: router.pathname, query: { [param]: next } },
        undefined,
        { shallow: true, scroll: false }
      );
    }

    if (next !== key) {
      setKey(next);
    } else if (!query.data?.success) {
      // Submitting the same thing after an error tries again.
      query.refetch();
    }
  }

  useEffect(() => {
    if (!router.isReady || !ready || ranFromUrl.current) return;
    ranFromUrl.current = true;

    const value = router.query[param];
    if (typeof value !== "string") return;
    // Redirects that copy a path can leave slashes around it.
    const trimmed = value.trim().replace(/^\/+|\/+$/g, "");
    if (!trimmed) return;

    setInitialValue(trimmed);
    lookup(trimmed);
    // lookup only needs to run once, with what the URL had.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [router.isReady, ready]);

  const result = inputError ? undefined : query.data;
  const error =
    inputError ?? (result && !result.success ? result.error : undefined);

  return {
    result,
    error,
    loading: query.isFetching,
    initialValue,
    lookup,
  };
}
