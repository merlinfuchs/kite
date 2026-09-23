import { beforeEach, expect, test } from "vitest";
import type {
  MessageComponentContainer,
  MessageComponentSection,
} from "./schema";
import { parseMessageData } from "./schemaRestore";
import {
  COMPONENTS_V2_FLAG,
  type ActionRowNode,
  type ButtonNode,
  type ContainerNode,
  createDocumentStore,
  type DocumentStore,
  type MessageNode,
  type NodeId,
  type SectionNode,
  type TextDisplayNode,
} from "./document";
import { toMessage } from "./documentConvert";

const containerMessage = parseMessageData({
  content: "",
  flags: COMPONENTS_V2_FLAG,
  components: [
    {
      type: 17,
      components: [
        {
          type: 9,
          components: [{ type: 10, content: "Section text" }],
          accessory: {
            type: 2,
            style: 1,
            label: "Accessory",
            flow_source_id: "accessory-flow",
          },
        },
        {
          type: 1,
          components: [
            {
              type: 2,
              style: 1,
              label: "Click",
              flow_source_id: "button-flow",
            },
          ],
        },
      ],
    },
  ],
});

let store: ReturnType<typeof createDocumentStore>;

function state(): DocumentStore {
  return store.getState();
}

function root(): MessageNode {
  return state().nodes[state().rootId] as MessageNode;
}

function nodeIdsOfType(type: string): NodeId[] {
  return Object.values(state().nodes)
    .filter((node) => node.type === type)
    .map((node) => node.id);
}

beforeEach(() => {
  store = createDocumentStore();
  state().replaceAll(containerMessage);
});

test("insert links the node to its parent", () => {
  const embedId = state().insert(state().rootId, "embeds", "end", {
    type: "embed",
    description: "Added",
  });

  expect(root().embedIds).toEqual([embedId]);
  expect(state().nodes[embedId].parentId).toBe(state().rootId);
  expect(toMessage(state()).message.embeds[0].description).toBe("Added");
});

test("insert at an index puts the node there", () => {
  const first = state().insert(state().rootId, "embeds", "end", {
    type: "embed",
    description: "First",
  });
  const second = state().insert(state().rootId, "embeds", 0, {
    type: "embed",
    description: "Second",
  });

  expect(root().embedIds).toEqual([second, first]);
});

test("inserting a button gives it its own flow source", () => {
  const [rowId] = nodeIdsOfType("actionRow");
  const buttonId = state().insert(rowId, "components", "end", {
    type: "button",
    style: 1,
    label: "New",
  });

  const button = state().nodes[buttonId] as ButtonNode;
  expect(button.flow_source_id).toBeTruthy();
  expect(button.flow_source_id).not.toBe("button-flow");
});

test("inserting into the accessory slot replaces the old accessory", () => {
  const [sectionId] = nodeIdsOfType("section");
  const oldAccessoryId = (state().nodes[sectionId] as SectionNode).accessoryId;

  const thumbnailId = state().insert(sectionId, "accessory", "end", {
    type: "thumbnail",
    media: { url: "https://kite.onl/thumb.png" },
  });

  expect((state().nodes[sectionId] as SectionNode).accessoryId).toBe(
    thumbnailId
  );
  expect(state().nodes[oldAccessoryId as NodeId]).toBeUndefined();
});

test("remove drops every descendant", () => {
  const [containerId] = nodeIdsOfType("container");

  state().remove(containerId);

  expect(Object.keys(state().nodes)).toEqual([state().rootId]);
  expect(root().componentIds).toEqual([]);
});

test("remove ignores the root", () => {
  const before = Object.keys(state().nodes).length;
  state().remove(state().rootId);
  expect(Object.keys(state().nodes)).toHaveLength(before);
});

test("move reorders within the parent and stops at the edges", () => {
  const [containerId] = nodeIdsOfType("container");
  const [sectionId, rowId] = (state().nodes[containerId] as ContainerNode)
    .childIds;

  state().move(rowId, -1);
  expect((state().nodes[containerId] as ContainerNode).childIds).toEqual([
    rowId,
    sectionId,
  ]);

  state().move(rowId, -1);
  expect((state().nodes[containerId] as ContainerNode).childIds).toEqual([
    rowId,
    sectionId,
  ]);
});

test("duplicate deep copies with fresh ids", () => {
  const [containerId] = nodeIdsOfType("container");
  const before = new Set(Object.keys(state().nodes));

  const copyId = state().duplicate(containerId);

  expect(before.has(copyId)).toBe(false);
  expect(root().componentIds).toEqual([containerId, copyId]);

  const copies = Object.keys(state().nodes).filter((id) => !before.has(id));
  expect(copies).toHaveLength(before.size - 1); // everything but the root
});

test("duplicate gives copied buttons new flow sources and discord ids", () => {
  const [rowId] = nodeIdsOfType("actionRow");
  const [buttonId] = (state().nodes[rowId] as ActionRowNode).childIds;
  const original = state().nodes[buttonId] as ButtonNode;

  const copyId = state().duplicate(buttonId);
  const copy = state().nodes[copyId] as ButtonNode;

  expect(copy.flow_source_id).not.toBe(original.flow_source_id);
  expect(copy.discordId).not.toBe(original.discordId);
});

test("duplicate leaves accessories alone", () => {
  const [sectionId] = nodeIdsOfType("section");
  const accessoryId = (state().nodes[sectionId] as SectionNode)
    .accessoryId as NodeId;
  const before = Object.keys(state().nodes).length;

  expect(state().duplicate(accessoryId)).toBe(accessoryId);
  expect(Object.keys(state().nodes)).toHaveLength(before);
});

test("moving a button keeps the discord id that flows address it by", () => {
  const [rowId] = nodeIdsOfType("actionRow");
  const [buttonId] = (state().nodes[rowId] as ActionRowNode).childIds;
  const before = (state().nodes[buttonId] as ButtonNode).discordId;

  state().insert(rowId, "components", 0, {
    type: "button",
    style: 2,
    label: "Before",
  });

  expect((state().nodes[buttonId] as ButtonNode).discordId).toBe(before);
});

test("update patches a single node", () => {
  const [textId] = nodeIdsOfType("textDisplay");
  state().update<TextDisplayNode>(textId, { content: "Changed" });

  const container = toMessage(state()).message
    .components[0] as MessageComponentContainer;
  const section = container.components[0] as MessageComponentSection;
  expect(section.components[0].content).toBe("Changed");
});

test("setComponentsV2 toggles the flag", () => {
  state().setComponentsV2(false);
  expect(root().flags ?? 0).toBe(0);

  state().setComponentsV2(true);
  expect((root().flags ?? 0) & COMPONENTS_V2_FLAG).toBe(COMPONENTS_V2_FLAG);
});

test("replaceAll swaps the whole document", () => {
  state().replaceAll(parseMessageData({ content: "Replaced" }));

  expect(Object.keys(state().nodes)).toEqual([state().rootId]);
  expect(toMessage(state()).message.content).toBe("Replaced");
});
