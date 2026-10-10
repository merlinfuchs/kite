import { ZodSchema } from "zod";
import { zodToJsonSchema } from "zod-to-json-schema";
import { getBlockDefinition } from "../blocks";
import { isNodeTypeAvailable, nodeCategories } from "./categories";
import { flowContextTypes } from "./context";
import { isTemplated, isUserPicked } from "./dataSchema";
import { getNodeOutputs, getOwnedChildTypes, nodeTypes } from "./nodes";

// A machine readable description of every block, generated from the schemas
// the editor uses. The service embeds it as kite-service/pkg/flow/catalog.json,
// which is regenerated with `pnpm test -u`.
export function buildFlowCatalog() {
  // Grouped like the block explorer, so e.g. all role blocks are together.
  const explorerOrder = Object.values(nodeCategories)
    .flat()
    .flatMap((s) => s.nodeTypes);
  const position = (type: string) => {
    const i = explorerOrder.indexOf(type);
    return i === -1 ? explorerOrder.length : i;
  };

  return {
    nodes: Object.fromEntries(
      Object.entries(nodeTypes)
        .sort(([a], [b]) => position(a) - position(b))
        .map(([type, values]) => [
          type,
          {
            title: values.defaultTitle,
            description: values.defaultDescription,
            contexts: flowContextTypes.filter((c) =>
              isNodeTypeAvailable(type, c)
            ),
            outputs: getNodeOutputs({ type, data: {} }),
            // Blocks like conditions and loops are created together with the
            // blocks they own and are connected to them with fixed edges.
            owned_children: getOwnedChildTypes(type),
            fixed: !!values.fixed,
            data_schema: values.dataSchema
              ? withoutHiddenFields(type, toJsonSchema(values.dataSchema))
              : null,
            result_schema: values.resultSchema
              ? toJsonSchema(values.resultSchema)
              : null,
          },
        ])
    ),
  };
}

function withoutHiddenFields(
  type: string,
  schema: ReturnType<typeof toJsonSchema>
) {
  const { properties } = schema as { properties?: Record<string, unknown> };
  for (const field of getBlockDefinition(type)?.fields ?? []) {
    if (field.hidden_from_ai) delete properties?.[field.name];
  }
  return schema;
}

export function toJsonSchema(schema: ZodSchema) {
  const { $schema, ...res } = zodToJsonSchema(schema, {
    $refStrategy: "none",
    postProcess: (json, def) => {
      if (!json) return json;

      // Some defaults are generated IDs, which would make the output random.
      const { default: _, ...rest } = json as Record<string, unknown>;
      return {
        ...rest,
        ...(isTemplated(def) && { "x-templated": true }),
        ...(isUserPicked(def) && { "x-user-picked": true }),
      } as typeof json;
    },
  });
  return res;
}
