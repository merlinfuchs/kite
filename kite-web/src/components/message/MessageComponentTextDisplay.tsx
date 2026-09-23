import { NodeId, TextDisplayNode } from "@/lib/message/document";
import {
  useDocumentStoreApi,
  useNode,
  useNodeActions,
  useNodeIndex,
} from "@/lib/message/state";
import { nodeField, nodeScope } from "@/lib/message/validationStore";
import { Card } from "../ui/card";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import MessageInput from "./MessageInput";
import MessageNodeActions from "./MessageNodeActions";

export default function MessageComponentTextDisplay({ id }: { id: NodeId }) {
  const data = useNode<TextDisplayNode>(id);
  const { index } = useNodeIndex(id);
  const actions = useNodeActions(id);
  const { update } = useDocumentStoreApi().getState();

  if (!data) return null;

  return (
    <Card className="p-3">
      <MessageCollapsibleSection
        title={`Text ${index + 1}`}
        size="md"
        validation={nodeScope(id)}
        actions={<MessageNodeActions actions={actions} />}
      >
        <MessageInput
          type="textarea"
          label="Content"
          maxLength={4000}
          value={data.content}
          onChange={(content) => update<TextDisplayNode>(id, { content })}
          validation={nodeField<TextDisplayNode>(id, "content")}
          placeholders
        />
      </MessageCollapsibleSection>
    </Card>
  );
}
