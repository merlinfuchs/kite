import { NodeId, SeparatorNode } from "@/lib/message/document";
import { useDocumentStoreApi, useNode } from "@/lib/message/state";
import { nodeField } from "@/lib/message/validationStore";
import MessageComponentCard from "./MessageComponentCard";
import MessageInput from "./MessageInput";

export default function MessageComponentSeparator({ id }: { id: NodeId }) {
  const data = useNode<SeparatorNode>(id);
  const { update } = useDocumentStoreApi().getState();

  if (!data) return null;

  return (
    <MessageComponentCard id={id} label="Separator" defaultOpen={false}>
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
    </MessageComponentCard>
  );
}
