import debounce from "just-debounce-it";
import { temporal } from "zundo";
import { create } from "zustand";
import { immer } from "zustand/middleware/immer";
import { getUniqueId } from "@/lib/utils";
import {
  COMPONENTS_V2_FLAG,
  MAX_COMPONENTS_V2,
  hasComponentsV2Flag,
  type EmbedAuthor,
  type EmbedFooter,
  type EmbedImage,
  type EmbedProvider,
  type EmbedThumbnail,
  type Emoji,
  type Message,
  type MessageAttachment,
  type MessageComponentButtonStyle,
  type UnfurledMediaItem,
} from "./schema";
import {
  type ChildSlot,
  type RestoredMessage,
  childIds,
  childSlots,
  fromMessage,
  setChildIds,
} from "./documentConvert";

export type NodeId = string;

/**
 * `discordId` is the numeric `id` that lives on embeds, fields and components in
 * the payload. It's kept separate from `NodeId` so the wire format can change
 * without touching how the editor addresses nodes. Flows address buttons by it
 * through their `component_<id>` handles, so it survives edits and moves.
 */
interface BaseNode {
  id: NodeId;
  parentId: NodeId | null;
  discordId: number;
}

export type MessageNode = BaseNode & {
  type: "message";
  content: string;
  username?: string;
  avatar_url?: string;
  tts: boolean;
  thread_name?: string;
  flags?: number;
  allowed_mentions?: Message["allowed_mentions"];
  attachments: MessageAttachment[];
  embedIds: NodeId[];
  componentIds: NodeId[];
};

export type EmbedNode = BaseNode & {
  type: "embed";
  title?: string;
  description?: string;
  url?: string;
  timestamp?: string;
  color?: number;
  footer?: EmbedFooter;
  author?: EmbedAuthor;
  provider?: EmbedProvider;
  image?: EmbedImage;
  thumbnail?: EmbedThumbnail;
  fieldIds: NodeId[];
};

export type EmbedFieldNode = BaseNode & {
  type: "embedField";
  name: string;
  value: string;
  inline?: boolean;
};

export type ActionRowNode = BaseNode & {
  type: "actionRow";
  childIds: NodeId[];
};

export type ButtonNode = BaseNode & {
  type: "button";
  style: MessageComponentButtonStyle;
  label: string;
  emoji?: Emoji;
  url?: string;
  disabled?: boolean;
  flow_source_id: string;
};

export type SelectMenuNode = BaseNode & {
  type: "selectMenu";
  placeholder?: string;
  min_values?: number;
  max_values?: number;
  disabled?: boolean;
  optionIds: NodeId[];
  flow_source_id: string;
};

export type SelectOptionNode = BaseNode & {
  type: "selectOption";
  label: string;
  value?: string;
  description?: string;
  emoji?: Emoji;
};

export type ContainerNode = BaseNode & {
  type: "container";
  accent_color?: number;
  spoiler?: boolean;
  childIds: NodeId[];
};

export type SectionNode = BaseNode & {
  type: "section";
  childIds: NodeId[];
  accessoryId: NodeId | null;
};

export type TextDisplayNode = BaseNode & {
  type: "textDisplay";
  content: string;
};

export type ThumbnailNode = BaseNode & {
  type: "thumbnail";
  media: UnfurledMediaItem;
  description?: string;
  spoiler?: boolean;
};

export type MediaGalleryNode = BaseNode & {
  type: "mediaGallery";
  itemIds: NodeId[];
};

export type MediaGalleryItemNode = BaseNode & {
  type: "mediaGalleryItem";
  media: UnfurledMediaItem;
  description?: string;
  spoiler?: boolean;
};

export type FileNode = BaseNode & {
  type: "file";
  file: UnfurledMediaItem;
  spoiler?: boolean;
};

export type SeparatorNode = BaseNode & {
  type: "separator";
  divider: boolean;
  spacing: 1 | 2;
};

