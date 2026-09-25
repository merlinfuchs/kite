import { describe, expect, it } from "vitest";
import { buildFlowCatalog } from "./catalog";
import { getNodeValues, nodeTypes } from "./nodes";
import { getTemplates } from "./templates";

const catalog = buildFlowCatalog();

type JsonSchema = {
  description?: string;
  properties?: Record<string, JsonSchema>;
  items?: JsonSchema;
  anyOf?: JsonSchema[];
};

// Returns the paths of properties without a description.
function undescribed(schema: JsonSchema, path: string): string[] {
  const res: string[] = [];
  for (const [key, prop] of Object.entries(schema.properties ?? {})) {
    const propPath = `${path}.${key}`;
    if (!prop.description) res.push(propPath);
    res.push(...undescribed(prop, propPath));
  }
  if (schema.items) res.push(...undescribed(schema.items, `${path}[]`));
  for (const s of schema.anyOf ?? []) res.push(...undescribed(s, path));
  return res;
}

describe("flow catalog", () => {
  it("matches the file embedded in the service", async () => {
    await expect(JSON.stringify(catalog, null, 2) + "\n").toMatchFileSnapshot(
      "../../../../kite-service/pkg/flow/catalog.json"
    );
  });

  it("describes every field", () => {
    const missing = Object.entries(catalog.nodes).flatMap(([type, node]) =>
      node.data_schema ? undescribed(node.data_schema as JsonSchema, type) : []
    );
    expect(missing).toEqual([]);
  });

  it("only lets blocks with a temporary variable field set one", () => {
    for (const [type, values] of Object.entries(nodeTypes)) {
      const properties =
        (catalog.nodes[type].data_schema as JsonSchema | null)?.properties ??
        {};
      expect("temporary_name" in properties, `${type} temporary_name`).toBe(
        values.dataFields.includes("temporary_name")
      );
    }
  });
});

describe("data schemas", () => {
  it("accept the blocks of the built-in templates", () => {
    const flows = getTemplates().flatMap((template) => {
      // Inputs are things like channel IDs, so a numeric string fits all.
      const inputs = Object.fromEntries(
        template.inputs.map((input) => [input.key, "123"])
      );
      return [
        ...template.commands.map((c) => c.flowSource(inputs)),
        ...template.eventListeners.map((l) => l.flowSource(inputs)),
      ];
    });

    for (const node of flows.flatMap((f) => f.nodes)) {
      const res = getNodeValues(node.type!).dataSchema?.safeParse(node.data);
      expect(res?.error?.issues ?? [], node.type).toEqual([]);
    }
  });
});
