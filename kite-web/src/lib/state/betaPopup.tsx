import { create } from "zustand";
import { persist } from "zustand/middleware";
// import { persist } from "zustand/middleware";

export interface BetaPopupStore {
  popupClosed: boolean;

  setPopupClosed(popupClosed: boolean): void;
}

export const useBetaPopupStore = create<BetaPopupStore>()(
  persist(
    (set) => ({
      popupClosed: false,

      setPopupClosed: (popupClosed) => {
        set({ popupClosed });
      },
    }),
    { name: "kite-beta-podpup" }
  )
);
