import CollapsibleSection from "./MessageCollapsibleSection";
import { useDocumentStoreApi, useNode } from "@/lib/message/state";
import { EmbedNode, NodeId } from "@/lib/message/document";
import { nodeField, nodeScope } from "@/lib/message/validationStore";
import MessageInput from "./MessageInput";

export default function MessageEmbedAuthor({ embedId }: { embedId: NodeId }) {
  const author = useNode<EmbedNode>(embedId)?.author;
  const { update } = useDocumentStoreApi().getState();

  const setAuthor = (patch: Partial<NonNullable<EmbedNode["author"]>>) => {
    const next = { name: "", ...author, ...patch };
    update<EmbedNode>(embedId, {
      author: next.name || next.url || next.icon_url ? next : undefined,
    });
  };

  return (
    <CollapsibleSection
      title="Author"
      size="md"
      validation={nodeScope<EmbedNode>(embedId, ["author"])}
      className="space-y-3"
    >
      <MessageInput
        type="text"
        label="Name"
        maxLength={256}
        value={author?.name || ""}
        onChange={(name) => setAuthor({ name })}
        validation={nodeField<EmbedNode>(embedId, "author.name")}
        placeholders
      />
      <div className="flex space-x-3">
        <MessageInput
          type="url"
          label="URL"
          value={author?.url || ""}
          onChange={(v) => setAuthor({ url: v || undefined })}
          validation={nodeField<EmbedNode>(embedId, "author.url")}
          placeholders
        />
        <MessageInput
          type="url"
          label="Icon URL"
          value={author?.icon_url || ""}
          onChange={(v) => setAuthor({ icon_url: v || undefined })}
          validation={nodeField<EmbedNode>(embedId, "author.icon_url")}
          imageUpload
        />
      </div>
    </CollapsibleSection>
  );
}
