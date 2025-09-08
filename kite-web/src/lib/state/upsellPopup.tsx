import { create } from "zustand";
import { persist } from "zustand/middleware";

const upsellAfterSeconds = 60 * 60;

export interface UpsellStateStore {
  pageFirstOpenedAt: number;
  upsellClosed: boolean;

  initialize: () => void;

  setUpsellClosed: (upsellClosed: boolean) => void;
  shouldUpsell: () => boolean;
}

export const useUpsellStateStore = create<UpsellStateStore>()(
  persist(
    (set, get) => ({
      pageFirstOpenedAt: 0,
      upsellClosed: false,

      initialize: () => {
        if (get().pageFirstOpenedAt === 0) {
          set({ pageFirstOpenedAt: Date.now() });
        }
      },

      setUpsellClosed: (upsellClosed: boolean) => set({ upsellClosed }),
      shouldUpsell: () => {
        if (get().upsellClosed) {
          return false;
        }

        return Date.now() - get().pageFirstOpenedAt > upsellAfterSeconds * 1000;
      },
    }),
    { name: "kite-upsell", version: 0 }
  )
);
