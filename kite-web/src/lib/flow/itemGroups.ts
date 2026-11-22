// An item's group in the command, event listener and message lists and in
// pickers: its folder, or else the template it was added with.
export type ItemGroup = {
  name: string;
  // Whether it's the template the item was added with, which has no folder.
  template: boolean;
};

export type GroupableItem = {
  folder?: string | null;
  bundle_install_id?: string | null;
};

export function itemGroup(
  item: GroupableItem,
  installs: ({ id: string; name: string } | null | undefined)[] | undefined
): ItemGroup | undefined {
  if (item.folder) return { name: item.folder, template: false };
  if (!item.bundle_install_id) return undefined;
  const install = installs?.find((i) => i?.id === item.bundle_install_id);
  return install ? { name: install.name, template: true } : undefined;
}

// sortByGroup puts the items without a group first, then the groups A to Z,
// keeping the order of the items within each.
export function sortByGroup<T>(
  items: T[],
  groupOf: (item: T) => string | undefined
): T[] {
  return items
    .map((item, index) => ({ item, index, group: groupOf(item) }))
    .sort((a, b) => {
      if (a.group !== b.group) {
        if (a.group === undefined) return -1;
        if (b.group === undefined) return 1;
        return a.group.localeCompare(b.group);
      }
      return a.index - b.index;
    })
    .map(({ item }) => item);
}
