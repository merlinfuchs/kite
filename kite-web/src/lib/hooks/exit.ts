import { DependencyList, useCallback, useEffect } from "react";

export function useBeforePageExit(
  callback: (e: BeforeUnloadEvent) => any,
  deps: DependencyList
) {
  const memoCallback = useCallback(callback, deps);

  useEffect(() => {
    window.addEventListener("beforeunload", memoCallback);
    return () => {
      window.removeEventListener("beforeunload", memoCallback);
    };
  }, [memoCallback]);
}
