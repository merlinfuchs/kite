import { useChildIds, useDocumentStoreApi } from "@/lib/message/state";
import { NodeId } from "@/lib/message/document";
import { slotScope } from "@/lib/message/validationStore";
import CollapsibleSection from "./MessageCollapsibleSection";
import { Button } from "@/components/ui/button";
import MessageEmbedField from "./MessageEmbedField";

export default function MessageEmbedFields({ embedId }: { embedId: NodeId }) {
  const fieldIds = useChildIds(embedId, "fields");
  const { insert, removeChildren } = useDocumentStoreApi().getState();

  return (
    <CollapsibleSection
      title="Fields"
      size="md"
      validation={slotScope(embedId, "fields")}
      className="space-y-3"
    >
      {fieldIds.map((id, i) => (
        <MessageEmbedField key={id} fieldId={id} fieldIndex={i} />
      ))}
      <div className="space-x-3">
        <Button
          onClick={() =>
            insert(embedId, "fields", "end", {
              type: "embedField",
              name: "",
              value: "",
            })
          }
          size="sm"
          disabled={fieldIds.length >= 25}
        >
          Add Field
        </Button>
        <Button
          onClick={() => removeChildren(embedId, "fields")}
          variant="destructive"
          size="sm"
        >
          Clear Fields
        </Button>
      </div>
    </CollapsibleSection>
  );
}
