import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface LegalUpdateStore {
  // legalUpdatedAt of the update the user dismissed the notice of.
  dismissedAt: number;
  dismiss: (updatedAt: number) => void;
}

export const useLegalUpdateStore = create<LegalUpdateStore>()(
  persist(
    (set) => ({
      dismissedAt: 0,
      dismiss: (updatedAt) => set({ dismissedAt: updatedAt }),
    }),
    { name: "kite-legal-update", version: 0 }
  )
);
