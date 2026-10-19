import { RefObject, useEffect } from "react";

/**
 * Reveals elements inside the root when they scroll into view. Mark an element
 * with data-reveal, or a container with data-reveal-stagger to reveal each of
 * its children one after another. The styles are in globals.css and only
 * apply inside an element with data-reveal-root.
 *
 * Uses data attributes instead of classes, so React doesn't remove them when
 * it re-renders an element's className.
 */
export function useScrollReveal(
  rootRef: RefObject<HTMLElement>,
  enabled = true
) {
  useEffect(() => {
    const root = rootRef.current;
    if (!root || !enabled) return;

    const revealAll = () => {
      root
        .querySelectorAll("[data-reveal], [data-reveal-stagger] > *")
        .forEach((el) => el.setAttribute("data-revealed", ""));
    };

    if (
      typeof IntersectionObserver === "undefined" ||
      window.matchMedia("(prefers-reduced-motion: reduce)").matches
    ) {
      revealAll();
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            entry.target.setAttribute("data-revealed", "");
            observer.unobserve(entry.target);
          }
        }
      },
      { rootMargin: "0px 0px -8% 0px", threshold: 0.1 }
    );

    // Tracked per effect run, so a re-run (like React's strict mode) observes
    // everything again with its new observer.
    const watched = new WeakSet<Element>();
    const watch = (el: Element, index?: number) => {
      if (watched.has(el) || el.hasAttribute("data-revealed")) return;
      watched.add(el);
      if (index !== undefined && el instanceof HTMLElement) {
        el.style.setProperty("--reveal-i", String(index));
      }
      observer.observe(el);
    };

    const scan = () => {
      root.querySelectorAll("[data-reveal]").forEach((el) => watch(el));
      root.querySelectorAll("[data-reveal-stagger]").forEach((parent) => {
        Array.from(parent.children).forEach((child, i) => watch(child, i));
      });
    };
    scan();

    // Sections like pricing render their content after loading data.
    let frame = 0;
    const mutations = new MutationObserver(() => {
      cancelAnimationFrame(frame);
      frame = requestAnimationFrame(scan);
    });
    mutations.observe(root, { childList: true, subtree: true });

    return () => {
      cancelAnimationFrame(frame);
      mutations.disconnect();
      observer.disconnect();
    };
  }, [rootRef, enabled]);
}
