import logo from "@/assets/logo/orange@1024.png";
import {
  AnimationEvent,
  ReactNode,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import { usePwaInstall, usePwaTransitionReady } from "../common/PwaExperience";

export default function BotLoadingScreen({
  isReady,
  onExited,
  onEntered,
  message = "Opening your bot…",
}: {
  isReady: boolean;
  onExited: () => void;
  onEntered?: () => void;
  message?: string;
}) {
  const [phase, setPhase] = useState<"entering" | "loading" | "exiting">(
    "entering"
  );
  const hasEntered = useRef(false);
  const notifyEntered = useCallback(() => {
    if (hasEntered.current) return;
    hasEntered.current = true;
    onEntered?.();
  }, [onEntered]);

  useEffect(() => {
    if (!isReady && phase === "exiting") {
      setPhase("loading");
    } else if (isReady && phase === "loading") {
      setPhase("exiting");
    }
  }, [isReady, phase]);

  useEffect(() => {
    if (phase === "entering") {
      const timeout = window.setTimeout(() => {
        notifyEntered();
        setPhase(isReady ? "exiting" : "loading");
      }, 1200);
      return () => window.clearTimeout(timeout);
    }

    if (phase === "exiting") {
      const timeout = window.setTimeout(onExited, 1200);
      return () => window.clearTimeout(timeout);
    }
  }, [isReady, notifyEntered, onExited, phase]);

  function handleAnimationEnd(event: AnimationEvent<HTMLDivElement>) {
    if (event.target !== event.currentTarget) return;

    if (phase === "entering") {
      notifyEntered();
      setPhase(isReady ? "exiting" : "loading");
    } else if (phase === "exiting") {
      onExited();
    }
  }

  return (
    <div
      className="bot-loading-screen fixed inset-0 z-[120] flex items-center justify-center overflow-hidden bg-[#141211] text-white"
      data-phase={phase}
      role="status"
      aria-live="polite"
      onAnimationEnd={handleAnimationEnd}
    >
      <div className="relative z-10 flex flex-col items-center gap-5">
        <img src={logo.src} alt="Kite" className="bot-loading-mark size-16" />
        <p className="text-sm text-white/70">{message}</p>
      </div>
      <div className="bot-loading-waves" aria-hidden="true">
        <svg
          className="bot-loading-wave bot-loading-wave-back"
          viewBox="0 0 1440 360"
          preserveAspectRatio="none"
        >
          <path d="M0 210C180 170 270 120 450 160s270 110 450 90 300-100 540-50v160H0Z" />
        </svg>
        <svg
          className="bot-loading-wave bot-loading-wave-front"
          viewBox="0 0 1440 360"
          preserveAspectRatio="none"
        >
          <path d="M0 250c180-70 300-50 480 0s300 80 480 20 300-100 480-70v160H0Z" />
        </svg>
      </div>
    </div>
  );
}

export function PwaRouteTransitionLayer({ children }: { children: ReactNode }) {
  const {
    isInstalled,
    pageTransition,
    finishPageTransition,
    onPageTransitionEntered,
  } = usePwaInstall();
  const transitionId = pageTransition?.id;
  const handleExited = useCallback(() => {
    if (transitionId !== undefined) finishPageTransition(transitionId);
  }, [finishPageTransition, transitionId]);
  const handleEntered = useCallback(() => {
    if (transitionId !== undefined) onPageTransitionEntered(transitionId);
  }, [onPageTransitionEntered, transitionId]);

  const isTransitioning = isInstalled && !!pageTransition;

  return (
    <>
      <div
        className="contents"
        inert={isTransitioning}
        aria-hidden={isTransitioning}
      >
        {children}
      </div>
      {isTransitioning && pageTransition && (
        <BotLoadingScreen
          key={pageTransition.id}
          isReady={pageTransition.routeComplete && pageTransition.pageReady}
          onExited={handleExited}
          onEntered={handleEntered}
          message="Loading Kite…"
        />
      )}
    </>
  );
}

export function BotLoadingGate({
  children,
  enabled,
  isLoading,
}: {
  children: ReactNode;
  enabled: boolean;
  isLoading: boolean;
}) {
  const { pageTransition } = usePwaInstall();
  usePwaTransitionReady(enabled && !isLoading);
  const [hasShown, setHasShown] = useState(enabled && isLoading);
  const showLoadingScreen =
    enabled && (isLoading || hasShown) && !pageTransition;
  const handleExited = useCallback(() => setHasShown(false), []);

  useEffect(() => {
    if (pageTransition) setHasShown(false);
    else if (enabled && isLoading) setHasShown(true);
  }, [enabled, isLoading, pageTransition]);

  if (!enabled) return children;

  return (
    <>
      <div inert={showLoadingScreen} aria-hidden={showLoadingScreen}>
        {children}
      </div>
      {showLoadingScreen && (
        <BotLoadingScreen isReady={!isLoading} onExited={handleExited} />
      )}
    </>
  );
}
