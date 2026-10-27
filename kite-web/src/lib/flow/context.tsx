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
import { Message, Variable } from "../types/wire.gen";
import { useMessages, useVariables } from "../hooks/api";

// Read-only previews, like the flows of a template, don't show the validation
// markers on blocks. Their parameters are only filled in when the template is
// added.
export const FlowReadOnlyContext = createContext(false);

// The stored variables and message templates the blocks of a preview use,
// like the ones of a template, which the app doesn't have until it's added.
// Blocks can still use the app's too.
export type FlowPreviewItems = {
  variables?: PreviewVariable[];
  messages?: PreviewMessage[];
};

type PreviewVariable = Pick<Variable, "id" | "name" | "scoped">;
// The data is only checked when the message is shown.
type PreviewMessage = Pick<Message, "id" | "name"> & { data: unknown };

export const FlowPreviewItemsContext = createContext<FlowPreviewItems>({});

// useFlowVariables are the stored variables blocks can use: the ones of a
// preview, then the app's.
export function useFlowVariables() {
  return withPreviewItems<PreviewVariable>(
    useContext(FlowPreviewItemsContext).variables,
    useVariables()
  );
}

// useFlowMessages are the message templates blocks can use: the ones of a
// preview, then the app's.
export function useFlowMessages() {
  return withPreviewItems<PreviewMessage>(
    useContext(FlowPreviewItemsContext).messages,
    useMessages()
  );
}

function withPreviewItems<T extends { id: string }>(
  preview: T[] | undefined,
  app: (T | null | undefined)[] | undefined
): T[] | undefined {
  if (!preview) return app?.flatMap((item) => (item ? [item] : []));
  const ids = new Set(preview.map((item) => item.id));
  return [
    ...preview,
    ...(app ?? []).flatMap((item) => (item && !ids.has(item.id) ? [item] : [])),
  ];
}

export const flowContextTypes = [
  "command",
  "component_button",
  "component_select_menu",
  "event_discord",
  "event_schedule",
  "event_webhook",
] as const;

export type FlowContextType = (typeof flowContextTypes)[number];

export interface FlowContextStore {
  type: FlowContextType;
  setType(type: FlowContextType): void;
  // The blocks the flow AI changed in its last answer, which stand out until
  // the user keeps or undoes the changes.
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
