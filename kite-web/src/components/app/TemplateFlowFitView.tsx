import { useNodesInitialized, useReactFlow, useStore } from "@xyflow/react";
import { useEffect } from "react";

// FitView fits the flow in once its blocks are measured and the preview has
// a size, which it may not have yet when the item or dialog it's in opens.
export default function FitView() {
  const { fitView } = useReactFlow();
  const initialized = useNodesInitialized();
  const width = useStore((s) => s.width);
  const height = useStore((s) => s.height);

  useEffect(() => {
    if (initialized && width && height) fitView({ maxZoom: 1 });
  }, [initialized, width, height, fitView]);

  return null;
}
