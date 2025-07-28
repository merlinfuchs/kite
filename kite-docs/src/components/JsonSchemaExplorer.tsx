import type { JsonSchema7Type } from "zod-to-json-schema";
import { JsonSchemaViewer } from "@stoplight/json-schema-viewer";
import { injectStyles } from "@stoplight/mosaic";
import { useEffect } from "react";
import "./JsonSchemaExplorer.css";

export default function JsonSchemaExplorer({
  schema,
}: {
  schema: JsonSchema7Type;
}) {
  return (
    <JsonSchemaViewer
      name="JSON Schema"
      schema={schema as any}
      expanded={true}
      hideTopBar={false}
      emptyText="No schema defined"
      defaultExpandedDepth={0}
    />
  );
}
