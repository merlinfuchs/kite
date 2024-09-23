import { TemporalState } from "zundo";
import { createMessageStore, MessageStore } from "./messageStore";
import { createContext, ReactNode, useContext, useMemo, useState } from "react";
import { useStore } from "zustand";
import {
  createValidationErrorStore,
  ValidationErrorStore,
} from "./validationStore";
import { createFlowStore, FlowStore } from "./flowStore";

type ContextValue = {
  messageStore: ReturnType<typeof createMessageStore>;
  validationStore: ReturnType<typeof createValidationErrorStore>;
  flowStore: ReturnType<typeof createFlowStore>;
};

const CurrentMessageStoreContext = createContext<ContextValue | null>(null);

export function CurrentMessageStoreProvider({
  children,
}: {
  children: ReactNode;
}) {
  const [messageStore] = useState(() => createMessageStore());
  const [validationStore] = useState(() => createValidationErrorStore());
  const [flowStore] = useState(() => createFlowStore());

  const value = useMemo(
    () => ({
      messageStore,
      validationStore,
      flowStore,
    }),
    [messageStore, validationStore, flowStore]
  );

  return (
    <CurrentMessageStoreContext.Provider value={value}>
      {children}
    </CurrentMessageStoreContext.Provider>
  );
}

export function useCurrentMessageStore() {
  const value = useContext(CurrentMessageStoreContext);
  if (!value) {
    throw new Error(
      "useCurrentMessageStore must be used within a CurrentMessageStoreProvider provider"
    );
  }

  return value.messageStore;
}

export function useCurrentMessage<T>(selector: (store: MessageStore) => T): T {
  const store = useCurrentMessageStore();
  return useStore(store, selector);
}

export function useCurrentMessageUndo<T>(
  selector: (state: TemporalState<MessageStore>) => T
) {
  const store = useCurrentMessageStore();
  return useStore(store.temporal, selector);
}

export function useValidationErrorStore() {
  const value = useContext(CurrentMessageStoreContext);
  if (!value) {
    throw new Error(
      "useValidationErrorStore must be used within a CurrentMessageStoreProvider provider"
    );
  }

  return value.validationStore;
}

export function useValidationErrors<T>(
  selector: (store: ValidationErrorStore) => T
): T {
  const store = useValidationErrorStore();
  return useStore(store, selector);
}

export function useCurrentFlowStore() {
  const value = useContext(CurrentMessageStoreContext);
  if (!value) {
    throw new Error(
      "useCurrentFlowStore must be used within a CurrentMessageStoreProvider provider"
    );
  }

  return value.flowStore;
}

export function useCurrentFlow<T>(selector: (store: FlowStore) => T): T {
  const store = useCurrentFlowStore();
  return useStore(store, selector);
}
