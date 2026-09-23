import { NodeId, SeparatorNode } from "@/lib/message/document";
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

export default function MessageComponentSeparator({ id }: { id: NodeId }) {
  const data = useNode<SeparatorNode>(id);
  const { index } = useNodeIndex(id);
  const actions = useNodeActions(id);
  const { update } = useDocumentStoreApi().getState();

  if (!data) return null;

  return (
    <Card className="p-3">
      <MessageCollapsibleSection
        title={`Separator ${index + 1}`}
        size="md"
        validation={nodeScope(id)}
        defaultOpen={false}
        actions={<MessageNodeActions actions={actions} />}
      >
        <div className="flex space-x-3">
          <MessageInput
            type="select"
            label="Spacing"
            value={data.spacing.toString()}
            options={[
              { label: "Small", value: "1" },
              { label: "Large", value: "2" },
            ]}
            placeholder="Select spacing"
            onChange={(v) =>
              update<SeparatorNode>(id, { spacing: v === "2" ? 2 : 1 })
            }
            validation={nodeField<SeparatorNode>(id, "spacing")}
          />
          <div className="flex-none">
            <MessageInput
              type="toggle"
              label="Divider"
              value={data.divider}
              onChange={(divider) => update<SeparatorNode>(id, { divider })}
              validation={nodeField<SeparatorNode>(id, "divider")}
            />
          </div>
        </div>
      </MessageCollapsibleSection>
    </Card>
  );
}
