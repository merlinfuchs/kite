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
import MessageEmojiPicker from "./MessageEmojiPicker";
import FlowPreview from "../flow/FlowPreview";
import FlowDialog from "../flow/FlowDialog";
import { FlowData } from "@/lib/flow/dataSchema";
import { useCallback, useRef } from "react";
import { getUniqueId } from "@/lib/utils";

function makeInitialFlow(): FlowData {
  return {
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
}

export default function MessageComponentSelectMenuOption({
  rowIndex,
  compIndex,
  optionIndex,
  optionCount,
  disableFlowEditor,
}: {
  rowIndex: number;
  compIndex: number;
  optionIndex: number;
  optionCount: number;
  disableFlowEditor?: boolean;
}) {
  const [label, setLabel] = useCurrentMessage(
    useShallow((state) => [
      state.getSelectMenu(rowIndex, compIndex)?.options?.[optionIndex]?.label || "",
      state.setSelectMenuOptionLabel,
    ])
  );

  const [description, setDescription] = useCurrentMessage(
    useShallow((state) => [
      state.getSelectMenu(rowIndex, compIndex)?.options?.[optionIndex]?.description,
      state.setSelectMenuOptionDescription,
    ])
  );

  const [emoji, setEmoji] = useCurrentMessage(
    useShallow((state) => [
      state.getSelectMenu(rowIndex, compIndex)?.options?.[optionIndex]?.emoji,
      state.setSelectMenuOptionEmoji,
    ])
  );

  const [moveUp, moveDown, duplicate, remove] = useCurrentMessage(
    useShallow((state) => [
      state.moveSelectMenuOptionUp,
      state.moveSelectMenuOptionDown,
      state.duplicateSelectMenuOption,
      state.deleteSelectMenuOption,
    ])
  );

  const flowSourceId = useCurrentMessage(
    (state) =>
      state.getSelectMenu(rowIndex, compIndex)?.options?.[optionIndex]?.flow_source_id
  );

  const [flowData, replaceFlow] = useCurrentFlow(
    useShallow((s) => [s.getFlow(flowSourceId || ""), s.replaceFlow])
  );

  // Stable initial flow per option mount
  const initialFlowRef = useRef<FlowData | null>(null);
  if (!initialFlowRef.current) {
    initialFlowRef.current = makeInitialFlow();
  }

  const onFlowDialogClose = useCallback(
    (d: FlowData) => {
      if (flowSourceId) {
        replaceFlow(flowSourceId, d);
      }
    },
    [replaceFlow, flowSourceId]
  );

  return (
    <Card className="p-3 border-l-[3px] rounded-l-[5px] border-l-indigo-400">
      <MessageCollapsibleSection
        title={`Option ${optionIndex + 1}${label ? `: ${label}` : ""}`}
        size="md"
        valiationPathPrefix={`components.${rowIndex}.components.${compIndex}.options.${optionIndex}`}
        className="space-y-3"
        animate={false}
        defaultOpen={false}
        actions={
          <>
            {optionIndex > 0 && (
              <ChevronUpIcon
                className="h-5 w-5"
                onClick={() => moveUp(rowIndex, compIndex, optionIndex)}
                role="button"
              />
            )}
            {optionIndex < optionCount - 1 && (
              <ChevronDownIcon
                className="h-5 w-5"
                onClick={() => moveDown(rowIndex, compIndex, optionIndex)}
                role="button"
              />
            )}
            {optionCount < 25 && (
              <CopyIcon
                className="h-4 w-4"
                onClick={() => duplicate(rowIndex, compIndex, optionIndex)}
                role="button"
              />
            )}
            <TrashIcon
              className="h-4 w-4"
              onClick={() => remove(rowIndex, compIndex, optionIndex)}
              role="button"
            />
          </>
        }
      >
        <div className="flex space-x-3">
          <MessageEmojiPicker
            emoji={emoji}
            onChange={(v) => setEmoji(rowIndex, compIndex, optionIndex, v)}
          />
          <MessageInput
            type="text"
            label="Label"
            maxLength={100}
            value={label}
            onChange={(v) => setLabel(rowIndex, compIndex, optionIndex, v)}
            validationPath={`components.${rowIndex}.components.${compIndex}.options.${optionIndex}.label`}
            placeholders
          />
        </div>
        <MessageInput
          type="text"
          label="Description"
          maxLength={100}
          value={description || ""}
          onChange={(v) =>
            setDescription(rowIndex, compIndex, optionIndex, v || undefined)
          }
          validationPath={`components.${rowIndex}.components.${compIndex}.options.${optionIndex}.description`}
          placeholders
        />
        {!disableFlowEditor && (
          <FlowDialog
            flowData={flowData || initialFlowRef.current}
            context="component_button"
            onClose={onFlowDialogClose}
          >
            <FlowPreview className="h-64 p-16 w-full" onClick={() => {}} />
          </FlowDialog>
        )}
      </MessageCollapsibleSection>
    </Card>
  );
}
