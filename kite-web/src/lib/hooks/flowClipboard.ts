import { track } from "@/lib/analytics";
import { isNodeTypeAvailable } from "@/lib/flow/categories";
import {
  copyFlowNodes,
  FlowClipboard,
  parseFlowClipboard,
  pasteFlowNodes,
} from "@/lib/flow/clipboard";
import { useFlowContext } from "@/lib/flow/context";
import { NodeData } from "@/lib/flow/dataSchema";
import { getDeletedNodeIds } from "@/lib/flow/nodes";
import { Edge, Node, useReactFlow } from "@xyflow/react";
import {
  MouseEvent,
  useCallback,
  useEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import { toast } from "sonner";

const STORAGE_KEY = "kite-flow-clipboard";

// Shared clipboard backed by localStorage so copy and paste works across tabs
// and flows without relying on OS clipboard permissions.
let sharedClipboard: FlowClipboard | null = null;
const clipboardListeners = new Set<() => void>();

export function getSharedClipboard(): FlowClipboard | null {
  if (sharedClipboard) return sharedClipboard;
  if (typeof window !== "undefined") {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        const parsed = parseFlowClipboard(stored);
        if (parsed) {
          sharedClipboard = parsed;
          return parsed;
        }
      }
    } catch {}
  }
  return null;
}

export function setSharedClipboard(next: FlowClipboard | null) {
  sharedClipboard = next;
  if (typeof window !== "undefined") {
    try {
      if (next) {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
      } else {
        localStorage.removeItem(STORAGE_KEY);
      }
    } catch {}
  }
  clipboardListeners.forEach((l) => l());
}

if (typeof window !== "undefined") {
  window.addEventListener("storage", (e) => {
    if (e.key === STORAGE_KEY) {
      if (e.newValue) {
        sharedClipboard = parseFlowClipboard(e.newValue);
      } else {
        sharedClipboard = null;
      }
      clipboardListeners.forEach((l) => l());
    }
  });
}

function subscribeClipboard(listener: () => void) {
  clipboardListeners.add(listener);
  return () => clipboardListeners.delete(listener);
}

interface UseFlowClipboardOptions {
  commit?: () => void;
  setNodesState?: React.Dispatch<React.SetStateAction<Node<NodeData>[]>>;
  setEdgesState?: React.Dispatch<React.SetStateAction<Edge[]>>;
}

