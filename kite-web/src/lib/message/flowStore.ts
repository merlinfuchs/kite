import { create } from "zustand";
import { FlowData } from "../flow/dataSchema";
import { immer } from "zustand/middleware/immer";

export interface FlowStore {
  flowSources: Record<string, FlowData>;
  replaceAll(data: Record<string, FlowData>): void;
  replaceFlow(id: string, data: FlowData): void;
  getFlow(id: string): FlowData | undefined;
}

export const createFlowStore = (initialData?: Record<string, FlowData>) => {
  return create<FlowStore>()(
    immer((set, get) => ({
      flowSources: initialData || {},

      replaceAll: (data) => set({ flowSources: data }),
      replaceFlow: (id: string, data: FlowData) =>
        set((state) => {
          state.flowSources[id] = data;
        }),
      getFlow: (id: string) => get().flowSources[id],
    }))
  );
};
