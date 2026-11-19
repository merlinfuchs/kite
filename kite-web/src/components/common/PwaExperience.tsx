import {
  createContext,
  ReactNode,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";
import { useRouter } from "next/router";
import { Button } from "@/components/ui/button";
import {
  DownloadIcon,
  LoaderCircleIcon,
  RefreshCwIcon,
  WifiOffIcon,
} from "lucide-react";
import { toast } from "sonner";

type InstallPromptEvent = Event & {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed"; platform: string }>;
};

type PwaContextValue = {
  canInstall: boolean;
  isInstalled: boolean;
  install: () => Promise<void>;
  pageTransition: PwaPageTransition | null;
  markPageTransitionReady: (path: string) => void;
  finishPageTransition: (id: number) => void;
  navigateWithPwaTransition: (path: string, navigate: () => void) => void;
  onPageTransitionEntered: (id: number) => void;
};

export type PwaPageTransition = {
  id: number;
  path: string;
  routeComplete: boolean;
  pageReady: boolean;
};

const PwaContext = createContext<PwaContextValue | null>(null);

function routePath(route: string) {
  return route.split(/[?#]/, 1)[0].replace(/\/+$/, "") || "/";
}

export function usePwaInstall() {
  const context = useContext(PwaContext);
  if (!context)
    throw new Error("usePwaInstall must be used within PwaProvider");
  return context;
}

export function isAllowedPwaRoute(route: string) {
  const pathname = route.split(/[?#]/, 1)[0].replace(/\/+$/, "") || "/";
  return ["/login", "/apps", "/marketplace", "/bundles"].some(
    (base) => pathname === base || pathname.startsWith(`${base}/`)
  );
}

export function isOpeningBotRoute(route: string) {
  const pathname = route.split(/[?#]/, 1)[0].replace(/\/+$/, "") || "/";
  return /^\/apps\/[^/]+$/.test(pathname);
}

export function isFlowEditorRoute(pathname: string) {
  const route = routePath(pathname);
  return (
    route === "/apps/[appId]/commands/[cmdId]" ||
    route === "/apps/[appId]/events/[eventId]" ||
    /^\/apps\/[^/]+\/commands\/[^/]+$/.test(route) ||
    /^\/apps\/[^/]+\/events\/[^/]+$/.test(route)
  );
}

export function isPwaMarketplaceRoute(route: string) {
  const pathname = routePath(route);
  return (
    pathname === "/marketplace" ||
    pathname.startsWith("/marketplace/") ||
    /^\/apps\/[^/]+\/marketplace(?:\/|$)/.test(pathname)
  );
}

export function isPwaDashboardRoute(route: string) {
  const pathname = routePath(route);
  return (
    pathname === "/apps" ||
    isOpeningBotRoute(pathname) ||
    /^\/apps\/[^/]+\/(?:commands|events|schedules|webhooks)$/.test(pathname)
  );
}

export function isPwaTransitionRoute(route: string) {
  const pathname = routePath(route);
  return (
    pathname === "/apps" ||
    isOpeningBotRoute(pathname) ||
    isFlowEditorRoute(pathname)
  );
}

export function PwaProvider({ children }: { children: ReactNode }) {
  const [installPrompt, setInstallPrompt] = useState<InstallPromptEvent | null>(
    null
  );
  const [isInstalled, setIsInstalled] = useState(false);
  const [isPwaModeKnown, setIsPwaModeKnown] = useState(false);
  const [isRedirectingToApp, setIsRedirectingToApp] = useState(false);
  const [redirectError, setRedirectError] = useState(false);
  const [updateWorker, setUpdateWorker] = useState<ServiceWorker | null>(null);
  const [isApplyingUpdate, setIsApplyingUpdate] = useState(false);
  const [isOnline, setIsOnline] = useState(true);
  const [isNavigating, setIsNavigating] = useState(false);
  const [isNavigatingToFullscreen, setIsNavigatingToFullscreen] =
    useState(false);
  const [pageTransition, setPageTransition] =
    useState<PwaPageTransition | null>(null);
  const pageTransitionId = useRef(0);
  const pendingNavigation = useRef<{
    id: number;
    navigate: () => void;
  } | null>(null);
  const reloadOnControllerChange = useRef(false);
  const router = useRouter();

  useEffect(() => {
    const standalone =
      window.matchMedia("(display-mode: standalone)").matches ||
      ("standalone" in navigator &&
        (navigator as Navigator & { standalone?: boolean }).standalone ===
          true);
    setIsInstalled(standalone);
    setIsPwaModeKnown(true);

    const handleInstallPrompt = (event: Event) => {
      event.preventDefault();
      setInstallPrompt(event as InstallPromptEvent);
    };
    const handleInstalled = () => {
      setInstallPrompt(null);
      setIsInstalled(true);
      setIsPwaModeKnown(true);
    };
    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);

    setIsOnline(navigator.onLine);
    window.addEventListener("beforeinstallprompt", handleInstallPrompt);
    window.addEventListener("appinstalled", handleInstalled);
    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);

    return () => {
      window.removeEventListener("beforeinstallprompt", handleInstallPrompt);
      window.removeEventListener("appinstalled", handleInstalled);
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }, []);

  const redirectToApp = useCallback(() => {
    setIsRedirectingToApp(true);
    setRedirectError(false);
    void router
      .replace("/apps")
      .then((succeeded) => {
        if (!succeeded) throw new Error("Navigation to /apps was cancelled");
      })
      .catch((error) => {
        console.error("Could not open Kite in the installed app:", error);
        document.getElementById("pwa-startup")?.remove();
        setIsRedirectingToApp(false);
        setRedirectError(true);
      });
  }, [router]);

  const markPageTransitionReady = useCallback((path: string) => {
    const targetPath = routePath(path);
    setPageTransition((current) =>
      current?.routeComplete &&
      !current.pageReady &&
      current.path === targetPath
        ? { ...current, pageReady: true }
        : current
    );
  }, []);
  const finishPageTransition = useCallback((id: number) => {
    setPageTransition((current) => (current?.id === id ? null : current));
  }, []);
  const navigateWithPwaTransition = useCallback(
    (path: string, navigate: () => void) => {
      if (
        !isInstalled ||
        !isFlowEditorRoute(router.asPath) ||
        !isPwaDashboardRoute(path)
      ) {
        navigate();
        return;
      }

      const id = ++pageTransitionId.current;
      pendingNavigation.current = { id, navigate };
      setPageTransition({
        id,
        path: routePath(path),
        routeComplete: false,
        pageReady: false,
      });
    },
    [isInstalled, router.asPath]
  );
  const onPageTransitionEntered = useCallback((id: number) => {
    const pending = pendingNavigation.current;
    if (pending?.id !== id) return;
    pendingNavigation.current = null;
    pending.navigate();
  }, []);

  useEffect(() => {
    if (!isPwaModeKnown || !router.isReady || !isInstalled) {
      if (isPwaModeKnown && !isInstalled) setIsRedirectingToApp(false);
      return;
    }

    if (isAllowedPwaRoute(router.asPath)) {
      setIsRedirectingToApp(false);
      setRedirectError(false);
      return;
    }

    redirectToApp();
  }, [
    isInstalled,
    isPwaModeKnown,
    redirectToApp,
    router.asPath,
    router.isReady,
  ]);

  useEffect(() => {
    const splash = document.getElementById("pwa-startup");
    if (
      !splash ||
      !isPwaModeKnown ||
      !router.isReady ||
      (isInstalled && !isAllowedPwaRoute(router.asPath))
    ) {
      return;
    }

    requestAnimationFrame(() => splash.classList.add("pwa-startup-hidden"));
    const timeout = window.setTimeout(() => splash.remove(), 220);
    return () => window.clearTimeout(timeout);
  }, [isInstalled, isPwaModeKnown, router.asPath, router.isReady]);

  useEffect(() => {
    const handleRouteStart = (url: string) => {
      setIsNavigating(true);
      setIsNavigatingToFullscreen(
        isOpeningBotRoute(url) || isFlowEditorRoute(url)
      );
      const isMarketplaceNavigation =
        isPwaMarketplaceRoute(url) || isPwaMarketplaceRoute(router.asPath);
      const shouldAnimate =
        !isMarketplaceNavigation &&
        (isPwaTransitionRoute(url) ||
          (isFlowEditorRoute(router.asPath) && isPwaDashboardRoute(url)));
      if (isInstalled && shouldAnimate) {
        const path = routePath(url);
        setPageTransition((current) =>
          current?.path === path && !current.routeComplete
            ? current
            : {
                id: ++pageTransitionId.current,
                path,
                routeComplete: false,
                pageReady: false,
              }
        );
      }
      if (isInstalled && !isAllowedPwaRoute(url)) {
        setIsRedirectingToApp(true);
        setRedirectError(false);
      }
    };
    const handleRouteEnd = (url: string) => {
      setIsNavigating(false);
      setIsNavigatingToFullscreen(false);
      const path = routePath(url);
      setPageTransition((current) =>
        current?.path === path ? { ...current, routeComplete: true } : current
      );
    };
    const handleRouteError = (_error?: Error, url?: string) => {
      setIsNavigating(false);
      setIsNavigatingToFullscreen(false);
      setPageTransition((current) => {
        if (!current || (url && current.path !== routePath(url)))
          return current;
        return { ...current, routeComplete: true, pageReady: true };
      });
      if (isInstalled && isAllowedPwaRoute(router.asPath)) {
        setIsRedirectingToApp(false);
      }
    };

    router.events.on("routeChangeStart", handleRouteStart);
    router.events.on("routeChangeComplete", handleRouteEnd);
    router.events.on("routeChangeError", handleRouteError);
    return () => {
      router.events.off("routeChangeStart", handleRouteStart);
      router.events.off("routeChangeComplete", handleRouteEnd);
      router.events.off("routeChangeError", handleRouteError);
    };
  }, [isInstalled, router.asPath, router.events]);

  useEffect(() => {
    if (
      process.env.NODE_ENV !== "production" ||
      !("serviceWorker" in navigator)
    ) {
      return;
    }

    let registration: ServiceWorkerRegistration | undefined;
    let updateInterval: number | undefined;
    const showWaitingWorker = () => {
      if (registration?.waiting && navigator.serviceWorker.controller) {
        setUpdateWorker(registration.waiting);
      }
    };
    const handleControllerChange = () => {
      if (reloadOnControllerChange.current) window.location.reload();
    };
    const checkForUpdate = () => {
      void registration?.update().catch((error) => {
        console.error("Could not check for a Kite update:", error);
      });
    };

    navigator.serviceWorker.addEventListener(
      "controllerchange",
      handleControllerChange
    );

    navigator.serviceWorker
      .register("/service-worker.js", {
        scope: "/",
        updateViaCache: "none",
      })
      .then((result) => {
        registration = result;
        showWaitingWorker();
        registration.addEventListener("updatefound", () => {
          const worker = registration?.installing;
          worker?.addEventListener("statechange", showWaitingWorker);
        });
        checkForUpdate();
        window.addEventListener("focus", checkForUpdate);
        updateInterval = window.setInterval(checkForUpdate, 60 * 60 * 1000);
      })
      .catch((error) => {
        console.error("Could not register Kite's service worker:", error);
      });

    return () => {
      window.removeEventListener("focus", checkForUpdate);
      if (updateInterval) window.clearInterval(updateInterval);
      navigator.serviceWorker.removeEventListener(
        "controllerchange",
        handleControllerChange
      );
    };
  }, []);

  async function install() {
    if (!installPrompt) return;
    const prompt = installPrompt;
    setInstallPrompt(null);
    try {
      await prompt.prompt();
      await prompt.userChoice;
    } catch (error) {
      console.error("Could not open Kite's install prompt:", error);
      toast.error(
        "Could not open the install prompt. Follow the steps below to install Kite."
      );
    }
  }

  function applyUpdate() {
    if (!updateWorker) return;
    setIsApplyingUpdate(true);
    reloadOnControllerChange.current = true;
    updateWorker.postMessage("SKIP_WAITING");
  }

  return (
    <PwaContext.Provider
      value={{
        canInstall: !!installPrompt,
        isInstalled,
        install,
        pageTransition,
        markPageTransitionReady,
        finishPageTransition,
        navigateWithPwaTransition,
        onPageTransitionEntered,
      }}
    >
      {children}
      {isNavigating && !isNavigatingToFullscreen && (
        <div
          className="route-progress"
          role="progressbar"
          aria-label="Loading page"
        />
      )}
      {!isOnline && (
        <div
          className="fixed inset-x-0 top-0 z-[100] flex items-center justify-center gap-2 border-b bg-background/95 px-4 py-3 text-sm shadow-sm backdrop-blur"
          role="status"
          aria-live="polite"
        >
          <WifiOffIcon className="size-4 shrink-0 text-muted-foreground" />
          <span>
            You&apos;re offline. Kite will reconnect when your connection is
            restored.
          </span>
        </div>
      )}
      {(isRedirectingToApp || redirectError) && (
        <div
          className="fixed inset-0 z-[115] flex flex-col items-center justify-center gap-4 bg-background text-foreground"
          role={redirectError ? "alert" : "status"}
          aria-live={redirectError ? "assertive" : "polite"}
        >
          <img src="/icons/icon-192x192.png" alt="" className="size-16" />
          {redirectError ? (
            <>
              <p className="text-sm text-muted-foreground">
                Kite couldn&apos;t open the app. Check your connection and try
                again.
              </p>
              <Button size="sm" onClick={redirectToApp}>
                Try again
              </Button>
            </>
          ) : (
            <>
              <LoaderCircleIcon className="size-6 animate-spin text-primary motion-reduce:animate-none" />
              <p className="text-sm text-muted-foreground">Opening Kite…</p>
            </>
          )}
        </div>
      )}
      {updateWorker && !isApplyingUpdate && (
        <div
          className="fixed inset-x-3 bottom-[max(0.75rem,env(safe-area-inset-bottom))] z-[100] mx-auto flex max-w-lg flex-wrap items-center justify-between gap-3 rounded-lg border bg-background p-4 shadow-lg"
          role="status"
          aria-live="polite"
        >
          <div className="flex items-center gap-3">
            <RefreshCwIcon className="size-5 shrink-0 text-primary" />
            <div>
              <p className="font-medium">A Kite update is ready</p>
              <p className="text-sm text-muted-foreground">
                Save any unsaved changes before reloading.
              </p>
            </div>
          </div>
          <Button size="sm" onClick={applyUpdate}>
            Reload to update
          </Button>
        </div>
      )}
      {isApplyingUpdate && (
        <div
          className="fixed inset-0 z-[110] flex flex-col items-center justify-center gap-4 bg-background text-foreground"
          role="status"
          aria-live="assertive"
        >
          <img src="/icons/icon-192x192.png" alt="" className="size-16" />
          <LoaderCircleIcon className="size-6 animate-spin text-primary motion-reduce:animate-none" />
          <p className="text-sm text-muted-foreground">Updating Kite…</p>
        </div>
      )}
    </PwaContext.Provider>
  );
}

export function InstallKiteButton({ className }: { className?: string }) {
  const { canInstall, isInstalled, install } = usePwaInstall();

  if (isInstalled) {
    return (
      <p className={className} role="status">
        Kite is already installed on this device.
      </p>
    );
  }

  if (!canInstall) return null;

  return (
    <Button className={className} onClick={() => void install()}>
      <DownloadIcon className="mr-2 size-4" />
      Install Kite
    </Button>
  );
}

export function usePwaTransitionReady(isReady: boolean) {
  const { isInstalled, pageTransition, markPageTransitionReady } =
    usePwaInstall();
  const router = useRouter();

  useEffect(() => {
    if (
      isInstalled &&
      isReady &&
      pageTransition?.routeComplete &&
      routePath(router.asPath) === pageTransition.path
    ) {
      markPageTransitionReady(router.asPath);
    }
  }, [
    isInstalled,
    isReady,
    markPageTransitionReady,
    pageTransition,
    router.asPath,
  ]);
}
