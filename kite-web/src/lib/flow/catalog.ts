import { ZodSchema } from "zod";
import { zodToJsonSchema } from "zod-to-json-schema";
import { isNodeTypeAvailable } from "./categories";
import { flowContextTypes } from "./context";
import { isTemplated } from "./dataSchema";
import { createNode, nodeTypes } from "./nodes";

// A machine readable description of every block, generated from the schemas
// the editor uses. The service embeds it as kite-service/pkg/flow/catalog.json,
// which is regenerated with `pnpm test -u`.
export function buildFlowCatalog() {
  return {
    nodes: Object.fromEntries(
      Object.entries(nodeTypes).map(([type, values]) => [
        type,
        {
          title: values.defaultTitle,
          description: values.defaultDescription,
          contexts: flowContextTypes.filter((c) =>
            isNodeTypeAvailable(type, c)
          ),
          outputs: values.outputs ?? ["default"],
          // Blocks like conditions and loops are created together with the
          // blocks they own and are connected to them with fixed edges.
          owned_children: createNode(type, { x: 0, y: 0 })[0]
            .slice(1)
            .map((n) => n.type),
          fixed: !!values.fixed,
          data_schema: values.dataSchema
            ? toJsonSchema(values.dataSchema)
            : null,
          result_schema: values.resultSchema
            ? toJsonSchema(values.resultSchema)
            : null,
        },
      ])
    ),
  };
}

export function toJsonSchema(schema: ZodSchema) {
  const { $schema, ...res } = zodToJsonSchema(schema, {
    $refStrategy: "none",
    postProcess: (json, def) => {
      if (!json) return json;

      // Some defaults are generated IDs, which would make the output random.
      const { default: _, ...rest } = json as Record<string, unknown>;
      return isTemplated(def) ? { ...rest, "x-templated": true } : rest;
    },
  });
  return res;
}