export type Node =
  | MessageNode
  | EmbedNode
  | EmbedFieldNode
  | ActionRowNode
  | ButtonNode
  | SelectMenuNode
  | SelectOptionNode
  | ContainerNode
  | SectionNode
  | TextDisplayNode
  | ThumbnailNode
  | MediaGalleryNode
  | MediaGalleryItemNode
  | FileNode
  | SeparatorNode;

export type NodeType = Node["type"];

type Primitive = string | number | boolean | bigint | symbol;

/** Bookkeeping that is never addressed by validation. */
type Bookkeeping = "id" | "parentId" | "discordId" | "type";

/** The id lists standing in for the payload's child arrays. */
type ChildRefs =
  | "embedIds"
  | "componentIds"
  | "fieldIds"
  | "childIds"
  | "optionIds"
  | "itemIds"
  | "accessoryId";

/**
 * The fields of a node as zod issue paths, e.g. `"author.name"`. Keeps a typo
 * from compiling into a lookup that silently matches nothing. Two levels deep,
 * which is as far as the message schema nests inside a node.
 *
 * Children are left out: the payload calls them `fields`, the node holds
 * `fieldIds`, and neither name is a path an issue ever sits at. `slotScope`
 * addresses those.
 */
export type FieldPath<T> = {
  [K in keyof Omit<T, Bookkeeping | ChildRefs> & string]: NonNullable<
    T[K]
  > extends Primitive | readonly unknown[]
    ? K
    : K | `${K}.${keyof NonNullable<T[K]> & string}`;
}[keyof Omit<T, Bookkeeping | ChildRefs> & string];

type DistributiveOmit<T, K extends keyof never> = T extends unknown
  ? Omit<T, K>
  : never;

/** Everything `insert` fills in itself. */
type DerivedKeys = Exclude<Bookkeeping, "type"> | ChildRefs | "flow_source_id";

export type NewNode = DistributiveOmit<Node, DerivedKeys>;

export interface DocumentData {
  nodes: Record<NodeId, Node>;
  rootId: NodeId;
}

export interface DocumentStore extends DocumentData {
  update<T extends Node>(
    id: NodeId,
    patch: Partial<Omit<T, "id" | "type" | "parentId">>
  ): void;
  insert(
    parentId: NodeId,
    slot: ChildSlot,
    index: number | "end",
    node: NewNode
  ): NodeId;
  remove(id: NodeId): void;
  removeChildren(parentId: NodeId, slot: ChildSlot): void;
  move(id: NodeId, delta: -1 | 1): void;
  duplicate(id: NodeId): NodeId;
  replaceAll(message: RestoredMessage): void;
  clear(): void;
  setComponentsV2(enabled: boolean): void;
}

export const emptyMessage: RestoredMessage = {
  content: "",
  tts: false,
  attachments: [],
  embeds: [],
  components: [],
};

/** What enabling components v2 replaces the message with. */
const emptyComponentsV2Message: RestoredMessage = {
  ...emptyMessage,
  flags: COMPONENTS_V2_FLAG,
};

/**
 * How many children a slot holds, as the message schema enforces it. Kept here
 * so the counter, the add button and the duplicate button agree.
 */
const SLOT_LIMITS: Record<string, number> = {
  "message.embeds": 10,
  "message.components": 5,
  "embed.fields": 25,
  "actionRow.components": 5,
  "section.components": 3,
  "container.components": 10,
  "selectMenu.options": 25,
  "mediaGallery.items": 10,
};

export function slotLimit(
  parentType: NodeType,
  slot: ChildSlot,
  componentsV2 = false,
  childType?: NodeType
): number {
  // A select menu fills its row on its own.
  if (parentType === "actionRow" && childType === "selectMenu") {
    return 1;
  }
  if (componentsV2 && parentType === "message" && slot === "components") {
    // Components v2 only caps the total component count, which the schema checks.
    return MAX_COMPONENTS_V2;
  }
  return SLOT_LIMITS[`${parentType}.${slot}`] ?? 1;
}

