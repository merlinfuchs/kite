import { create } from "zustand";
import { persist } from "zustand/middleware";

const upsellAfterSeconds = 60 * 60;
// Closing the popup only hides it for a week.
const upsellAgainAfterSeconds = 7 * 24 * 60 * 60;

export interface UpsellStateStore {
  pageFirstOpenedAt: number;
  upsellClosedAt: number;

  initialize: () => void;

  closeUpsell: () => void;
  shouldUpsell: () => boolean;
}

export const useUpsellStateStore = create<UpsellStateStore>()(
  persist(
    (set, get) => ({
      pageFirstOpenedAt: 0,
      upsellClosedAt: 0,

      initialize: () => {
        if (get().pageFirstOpenedAt === 0) {
          set({ pageFirstOpenedAt: Date.now() });
        }
      },

      closeUpsell: () => set({ upsellClosedAt: Date.now() }),
      shouldUpsell: () => {
        const now = Date.now();
        if (now - get().upsellClosedAt < upsellAgainAfterSeconds * 1000) {
          return false;
        }

        return now - get().pageFirstOpenedAt > upsellAfterSeconds * 1000;
      },
    }),
    {
      name: "kite-upsell",
      version: 1,
      // Version 0 hid the popup for good once closed. Those users see it
      // again a week from now.
      migrate: (state, version) => {
        const old = state as {
          pageFirstOpenedAt?: number;
          upsellClosed?: boolean;
        };
        if (version === 0) {
          return {
            pageFirstOpenedAt: old.pageFirstOpenedAt ?? 0,
            upsellClosedAt: old.upsellClosed ? Date.now() : 0,
          };
        }
        return state;
      },
    }
  )
);
