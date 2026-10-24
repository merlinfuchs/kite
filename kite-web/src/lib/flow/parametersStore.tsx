import {
  createContext,
  ReactNode,
  useCallback,
  useContext,
  useMemo,
  useState,
} from "react";
import { useReactFlow, useStore as useFlowStore } from "@xyflow/react";
import { create, StoreApi, UseBoundStore, useStore } from "zustand";
import { shallow } from "zustand/shallow";
import { FlowParameter } from "../types/flow.gen";
import { FlowData, NodeType } from "./dataSchema";
import {
  needsValue,
  referencedParameters,
  syncParameters,
  withParameterKeys,
} from "./parameters";

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

// useReferencedParameters are the keys of the parameters the blocks use.
// Moving blocks doesn't change them, so it doesn't re-render.
export function useReferencedParameters() {
  return useFlowStore((s) => referencedParameters({ nodes: s.nodes }), shallow);
}

// useSyncedParameters are the parameters with one for every key a block uses.
export function useSyncedParameters() {
  const keys = useReferencedParameters();
  const stored = useFlowParameters();
  return useMemo(() => withParameterKeys(stored, keys), [keys, stored]);
}

// useMissingParameterCount is how many parameters blocks use that need a
// value they don't have.
export function useMissingParameterCount() {
  const keys = useReferencedParameters();
  const parameters = useSyncedParameters();
  return useMemo(
    () =>
      parameters.filter((p) => keys.includes(p.key) && needsValue(p)).length,
    [keys, parameters]
  );
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