// Copy, cut and paste for flow blocks.
export function useFlowClipboard(options?: UseFlowClipboardOptions) {
  const { getNodes, getEdges, setNodes, setEdges, screenToFlowPosition } =
    useReactFlow<Node<NodeData>, Edge>();
  const contextType = useFlowContext((c) => c.type);
  const mousePosition = useRef<{ x: number; y: number } | null>(null);
  const commit = options?.commit;
  const applyNodes = options?.setNodesState ?? setNodes;
  const applyEdges = options?.setEdgesState ?? setEdges;

  const currentClipboard = useSyncExternalStore(
    subscribeClipboard,
    getSharedClipboard,
    () => null
  );

  const onMouseMove = useCallback((e: MouseEvent) => {
    mousePosition.current = { x: e.clientX, y: e.clientY };
  }, []);

  const writeSystemClipboard = useCallback((clipboard: FlowClipboard) => {
    setSharedClipboard(clipboard);
  }, []);

  const insertFromClipboard = useCallback(
    (
      clipboard: FlowClipboard,
      screenPosition?: { x: number; y: number } | null
    ) => {
      const existingNodes = getNodes();
      let flowPosition = screenPosition
        ? screenToFlowPosition(screenPosition)
        : null;

      if (!flowPosition && clipboard.nodes.length > 0) {
        const minX = Math.min(...clipboard.nodes.map((n) => n.position.x));
        const minY = Math.min(...clipboard.nodes.map((n) => n.position.y));
        let offset = 50;
        while (
          existingNodes.some(
            (n) =>
              Math.abs(n.position.x - (minX + offset)) < 10 &&
              Math.abs(n.position.y - (minY + offset)) < 10
          ) &&
          offset < 500
        ) {
          offset += 40;
        }
        flowPosition = { x: minX + offset, y: minY + offset };
      }

      const [newNodes, newEdges] = pasteFlowNodes(
        clipboard,
        flowPosition,
        (type) => isNodeTypeAvailable(type, contextType)
      );
      if (newNodes.length < clipboard.nodes.length) {
        toast.warning("Some blocks aren't available here and weren't pasted.");
      }
      if (newNodes.length === 0) return false;
      track("blocks_pasted", { count: newNodes.length });

      commit?.();
      applyNodes((nds) => [
        ...nds.map((n) => (n.selected ? { ...n, selected: false } : n)),
        ...newNodes,
      ]);
      applyEdges((eds) => [
        ...eds.map((e) => (e.selected ? { ...e, selected: false } : e)),
        ...newEdges,
      ]);
      return true;
    },
    [
      getNodes,
      screenToFlowPosition,
      contextType,
      commit,
      applyNodes,
      applyEdges,
    ]
  );

  const removeSelectedNodes = useCallback(() => {
    const nodes = getNodes();
    const edges = getEdges();
    const selectedIds = nodes.filter((n) => n.selected).map((n) => n.id);
    const removed = getDeletedNodeIds(selectedIds, nodes, edges);
    if (removed.size === 0) return false;

    commit?.();
    applyEdges((eds) =>
      eds.filter((e) => !removed.has(e.source) && !removed.has(e.target))
    );
    applyNodes((nds) => nds.filter((n) => !removed.has(n.id)));
    return true;
  }, [getNodes, getEdges, commit, applyEdges, applyNodes]);

  const copy = useCallback(() => {
    const clipboard = copyFlowNodes(getNodes(), getEdges());
    if (!clipboard) return false;
    writeSystemClipboard(clipboard);
    return true;
  }, [getNodes, getEdges, writeSystemClipboard]);

  const cut = useCallback(() => {
    const clipboard = copyFlowNodes(getNodes(), getEdges());
    if (!clipboard) return false;
    writeSystemClipboard(clipboard);
    return removeSelectedNodes();
  }, [getNodes, getEdges, writeSystemClipboard, removeSelectedNodes]);

  const duplicate = useCallback(() => {
    const clipboard = copyFlowNodes(getNodes(), getEdges());
    if (!clipboard) return false;
    return insertFromClipboard(clipboard, null);
  }, [getNodes, getEdges, insertFromClipboard]);

  const paste = useCallback(
    async (screenPosition?: { x: number; y: number } | null) => {
      let clipboard: FlowClipboard | null = getSharedClipboard();

      if (
        !clipboard &&
        typeof navigator !== "undefined" &&
        navigator.clipboard?.readText
      ) {
        try {
          const text = await navigator.clipboard.readText();
          clipboard = parseFlowClipboard(text);
        } catch {}
      }

      if (!clipboard) return false;

      if (clipboard !== sharedClipboard) {
        setSharedClipboard(clipboard);
      }
      return insertFromClipboard(clipboard, screenPosition);
    },
    [insertFromClipboard]
  );

  useEffect(() => {
    const shouldIgnore = (e: Event) => {
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
      setSharedClipboard(clipboard);
    };

    const onCut = (e: ClipboardEvent) => {
      if (shouldIgnore(e)) return;

      const clipboard = copyFlowNodes(getNodes(), getEdges());
      if (!clipboard) return;

      e.preventDefault();
      setSharedClipboard(clipboard);
      removeSelectedNodes();
    };

    const onPaste = (e: ClipboardEvent) => {
      if (shouldIgnore(e)) return;

      const clipboard =
        getSharedClipboard() ??
        parseFlowClipboard(e.clipboardData?.getData("text/plain") || "");
      if (!clipboard) return;

      e.preventDefault();
      setSharedClipboard(clipboard);
      insertFromClipboard(clipboard, mousePosition.current);
    };

    const onKeyDown = (e: KeyboardEvent) => {
      if (!(e.metaKey || e.ctrlKey) || e.altKey || e.shiftKey) return;
      if (e.key.toLowerCase() !== "d") return;
      if (shouldIgnore(e)) return;

      const clipboard = copyFlowNodes(getNodes(), getEdges());
      if (!clipboard) return;

      e.preventDefault();
      insertFromClipboard(clipboard, null);
    };

    document.addEventListener("copy", onCopy);
    document.addEventListener("cut", onCut);
    document.addEventListener("paste", onPaste);
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("copy", onCopy);
      document.removeEventListener("cut", onCut);
      document.removeEventListener("paste", onPaste);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [getNodes, getEdges, insertFromClipboard, removeSelectedNodes]);

  return {
    onMouseMove,
    copy,
    cut,
    duplicate,
    paste,
    hasClipboard: !!currentClipboard,
  };
}
