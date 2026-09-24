import CollapsibleSection from "./MessageCollapsibleSection";
import { useDocumentStoreApi, useNode } from "@/lib/message/state";
import { EmbedNode, NodeId } from "@/lib/message/document";
import { nodeField, nodeScope } from "@/lib/message/validationStore";
import MessageInput from "./MessageInput";

export default function MessageEmbedBody({ embedId }: { embedId: NodeId }) {
  const embed = useNode<EmbedNode>(embedId);
  const { update } = useDocumentStoreApi().getState();

  return (
    <CollapsibleSection
      title="Body"
      size="md"
      validation={nodeScope<EmbedNode>(embedId, [
        "title",
        "description",
        "url",
        "color",
      ])}
      className="space-y-3"
    >
      <MessageInput
        type="text"
        label="Title"
        maxLength={256}
        value={embed?.title || ""}
        onChange={(v) => update<EmbedNode>(embedId, { title: v || undefined })}
        validation={nodeField<EmbedNode>(embedId, "title")}
        placeholders
      />
      <MessageInput
        type="textarea"
        label="Description"
        maxLength={4000}
        value={embed?.description || ""}
        onChange={(v) =>
          update<EmbedNode>(embedId, { description: v || undefined })
        }
        validation={nodeField<EmbedNode>(embedId, "description")}
        placeholders
      />
      <div className="flex space-x-3">
        <MessageInput
          type="url"
          label="URL"
          value={embed?.url || ""}
          onChange={(v) => update<EmbedNode>(embedId, { url: v || undefined })}
          validation={nodeField<EmbedNode>(embedId, "url")}
          placeholders
        />
        <MessageInput
          type="color"
          label="Color"
          value={embed?.color}
          onChange={(color) => update<EmbedNode>(embedId, { color })}
          validation={nodeField<EmbedNode>(embedId, "color")}
        />
      </div>
    </CollapsibleSection>
  );
}
