import { DependencyList, useCallback, useEffect, useRef } from "react";
import { useRouter } from "next/router";

export const UNSAVED_CHANGES_WARNING =
  "You have unsaved changes. Leave anyway?";

let bypassNextNavigation = false;

export function bypassUnsavedChangesWarning(action: () => void) {
  bypassNextNavigation = true;
  try {
    action();
  } finally {
    setTimeout(() => {
      bypassNextNavigation = false;
    }, 100);
  }
}

export function useBeforePageExit(
  callback: (e: BeforeUnloadEvent) => any,
  deps: DependencyList
) {
  const memoCallback = useCallback(callback, [callback, ...deps]);

  useEffect(() => {
    window.addEventListener("beforeunload", memoCallback);
    return () => {
      window.removeEventListener("beforeunload", memoCallback);
    };
  }, [memoCallback]);
}

export function useUnsavedChangesWarning(hasUnsavedChanges: boolean) {
  const router = useRouter();
  const isNavigatingRef = useRef(false);

  useBeforePageExit(
    (e) => {
      if (hasUnsavedChanges) {
        e.preventDefault();
        return UNSAVED_CHANGES_WARNING;
      }
    },
    [hasUnsavedChanges]
  );

  useEffect(() => {
    const handleRouteChangeStart = (url: string) => {
      if (bypassNextNavigation) return;
      if (!hasUnsavedChanges || isNavigatingRef.current) return;
      if (url === router.asPath) return;

      const ok = window.confirm(UNSAVED_CHANGES_WARNING);
      if (!ok) {
        router.events.emit("routeChangeError");
        throw "Abort route change due to unsaved changes";
      } else {
        isNavigatingRef.current = true;
      }
    };

    router.events.on("routeChangeStart", handleRouteChangeStart);
    return () => {
      router.events.off("routeChangeStart", handleRouteChangeStart);
    };
  }, [hasUnsavedChanges, router]);
}
