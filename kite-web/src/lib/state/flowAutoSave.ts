import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface FlowAutoSaveStore {
  // Flows that save on their own, by their viewport key like "command:<id>".
  enabledKeys: string[];
  setEnabled: (key: string, enabled: boolean) => void;
}

export const useFlowAutoSaveStore = create<FlowAutoSaveStore>()(
  persist(
    (set) => ({
      enabledKeys: [],
      setEnabled: (key, enabled) =>
        set((state) => {
          const rest = state.enabledKeys.filter((k) => k !== key);
          return { enabledKeys: enabled ? [...rest, key] : rest };
        }),
    }),
    { name: "kite-flow-auto-save", version: 0 }
  )
);
