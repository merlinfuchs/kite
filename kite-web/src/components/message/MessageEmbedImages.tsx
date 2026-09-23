import CollapsibleSection from "./MessageCollapsibleSection";
import { useDocumentStoreApi, useNode } from "@/lib/message/state";
import { EmbedNode, NodeId } from "@/lib/message/document";
import { nodeField, nodeScope } from "@/lib/message/validationStore";
import MessageInput from "./MessageInput";

export default function MessageEmbedImages({ embedId }: { embedId: NodeId }) {
  const embed = useNode<EmbedNode>(embedId);
  const { update } = useDocumentStoreApi().getState();

  return (
    <CollapsibleSection
      title="Images"
      size="md"
      validation={nodeScope<EmbedNode>(embedId, ["image", "thumbnail"])}
      className="space-y-3"
    >
      <MessageInput
        type="url"
        label="Image URL"
        value={embed?.image?.url || ""}
        onChange={(url) =>
          update<EmbedNode>(embedId, { image: url ? { url } : undefined })
        }
        validation={nodeField<EmbedNode>(embedId, "image.url")}
        imageUpload
      />
      <MessageInput
        type="url"
        label="Thumbnail URL"
        value={embed?.thumbnail?.url || ""}
        onChange={(url) =>
          update<EmbedNode>(embedId, { thumbnail: url ? { url } : undefined })
        }
        validation={nodeField<EmbedNode>(embedId, "thumbnail.url")}
        imageUpload
      />
    </CollapsibleSection>
  );
}
