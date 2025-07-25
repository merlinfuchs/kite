import { useEffect, useRef, useState } from "react";
import { useColorMode } from "@docusaurus/theme-common";
import { useInView } from "react-intersection-observer";

export default function EmbedFlowNode({ type }: { type: string }) {
  const { colorMode } = useColorMode();

  const { ref, inView, entry } = useInView({
    triggerOnce: true,
    threshold: 0,
  });

  const iframeRef = useRef<HTMLIFrameElement>(null);
  const [dimensions, setDimensions] = useState({
    width: "100%",
    height: "0",
  });

  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      if (
        event.data.type === "resize" &&
        event.data.width &&
        event.data.height
      ) {
        setDimensions({
          width: `${event.data.width}px`,
          height: `${event.data.height}px`,
        });
      }
    };

    window.addEventListener("message", handleMessage);

    return () => {
      window.removeEventListener("message", handleMessage);
    };
  }, []);

  const queryParams = new URLSearchParams();
  queryParams.set("type", type);
  if (colorMode === "dark") {
    queryParams.set("theme", "dark");
  }

  return (
    <div style={{ margin: "10px 0" }} ref={ref}>
      {inView && (
        <iframe
          ref={iframeRef}
          src={`http://localhost:3000/embed/flow/node?${queryParams.toString()}`}
          style={{
            width: dimensions.width,
            height: dimensions.height,
            border: "none",
            borderRadius: "10px",
            overflow: "hidden",
            display: "block",
            minWidth: "0",
            minHeight: "0",
          }}
        />
      )}
    </div>
  );
}
