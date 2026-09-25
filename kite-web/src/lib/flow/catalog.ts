import { ZodSchema } from "zod";
import { zodToJsonSchema } from "zod-to-json-schema";
import { isNodeTypeAvailable } from "./categories";
import { flowContextTypes } from "./context";
import { isTemplated, isUserPicked } from "./dataSchema";
import { getNodeOutputs, getOwnedChildTypes, nodeTypes } from "./nodes";

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
          outputs: getNodeOutputs({ type, data: {} }),
          // Blocks like conditions and loops are created together with the
          // blocks they own and are connected to them with fixed edges.
          owned_children: getOwnedChildTypes(type),
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
      return {
        ...rest,
        ...(isTemplated(def) && { "x-templated": true }),
        ...(isUserPicked(def) && { "x-user-picked": true }),
      } as typeof json;
    },
  });
  return res;
}

// A short list of the blocks that can be added, with the settings they need,
// for the model that checks prompts before they go to the flow AI. The service
// embeds it as kite-service/pkg/flow/catalog_summary.txt.
export function buildFlowCatalogSummary() {
  const { nodes } = buildFlowCatalog();
  const lines = Object.entries(nodes)
    // Entries come with the flow, and fixed blocks with their owner.
    .filter(([type, node]) => !type.startsWith("entry_") && !node.fixed)
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([, node]) => {
      let line = `- ${node.title}: ${node.description}`;
      if (node.contexts.length < flowContextTypes.length) {
        line += ` (only in ${node.contexts.join(", ")} flows)`;
      }

      const schema = node.data_schema as {
        required?: string[];
        properties?: Record<
          string,
          { description?: string; "x-user-picked"?: boolean }
        >;
      } | null;
      // The user picks those, so the check doesn't need to ask for them.
      const required = (schema?.required ?? [])
        .filter((name) => !schema?.properties?.[name]?.["x-user-picked"])
        .map((name) => {
          const description = schema?.properties?.[name]?.description ?? "";
          return `${name} (${description.replace(/\.$/, "")})`;
        });
      if (required.length > 0) line += `. Needs ${required.join(", ")}`;
      return line;
    });
  return lines.join("\n") + "\n";
}
