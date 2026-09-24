import {
  useDocumentStoreApi,
  useNode,
  useNodeActions,
} from "@/lib/message/state";
import { ButtonNode, NodeId } from "@/lib/message/document";
import { MessageComponentButtonStyle } from "@/lib/message/schema";
import { nodeField, nodeScope } from "@/lib/message/validationStore";
import MessageNodeActions from "./MessageNodeActions";
import { Card } from "../ui/card";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import MessageInput from "./MessageInput";
import { useMemo } from "react";
import MessageEmojiPicker from "./MessageEmojiPicker";
import MessageComponentFlow from "./MessageComponentFlow";

export const buttonColors = {
  1: "#5865F2",
  2: "#4E5058",
  3: "#57F287",
  4: "#ED4245",
  5: "#4E5058",
};

export default function MessageComponentButton({
  buttonId,
  title,
  disableFlowEditor,
}: {
  buttonId: NodeId;
  title?: string;
  disableFlowEditor?: boolean;
}) {
  const button = useNode<ButtonNode>(buttonId);
  const actions = useNodeActions(buttonId);
  const { index } = actions;
  const { update } = useDocumentStoreApi().getState();

  const style = button?.style;

  const color = useMemo(
    () => (style ? buttonColors[style] : buttonColors[1]),
    [style]
  );

  if (!button || !style) {
    // This is not a button (should never happen)
    return <div></div>;
  }

  return (
    <Card
      className="p-3 border-l-[3px] rounded-l-[5px]"
      style={{
        borderLeftColor: color,
      }}
    >
      <MessageCollapsibleSection
        title={title ?? `Button ${index + 1}`}
        size="md"
        validation={nodeScope(buttonId)}
        className="space-y-3"
        animate={false}
        defaultOpen={false}
        actions={<MessageNodeActions actions={actions} />}
      >
        <div className="flex space-x-3">
          <div className="w-full">
            <MessageInput
              type="select"
              label="Style"
              value={style.toString()}
              options={[
                { label: "Blurple", value: "1" },
                { label: "Gray", value: "2" },
                { label: "Green", value: "3" },
                { label: "Red", value: "4" },
                { label: "Direct Link", value: "5" },
              ]}
              placeholder="Select a button style"
              onChange={(v) =>
                update<ButtonNode>(buttonId, {
                  style: parseInt(v) as MessageComponentButtonStyle,
                })
              }
              validation={nodeField<ButtonNode>(buttonId, "style")}
            />
          </div>
          <div className="flex-none">
            <MessageInput
              type="toggle"
              label="Disabled"
              value={button.disabled || false}
              onChange={(v) =>
                update<ButtonNode>(buttonId, { disabled: v || undefined })
              }
              validation={nodeField<ButtonNode>(buttonId, "disabled")}
            />
          </div>
        </div>
        <div className="flex space-x-3">
          <MessageEmojiPicker
            emoji={button.emoji}
            onChange={(emoji) => update<ButtonNode>(buttonId, { emoji })}
          />
          <MessageInput
            type="text"
            label="Label"
            maxLength={80}
            value={button.label}
            onChange={(label) => update<ButtonNode>(buttonId, { label })}
            validation={nodeField<ButtonNode>(buttonId, "label")}
            placeholders
          />
        </div>
        {style === 5 ? (
          <MessageInput
            type="url"
            label="URL"
            value={button.url ?? ""}
            onChange={(url) => update<ButtonNode>(buttonId, { url })}
            validation={nodeField<ButtonNode>(buttonId, "url")}
            placeholders
          />
        ) : (
          !disableFlowEditor && (
            <MessageComponentFlow
              flowSourceId={button.flow_source_id}
              context="component_button"
            />
          )
        )}
      </MessageCollapsibleSection>
    </Card>
  );
}
