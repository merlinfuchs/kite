import {
  useChildIds,
  useDocumentStoreApi,
  useRootId,
} from "@/lib/message/state";
import { slotScope } from "@/lib/message/validationStore";
import CollapsibleSection from "./MessageCollapsibleSection";
import MessageEmbed from "./MessageEmbed";
import { Button } from "@/components/ui/button";

export default function MessageEmbedSection() {
  const rootId = useRootId();
  const embedIds = useChildIds(rootId, "embeds");
  const { insert, removeChildren } = useDocumentStoreApi().getState();

  return (
    <CollapsibleSection
      title="Embeds"
      validation={slotScope(rootId, "embeds")}
      className="space-y-4"
    >
      {embedIds.map((id, i) => (
        <MessageEmbed key={id} embedId={id} embedIndex={i} />
      ))}
      <div className="space-x-3">
        <Button
          onClick={() =>
            insert(rootId, "embeds", "end", {
              type: "embed",
              description: "",
            })
          }
          disabled={embedIds.length >= 10}
        >
          Add Embed
        </Button>
        <Button
          onClick={() => removeChildren(rootId, "embeds")}
          variant="outline"
        >
          Clear Embeds
        </Button>
      </div>
    </CollapsibleSection>
  );
}
