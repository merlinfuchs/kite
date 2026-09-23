import { describe, expect, test } from "vitest";
import { messageSchema } from "./schema";
import { parseMessageData } from "./schemaRestore";
import type { DocumentData, Node, NodeId } from "./document";
import {
  RestoredMessage,
  childIds,
  childSlots,
  fromMessage,
  toMessage,
} from "./documentConvert";

const parse = (raw: unknown): RestoredMessage => parseMessageData(raw);

const emptyMessage = parse({ content: "Hello" });

const embedMessage = parse({
  content: "",
  attachments: [{ asset_id: "asset-1" }],
  embeds: [
    {
      title: "First",
      description: "Description",
      url: "https://kite.onl",
      timestamp: "2026-01-01T00:00:00.000Z",
      color: 2326507,
      author: { name: "Author", url: "https://kite.onl" },
      footer: { text: "Footer" },
      image: { url: "https://kite.onl/image.png" },
      thumbnail: { url: "https://kite.onl/thumb.png" },
      fields: [
        { name: "A", value: "1", inline: true },
        { name: "B", value: "2" },
      ],
    },
    {
      description: "Second",
      fields: [{ name: "C", value: "3" }],
    },
  ],
});

const actionRowMessage = parse({
  content: "Components",
  components: [
    {
      type: 1,
      components: [
        { type: 2, style: 1, label: "Click", flow_source_id: "flow-1" },
        { type: 2, style: 5, label: "Link", url: "https://kite.onl" },
      ],
    },
    {
      type: 1,
      components: [
        {
          type: 3,
          placeholder: "Pick a color",
          min_values: 1,
          max_values: 2,
          flow_source_id: "flow-select",
          options: [
            { label: "Red", value: "red", description: "warm" },
            { label: "Blue" },
          ],
        },
      ],
    },
  ],
});

const componentsV2Message = parse({
  content: "",
  flags: 1 << 15,
  components: [
    {
      type: 17,
      accent_color: 2326507,
      components: [
        {
          type: 9,
          components: [{ type: 10, content: "Section text" }],
          accessory: {
            type: 11,
            media: { url: "https://kite.onl/thumb.png" },
            description: "thumb",
          },
        },
        { type: 10, content: "Standalone text" },
        {
          type: 12,
          items: [
            { media: { url: "https://kite.onl/one.png" } },
            { media: { url: "https://kite.onl/two.png" }, spoiler: true },
          ],
        },
        { type: 13, file: { url: "https://kite.onl/file.txt" } },
        { type: 14, divider: false, spacing: 2 },
        {
          type: 1,
          components: [{ type: 2, style: 2, label: "In container" }],
        },
        {
          type: 9,
          components: [{ type: 10, content: "With a button accessory" }],
          accessory: {
            type: 2,
            style: 1,
            label: "Accessory",
            flow_source_id: "flow-2",
          },
        },
      ],
    },
  ],
});

const fixtures: [string, RestoredMessage][] = [
  ["empty message", emptyMessage],
  ["embeds with fields", embedMessage],
  ["v1 action rows", actionRowMessage],
  ["v2 container", componentsV2Message],
];

/** Path of every array element in the payload, as a zod-style path. */
function elementPaths(value: unknown, path = "", inArray = false): string[] {
  if (Array.isArray(value)) {
    return value.flatMap((entry, i) =>
      elementPaths(entry, `${path}.${i}`, true)
    );
  }
  if (value && typeof value === "object") {
    const paths = inArray ? [path] : [];
    for (const [key, entry] of Object.entries(value)) {
      // attachments are plain values on the message node, not nodes
      if (key === "attachments") continue;
      paths.push(...elementPaths(entry, path ? `${path}.${key}` : key));
    }
    return paths;
  }
  return [];
}

function allNodeIds(state: DocumentData): NodeId[] {
  return Object.keys(state.nodes);
}

function subtree(state: DocumentData, id: NodeId): NodeId[] {
  const node = state.nodes[id] as Node;
  return [
    id,
    ...childSlots(node).flatMap((slot) =>
      childIds(node, slot).flatMap((childId) => subtree(state, childId))
    ),
  ];
}

