import { create } from "zustand";
import { persist } from "zustand/middleware";

const dismissDays = 3;

export interface SidebarSupportStore {
  dismissedUntil: number;
  dismiss: () => void;
}

export const useSidebarSupportStore = create<SidebarSupportStore>()(
  persist(
    (set) => ({
      dismissedUntil: 0,
      dismiss: () =>
        set({ dismissedUntil: Date.now() + dismissDays * 24 * 60 * 60 * 1000 }),
    }),
    { name: "kite-sidebar-support", version: 0 }
  )
);
