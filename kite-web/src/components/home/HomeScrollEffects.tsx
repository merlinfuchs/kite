import { useEffect } from "react";

/**
 * Keeps `--scroll-y` up to date for the hero parallax, and turns on smooth
 * scrolling for the anchor links while it is mounted (see home.css).
 */
export default function HomeScrollEffects() {
  useEffect(() => {
    const root = document.documentElement;
    let frame = 0;

    function update() {
      frame = 0;
      root.style.setProperty("--scroll-y", `${window.scrollY}`);
    }

    function onScroll() {
      if (!frame) frame = requestAnimationFrame(update);
    }

    update();
    window.addEventListener("scroll", onScroll, { passive: true });

    return () => {
      window.removeEventListener("scroll", onScroll);
      if (frame) cancelAnimationFrame(frame);
      root.style.removeProperty("--scroll-y");
    };
  }, []);

  return <div aria-hidden className="home-smooth-scroll hidden" />;
}
