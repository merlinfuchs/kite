import {
  createContext,
  ReactNode,
  useCallback,
  useContext,
  useMemo,
  useState,
} from "react";
import { useNodes, useReactFlow } from "@xyflow/react";
import { create, StoreApi, UseBoundStore, useStore } from "zustand";
import { FlowParameter } from "../types/flow.gen";
import { FlowData, NodeType } from "./dataSchema";
import { syncParameters } from "./parameters";

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

// useSyncedParameters are the parameters with one for every key a block uses.
export function useSyncedParameters() {
  const nodes = useNodes();
  const stored = useFlowParameters();
  return useMemo(() => syncParameters({ nodes }, stored), [nodes, stored]);
}

// useGetFlowData returns a function for the flow as it's saved, its blocks and
// the parameters they use.
export function useGetFlowData() {
  const { getNodes, getEdges } = useReactFlow<NodeType>();
  const store = useFlowParametersStore();
  return useCallback((): FlowData => {
    const nodes = getNodes();
    return {
      nodes,
      edges: getEdges(),
      parameters: syncParameters({ nodes }, store.getState().parameters),
    };
  }, [getNodes, getEdges, store]);
}
