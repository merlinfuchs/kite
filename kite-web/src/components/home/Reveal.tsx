import { cn } from "@/lib/utils";
import { CSSProperties, ReactNode, useEffect, useRef, useState } from "react";

type RevealDirection = "up" | "down" | "left" | "right" | "scale" | "fade";

interface RevealProps {
  children: ReactNode;
  className?: string;
  /** Where the element comes from as it enters the viewport. */
  from?: RevealDirection;
  /** Delay in milliseconds, used to stagger siblings. */
  delay?: number;
  /** Duration in milliseconds. */
  duration?: number;
}

/**
 * Fades and slides its children into place the first time they scroll into
 * view. The motion itself lives in home.css, this only flips `data-visible`.
 */
export default function Reveal({
  children,
  className,
  from = "up",
  delay = 0,
  duration,
}: RevealProps) {
  const ref = useRef<HTMLDivElement>(null);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    if (typeof IntersectionObserver === "undefined") {
      setVisible(true);
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          setVisible(true);
          observer.disconnect();
        }
      },
      // Wait until the element is a little way into the viewport, so the
      // animation isn't already over by the time it is noticed.
      { rootMargin: "0px 0px -8% 0px", threshold: 0.08 }
    );

    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  const style: CSSProperties & Record<string, string> = {};
  if (delay) style["--reveal-delay"] = `${delay}ms`;
  if (duration) style["--reveal-duration"] = `${duration}ms`;

  return (
    <div
      ref={ref}
      className={cn("reveal", className)}
      data-from={from}
      data-visible={visible}
      style={style}
    >
      {children}
    </div>
  );
}
