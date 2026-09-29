import {
  createContext,
  ReactNode,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { create, useStore } from "zustand";
import { immer } from "zustand/middleware/immer";

export const flowContextTypes = [
  "command",
  "component_button",
  "component_select_menu",
  "event_discord",
  "event_schedule",
] as const;

export type FlowContextType = (typeof flowContextTypes)[number];

export interface FlowContextStore {
  type: FlowContextType;
  setType(type: FlowContextType): void;
  // The blocks the flow AI changed in its last answer, which stand out until
  // the user clicks into the editor.
  aiChangedNodeIds: string[];
  setAIChangedNodeIds(ids: string[]): void;
}

export const createFlowContextStore = () => {
  return create<FlowContextStore>()(
    immer((set, get) => ({
      type: "command",

      setType: (type) => set({ type }),
      aiChangedNodeIds: [],
      setAIChangedNodeIds: (ids) => set({ aiChangedNodeIds: ids }),
    }))
  );
};

const FlowContextStoreContext = createContext<ReturnType<
  typeof createFlowContextStore
> | null>(null);

export function FlowContextStoreProvider({
  children,
  type,
}: {
  children: ReactNode;
  type: FlowContextType;
}) {
  const [contextStore] = useState(() => createFlowContextStore());

  useEffect(() => {
    contextStore.getState().setType(type);
  }, [type, contextStore]);

  return (
    <FlowContextStoreContext.Provider value={contextStore}>
      {children}
    </FlowContextStoreContext.Provider>
  );
}

export function useFlowContextStore() {
  const value = useContext(FlowContextStoreContext);
  if (!value) {
    throw new Error(
      "useFlowContextStore must be used within a FlowContextStore provider"
    );
  }
  return value;
}

export function useFlowContext<T>(selector: (store: FlowContextStore) => T): T {
  const store = useFlowContextStore();
  return useStore(store, selector);
}

// Blocks are also shown outside the editor, e.g. in examples, where nothing
// was changed by the AI.
const emptyContextStore = createFlowContextStore();

export function useChangedByAI(nodeId: string) {
  const store = useContext(FlowContextStoreContext);
  return useStore(store ?? emptyContextStore, (s) =>
    s.aiChangedNodeIds.includes(nodeId)
  );
}