describe("round trip", () => {
  for (const [name, message] of fixtures) {
    test(name, () => {
      expect(toMessage(fromMessage(message)).message).toEqual(message);
    });
  }
});

describe("round-tripped messages pass validation", () => {
  for (const [name, message] of fixtures) {
    test(name, () => {
      const res = messageSchema.safeParse(
        toMessage(fromMessage(message)).message
      );
      expect(res.error?.issues).toBeUndefined();
    });
  }
});

describe("pathToId", () => {
  for (const [name, message] of fixtures) {
    test(name, () => {
      const {
        message: payload,
        pathToId,
        idToPath,
      } = toMessage(fromMessage(message));

      for (const path of elementPaths(payload)) {
        expect(pathToId.has(path), `missing path ${path}`).toBe(true);
      }
      for (const [path, id] of pathToId) {
        expect(idToPath.get(id)).toBe(path);
      }
    });
  }
});

test("toMessage is memoized on the nodes object", () => {
  const state = fromMessage(embedMessage);
  expect(toMessage(state)).toBe(toMessage(state));
});

test("parents point at their children", () => {
  const state = fromMessage(componentsV2Message);

  for (const id of allNodeIds(state)) {
    const node = state.nodes[id];
    for (const slot of childSlots(node)) {
      for (const childId of childIds(node, slot)) {
        expect(state.nodes[childId].parentId).toBe(id);
      }
    }
  }
});

test("a message with no embeds or components has only the root", () => {
  const state = fromMessage(emptyMessage);
  expect(allNodeIds(state)).toEqual([state.rootId]);
});

test("subtree ids are unique", () => {
  const state = fromMessage(componentsV2Message);
  const ids = subtree(state, state.rootId);
  expect(new Set(ids).size).toBe(ids.length);
  expect(ids.length).toBe(allNodeIds(state).length);
});

test("rows saved without a type load as action rows", () => {
  const legacy = parse({
    content: "",
    components: [
      { id: 1, components: [{ id: 2, type: 2, style: 1, label: "Old" }] },
    ],
  });

  const { message } = toMessage(fromMessage(legacy));
  expect(message.components[0].type).toBe(1);
});

test("a section without an accessory is reported instead of faked", () => {
  const broken = parse({
    flags: 1 << 15,
    components: [{ type: 9, components: [{ type: 10, content: "text" }] }],
  });

  const { message } = toMessage(fromMessage(broken));
  const res = messageSchema.safeParse(message);

  expect(res.success).toBe(false);
  expect(res.error?.issues.map((i) => i.path.join("."))).toContain(
    "components.0.accessory"
  );
});

describe("select menu validation", () => {
  const withMenu = (menu: Record<string, unknown>, extra: unknown[] = []) =>
    messageSchema.safeParse(
      toMessage(
        fromMessage(
          parse({
            content: "",
            components: [
              {
                type: 1,
                components: [
                  {
                    type: 3,
                    options: [{ label: "A" }, { label: "B" }],
                    ...menu,
                  },
                  ...extra,
                ],
              },
            ],
          })
        )
      ).message
    );

  const paths = (res: ReturnType<typeof withMenu>) =>
    res.error?.issues.map((i) => i.path.join(".")) ?? [];

  test("a valid menu passes", () => {
    expect(paths(withMenu({ min_values: 1, max_values: 2 }))).toEqual([]);
  });

  test("max can't be below min or above the option count", () => {
    expect(paths(withMenu({ min_values: 2, max_values: 1 }))).toContain(
      "components.0.components.0.max_values"
    );
    expect(paths(withMenu({ max_values: 3 }))).toContain(
      "components.0.components.0.max_values"
    );
  });

  test("option values have to be unique, falling back to the label", () => {
    expect(
      paths(withMenu({ options: [{ label: "A" }, { label: "B", value: "A" }] }))
    ).toContain("components.0.components.0.options.1.value");
  });

  test("a select menu has to be alone in its row", () => {
    expect(
      paths(withMenu({}, [{ type: 2, style: 1, label: "Click" }]))
    ).toContain("components.0.components");
  });
});
