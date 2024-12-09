import { useCurrentFlow, useCurrentMessage } from "@/lib/message/state";
import { useShallow } from "zustand/react/shallow";
import { Card } from "../ui/card";
import MessageCollapsibleSection from "./MessageCollapsibleSection";
import {
  ChevronDownIcon,
  ChevronUpIcon,
  CopyIcon,
  TrashIcon,
} from "lucide-react";
import MessageInput from "./MessageInput";
import { useCallback, useMemo } from "react";
import MessageEmojiPicker from "./MessageEmojiPicker";
import FlowPreview from "../flow/FlowPreview";
import FlowDialog from "../flow/FlowDialog";
import { FlowData } from "@/lib/flow/data";
import { getUniqueId } from "@/lib/utils";

const buttonColors = {
  1: "#5865F2",
  2: "#4E5058",
  3: "#57F287",
  4: "#ED4245",
  5: "#4E5058",
};

const initialFlow = {
  nodes: [
    {
      id: getUniqueId().toString(),
      position: { x: 0, y: 0 },
      data: {},
      type: "entry_component_button",
    },
  ],
  edges: [],
};

export default function MessageComponentButton({
  rowIndex,
  compIndex,
}: {
  rowIndex: number;
  rowId: number;
  compIndex: number;
  compId: number;
}) {
  const buttonCount = useCurrentMessage(
    (state) => state.components[rowIndex].components.length
  );

  const [label, setLabel] = useCurrentMessage(
    useShallow((state) => [
      state.getButton(rowIndex, compIndex)?.label || "",
      state.setButtonLabel,
    ])
  );

  const [emoji, setEmoji] = useCurrentMessage(
    useShallow((state) => [
      state.getButton(rowIndex, compIndex)?.emoji,
      state.setButtonEmoji,
    ])
  );

  const [url, setUrl] = useCurrentMessage(
    useShallow((state) => {
      const button = state.getButton(rowIndex, compIndex);
      return [button?.style === 5 ? button.url : "", state.setButtonUrl];
    })
  );

  const [style, setStyle] = useCurrentMessage(
    useShallow((state) => [
      state.getButton(rowIndex, compIndex)?.style,
      state.setButtonStyle,
    ])
  );

  const [disabled, setDisabled] = useCurrentMessage((state) => [
    state.getButton(rowIndex, compIndex)?.disabled,
    state.setButtonDisabled,
  ]);

  const [moveUp, moveDown, duplicate, remove] = useCurrentMessage(
    useShallow((state) => [
      state.moveButtonUp,
      state.moveButtonDown,
      state.duplicateButton,
      state.deleteButton,
    ])
  );

  const color = useMemo(
    () => (style ? buttonColors[style] : buttonColors[1]),
    [style]
  );

  const flowSourceId = useCurrentMessage(
    (state) => state.getButton(rowIndex, compIndex)?.flow_source_id
  );

  const [flowData, replaceFlow] = useCurrentFlow(
    useShallow((s) => [s.getFlow(flowSourceId || ""), s.replaceFlow])
  );

  const onFlowDialogClose = useCallback(
    (d: FlowData) => {
      if (flowSourceId) {
        replaceFlow(flowSourceId, d);
      }
    },
    [replaceFlow, flowSourceId]
  );

  if (!style) {
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
        title={`Button ${compIndex + 1}`}
        size="md"
        valiationPathPrefix={`components.${rowIndex}.components.${compIndex}`}
        className="space-y-3"
        animate={false}
        defaultOpen={false}
        actions={
          <>
            {compIndex > 0 && (
              <ChevronUpIcon
                className="h-5 w-5"
                onClick={() => moveUp(rowIndex, compIndex)}
                role="button"
              />
            )}
            {compIndex < buttonCount - 1 && (
              <ChevronDownIcon
                className="h-5 w-5"
                onClick={() => moveDown(rowIndex, compIndex)}
                role="button"
              />
            )}
            {buttonCount < 5 && (
              <CopyIcon
                className="h-4 w-4"
                onClick={() => duplicate(rowIndex, compIndex)}
                role="button"
              />
            )}
            <TrashIcon
              className="h-4 w-4"
              onClick={() => remove(rowIndex, compIndex)}
              role="button"
            />
          </>
        }
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
                setStyle(rowIndex, compIndex, parseInt(v) as any)
              }
              validationPath={`components.${rowIndex}.components.${compIndex}.style`}
            />
          </div>
          <div className="flex-none">
            <MessageInput
              type="toggle"
              label="Disabled"
              value={disabled || false}
              onChange={(v) => setDisabled(rowIndex, compIndex, v || undefined)}
              validationPath={`components.${rowIndex}.components.${compIndex}.disabled`}
            />
          </div>
        </div>
        <div className="flex space-x-3">
          <MessageEmojiPicker
            emoji={emoji}
            onChange={(v) => setEmoji(rowIndex, compIndex, v)}
          />
          <MessageInput
            type="text"
            label="Label"
            maxLength={80}
            value={label}
            onChange={(v) => setLabel(rowIndex, compIndex, v)}
            validationPath={`components.${rowIndex}.components.${compIndex}.label`}
          />
        </div>
        {style === 5 ? (
          <MessageInput
            type="url"
            label="URL"
            value={url}
            onChange={(v) => setUrl(rowIndex, compIndex, v)}
            validationPath={`components.${rowIndex}.components.${compIndex}.url`}
          />
        ) : (
          <>
            <FlowDialog
              flowData={flowData || initialFlow}
              context="component_button"
              onClose={onFlowDialogClose}
            >
              <FlowPreview className="h-64 p-16 w-full" onClick={() => {}} />
            </FlowDialog>
          </>
        )}
      </MessageCollapsibleSection>
    </Card>
  );
}
