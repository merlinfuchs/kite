import { TemporalState } from "zundo";
import {
  createDocumentStore,
  DocumentData,
  DocumentStore,
  isComponentsV2,
  Node,
  NodeId,
  slotLimit,
  slotOfChild,
} from "./document";
import { ChildSlot, childIds, toMessage } from "./documentConvert";
import { createContext, ReactNode, useContext, useMemo, useState } from "react";
import { useStore } from "zustand";
import { useShallow } from "zustand/react/shallow";
import {
  createValidationErrorStore,
  ValidationErrorStore,
} from "./validationStore";
import { createFlowStore, FlowStore } from "./flowStore";
import { Message } from "./schema";

type ContextValue = {
  documentStore: ReturnType<typeof createDocumentStore>;
  validationStore: ReturnType<typeof createValidationErrorStore>;
  flowStore: ReturnType<typeof createFlowStore>;
};

const CurrentMessageStoreContext = createContext<ContextValue | null>(null);

export function CurrentMessageStoreProvider({
  children,
}: {
  children: ReactNode;
}) {
  const [documentStore] = useState(() => createDocumentStore());
  const [validationStore] = useState(() => createValidationErrorStore());
  const [flowStore] = useState(() => createFlowStore());

  const value = useMemo(
    () => ({
      documentStore,
      validationStore,
      flowStore,
    }),
    [documentStore, validationStore, flowStore]
  );

  return (
    <CurrentMessageStoreContext.Provider value={value}>
      {children}
    </CurrentMessageStoreContext.Provider>
  );
}

function useStoreContext(hook: string) {
  const value = useContext(CurrentMessageStoreContext);
  if (!value) {
    throw new Error(
      `${hook} must be used within a CurrentMessageStoreProvider provider`
    );
  }
  return value;
}

export function useDocumentStoreApi() {
  return useStoreContext("useDocumentStoreApi").documentStore;
}

export function useDocument<T>(selector: (state: DocumentStore) => T): T {
  return useStore(useDocumentStoreApi(), selector);
}

/** The undo stack only tracks the document itself, not the store actions. */
export function useDocumentUndo<T>(
  selector: (state: TemporalState<DocumentData>) => T
) {
  return useStore(useDocumentStoreApi().temporal, selector);
}

/** The message payload the document currently describes. */
export function getMessage(store: { getState(): DocumentData }): Message {
  return toMessage(store.getState()).message;
}

export const useRootId = () => useDocument((state) => state.rootId);

export const useNode = <T extends Node>(id: NodeId) =>
  useDocument((state) => state.nodes[id] as T | undefined);

export const useChildIds = (id: NodeId, slot: ChildSlot) =>
  useDocument(useShallow((state) => childIds(state.nodes[id], slot)));

export const useComponentsV2Enabled = () => useDocument(isComponentsV2);

/** Position of a node among its siblings, for move and duplicate buttons. */
export const useNodeIndex = (id: NodeId) =>
  useDocument(
    useShallow((state) => {
      const node = state.nodes[id];
      const parent = node?.parentId ? state.nodes[node.parentId] : undefined;
      const slot = parent && slotOfChild(parent, id);
      const ids = slot ? childIds(parent, slot) : [];

      return { index: ids.indexOf(id), count: ids.length };
    })
  );

/** How many children the slot a node sits in can hold. */
export const useSlotLimit = (id: NodeId) =>
  useDocument((state) => {
    const node = state.nodes[id];
    const parent = node?.parentId ? state.nodes[node.parentId] : undefined;
    const slot = parent && slotOfChild(parent, id);

    return parent && slot
      ? slotLimit(parent.type, slot, isComponentsV2(state))
      : 1;
  });

/**
 * The move, duplicate and remove handlers of a node, left undefined at the ends
 * of its slot and once the slot is full.
 */
export function useNodeActions(id: NodeId) {
  const { index, count } = useNodeIndex(id);
  const max = useSlotLimit(id);
  const { move, duplicate, remove } = useDocumentStoreApi().getState();

  return {
    moveUp: index > 0 ? () => move(id, -1) : undefined,
    moveDown: index < count - 1 ? () => move(id, 1) : undefined,
    duplicate: count < max ? () => duplicate(id) : undefined,
    remove: () => remove(id),
  };
}

export function useValidationErrorStore() {
  return useStoreContext("useValidationErrorStore").validationStore;
}

export function useValidationErrors<T>(
  selector: (store: ValidationErrorStore) => T
): T {
  return useStore(useValidationErrorStore(), selector);
}

export function useCurrentFlowStore() {
  return useStoreContext("useCurrentFlowStore").flowStore;
}

export function useCurrentFlow<T>(selector: (store: FlowStore) => T): T {
  return useStore(useCurrentFlowStore(), selector);
}
