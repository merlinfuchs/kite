import { TemporalState } from "zundo";
import { createMessageStore, MessageStore } from "./messageStore";
import { createContext, ReactNode, useContext, useMemo, useState } from "react";
import { useStore } from "zustand";
import {
  createValidationErrorStore,
  ValidationErrorStore,
} from "./validationStore";

type ContextValue = {
  messageStore: ReturnType<typeof createMessageStore>;
  validationStore: ReturnType<typeof createValidationErrorStore>;
};

const CurrentMessageStoreContext = createContext<ContextValue | null>(null);

export function CurrentMessageStoreProvider({
  children,
}: {
  children: ReactNode;
}) {
  const [messageStore] = useState(() => createMessageStore());
  const [validationStore] = useState(() => createValidationErrorStore());

  const value = useMemo(
    () => ({
      messageStore,
      validationStore,
    }),
    [messageStore, validationStore]
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
