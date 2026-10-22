import {
  useChildIds,
  useDocument,
  useDocumentStoreApi,
  useRootId,
} from "@/lib/message/state";
import { EmbedFieldNode, EmbedNode, NodeId } from "@/lib/message/document";
import { childIds } from "@/lib/message/documentConvert";
import { MAX_EMBEDS_TEXT, embedsTextLength } from "@/lib/message/schema";
import { slotScope } from "@/lib/message/validationStore";
import MessageValidationError from "./MessageValidationError";
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
      count={
        embedIds.length > 0 ? <EmbedsTextCount rootId={rootId} /> : undefined
      }
      className="space-y-4"
    >
      {embedIds.map((id) => (
        <MessageEmbed key={id} embedId={id} />
      ))}
      <MessageValidationError target="embeds" />
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

// Its own component, so typing in an embed only rerenders the count and not
// the whole section.
function EmbedsTextCount({ rootId }: { rootId: NodeId }) {
  const textLength = useDocument(({ nodes }) =>
    embedsTextLength(
      childIds(nodes[rootId], "embeds").map((id) => {
        const embed = nodes[id] as EmbedNode;
        return {
          ...embed,
          fields: embed.fieldIds.map((id) => nodes[id] as EmbedFieldNode),
        };
      })
    )
  );

  return (
    <span title="Discord's limit for the text of all embeds together. Placeholders aren't counted, as they're only filled in when the message is sent.">
      {textLength.toLocaleString()} / {MAX_EMBEDS_TEXT.toLocaleString()}
    </span>
  );
}
