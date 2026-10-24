import { track } from "@/lib/analytics";
import { isNodeTypeAvailable } from "@/lib/flow/categories";
import {
  copyFlowNodes,
  parseFlowClipboard,
  pasteFlowNodes,
} from "@/lib/flow/clipboard";
import { useFlowContext } from "@/lib/flow/context";
import { NodeData } from "@/lib/flow/dataSchema";
import { Edge, Node, useReactFlow, XYPosition } from "@xyflow/react";
import { MouseEvent, useCallback, useEffect, useRef } from "react";
import { toast } from "sonner";

// Copy, cut and paste for flow blocks. Uses the system clipboard so the blocks
// can be pasted into other commands and event listeners. Returns a mouse move
// handler for the flow so pasted blocks land at the cursor, and the actions
// for the context menu.
export function useFlowClipboard() {
  const {
    getNodes,
    getEdges,
    setNodes,
    setEdges,
    deleteElements,
    screenToFlowPosition,
  } = useReactFlow<Node<NodeData>, Edge>();
  const contextType = useFlowContext((c) => c.type);
  const mousePosition = useRef<XYPosition | null>(null);

  const onMouseMove = useCallback((e: MouseEvent) => {
    mousePosition.current = { x: e.clientX, y: e.clientY };
  }, []);

  // The selected blocks as clipboard text, or null if none can be copied.
  const copySelected = useCallback(() => {
    const clipboard = copyFlowNodes(getNodes(), getEdges());
    return clipboard && JSON.stringify(clipboard);
  }, [getNodes, getEdges]);

  const deleteSelected = useCallback(() => {
    deleteElements({ nodes: getNodes().filter((n) => n.selected) });
  }, [getNodes, deleteElements]);

  // Pastes the blocks at the given screen position. Returns false if the text
  // isn't copied blocks.
  const pasteText = useCallback(
    (text: string, screenPosition: XYPosition | null) => {
      const clipboard = parseFlowClipboard(text);
      if (!clipboard) return false;

      const [newNodes, newEdges] = pasteFlowNodes(
        clipboard,
        screenPosition && screenToFlowPosition(screenPosition),
        (type) => isNodeTypeAvailable(type, contextType)
      );
      if (newNodes.length < clipboard.nodes.length) {
        toast.warning("Some blocks aren't available here and weren't pasted.");
      }
      if (newNodes.length === 0) return true;
      track("blocks_pasted", { count: newNodes.length });

      setNodes((nds) => [
        ...nds.map((n) => ({ ...n, selected: false })),
        ...newNodes,
      ]);
      setEdges((eds) => [
        ...eds.map((e) => ({ ...e, selected: false })),
        ...newEdges,
      ]);
      return true;
    },
    [screenToFlowPosition, contextType, setNodes, setEdges]
  );

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

      const text = copySelected();
      if (!text) return;

      e.preventDefault();
      e.clipboardData?.setData("text/plain", text);
    };

    const onCut = (e: ClipboardEvent) => {
      if (shouldIgnore(e)) return;

      const text = copySelected();
      if (!text) return;

      e.preventDefault();
      e.clipboardData?.setData("text/plain", text);
      deleteSelected();
    };

    const onPaste = (e: ClipboardEvent) => {
      if (shouldIgnore(e)) return;

      if (
        pasteText(
          e.clipboardData?.getData("text/plain") || "",
          mousePosition.current
        )
      ) {
        e.preventDefault();
      }
    };

    document.addEventListener("copy", onCopy);
    document.addEventListener("cut", onCut);
    document.addEventListener("paste", onPaste);
    return () => {
      document.removeEventListener("copy", onCopy);
      document.removeEventListener("cut", onCut);
      document.removeEventListener("paste", onPaste);
    };
  }, [copySelected, deleteSelected, pasteText]);

  const copy = useCallback(async () => {
    const text = copySelected();
    if (!text) return false;

    try {
      await navigator.clipboard.writeText(text);
      return true;
    } catch {
      toast.error("Couldn't copy the blocks to the clipboard.");
      return false;
    }
  }, [copySelected]);

  const cut = useCallback(async () => {
    if (await copy()) deleteSelected();
  }, [copy, deleteSelected]);

  const paste = useCallback(
    async (screenPosition: XYPosition) => {
      let text: string;
      try {
        text = await navigator.clipboard.readText();
      } catch {
        toast.error(
          "Couldn't read the clipboard, try pasting with the keyboard."
        );
        return;
      }

      if (!pasteText(text, screenPosition)) {
        toast.info("There are no blocks in the clipboard.");
      }
    },
    [pasteText]
  );

  return { onMouseMove, copy, cut, paste };
}
