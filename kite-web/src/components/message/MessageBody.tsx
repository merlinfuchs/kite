import { useDocumentStoreApi, useNode, useRootId } from "@/lib/message/state";
import { MessageNode } from "@/lib/message/document";
import { nodeField } from "@/lib/message/validationStore";
import MessageInput from "./MessageInput";

export default function MessageBody() {
  const rootId = useRootId();
  const message = useNode<MessageNode>(rootId);
  const { update } = useDocumentStoreApi().getState();

  return (
    <div className="space-y-5">
      <MessageInput
        label="Content"
        type="textarea"
        value={message?.content ?? ""}
        onChange={(content) => update<MessageNode>(rootId, { content })}
        maxLength={2000}
        validation={nodeField<MessageNode>(rootId, "content")}
        placeholders
      />
    </div>
  );
}
