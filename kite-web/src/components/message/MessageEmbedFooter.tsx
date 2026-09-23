import CollapsibleSection from "./MessageCollapsibleSection";
import { useDocumentStoreApi, useNode } from "@/lib/message/state";
import { EmbedNode, NodeId } from "@/lib/message/document";
import { nodeField, nodeScope } from "@/lib/message/validationStore";
import MessageInput from "./MessageInput";

export default function MessageEmbedFooter({ embedId }: { embedId: NodeId }) {
  const embed = useNode<EmbedNode>(embedId);
  const { update } = useDocumentStoreApi().getState();

  const setFooter = (patch: Partial<NonNullable<EmbedNode["footer"]>>) => {
    const next = { ...embed?.footer, ...patch };
    update<EmbedNode>(embedId, {
      footer: next.text || next.icon_url ? next : undefined,
    });
  };

  return (
    <CollapsibleSection
      title="Footer"
      size="md"
      validation={nodeScope<EmbedNode>(embedId, ["footer"])}
      className="space-y-3"
    >
      <MessageInput
        type="text"
        label="Footer"
        maxLength={2048}
        value={embed?.footer?.text || ""}
        onChange={(v) => setFooter({ text: v || undefined })}
        validation={nodeField<EmbedNode>(embedId, "footer.text")}
        placeholders
      />
      <div className="flex space-x-3">
        <MessageInput
          type="url"
          label="Footer Icon URL"
          value={embed?.footer?.icon_url || ""}
          onChange={(v) => setFooter({ icon_url: v || undefined })}
          validation={nodeField<EmbedNode>(embedId, "footer.icon_url")}
          imageUpload
        />
        <MessageInput
          type="date"
          label="Timestamp"
          value={embed?.timestamp}
          onChange={(timestamp) => update<EmbedNode>(embedId, { timestamp })}
          validation={nodeField<EmbedNode>(embedId, "timestamp")}
        />
      </div>
    </CollapsibleSection>
  );
}
