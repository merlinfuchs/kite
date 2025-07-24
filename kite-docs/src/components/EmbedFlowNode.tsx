import { useEffect, useRef, useState } from "react";

export default function EmbedFlowNode({ type }: { type: string }) {
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

  return (
    <iframe
      ref={iframeRef}
      src={`http://localhost:3000/embed/flow/node?type=${type}`}
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
  );
}
