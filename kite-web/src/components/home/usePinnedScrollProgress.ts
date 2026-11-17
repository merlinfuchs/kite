import { RefObject, useEffect, useRef, useState } from "react";

export type PinState = "before" | "pinned" | "after";

function clamp01(n: number) {
  return Math.min(1, Math.max(0, n));
}

// Tracks scroll progress (0..1) and viewport pinning state across a tall
// desktop runway section while falling back to static layout on mobile or when
// reduced motion is preferred.
export function usePinnedScrollProgress<
  TSection extends HTMLElement = HTMLElement,
  TStage extends HTMLElement = HTMLDivElement
>(): {
  sectionRef: RefObject<TSection | null>;
  stageRef: RefObject<TStage | null>;
  progress: number;
  pinState: PinState;
  isDesktop: boolean;
  reducedMotion: boolean;
  scrollToProgress: (targetProgress: number) => void;
} {
  const sectionRef = useRef<TSection>(null);
  const stageRef = useRef<TStage>(null);
  const [progress, setProgress] = useState(0);
  const [pinState, setPinState] = useState<PinState>("before");
  const [isDesktop, setIsDesktop] = useState(true);
  const [reducedMotion, setReducedMotion] = useState(false);

  useEffect(() => {
    const section = sectionRef.current;
    const stage = stageRef.current;
    if (!section || !stage) return;

    const prefersReduced = !window.matchMedia(
      "(prefers-reduced-motion: no-preference)"
    ).matches;
    setReducedMotion(prefersReduced);
    setIsDesktop(window.innerWidth >= 1024);
    if (prefersReduced) {
      setProgress(1);
      setPinState("before");
      return;
    }

    let rafId = 0;
    const measure = () => {
      rafId = 0;
      const desktop = window.innerWidth >= 1024;
      setIsDesktop(desktop);
      if (!desktop) {
        setPinState("before");
        setProgress(1);
        return;
      }

      const rect = section.getBoundingClientRect();
      const stageH = stage.offsetHeight;
      const pinTop = 56; // Matches sticky navbar height (h-14)
      const travel = Math.max(1, rect.height - stageH);

      if (rect.top > pinTop) {
        setPinState("before");
        setProgress(0);
      } else if (rect.bottom < pinTop + stageH) {
        setPinState("after");
        setProgress(1);
      } else {
        setPinState("pinned");
        setProgress(clamp01((pinTop - rect.top) / travel));
      }
    };

    const schedule = () => {
      if (!rafId) rafId = window.requestAnimationFrame(measure);
    };

    measure();
    window.addEventListener("scroll", schedule, { passive: true });
    window.addEventListener("resize", schedule, { passive: true });
    return () => {
      window.removeEventListener("scroll", schedule);
      window.removeEventListener("resize", schedule);
      if (rafId) window.cancelAnimationFrame(rafId);
    };
  }, []);

  function scrollToProgress(targetProgress: number) {
    const section = sectionRef.current;
    const stage = stageRef.current;
    if (!section || !stage || window.innerWidth < 1024 || reducedMotion) return;

    const rect = section.getBoundingClientRect();
    const travel = Math.max(1, rect.height - stage.offsetHeight);
    const targetTop =
      window.scrollY + rect.top - 56 + clamp01(targetProgress) * travel;
    window.scrollTo({
      top: targetTop,
      behavior: reducedMotion ? "auto" : "smooth",
    });
  }

  return {
    sectionRef,
    stageRef,
    progress,
    pinState,
    isDesktop,
    reducedMotion,
    scrollToProgress,
  };
}
