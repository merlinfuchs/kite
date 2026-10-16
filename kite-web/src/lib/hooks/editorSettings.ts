import { NodeData } from "@/lib/flow/dataSchema";
import {
  Node,
  OnMove,
  useNodesInitialized,
  useReactFlow,
  useStoreApi,
  Viewport,
} from "@xyflow/react";
import { useRouter } from "next/router";
import {
  RefObject,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { create } from "zustand";
import { persist } from "zustand/middleware";
import { useIsMobile } from "./use-mobile";

export type ReturnToCanvasBehavior = "root" | "previous" | "center";

export interface KeyboardMovementKeybinds {
  up: string;
  down: string;
  left: string;
  right: string;
  fast: string;
}

export const defaultKeyboardKeybinds: KeyboardMovementKeybinds = {
  up: "w",
  down: "s",
  left: "a",
  right: "d",
  fast: "shift",
};

const viewportStorageKey = "kite-canvas-viewports";
const maxSavedViewports = 30;
const basePanSpeed = 600;
const fastPanSpeed = 1400;
const velocitySmoothing = 20;
const minVelocityEpsilon = 8;

let cachedViewports: Record<string, Viewport> | null = null;
let saveIdleHandle: number | null = null;

function readViewports(): Record<string, Viewport> {
  if (cachedViewports) return cachedViewports;
  if (typeof window === "undefined") return {};
  try {
    const raw = window.localStorage.getItem(viewportStorageKey);
    cachedViewports = raw ? JSON.parse(raw) : {};
  } catch {
    cachedViewports = {};
  }
  return cachedViewports!;
}

function getSavedCanvasViewport(key: string): Viewport | undefined {
  return readViewports()[key];
}

// Stored outside reactive state and flushed during browser idle time so panning
// never triggers React re-renders or main-thread stalls on low-end devices.
function saveCanvasViewport(key: string, viewport: Viewport) {
  if (typeof window === "undefined") return;
  const map = readViewports();
  delete map[key];
  map[key] = {
    x: Math.round(viewport.x * 10) / 10,
    y: Math.round(viewport.y * 10) / 10,
    zoom: Math.round(viewport.zoom * 1000) / 1000,
  };

  const keys = Object.keys(map);
  while (keys.length > maxSavedViewports) {
    const oldest = keys.shift();
    if (oldest) delete map[oldest];
  }

  if (saveIdleHandle !== null) return;

  const flush = () => {
    saveIdleHandle = null;
    try {
      window.localStorage.setItem(viewportStorageKey, JSON.stringify(map));
    } catch {
      // Ignore storage quota errors
    }
  };

  if ("requestIdleCallback" in window) {
    saveIdleHandle = window.requestIdleCallback(flush, { timeout: 1000 });
  } else {
    saveIdleHandle = setTimeout(flush, 250) as unknown as number;
  }
}

export interface EditorSettingsState {
  showShare: boolean;
  keyboardMovement: boolean;
  fastMovement: boolean;
  keybinds: KeyboardMovementKeybinds;
  returnToCanvas: ReturnToCanvasBehavior;
}

export interface EditorSettingsStore extends EditorSettingsState {
  setShowShare: (showShare: boolean) => void;
  setKeyboardMovement: (keyboardMovement: boolean) => void;
  setFastMovement: (fastMovement: boolean) => void;
  setKeybind: (action: keyof KeyboardMovementKeybinds, key: string) => boolean;
  resetKeybinds: () => void;
  setReturnToCanvas: (returnToCanvas: ReturnToCanvasBehavior) => void;
}

function sanitizeKeybinds(
  raw?: Partial<KeyboardMovementKeybinds>
): KeyboardMovementKeybinds {
  const merged: KeyboardMovementKeybinds = {
    ...defaultKeyboardKeybinds,
    ...raw,
  };
  const seen = new Set<string>();
  for (const action of Object.keys(
    defaultKeyboardKeybinds
  ) as (keyof KeyboardMovementKeybinds)[]) {
    const key = (
      merged[action] || defaultKeyboardKeybinds[action]
    ).toLowerCase();
    if (seen.has(key)) {
      return { ...defaultKeyboardKeybinds };
    }
    seen.add(key);
    merged[action] = key;
  }
  return merged;
}

export const useEditorSettingsStore = create<EditorSettingsStore>()(
  persist(
    (set, get) => ({
      showShare: false,
      keyboardMovement: false,
      fastMovement: true,
      keybinds: defaultKeyboardKeybinds,
      returnToCanvas: "center",
      setShowShare: (showShare) => set({ showShare }),
      setKeyboardMovement: (keyboardMovement) => set({ keyboardMovement }),
      setFastMovement: (fastMovement) => set({ fastMovement }),
      setKeybind: (action, key) => {
        const normalized = key.toLowerCase();
        const current = get().keybinds;
        const isDuplicate = (
          Object.entries(current) as [keyof KeyboardMovementKeybinds, string][]
        ).some(([k, v]) => k !== action && v.toLowerCase() === normalized);
        if (isDuplicate) {
          return false;
        }
        set((state) => ({
          keybinds: {
            ...state.keybinds,
            [action]: normalized,
          },
        }));
        return true;
      },
      resetKeybinds: () => set({ keybinds: defaultKeyboardKeybinds }),
      setReturnToCanvas: (returnToCanvas) => set({ returnToCanvas }),
    }),
    {
      name: "kite-editor-settings",
      version: 1,
      merge: (persistedState, currentState) => {
        const persisted = persistedState as
          | Partial<EditorSettingsState>
          | undefined;
        return {
          ...currentState,
          ...persisted,
          keybinds: sanitizeKeybinds(persisted?.keybinds),
        };
      },
    }
  )
);

export function useEditorSettings(): EditorSettingsStore {
  const store = useEditorSettingsStore();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  return {
    showShare: mounted ? store.showShare : false,
    keyboardMovement: mounted ? store.keyboardMovement : false,
    fastMovement: mounted ? store.fastMovement : true,
    keybinds: mounted ? store.keybinds : defaultKeyboardKeybinds,
    returnToCanvas: mounted ? store.returnToCanvas : "center",
    setShowShare: store.setShowShare,
    setKeyboardMovement: store.setKeyboardMovement,
    setFastMovement: store.setFastMovement,
    setKeybind: store.setKeybind,
    resetKeybinds: store.resetKeybinds,
    setReturnToCanvas: store.setReturnToCanvas,
  };
}

export function useFlowNavigation({
  containerRef,
}: {
  containerRef: RefObject<HTMLElement>;
}) {
  const router = useRouter();
  const { fitView, getNodes, getViewport, setViewport } =
    useReactFlow<Node<NodeData>>();
  const store = useStoreApi<Node<NodeData>>();
  const nodesInitialized = useNodesInitialized();

  const isMobile = useIsMobile();
  const keyboardMovement = useEditorSettingsStore((s) => s.keyboardMovement);
  const fastMovement = useEditorSettingsStore((s) => s.fastMovement);
  const keybinds = useEditorSettingsStore((s) => s.keybinds);

  const flowKey = router.asPath;

  const initialDefaultViewport = useMemo(() => {
    const state = useEditorSettingsStore.getState();
    if (state.returnToCanvas === "previous") {
      return getSavedCanvasViewport(flowKey);
    }
    return undefined;
  }, [flowKey]);

  const hasInitializedRef = useRef(false);
  const isReadyToSaveRef = useRef(false);
  const isKeyboardPanningRef = useRef(false);

  useEffect(() => {
    if (!nodesInitialized || hasInitializedRef.current) return;
    hasInitializedRef.current = true;

    const behavior = useEditorSettingsStore.getState().returnToCanvas;
    const saved = getSavedCanvasViewport(flowKey);
    const nodes = getNodes();
    const entry = nodes.find((n) => n.type?.startsWith("entry_")) ?? nodes[0];

    if (behavior === "previous" && saved) {
      setViewport(saved);
    } else if (behavior === "root" && entry) {
      fitView({
        nodes: [{ id: entry.id }],
        padding: 0.5,
        maxZoom: 1,
      });
    } else {
      fitView();
    }

    const frameId = requestAnimationFrame(() => {
      isReadyToSaveRef.current = true;
      if (!saved) {
        saveCanvasViewport(flowKey, getViewport());
      }
    });

    return () => cancelAnimationFrame(frameId);
  }, [nodesInitialized, flowKey, fitView, getNodes, setViewport, getViewport]);

  const onMoveEnd: OnMove = useCallback(
    (_event, viewport) => {
      if (!isReadyToSaveRef.current || isKeyboardPanningRef.current) return;
      saveCanvasViewport(flowKey, viewport);
    },
    [flowKey]
  );

  useEffect(() => {
    if (!keyboardMovement || isMobile) return;

    const pressedKeys = new Set<string>();
    let frameId: number | null = null;
    let lastTime = 0;
    let vx = 0;
    let vy = 0;
    let isHoldingNode = false;
    const lastMouse = { x: 0, y: 0, initialized: false };

    const movementKeys = new Set([
      keybinds.up,
      keybinds.down,
      keybinds.left,
      keybinds.right,
    ]);

    const step = (now: number) => {
      if (!document.hasFocus()) {
        pressedKeys.clear();
        isHoldingNode = false;
      }

      const dt = lastTime ? Math.min((now - lastTime) / 1000, 0.05) : 1 / 60;
      lastTime = now;

      let dirX = 0;
      let dirY = 0;
      if (pressedKeys.has(keybinds.up)) dirY += 1;
      if (pressedKeys.has(keybinds.down)) dirY -= 1;
      if (pressedKeys.has(keybinds.left)) dirX += 1;
      if (pressedKeys.has(keybinds.right)) dirX -= 1;

      const isFast = fastMovement && pressedKeys.has(keybinds.fast);
      const speed = isFast ? fastPanSpeed : basePanSpeed;
      const len = Math.hypot(dirX, dirY);
      const targetVx = len > 0 ? (dirX / len) * speed : 0;
      const targetVy = len > 0 ? (dirY / len) * speed : 0;

      const alpha = 1 - Math.exp(-velocitySmoothing * dt);
      vx += (targetVx - vx) * alpha;
      vy += (targetVy - vy) * alpha;

      if (Math.hypot(vx, vy) < minVelocityEpsilon && len === 0) {
        vx = 0;
        vy = 0;
        frameId = null;
        lastTime = 0;
        if (isKeyboardPanningRef.current) {
          isKeyboardPanningRef.current = false;
          store.setState({ autoPanOnNodeDrag: true });
        }
        if (isReadyToSaveRef.current) {
          saveCanvasViewport(flowKey, getViewport());
        }
        return;
      }

      if (!isKeyboardPanningRef.current) {
        isKeyboardPanningRef.current = true;
        store.setState({ autoPanOnNodeDrag: false });
      }

      const state = store.getState();
      const [tx, ty, tZoom] = state.transform;
      const nextX = tx + vx * dt;
      const nextY = ty + vy * dt;
      const nextTransform: [number, number, number] = [nextX, nextY, tZoom];

      state.panZoom?.syncViewport({ x: nextX, y: nextY, zoom: tZoom });
      state.transform = nextTransform;

      // Keep held/dragged block(s) locked to the cursor while WASD pans the canvas,
      // even when the user's finger is stationary on a touchpad.
      if (isHoldingNode && lastMouse.initialized) {
        window.dispatchEvent(
          new MouseEvent("mousemove", {
            clientX: lastMouse.x,
            clientY: lastMouse.y,
            buttons: 1,
            bubbles: true,
            view: window,
          })
        );
      }

      store.setState({ transform: nextTransform });

      frameId = requestAnimationFrame(step);
    };

    const releaseAllKeys = () => {
      pressedKeys.clear();
      isHoldingNode = false;
    };

    const isBlockedByInputOrDialog = (e?: KeyboardEvent) => {
      const target = (e?.target ??
        document.activeElement) as HTMLElement | null;
      if (
        target?.closest(
          "input, textarea, select, [contenteditable], [role=textbox], [role=combobox], [role=listbox], [role=menu]"
        )
      ) {
        return true;
      }

      const dialogs = document.querySelectorAll("[role=dialog]");
      for (const dialog of dialogs) {
        if (!dialog.contains(containerRef.current)) {
          return true;
        }
      }

      return false;
    };

    const onKeyDown = (e: KeyboardEvent) => {
      const key = e.key.toLowerCase();
      const isFastKey = fastMovement && key === keybinds.fast;
      const isMoveKey = movementKeys.has(key);

      if (!isMoveKey && !isFastKey) return;
      if ((e.ctrlKey || e.metaKey) && key !== "control" && key !== "meta") {
        pressedKeys.clear();
        return;
      }
      if (e.altKey && key !== "alt") {
        pressedKeys.clear();
        return;
      }

      if (pressedKeys.has(key)) {
        if (isMoveKey) {
          e.preventDefault();
          e.stopPropagation();
        }
        return;
      }

      if (isBlockedByInputOrDialog(e)) {
        pressedKeys.clear();
        return;
      }

      if (isMoveKey) {
        e.preventDefault();
        e.stopPropagation();
      }
      if (e.shiftKey && fastMovement && keybinds.fast === "shift") {
        pressedKeys.add("shift");
      }
      pressedKeys.add(key);

      if (isMoveKey && frameId === null) {
        lastTime = performance.now();
        frameId = requestAnimationFrame(step);
      }
    };

    const onKeyUp = (e: KeyboardEvent) => {
      const key = e.key.toLowerCase();
      pressedKeys.delete(key);
      if (!e.shiftKey && keybinds.fast === "shift") {
        pressedKeys.delete("shift");
      }
    };

    const onMouseMove = (e: MouseEvent) => {
      if (!e.isTrusted) return;
      lastMouse.x = e.clientX;
      lastMouse.y = e.clientY;
      lastMouse.initialized = true;
    };

    const onMouseDown = (e: MouseEvent) => {
      if (!e.isTrusted) return;
      lastMouse.x = e.clientX;
      lastMouse.y = e.clientY;
      lastMouse.initialized = true;

      if (e.button !== 0) {
        releaseAllKeys();
        return;
      }

      const target = e.target as HTMLElement | null;
      if (
        target?.closest(
          ".react-flow__node, .react-flow__nodesselection-rect"
        ) &&
        !target.closest(".nodrag, .react-flow__handle")
      ) {
        isHoldingNode = true;
      }
    };

    const onMouseUp = (e: MouseEvent) => {
      if (!e.isTrusted) return;
      isHoldingNode = false;
    };

    const onVisibilityChange = () => {
      if (document.hidden) {
        releaseAllKeys();
      }
    };

    window.addEventListener("keydown", onKeyDown, true);
    window.addEventListener("keyup", onKeyUp, true);
    window.addEventListener("mousemove", onMouseMove, {
      capture: true,
      passive: true,
    });
    window.addEventListener("mousedown", onMouseDown, true);
    window.addEventListener("mouseup", onMouseUp, true);
    window.addEventListener("blur", releaseAllKeys);
    window.addEventListener("contextmenu", releaseAllKeys);
    document.addEventListener("visibilitychange", onVisibilityChange);

    return () => {
      if (frameId !== null) {
        cancelAnimationFrame(frameId);
      }
      if (isKeyboardPanningRef.current) {
        isKeyboardPanningRef.current = false;
        store.setState({ autoPanOnNodeDrag: true });
      }
      window.removeEventListener("keydown", onKeyDown, true);
      window.removeEventListener("keyup", onKeyUp, true);
      window.removeEventListener("mousemove", onMouseMove, true);
      window.removeEventListener("mousedown", onMouseDown, true);
      window.removeEventListener("mouseup", onMouseUp, true);
      window.removeEventListener("blur", releaseAllKeys);
      window.removeEventListener("contextmenu", releaseAllKeys);
      document.removeEventListener("visibilitychange", onVisibilityChange);
    };
  }, [
    keyboardMovement,
    isMobile,
    fastMovement,
    keybinds,
    containerRef,
    flowKey,
    store,
    getViewport,
  ]);

  return {
    defaultViewport: initialDefaultViewport,
    onMoveEnd,
  };
}
