import { useEffect, useRef, useState } from "react";
import { useColorMode } from "@docusaurus/theme-common";
import { useInView } from "react-intersection-observer";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";

export default function EmbedBase({
  src,
  params,
}: {
  src: string;
  params?: Record<string, string>;
}) {
  const { colorMode } = useColorMode();

  const {
    siteConfig: { customFields },
  } = useDocusaurusContext();

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
  if (params) {
    for (const [key, value] of Object.entries(params)) {
      queryParams.set(key, value);
    }
  }

  if (colorMode === "dark") {
    queryParams.set("theme", "dark");
  }

  const url = `${customFields.appBaseUrl}${src}?${queryParams.toString()}`;

  return (
    <div style={{ margin: "20px 0" }} ref={ref}>
      {inView && (
        <iframe
          ref={iframeRef}
          src={url}
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
