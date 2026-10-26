import { useCallback, useEffect, useRef } from "react";
import { useRouter } from "next/router";

const UNSAVED_CHANGES_WARNING = "You have unsaved changes. Leave anyway?";

// Warns before leaving with unsaved changes. Returns a function to navigate
// without the warning, for when the user already confirmed leaving.
export function useUnsavedChangesWarning(hasUnsavedChanges: boolean) {
  const router = useRouter();
  const bypassRef = useRef(false);

  useEffect(() => {
    if (!hasUnsavedChanges) return;

    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      e.preventDefault();
      return UNSAVED_CHANGES_WARNING;
    };
    window.addEventListener("beforeunload", handleBeforeUnload);
    return () => {
      window.removeEventListener("beforeunload", handleBeforeUnload);
    };
  }, [hasUnsavedChanges]);

  useEffect(() => {
    const handleRouteChangeStart = (url: string) => {
      if (!hasUnsavedChanges || bypassRef.current) return;
      if (url === router.asPath) return;

      if (!window.confirm(UNSAVED_CHANGES_WARNING)) {
        router.events.emit("routeChangeError");
        throw "Abort route change due to unsaved changes";
      }
      bypassRef.current = true;
    };
    const handleRouteChangeError = () => {
      bypassRef.current = false;
    };

    // The browser has already changed the URL when back or forward is pressed,
    // so on cancel we push the entry of this page again.
    const state = window.history.state;
    router.beforePopState(({ as }) => {
      if (!hasUnsavedChanges || bypassRef.current || as === router.asPath) {
        return true;
      }
      if (window.confirm(UNSAVED_CHANGES_WARNING)) {
        bypassRef.current = true;
        return true;
      }
      window.history.pushState(state, "", router.asPath);
      return false;
    });

    router.events.on("routeChangeStart", handleRouteChangeStart);
    router.events.on("routeChangeError", handleRouteChangeError);
    return () => {
      router.beforePopState(() => true);
      router.events.off("routeChangeStart", handleRouteChangeStart);
      router.events.off("routeChangeError", handleRouteChangeError);
    };
  }, [hasUnsavedChanges, router]);

  return useCallback(
    (url: Parameters<typeof router.push>[0]) => {
      bypassRef.current = true;
      router.push(url);
    },
    [router]
  );
}
