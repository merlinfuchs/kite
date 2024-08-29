import { useCurrentMessage } from "@/lib/message/state";
import CollapsibleSection from "./MessageCollapsibleSection";
import { useShallow } from "zustand/react/shallow";
import { Button } from "@/components/ui/button";
import { getUniqueId } from "@/lib/utils";
import MessageEmbedField from "./MessageEmbedField";

export default function MessageEmbedFields({
  embedId,
  embedIndex,
}: {
  embedId: number;
  embedIndex: number;
}) {
  const fields = useCurrentMessage(
    useShallow((state) => state.embeds[embedIndex].fields.map((e) => e.id))
  );

  const [addField, clearFields] = useCurrentMessage(
    useShallow((state) => [state.addEmbedField, state.clearEmbedFields])
  );

  return (
    <CollapsibleSection
      title="Fields"
      size="md"
      valiationPathPrefix={`embeds.${embedIndex}.fields`}
      className="space-y-3"
    >
      {fields.map((id, i) => (
        <MessageEmbedField
          key={id}
          embedIndex={embedIndex}
          embedId={embedId}
          fieldIndex={i}
          fieldId={id}
        />
      ))}
      <div className="space-x-3">
        <Button
          onClick={() =>
            addField(embedIndex, {
              id: getUniqueId(),
              name: "",
              value: "",
            })
          }
          size="sm"
        >
          Add Field
        </Button>
        <Button
          onClick={() => clearFields(embedIndex)}
          variant="destructive"
          size="sm"
        >
          Clear Fields
        </Button>
      </div>
    </CollapsibleSection>
  );
}