export function isComponentsV2(state: DocumentData): boolean {
  const root = state.nodes[state.rootId];
  return root?.type === "message" && hasComponentsV2Flag(root.flags);
}

function freshId(nodes: Record<NodeId, Node>): NodeId {
  let id = getUniqueId().toString();
  while (nodes[id]) {
    id = getUniqueId().toString();
  }
  return id;
}

/** The empty child arrays that a node of this type needs. */
function emptyChildren(type: NodeType) {
  switch (type) {
    case "message":
      return { embedIds: [], componentIds: [] };
    case "embed":
      return { fieldIds: [] };
    case "actionRow":
    case "container":
      return { childIds: [] };
    case "section":
      return { childIds: [], accessoryId: null };
    case "selectMenu":
      return { optionIds: [] };
    case "mediaGallery":
      return { itemIds: [] };
    default:
      return {};
  }
}

function usesFlowSource(node: { type: NodeType }): boolean {
  return node.type === "button" || node.type === "selectMenu";
}

/** Collects `id` and every node below it, deepest last. */
function descendants(nodes: Record<NodeId, Node>, id: NodeId): NodeId[] {
  const node = nodes[id];
  if (!node) return [];

  const found = [id];
  for (const slot of childSlots(node)) {
    for (const childId of childIds(node, slot)) {
      found.push(...descendants(nodes, childId));
    }
  }
  return found;
}

export function slotOfChild(parent: Node, childId: NodeId): ChildSlot | null {
  for (const slot of childSlots(parent)) {
    if (childIds(parent, slot).includes(childId)) return slot;
  }
  return null;
}

export const createDocumentStore = (
  initialMessage: RestoredMessage = emptyMessage
) =>
  create<DocumentStore>()(
    immer(
      temporal(
        (set, get) => ({
          ...fromMessage(initialMessage),

          update: (id, patch) =>
            set((state) => {
              const node = state.nodes[id];
              if (!node) return;
              Object.assign(node, patch);
            }),

          insert: (parentId, slot, index, node) => {
            let id = "";
            set((state) => {
              id = insertNode(state, parentId, slot, index, node);
            });
            return id;
          },

          remove: (id) =>
            set((state) => {
              if (id === state.rootId) return;
              removeSubtree(state, id);
            }),

          removeChildren: (parentId, slot) =>
            set((state) => {
              const parent = state.nodes[parentId];
              if (!parent) return;

              for (const childId of [...childIds(parent, slot)]) {
                removeSubtree(state, childId);
              }
            }),

          move: (id, delta) =>
            set((state) => {
              const node = state.nodes[id];
              if (!node?.parentId) return;

              const parent = state.nodes[node.parentId];
              if (!parent) return;

              const slot = slotOfChild(parent, id);
              if (!slot) return;

              const ids = [...childIds(parent, slot)];
              const index = ids.indexOf(id);
              const target = index + delta;
              if (target < 0 || target >= ids.length) return;

              ids.splice(index, 1);
              ids.splice(target, 0, id);
              setChildIds(parent, slot, ids);
            }),

          duplicate: (id) => {
            const copyId = freshId(get().nodes);
            let copied = false;

            set((state) => {
              const node = state.nodes[id];
              if (!node?.parentId) return;

              const parent = state.nodes[node.parentId];
              if (!parent) return;

              const slot = slotOfChild(parent, id);
              // An accessory slot only holds one node, nothing to duplicate into.
              if (!slot || slot === "accessory") return;

              copySubtree(state, id, node.parentId, copyId);
              copied = true;

              const ids = [...childIds(parent, slot)];
              ids.splice(ids.indexOf(id) + 1, 0, copyId);
              setChildIds(parent, slot, ids);
            });

            // Nothing was copied for a root or accessory node, so the caller
            // gets the node it asked about rather than a dangling id.
            return copied ? copyId : id;
          },

          replaceAll: (message) => set(fromMessage(message)),

          clear: () => set(fromMessage(emptyMessage)),

          // The two modes cannot hold each other's content, so the toggle
          // replaces the message rather than editing it.
          setComponentsV2: (enabled) =>
            set(fromMessage(enabled ? emptyComponentsV2Message : emptyMessage)),
        }),
        {
          limit: 10,
          handleSet: (handleSet) => debounce(handleSet, 1000, true),
          partialize: (state) => ({
            nodes: state.nodes,
            rootId: state.rootId,
          }),
        }
      )
    )
  );

