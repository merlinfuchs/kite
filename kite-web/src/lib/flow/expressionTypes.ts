import generated from "./expressionTypes.gen.json";

// The types of the values expressions can use, like the result of a Get
// Member block. They are generated from the service, run
// `go test ./pkg/eval -update` in kite-service to update them.

export type ThingType = keyof typeof generated.things;
export type ExpressionTypeName = keyof typeof generated.types;

export const expressionTypeNames = Object.keys(
  generated.types
) as ExpressionTypeName[];

export function isExpressionTypeName(type: string): type is ExpressionTypeName {
  return type in generated.types;
}

export interface ExpressionField {
  name: string;
  // string, int, float, bool, any or an ExpressionTypeName.
  type: string;
  list?: boolean;
  // The field is a function, like result('id').data().
  call?: boolean;
  // The field fails for a member from a block result, as it doesn't know its
  // server.
  needs_server?: boolean;
  description?: string;
}

export function thingTypeName(thing: ThingType): ExpressionTypeName {
  return generated.things[thing] as ExpressionTypeName;
}

export function expressionFields(type: ExpressionTypeName): ExpressionField[] {
  return generated.types[type];
}

// resultFields are the fields of a type that work for a block result.
export function resultFields(type: ExpressionTypeName): ExpressionField[] {
  return expressionFields(type).filter((field) => !field.needs_server);
}

export type JsonSchema = Record<string, unknown>;

const primitiveJsonSchemas: Record<string, JsonSchema> = {
  string: { type: "string" },
  int: { type: "integer" },
  float: { type: "number" },
  bool: { type: "boolean" },
  any: {},
};

// The JSON Schema of a type as a block result, with what ref returns for the
// fields of other types, so the catalog can describe each type once.
export function typeJsonSchema(
  type: ExpressionTypeName,
  ref: (type: ExpressionTypeName) => JsonSchema
): JsonSchema {
  return {
    type: "object",
    properties: Object.fromEntries(
      resultFields(type).map((field) => {
        let schema = isExpressionTypeName(field.type)
          ? ref(field.type)
          : primitiveJsonSchemas[field.type];
        if (field.list) schema = { type: "array", items: schema };
        const description = field.call
          ? `Call it as ${field.name}(). ${field.description ?? ""}`.trim()
          : field.description;
        return [field.name, description ? { ...schema, description } : schema];
      })
    ),
  };
}

export function inlineTypeJsonSchema(type: ExpressionTypeName): JsonSchema {
  return typeJsonSchema(type, inlineTypeJsonSchema);
}

export function refTypeJsonSchema(type: ExpressionTypeName): JsonSchema {
  return { $ref: `#/types/${type}` };
}
