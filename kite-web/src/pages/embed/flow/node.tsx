import { useNodeValues } from "@/lib/flow/nodes";
import { useRouter } from "next/router";
import { useEffect, useRef } from "react";

export default function EmbedFlowNodePage() {
  const router = useRouter();
  const containerRef = useRef<HTMLDivElement>(null);

  const { type, theme = "light" } = router.query;

  const values = useNodeValues(type as string);

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
      className="p-3 bg-muted relative select-none inline-block w-[400px]"
      style={{ maxWidth: "100%", overflow: "hidden" }}
    >
      <div className="flex items-start space-x-3">
        <div
          className="rounded-md w-8 h-8 flex justify-center items-center flex-none"
          style={{ backgroundColor: values.color }}
        >
          <values.icon className="h-5 w-5 text-white" />
        </div>
        <div className="overflow-hidden flex-1 min-w-0">
          <div className="font-medium text-foreground leading-5 mb-1 truncate">
            {values.defaultTitle}
          </div>
          <div className="text-sm text-muted-foreground">
            {values.defaultDescription}
          </div>
        </div>
      </div>
    </div>
  );
}
