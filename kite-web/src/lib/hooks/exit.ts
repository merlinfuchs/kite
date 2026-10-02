import { DependencyList, useCallback, useEffect, useRef } from "react";
import { useRouter } from "next/router";

export const UNSAVED_CHANGES_WARNING =
  "Looks like you forgot to save and are trying to leave, do you want to do this? If you do not have, all work since last save will be lost.";

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

export function useUnsavedChangesWarning(
  hasUnsavedChanges: boolean,
  onTriggerExit?: (proceed: () => void) => void
) {
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
      if (!hasUnsavedChanges || isNavigatingRef.current) return;
      if (url === router.asPath) return;

      if (onTriggerExit) {
        router.events.emit("routeChangeError");
        onTriggerExit(() => {
          isNavigatingRef.current = true;
          router.push(url);
        });
        throw "Abort route change due to unsaved changes";
      }

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
  }, [hasUnsavedChanges, onTriggerExit, router]);

  useEffect(() => {
    if (!hasUnsavedChanges) return;

    const stateObj = { __kite_guard: true };
    window.history.pushState(stateObj, "", window.location.href);

    const handlePopState = () => {
      if (isNavigatingRef.current) return;

      if (onTriggerExit) {
        window.history.pushState(stateObj, "", window.location.href);
        onTriggerExit(() => {
          isNavigatingRef.current = true;
          window.history.back();
        });
        return;
      }

      const ok = window.confirm(UNSAVED_CHANGES_WARNING);
      if (!ok) {
        window.history.pushState(stateObj, "", window.location.href);
      } else {
        isNavigatingRef.current = true;
        window.history.back();
      }
    };

    window.addEventListener("popstate", handlePopState);
    return () => {
      window.removeEventListener("popstate", handlePopState);
    };
  }, [hasUnsavedChanges, onTriggerExit]);
}
