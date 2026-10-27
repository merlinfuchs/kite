import { cn } from "@/lib/utils";
import { ReactNode, RefObject, useEffect, useRef, useState } from "react";

// useInView is true once the element was scrolled into view.
export function useInView<T extends Element>(
  threshold = 0.2
): [RefObject<T | null>, boolean] {
  const ref = useRef<T>(null);
  const [inView, setInView] = useState(false);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setInView(true);
          observer.disconnect();
        }
      },
      { threshold }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [threshold]);

  return [ref, inView];
}

// Fades its children in and up when they're scrolled into view.
export default function Reveal({
  children,
  className,
  delay = 0,
}: {
  children: ReactNode;
  className?: string;
  delay?: number;
}) {
  const [ref, inView] = useInView<HTMLDivElement>();

  return (
    <div
      ref={ref}
      style={{ transitionDelay: `${delay}ms` }}
      className={cn(
        "transition duration-700 ease-out motion-reduce:transition-none",
        inView
          ? "opacity-100 translate-y-0"
          : "opacity-0 translate-y-6 motion-reduce:opacity-100 motion-reduce:translate-y-0",
        className
      )}
    >
      {children}
    </div>
  );
}
