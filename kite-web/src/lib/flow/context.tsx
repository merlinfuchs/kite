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

export type FlowContextType = "command" | "component_button" | "event_discord";

export interface FlowContextStore {
  type: FlowContextType;
  setType(type: FlowContextType): void;
}

export const createFlowContextStore = () => {
  return create<FlowContextStore>()(
    immer((set, get) => ({
      type: "command",

      setType: (type) => set({ type }),
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
