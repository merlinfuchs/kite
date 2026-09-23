import { NodeId, TextDisplayNode } from "@/lib/message/document";
import { useDocumentStoreApi, useNode } from "@/lib/message/state";
import { nodeField } from "@/lib/message/validationStore";
import MessageComponentCard from "./MessageComponentCard";
import MessageInput from "./MessageInput";

export default function MessageComponentTextDisplay({ id }: { id: NodeId }) {
  const data = useNode<TextDisplayNode>(id);
  const { update } = useDocumentStoreApi().getState();

  if (!data) return null;

  return (
    <MessageComponentCard id={id} label="Text">
      <MessageInput
        type="textarea"
        label="Content"
        maxLength={4000}
        value={data.content}
        onChange={(content) => update<TextDisplayNode>(id, { content })}
        validation={nodeField<TextDisplayNode>(id, "content")}
        placeholders
      />
    </MessageComponentCard>
  );
}
