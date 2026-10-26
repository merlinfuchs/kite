import { Viewport } from "@xyflow/react";
import { create } from "zustand";
import { persist } from "zustand/middleware";

const maxViewports = 30;

export interface FlowViewportsStore {
  // Ordered from least to most recently moved.
  viewports: Record<string, Viewport>;
  save: (key: string, viewport: Viewport) => void;
}

export const useFlowViewportsStore = create<FlowViewportsStore>()(
  persist(
    (set) => ({
      viewports: {},
      save: (key, viewport) =>
        set((state) => {
          const { [key]: _, ...rest } = state.viewports;
          const entries = Object.entries(rest).slice(1 - maxViewports);
          return {
            viewports: { ...Object.fromEntries(entries), [key]: viewport },
          };
        }),
    }),
    { name: "kite-flow-viewports", version: 0 }
  )
);
