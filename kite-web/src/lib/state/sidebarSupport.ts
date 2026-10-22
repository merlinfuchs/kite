import { create } from "zustand";
import { persist } from "zustand/middleware";

export const THREE_DAYS_MS = 3 * 24 * 60 * 60 * 1000;

export interface SidebarSupportStore {
  dismissedUntil: number | null;
  dismiss: () => void;
  isDismissed: () => boolean;
}

export const useSidebarSupportStore = create<SidebarSupportStore>()(
  persist(
    (set, get) => ({
      dismissedUntil: null,
      dismiss: () => set({ dismissedUntil: Date.now() + THREE_DAYS_MS }),
      isDismissed: () => {
        const until = get().dismissedUntil;
        return typeof until === "number" && Date.now() < until;
      },
    }),
    { name: "kite-sidebar-support", version: 0 }
  )
);
