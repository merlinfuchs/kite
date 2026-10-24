import { createContext, ReactNode, useContext, useState } from "react";
import { create, StoreApi, UseBoundStore, useStore } from "zustand";
import { FlowParameter } from "../types/flow.gen";

interface FlowParametersStore {
  parameters: FlowParameter[];
  setParameters(parameters: FlowParameter[]): void;
}

type Store = UseBoundStore<StoreApi<FlowParametersStore>>;

const createParametersStore = (parameters: FlowParameter[]) =>
  create<FlowParametersStore>()((set) => ({
    parameters,
    setParameters: (parameters) => set({ parameters }),
  }));

const FlowParametersContext = createContext<Store | null>(null);

// FlowParametersProvider holds the parameters of the flow that's edited. They
// aren't blocks, so react-flow doesn't keep them, and saving reads them from
// here.
export function FlowParametersProvider({
  initial,
  children,
}: {
  initial?: FlowParameter[];
  children: ReactNode;
}) {
  const [store] = useState(() => createParametersStore(initial ?? []));
  return (
    <FlowParametersContext.Provider value={store}>
      {children}
    </FlowParametersContext.Provider>
  );
}

// Flows shown outside an editor have no parameters to change.
const emptyStore = createParametersStore([]);

export function useFlowParametersStore(): Store {
  return useContext(FlowParametersContext) ?? emptyStore;
}

export function useFlowParameters() {
  return useStore(useFlowParametersStore(), (s) => s.parameters);
}
