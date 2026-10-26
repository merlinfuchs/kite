import { track } from "@/lib/analytics";
import { isNodeTypeAvailable } from "@/lib/flow/categories";
import {
  copyFlowNodes,
  duplicateFlowNodes,
  parseFlowClipboard,
  pasteFlowNodes,
} from "@/lib/flow/clipboard";
import { useFlowContext } from "@/lib/flow/context";
import { NodeData } from "@/lib/flow/dataSchema";
import { getDeletedNodeIds } from "@/lib/flow/nodes";
import { Edge, Node, useReactFlow, XYPosition } from "@xyflow/react";
import { MouseEvent, RefObject, useCallback, useEffect, useRef } from "react";
import { toast } from "sonner";

// Copy, cut and paste for flow blocks. Uses the system clipboard so the blocks
// can be pasted into other commands and event listeners. Returns a mouse move
// handler for the flow so pasted blocks land at the cursor, and the actions
// for the context menu, which take the IDs of the blocks they act on.
export function useFlowClipboard(containerRef: RefObject<HTMLElement | null>) {
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

  const getSelectedIds = useCallback(
    () =>
      getNodes()
        .filter((n) => n.selected)
        .map((n) => n.id),
    [getNodes]
  );

  // The blocks as clipboard text, or null if none can be copied.
  const copyText = useCallback(
    (ids: string[]) => {
      const clipboard = copyFlowNodes(getNodes(), getEdges(), ids);
      return clipboard && JSON.stringify(clipboard);
    },
    [getNodes, getEdges]
  );

  // Fixed blocks like the entry are left out, otherwise react-flow would
  // remove their wires.
  const deleteNodes = useCallback(
    (ids: string[]) => {
      const deletedIds = getDeletedNodeIds(ids, getNodes(), getEdges());
      deleteElements({ nodes: [...deletedIds].map((id) => ({ id })) });
    },
    [getNodes, getEdges, deleteElements]
  );

  // Adds the blocks as the new selection.
  const insertNodes = useCallback(
    (newNodes: Node<NodeData>[], newEdges: Edge[]) => {
      setNodes((nds) => [
        ...nds.map((n) => (n.selected ? { ...n, selected: false } : n)),
        ...newNodes,
      ]);
      setEdges((eds) => [
        ...eds.map((e) => (e.selected ? { ...e, selected: false } : e)),
        ...newEdges,
      ]);
    },
    [setNodes, setEdges]
  );

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

      insertNodes(newNodes, newEdges);
      return true;
    },
    [screenToFlowPosition, contextType, insertNodes]
  );

  useEffect(() => {
    const shouldIgnore = (e: ClipboardEvent) => {
      const target = e.target as HTMLElement | null;
      if (
        target?.closest("input, textarea, [contenteditable]") ||
        window.getSelection()?.toString()
      ) {
        return true;
      }

      // Dialogs opened on top of the flow, like the message editor, have
      // their own clipboard.
      const dialog = target?.closest("[role=dialog]");
      return !!dialog && !dialog.contains(containerRef.current);
    };

    // Writes the selected blocks to the clipboard of a copy or cut event.
    // Returns false if the event isn't for them.
    const writeSelection = (e: ClipboardEvent) => {
      if (shouldIgnore(e)) return false;

      const text = copyText(getSelectedIds());
      if (!text) return false;

      e.preventDefault();
      e.clipboardData?.setData("text/plain", text);
      return true;
    };

    const onCut = (e: ClipboardEvent) => {
      if (writeSelection(e)) deleteNodes(getSelectedIds());
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

    document.addEventListener("copy", writeSelection);
    document.addEventListener("cut", onCut);
    document.addEventListener("paste", onPaste);
    return () => {
      document.removeEventListener("copy", writeSelection);
      document.removeEventListener("cut", onCut);
      document.removeEventListener("paste", onPaste);
    };
  }, [containerRef, copyText, getSelectedIds, deleteNodes, pasteText]);

  const copy = useCallback(
    async (ids: string[]) => {
      const text = copyText(ids);
      if (!text) return false;

      try {
        await navigator.clipboard.writeText(text);
        return true;
      } catch {
        toast.error("Couldn't copy the blocks to the clipboard.");
        return false;
      }
    },
    [copyText]
  );

  const cut = useCallback(
    async (ids: string[]) => {
      if (await copy(ids)) deleteNodes(ids);
    },
    [copy, deleteNodes]
  );

  const duplicate = useCallback(
    (ids: string[]) => {
      const [newNodes, newEdges] = duplicateFlowNodes(
        ids,
        getNodes(),
        getEdges()
      );
      insertNodes(newNodes, newEdges);
    },
    [getNodes, getEdges, insertNodes]
  );

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

  return { onMouseMove, copy, cut, duplicate, deleteNodes, paste };
}
