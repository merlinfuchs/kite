import { useNodeValues } from "@/lib/flow/nodes";
import { useRouter } from "next/router";
import { ReactNode, useEffect, useRef } from "react";

export default function EmbeddablePage({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  const router = useRouter();
  const containerRef = useRef<HTMLDivElement>(null);

  const { nodeType, theme = "light" } = router.query;

  const values = useNodeValues(nodeType as string);

  useEffect(() => {
    // Remove any fixed width and ensure no horizontal scroll
    document.body.style.width = "auto";
    document.body.style.overflow = "hidden";
    document.documentElement.style.overflow = "hidden";

    setTimeout(() => {
      if (theme === "dark") {
        document.documentElement.classList.add("dark");
      } else {
        document.documentElement.classList.remove("dark");
      }
    }, 100);
  }, [theme]);

  useEffect(() => {
    const sendResizeMessage = () => {
      if (containerRef.current && window.parent !== window) {
        const rect = containerRef.current.getBoundingClientRect();
        window.parent.postMessage(
          {
            type: "resize",
            width: Math.ceil(containerRef.current.scrollWidth),
            height: Math.ceil(containerRef.current.scrollHeight),
          },
          "*"
        );
      }
    };

    // Send initial size
    sendResizeMessage();

    // Send size after a short delay to ensure content is rendered
    const timeoutId = setTimeout(sendResizeMessage, 100);

    // Listen for window resize events
    window.addEventListener("resize", sendResizeMessage);

    return () => {
      clearTimeout(timeoutId);
      window.removeEventListener("resize", sendResizeMessage);
    };
  }, [values]);

  return (
    <div
      ref={containerRef}
      className={className}
      style={{ maxWidth: "100%", overflow: "hidden" }}
    >
      {children}
    </div>
  );
}