/**
 * Children a node can't be valid without, created along with it so every way
 * of adding one (and its undo step) gets them.
 */
const REQUIRED_CHILDREN: Partial<
  Record<NodeType, { slot: ChildSlot; node: NewNode }[]>
> = {
  selectMenu: [
    { slot: "options", node: { type: "selectOption", label: "", value: "" } },
  ],
  section: [
    { slot: "components", node: { type: "textDisplay", content: "" } },
    { slot: "accessory", node: { type: "thumbnail", media: { url: "" } } },
  ],
};

/** Creates `node` under `parentId` inside a draft, returning its id. */
function insertNode(
  state: DocumentData,
  parentId: NodeId,
  slot: ChildSlot,
  index: number | "end",
  node: NewNode
): NodeId {
  const parent = state.nodes[parentId];
  if (!parent) return "";

  // An accessory slot holds a single node, so inserting replaces.
  if (slot === "accessory") {
    for (const existing of childIds(parent, slot)) {
      removeSubtree(state, existing);
    }
  }

  const id = freshId(state.nodes);
  const created = {
    ...node,
    ...emptyChildren(node.type),
    id,
    parentId,
    discordId: getUniqueId(),
  } as Node;

  if (usesFlowSource(created)) {
    (created as ButtonNode | SelectMenuNode).flow_source_id =
      getUniqueId().toString();
  }

  state.nodes[id] = created;

  const ids = [...childIds(parent, slot)];
  ids.splice(index === "end" ? ids.length : index, 0, id);
  setChildIds(parent, slot, ids);

  for (const child of REQUIRED_CHILDREN[node.type] ?? []) {
    insertNode(state, id, child.slot, "end", child.node);
  }

  return id;
}

/** Detaches `id` from its parent and drops it and everything below it. */
function removeSubtree(state: DocumentData, id: NodeId) {
  const node = state.nodes[id];
  if (!node) return;

  if (node.parentId) {
    const parent = state.nodes[node.parentId];
    const slot = parent && slotOfChild(parent, id);
    if (parent && slot) {
      setChildIds(
        parent,
        slot,
        childIds(parent, slot).filter((childId) => childId !== id)
      );
    }
  }

  for (const removedId of descendants(state.nodes, id)) {
    delete state.nodes[removedId];
  }
}

/**
 * Deep copies `id` under `parentId`, giving every node fresh ids. Copied buttons
 * get a new flow source, so the copy starts without a flow like it did before.
 */
function copySubtree(
  state: DocumentData,
  id: NodeId,
  parentId: NodeId | null,
  copyId: NodeId
): NodeId {
  const node = state.nodes[id];
  const copy = {
    ...node,
    id: copyId,
    parentId,
    discordId: getUniqueId(),
  } as Node;

  if (usesFlowSource(copy)) {
    (copy as ButtonNode | SelectMenuNode).flow_source_id =
      getUniqueId().toString();
  }

  state.nodes[copyId] = copy;

  for (const slot of childSlots(node)) {
    const copies = childIds(node, slot).map((childId) =>
      copySubtree(state, childId, copyId, freshId(state.nodes))
    );
    setChildIds(copy, slot, copies);
  }

  return copyId;
}
