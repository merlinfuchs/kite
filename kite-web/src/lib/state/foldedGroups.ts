import { create } from "zustand";
import { persist } from "zustand/middleware";

interface FoldedGroupsStore {
  // Keys of the folded groups of lists, see foldedGroupKey.
  folded: Record<string, true>;
  toggle: (key: string) => void;
}

// The groups of the command, event listener and message lists someone folded,
// remembered on their device.
export const useFoldedGroupsStore = create<FoldedGroupsStore>()(
  persist(
    (set) => ({
      folded: {},
      toggle: (key) =>
        set(({ folded }) => {
          const next = { ...folded };
          if (next[key]) delete next[key];
          else next[key] = true;
          return { folded: next };
        }),
    }),
    { name: "kite-folded-groups", version: 0 }
  )
);

export function foldedGroupKey(list: string, group: string) {
  return `${list}:${group}`;
}
