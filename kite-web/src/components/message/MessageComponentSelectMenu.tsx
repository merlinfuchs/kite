import { memo } from "react";
import {
  NodeId,
  SelectMenuNode,
  SelectOptionNode,
  slotLimit,
} from "@/lib/message/document";
import {
  useChildIds,
  useDocumentStoreApi,
  useNode,
  useNodeActions,
} from "@/lib/message/state";
import { nodeField, nodeScope, slotScope } from "@/lib/message/validationStore";
import { Button } from "../ui/button";
import { Card } from "../ui/card";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import MessageComponentCard from "./MessageComponentCard";
import MessageComponentFlow from "./MessageComponentFlow";
import MessageEmojiPicker from "./MessageEmojiPicker";
import MessageInput from "./MessageInput";
import MessageNodeActions from "./MessageNodeActions";

const valueCountOptions = (from: number) =>
  Array.from(
    { length: slotLimit("selectMenu", "options") + 1 - from },
    (_, i) => ({ label: (i + from).toString(), value: (i + from).toString() })
  );

const minValueOptions = valueCountOptions(0);
const maxValueOptions = valueCountOptions(1);

export default function MessageComponentSelectMenu({
  id,
  disableFlowEditor,
}: {
  id: NodeId;
  disableFlowEditor?: boolean;
}) {
  const data = useNode<SelectMenuNode>(id);
  const optionIds = useChildIds(id, "options");
  const actions = useNodeActions(id);
  const { update, insert, removeChildren } = useDocumentStoreApi().getState();

  if (!data) return null;

  return (
    <Card className="p-3">
      <MessageCollapsibleSection
        title="Select Menu"
        size="md"
        validation={nodeScope(id)}
        className="space-y-3"
        actions={<MessageNodeActions actions={actions} />}
      >
        <div className="flex space-x-3">
          <MessageInput
            type="text"
            label="Placeholder"
            maxLength={150}
            value={data.placeholder ?? ""}
            onChange={(v) =>
              update<SelectMenuNode>(id, { placeholder: v || undefined })
            }
            validation={nodeField<SelectMenuNode>(id, "placeholder")}
            placeholders
          />
          <div className="flex-none">
            <MessageInput
              type="toggle"
              label="Disabled"
              value={data.disabled ?? false}
              onChange={(v) =>
                update<SelectMenuNode>(id, { disabled: v || undefined })
              }
              validation={nodeField<SelectMenuNode>(id, "disabled")}
            />
          </div>
        </div>
        <div className="flex space-x-3">
          <MessageInput
            type="select"
            label="Min Selections"
            value={(data.min_values ?? 1).toString()}
            options={minValueOptions}
            placeholder="1"
            onChange={(v) =>
              update<SelectMenuNode>(id, { min_values: parseInt(v, 10) })
            }
            validation={nodeField<SelectMenuNode>(id, "min_values")}
          />
          <MessageInput
            type="select"
            label="Max Selections"
            value={(data.max_values ?? 1).toString()}
            options={maxValueOptions}
            placeholder="1"
            onChange={(v) =>
              update<SelectMenuNode>(id, { max_values: parseInt(v, 10) })
            }
            validation={nodeField<SelectMenuNode>(id, "max_values")}
          />
        </div>

        <MessageCollapsibleSection
          title="Options"
          size="md"
          validation={slotScope(id, "options")}
          className="space-y-3"
        >
          {optionIds.map((optionId) => (
            <MessageComponentSelectOption key={optionId} id={optionId} />
          ))}
          <div className="space-x-3">
            <Button
              size="sm"
              onClick={() =>
                insert(id, "options", "end", {
                  type: "selectOption",
                  label: "",
                })
              }
              disabled={optionIds.length >= slotLimit("selectMenu", "options")}
            >
              Add Option
            </Button>
            <Button
              size="sm"
              variant="destructive"
              onClick={() => removeChildren(id, "options")}
            >
              Clear Options
            </Button>
          </div>
        </MessageCollapsibleSection>

        {!disableFlowEditor && (
          <MessageComponentFlow
            flowSourceId={data.flow_source_id}
            context="component_select_menu"
          />
        )}
      </MessageCollapsibleSection>
    </Card>
  );
}

// Memoized so editing the menu doesn't re-render every option.
const MessageComponentSelectOption = memo(
  function MessageComponentSelectOption({ id }: { id: NodeId }) {
    const data = useNode<SelectOptionNode>(id);
    const { update } = useDocumentStoreApi().getState();

    if (!data) return null;

    return (
      <MessageComponentCard id={id} label="Option">
        <div className="flex space-x-3">
          <MessageEmojiPicker
            emoji={data.emoji}
            onChange={(emoji) => update<SelectOptionNode>(id, { emoji })}
          />
          <MessageInput
            type="text"
            label="Label"
            maxLength={100}
            value={data.label}
            onChange={(label) => update<SelectOptionNode>(id, { label })}
            validation={nodeField<SelectOptionNode>(id, "label")}
            placeholders
          />
        </div>
        <MessageInput
          type="text"
          label="Value"
          placeholder="Defaults to the label"
          maxLength={100}
          value={data.value ?? ""}
          onChange={(v) =>
            update<SelectOptionNode>(id, { value: v || undefined })
          }
          validation={nodeField<SelectOptionNode>(id, "value")}
          placeholders
        />
        <MessageInput
          type="text"
          label="Description"
          maxLength={100}
          value={data.description ?? ""}
          onChange={(v) =>
            update<SelectOptionNode>(id, { description: v || undefined })
          }
          validation={nodeField<SelectOptionNode>(id, "description")}
          placeholders
        />
      </MessageComponentCard>
    );
  }
);
