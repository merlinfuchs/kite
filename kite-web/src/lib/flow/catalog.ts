import { z } from "zod";
import { getBlockDefinition } from "../blocks";
import { BlockResult } from "../blocks/types";
import { isNodeTypeAvailable, nodeCategories } from "./categories";
import { flowContextTypes } from "./context";
import {
  ExpressionTypeName,
  expressionTypeNames,
  JsonSchema,
  refTypeJsonSchema,
  thingTypeName,
  typeJsonSchema,
} from "./expressionTypes";
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
            result_schema: values.result
              ? resultJsonSchema(values.result, refTypeJsonSchema)
              : null,
          },
        ])
    ),
    // The types result schemas refer to, like #/types/message.
    types: Object.fromEntries(
      expressionTypeNames.map((type) => [
        type,
        typeJsonSchema(type, refTypeJsonSchema),
      ])
    ),
  };
}

function withoutHiddenFields(
  type: string,
  schema: ReturnType<typeof toJsonSchema>
) {
  const res = schema as {
    properties?: Record<string, unknown>;
    required?: string[];
  };
  for (const field of getBlockDefinition(type)?.fields ?? []) {
    if (!field.hidden_from_ai) continue;
    delete res.properties?.[field.name];
    res.required = res.required?.filter((name) => name !== field.name);
  }
  return schema;
}

export function toJsonSchema(schema: z.ZodType): JsonSchema {
  const { $schema, ...res } = z.toJSONSchema(schema, {
    target: "draft-7",
    io: "input",
    unrepresentable: "any",
    override: ({ jsonSchema }) => {
      // Some defaults are generated IDs, which would make the output random.
      delete jsonSchema.default;
      // Like zod 3 did, so the AI doesn't make up settings that are dropped.
      if (
        jsonSchema.type === "object" &&
        !("additionalProperties" in jsonSchema)
      ) {
        jsonSchema.additionalProperties = false;
      }
    },
  });
  return res;
}

export function resultJsonSchema(
  result: BlockResult,
  ref: (type: ExpressionTypeName) => JsonSchema
): JsonSchema {
  if (result.schema) return toJsonSchema(result.schema);

  let schema = ref(thingTypeName(result.thing));
  if (result.list) schema = { type: "array", items: schema };
  return result.description
    ? { ...schema, description: result.description }
    : schema;
}
