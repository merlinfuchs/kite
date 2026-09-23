import {
  useDocumentStoreApi,
  useNode,
  useNodeActions,
} from "@/lib/message/state";
import { EmbedFieldNode, NodeId } from "@/lib/message/document";
import { nodeField, nodeScope } from "@/lib/message/validationStore";
import { Card } from "@/components/ui/card";
import MessageInput from "./MessageInput";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import MessageNodeActions from "./MessageNodeActions";

export default function MessageEmbedField({
  fieldId,
  fieldIndex,
}: {
  fieldId: NodeId;
  fieldIndex: number;
}) {
  const field = useNode<EmbedFieldNode>(fieldId);
  const actions = useNodeActions(fieldId);
  const { update } = useDocumentStoreApi().getState();

  return (
    <Card className="p-3">
      <MessageCollapsibleSection
        title={`Field ${fieldIndex + 1}`}
        size="md"
        validation={nodeScope(fieldId)}
        className="space-y-3"
        actions={<MessageNodeActions actions={actions} />}
      >
        <div className="flex space-x-3">
          <MessageInput
            type="text"
            label="Name"
            maxLength={256}
            value={field?.name ?? ""}
            onChange={(name) => update<EmbedFieldNode>(fieldId, { name })}
            validation={nodeField<EmbedFieldNode>(fieldId, "name")}
            placeholders
          />
          <MessageInput
            type="toggle"
            label="Inline"
            value={field?.inline || false}
            onChange={(v) =>
              update<EmbedFieldNode>(fieldId, { inline: v || undefined })
            }
            validation={nodeField<EmbedFieldNode>(fieldId, "inline")}
          />
        </div>
        <MessageInput
          type="textarea"
          label="Value"
          maxLength={1024}
          value={field?.value ?? ""}
          onChange={(value) => update<EmbedFieldNode>(fieldId, { value })}
          validation={nodeField<EmbedFieldNode>(fieldId, "value")}
          placeholders
        />
      </MessageCollapsibleSection>
    </Card>
  );
}
