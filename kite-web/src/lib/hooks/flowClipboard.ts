import { isNodeTypeAvailable } from "@/lib/flow/categories";
import {
  copyFlowNodes,
  parseFlowClipboard,
  pasteFlowNodes,
} from "@/lib/flow/clipboard";
import { useFlowContext } from "@/lib/flow/context";
import { NodeData } from "@/lib/flow/dataSchema";
import { Edge, Node, useReactFlow } from "@xyflow/react";
import {
  Dispatch,
  MouseEvent,
  SetStateAction,
  useCallback,
  useEffect,
  useRef,
} from "react";
import { toast } from "sonner";

// Copy and paste for flow blocks. Uses the native copy and paste events so the
// blocks go through the system clipboard and can be pasted into other commands
// and event listeners. Returns a mouse move handler for the flow so pasted
// blocks land at the cursor.
export function useFlowClipboard({
  setNodes,
  setEdges,
  onChange,
}: {
  setNodes: Dispatch<SetStateAction<Node<NodeData>[]>>;
  setEdges: Dispatch<SetStateAction<Edge[]>>;
  onChange: () => void;
}) {
  const { getNodes, getEdges, screenToFlowPosition } = useReactFlow<
    Node<NodeData>,
    Edge
  >();
  const contextType = useFlowContext((c) => c.type);
  const mousePosition = useRef<{ x: number; y: number } | null>(null);

  const onMouseMove = useCallback((e: MouseEvent) => {
    mousePosition.current = { x: e.clientX, y: e.clientY };
  }, []);

  useEffect(() => {
    const shouldIgnore = (e: ClipboardEvent) => {
      const target = e.target as HTMLElement | null;
      return (
        !!target?.closest(
          "input, textarea, [contenteditable], [role=dialog]"
        ) || !!window.getSelection()?.toString()
      );
    };

    const onCopy = (e: ClipboardEvent) => {
      if (shouldIgnore(e)) return;

      const clipboard = copyFlowNodes(getNodes(), getEdges());
      if (!clipboard) return;

      e.preventDefault();
      e.clipboardData?.setData("text/plain", JSON.stringify(clipboard));
    };

    const onPaste = (e: ClipboardEvent) => {
      if (shouldIgnore(e)) return;

      const clipboard = parseFlowClipboard(
        e.clipboardData?.getData("text/plain") || ""
      );
      if (!clipboard) return;

      e.preventDefault();

      const [newNodes, newEdges] = pasteFlowNodes(
        clipboard,
        mousePosition.current && screenToFlowPosition(mousePosition.current),
        (type) => isNodeTypeAvailable(type, contextType),
        getNodes().map((n) => n.id)
      );
      if (newNodes.length < clipboard.nodes.length) {
        toast.warning("Some blocks aren't available here and weren't pasted.");
      }
      if (newNodes.length === 0) return;

      setNodes((nds) => [
        ...nds.map((n) => ({ ...n, selected: false })),
        ...newNodes,
      ]);
      setEdges((eds) => [
        ...eds.map((e) => ({ ...e, selected: false })),
        ...newEdges,
      ]);
      onChange();
    };

    document.addEventListener("copy", onCopy);
    document.addEventListener("paste", onPaste);
    return () => {
      document.removeEventListener("copy", onCopy);
      document.removeEventListener("paste", onPaste);
    };
  }, [
    getNodes,
    getEdges,
    screenToFlowPosition,
    contextType,
    setNodes,
    setEdges,
    onChange,
  ]);

  return onMouseMove;
}
