import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface SetupChecklistStore {
  // Apps whose checklist was dismissed or completed. A completed step can
  // untick again, like when the monthly credits reset.
  hiddenAppIds: string[];
  hide: (appId: string) => void;
}

export const useSetupChecklistStore = create<SetupChecklistStore>()(
  persist(
    (set) => ({
      hiddenAppIds: [],
      hide: (appId) =>
        set((state) => ({
          hiddenAppIds: [...state.hiddenAppIds, appId],
        })),
    }),
    { name: "kite-setup-checklist", version: 0 }
  )
);
