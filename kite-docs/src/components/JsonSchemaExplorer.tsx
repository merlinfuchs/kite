import { JsonSchemaViewer } from "@stoplight/json-schema-viewer";
import type { JsonSchema7Type } from "zod-to-json-schema";
import "./JsonSchemaExplorer.css";

export default function JsonSchemaExplorer({
  schema,
}: {
  schema: JsonSchema7Type;
}) {
  return (
    <div className="json-schema-explorer">
      <JsonSchemaViewer
        name="JSON Schema"
        schema={schema}
        expanded={true}
        hideTopBar={true}
        emptyText="No schema defined"
        defaultExpandedDepth={1}
      />
    </div>
  );
}
